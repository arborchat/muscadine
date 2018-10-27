package main

import (
	"io"
	"log"
	"net"
	"os"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/client/tui"
)

// UI is all of the operations that an Arbor client front-end needs to support
// in order to be a drop-in replacement for the default.
type UI interface {
	Display(*arbor.ChatMessage) // adds a chat message to the UI
	AwaitExit()                 // blocks until UI exit
}

// Client can be used to communicate with an Arbor server.
type Client struct {
	recvChan       <-chan *arbor.ProtocolMessage
	recieveHandler func(*arbor.ChatMessage)
}

// listen monitors for messages from the server and handles them.
func (c *Client) listen() {
	for m := range c.recvChan {
		switch m.Type {
		case arbor.NewMessageType:
			if c.recieveHandler != nil {
				c.recieveHandler(m.ChatMessage)
			}
		}
	}
}

// Connect wraps the given io.ReadWriter in a Client with methods for
// interacting with a server on the other end.
func Connect(connection io.ReadWriter) (*Client, error) {
	c := &Client{}
	c.recvChan = arbor.MakeMessageReader(connection)
	return c, nil
}

// RecieveHandler sets a function to be invoked whenever the Client
// receives a chat message from the server. This handler function must
// be safe to be invoked concurrently.
func (c *Client) RecieveHandler(handler func(*arbor.ChatMessage)) {
	c.recieveHandler = handler
	go c.listen()
}

func main() {
	var (
		ui  UI
		err error
	)
	if len(os.Args) < 2 {
		log.Fatal("Usage: " + os.Args[0] + " <ip>:<port>")
	}
	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	ui, err = tui.NewTUI()
	if err != nil {
		log.Fatal(err)
	}
	client, err := Connect(conn)
	if err != nil {
		log.Fatal(err)
	}
	client.RecieveHandler(ui.Display)
	ui.AwaitExit()
}
