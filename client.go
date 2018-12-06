package main

import (
	"io"
	"net"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/archive"
	"github.com/arborchat/muscadine/tui"
)

// Connector is the type of functions that connect to a server over
// a given transport.
type Connector func(address string) (io.ReadWriteCloser, error)

// TCPDial makes an unencrypted TCP connection to the given address
func TCPDial(address string) (io.ReadWriteCloser, error) {
	return net.Dial("tcp", address)
}

// NetClient manages the connection to a server. It provides methods to configure
// event handlers and to connect and disconnect from the server. It also embeds
// the functionality of an Archive and Composer.
type NetClient struct {
	*archive.Archive
	Composer
	address string
	arbor.ReadWriteCloser
	connectFunc       Connector
	disconnectHandler func(tui.Connection)
	receiveHandler    func(*arbor.ChatMessage)
	stopSending       chan struct{}
	stopReceiving     chan struct{}
}

// NewNetClient creates a NetClient configured to communicate with the server at the
// given address and to use the provided archive to store the history.
func NewNetClient(address, username string, history *archive.Archive) (*NetClient, error) {
	composerOut := make(chan *arbor.ProtocolMessage)
	stopSending := make(chan struct{})
	stopReceiving := make(chan struct{})
	nc := &NetClient{
		address:       address,
		Archive:       history,
		connectFunc:   TCPDial,
		Composer:      Composer{username: username, sendChan: composerOut},
		stopSending:   stopSending,
		stopReceiving: stopReceiving,
	}
	return nc, nil
}

// SetConnector changes the function used to connect to a server address. This is
// useful both for testing purposes and to change the transport mechanism of the
// io.ReadWriteCloser. To avoid race conditions, change this before calling
// Connect() for the first time.
func (nc *NetClient) SetConnector(connector Connector) {
	nc.connectFunc = connector
}

// OnDisconnect sets the handler for disconnections. This should be done before
// calling Connect() for the first time to avoid race conditions.
func (nc *NetClient) OnDisconnect(handler func(tui.Connection)) {
	nc.disconnectHandler = handler
}

// OnReceive sets the handler for when ChatMessages are received. This should be done before
// calling Connect() for the first time to avoid race conditions.
func (nc *NetClient) OnReceive(handler func(*arbor.ChatMessage)) {
	nc.receiveHandler = handler
}

// Connect resolves the address of the NetClient and attempts to establish a connection.
func (nc *NetClient) Connect() error {
	conn, err := nc.connectFunc(nc.address)
	if err != nil {
		return err
	}
	nc.ReadWriteCloser, err = arbor.NewProtocolReadWriter(conn)
	if err != nil {
		return err
	}
	go nc.send()
	go nc.receive()
	return nil
}

// Disconnect stops all communication with the server and closes the connection. It invokes
// the handler set by OnDisconnect, if there is one.
func (nc *NetClient) Disconnect() error {
	err := nc.ReadWriteCloser.Close()
	if nc.disconnectHandler != nil {
		go nc.disconnectHandler(nc)
	}
	go func() {
		nc.stopSending <- struct{}{}
		nc.stopReceiving <- struct{}{}
	}()

	return err
}

func (nc *NetClient) send() {
	errored := false
	for {
		select {
		case p := <-nc.Composer.sendChan:
			err := nc.ReadWriteCloser.Write(p)
			if !errored && err != nil {
				errored = true
				go nc.Disconnect()
			} else if errored {
				continue
			}
		case <-nc.stopSending:
			return
		}
	}
}

func (nc *NetClient) receive() {
	errored := false
	for {
		m := new(arbor.ProtocolMessage)
		select {
		case <-nc.stopReceiving:
			return
		default:
			err := nc.ReadWriteCloser.Read(m)
			if !errored && err != nil {
				errored = true
				go nc.Disconnect()
			} else if errored {
				continue
			}
			switch m.Type {
			case arbor.NewMessageType:
				if nc.receiveHandler != nil {
					nc.receiveHandler(m.ChatMessage)
					// ask notificationEngine to display the message
					notificationEngine(nc, m.ChatMessage)
				}
			case arbor.WelcomeType:
				if !nc.Has(m.Root) {
					nc.Query(m.Root)
				}
				for _, recent := range m.Recent {
					if !nc.Has(recent) {
						nc.Query(recent)
					}
				}
			}
		}
	}
}
