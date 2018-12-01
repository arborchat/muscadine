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
