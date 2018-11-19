package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/archive"
	"github.com/arborchat/muscadine/tui"
	"github.com/gen2brain/beeep"
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
	Composer
}

// listen monitors for messages from the server and handles them.
func (c *Client) listen() {
	for m := range c.recvChan {
		switch m.Type {
		case arbor.NewMessageType:
			if c.recieveHandler != nil {
				c.recieveHandler(m.ChatMessage)
if c.Composer.Username != m.ChatMessage.Username {
	beeep.Notify("Muscadine", m.ChatMessage.Content, "")
}
			}
		case arbor.WelcomeType:
			c.Query(m.Root)
			for _, recent := range m.Recent {
				c.Query(recent)
			}
		}
	}
}

// Connect wraps the given io.ReadWriter in a Client with methods for
// interacting with a server on the other end.
func Connect(connection io.ReadWriteCloser) (*Client, error) {
	c := &Client{}
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

func loadHist(history tui.Archive, histfile string) {
	if histfile != "" {
		file, err := os.Open(histfile)
		if err != nil {
			log.Println("Error opening history", err)
		} else {
			defer file.Close()
			if err = history.Load(file); err != nil {
				log.Println("Error loading history", err)
			} else {
				log.Println("History loaded from", file.Name())
			}
		}
	}
}
func saveHist(history tui.Archive, histfile string) {
	if histfile != "" {
		file, err := os.OpenFile(histfile, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			log.Println("Error opening history", err)
		} else {
			defer file.Close()
			if err = file.Truncate(0); err != nil {
				log.Println("Error truncating history", err)
			}
			if err = history.Persist(file); err != nil {
				log.Println("Error saving history", err)
			} else {
				log.Println("History saved to", file.Name())
			}
		}
	}
}
func main() {
	var (
		ui       UI
		err      error
		username string
		histfile string
	)
	flag.StringVar(&username, "username", "muscadine", "Set your username on the server")
	flag.StringVar(&histfile, "histfile", "", "Load history from this file")
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage: " + os.Args[0] + " <ip>:<port>")
	}
	history := archive.New()
	loadHist(history, histfile)
	conn, err := net.Dial("tcp", flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	client, err := Connect(conn)
	if err != nil {
		log.Fatal(err)
	}
	client.username = username
	ui, err = tui.NewTUI(client, history)
	if err != nil {
		log.Fatal(err)
	}
	client.RecieveHandler(ui.Display)
	ui.AwaitExit()
	saveHist(history, histfile)
}
