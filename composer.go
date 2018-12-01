package main

import (
	"flag"
	"log"
	"net"
	"os"
	"time"

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

// This method makes notifications and handles all notification logic
func notificationEngine(cli *Client, msg *arbor.ChatMessage) {
	// is the message new?
	if msg.Timestamp > (time.Now().Unix() - int64(5)) {
		// do not reply to self
		if cli.username != msg.Username {
			toSend := msg.Username + ": " + msg.Content
			beeep.Notify("Muscadine", toSend, "")
		}
	}
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
	client, err := Connect(conn, history)
	if err != nil {
		log.Fatal(err)
	}
	client.username = username
	ui, err = tui.NewTUI(client)
	if err != nil {
		log.Fatal(err)
	}
	client.RecieveHandler(ui.Display)
	ui.AwaitExit()
	saveHist(history, histfile)
}
