package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/archive"
	"github.com/arborchat/muscadine/session"
	"github.com/arborchat/muscadine/types"
	uuid "github.com/nu7hatch/gouuid"
)

const timeout = 30 * time.Second

// Connector is the type of function that connects to a server over
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
	*archive.Manager
	Composer
	address string
	arbor.ReadWriteCloser
	connectFunc Connector
	*session.List
	session.Session
	disconnectHandler func(types.Connection)
	receiveHandler    func(*arbor.ChatMessage)
	stopSending       chan struct{}
	stopReceiving     chan struct{}
	// pingServer is used to request that we attempt to force a response from the server.
	// This allows us to guard against a stale connection.
	pingServer chan struct{}
}

// NewNetClient creates a NetClient configured to communicate with the server at the
// given address and to use the provided archive to store the history.
func NewNetClient(address, username string, history *archive.Manager) (*NetClient, error) {
	if address == "" {
		return nil, fmt.Errorf("Illegal address: \"%s\"", address)
	} else if username == "" {
		return nil, fmt.Errorf("Illegal username: \"%s\"", username)
	} else if history == nil {
		return nil, fmt.Errorf("Illegal archive: %v", history)
	}
	composerOut := make(chan *arbor.ProtocolMessage)
	stopSending := make(chan struct{})
	stopReceiving := make(chan struct{})

	sessionID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("Couldn't generate session id: %s", err)
	}
	nc := &NetClient{
		address:       address,
		Manager:       history,
		connectFunc:   TCPDial,
		Composer:      Composer{username: username, sendChan: composerOut},
		stopSending:   stopSending,
		stopReceiving: stopReceiving,
		List:          session.NewList(),
		Session:       session.Session{ID: sessionID.String()},
		pingServer:    make(chan struct{}),
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
func (nc *NetClient) OnDisconnect(handler func(types.Connection)) {
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

// send reads messages from the Composer and sends them to the server.
func (nc *NetClient) send() {
	errored := false
	for {
		select {
		case protoMessage := <-nc.Composer.sendChan:
			err := nc.ReadWriteCloser.Write(protoMessage)
			if !errored && err != nil {
				errored = true
				log.Println("Error writing to server:", err)
				go nc.Disconnect()
			} else if errored {
				continue
			}
		case <-nc.pingServer:
			// query for the root message
			root, _ := nc.Archive.Root()
			go nc.Composer.Query(root)
		case <-nc.stopSending:
			return
		}
	}
}

// readChannel spawns its own goroutine to read from the NetClient's connection.
// You can stop the goroutine by closing the channel that it returns.
func (nc *NetClient) readChannel() chan struct {
	*arbor.ProtocolMessage
	error
} {
	out := make(chan struct {
		*arbor.ProtocolMessage
		error
	})
	// read messages continuously and send results back on a channel
	go func() {
		defer func() {
			// ensure send on closed channel doesn't cause panic
			if err := recover(); err != nil {
				if _, ok := err.(runtime.Error); !ok {
					// silently cancel runtime errors, but allow other errors
					// to propagate.
					panic(err)
				}
			}
		}()
		for {
			m := new(arbor.ProtocolMessage)
			err := nc.ReadWriteCloser.Read(m)
			out <- struct {
				*arbor.ProtocolMessage
				error
			}{m, err}
		}
	}()
	return out
}

// handleMessage processes an Arbor ProtocolMessage. The actions
// taken vary by the type of ProtocolMessage.
func (nc *NetClient) handleMessage(m *arbor.ProtocolMessage) {
	switch m.Type {
	case arbor.NewMessageType:
		if !nc.Archive.Has(m.UUID) {
			if nc.receiveHandler != nil {
				nc.receiveHandler(m.ChatMessage)
				// ask notificationEngine to display the message
				notificationEngine(nc, m.ChatMessage)
			}
			if m.Parent != "" && !nc.Archive.Has(m.Parent) {
				nc.Query(m.Parent)
			}
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
	case arbor.MetaType:
		for key, value := range m.Meta {
			switch key {
			case "presence/who":
				nc.Composer.AnnouncePresence(nc.Session.ID)
			case "presence/here":
				parts := strings.Split(value, ";")
				if len(parts) < 3 {
					log.Println("invalid presence/here message:", value)
					continue
				}
				username := parts[0]
				sessionID := parts[1]
				timestamp := parts[2]
				timeInt, err := strconv.Atoi(timestamp)
				if err != nil {
					log.Println("Error decoding timestamp in presence/here message:", value)
					continue
				}
				if username == nc.username && sessionID == nc.Session.ID {
					// don't track our own session
					continue
				}
				err = nc.List.Track(username, session.Session{ID: sessionID, LastSeen: time.Unix(int64(timeInt), 0)})
				if err != nil {
					log.Println("Error updating session", err)
					continue
				}
				log.Printf("Tracking session (id=%s) for user %s\n", sessionID, username)
			default:
				log.Println("Unknown meta key:", key)
			}
		}
	}
}

// recieve monitors for new messages and for connection staleness.
// If the connection with the server gets too stale, receive will close
// it automatically.
func (nc *NetClient) receive() {
	errored := false
	tick := time.NewTimer(timeout)
	defer tick.Stop()
	ticks := 0
	out := nc.readChannel()
	defer close(out)
	for {
		select {
		case <-nc.stopReceiving:
			return
		case <-tick.C:
			ticks++
			if ticks == 1 {
				// we haven't heard from the server in 30 seconds,
				// try to interact.
				nc.pingServer <- struct{}{}
				log.Println("No server contact in 30 seconds, pinging...")
			} else if ticks > 1 {
				// we haven't heard from the server in a minute,
				// we're probably disconnected.
				go nc.Disconnect()
				log.Println("No server contact in 60 seconds, disconnecting")
			}
		case readMsg := <-out:
			// reset our ticker to wait until 30 seconds from when we
			// received this message.
			tick.Reset(timeout)
			ticks = 0

			// check for errors
			m := readMsg.ProtocolMessage
			err := readMsg.error
			if !errored && err != nil {
				errored = true
				log.Println("Error reading from server:", err)
				go nc.Disconnect()
			} else if errored {
				continue
			}
			// process the message
			nc.handleMessage(m)
		}
	}
}
