package main

import (
	"io"
	"log"
	"net"
	"os"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/tui"
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
	sendChan       chan<- *arbor.ProtocolMessage
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
		case arbor.WelcomeType:
			c.sendChan <- &arbor.ProtocolMessage{
				Type:        arbor.QueryType,
				ChatMessage: &arbor.ChatMessage{UUID: m.Root},
			}
			for _, recent := range m.Recent {
				c.sendChan <- &arbor.ProtocolMessage{
					Type:        arbor.QueryType,
					ChatMessage: &arbor.ChatMessage{UUID: recent},
				}
			}
		}
	}
}

// Connect wraps the given io.ReadWriter in a Client with methods for
// interacting with a server on the other end.
func Connect(connection io.ReadWriteCloser) (*Client, error) {
	c := &Client{}
	c.recvChan = arbor.MakeMessageReader(connection)
	c.sendChan = arbor.MakeMessageWriter(connection)
	return c, nil
}

// RecieveHandler sets a function to be invoked whenever the Client
// receives a chat message from the server. This handler function must
// be safe to be invoked concurrently.
func (c *Client) RecieveHandler(handler func(*arbor.ChatMessage)) {
	c.recieveHandler = handler
	go c.listen()
}

// Composer writes arbor protocol messages
type Composer struct {
	username string
	sendChan chan<- *arbor.ProtocolMessage
}

// Reply sends a reply to `parent` with the given message content.
func (c *Composer) Reply(parent, content string) error {
	chat, err := arbor.NewChatMessage(content)
	if err != nil {
		return err
	}
	chat.Parent = parent
	chat.Username = c.username
	proto := &arbor.ProtocolMessage{ChatMessage: chat, Type: arbor.NewMessageType}
	c.sendChan <- proto
	return nil
}

// Query sends a query for the message with the given ID.
func (c *Composer) Query(id string) {
	c.sendChan <- &arbor.ProtocolMessage{Type: arbor.QueryType, ChatMessage: &arbor.ChatMessage{UUID: id}}
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
	client, err := Connect(conn)
	if err != nil {
		log.Fatal(err)
	}
	composer := &Composer{username: "muscadine", sendChan: client.sendChan}
	ui, err = tui.NewTUI(composer)
	if err != nil {
		log.Fatal(err)
	}
	client.RecieveHandler(ui.Display)
	ui.AwaitExit()
}
