package main

import (
	"io"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/archive"
)

// Client can be used to communicate with an Arbor server.
type Client struct {
	*archive.Archive
	recvChan       <-chan *arbor.ProtocolMessage
	recieveHandler func(*arbor.ChatMessage)
	Composer
}

// listen monitors for messages from the server and handles them.
func (c *Client) listen() {
	for m := range c.recvChan {
		switch m.Type {
		case arbor.NewMessageType:
			if c.recieveHandler != nil {
				c.recieveHandler(m.ChatMessage)
				// ask notificationEngine to display the message
				notificationEngine(c, m.ChatMessage)
			}
		case arbor.WelcomeType:
			if !c.Has(m.Root) {
				c.Query(m.Root)
			}
			for _, recent := range m.Recent {
				if !c.Has(recent) {
					c.Query(recent)
				}
			}
		}
	}
}

// Connect wraps the given io.ReadWriter in a Client with methods for
// interacting with a server on the other end.
func Connect(connection io.ReadWriteCloser, history *archive.Archive) (*Client, error) {
	c := &Client{}
	c.Archive = history
	c.recvChan = arbor.MakeMessageReader(connection)
	c.Composer = Composer{sendChan: arbor.MakeMessageWriter(connection)}
	return c, nil
}

// RecieveHandler sets a function to be invoked whenever the Client
// receives a chat message from the server. This handler function must
// be safe to be invoked concurrently.
func (c *Client) RecieveHandler(handler func(*arbor.ChatMessage)) {
	c.recieveHandler = handler
	go c.listen()
}
