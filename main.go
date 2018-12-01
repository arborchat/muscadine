package main

import (
	"flag"
	"log"
	"os"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/archive"
	"github.com/arborchat/muscadine/tui"
)

// UI is all of the operations that an Arbor client front-end needs to support
// in order to be a drop-in replacement for the default.
type UI interface {
	Display(*arbor.ChatMessage) // adds a chat message to the UI
	AwaitExit()                 // blocks until UI exit
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
	client, err := NewNetClient(flag.Arg(0), username, history)
	if err != nil {
		log.Fatal(err)
	}
	ui, err = tui.NewTUI(client)
	if err != nil {
		log.Fatal(err)
	}
	ui.AwaitExit()
	saveHist(history, histfile)
}
