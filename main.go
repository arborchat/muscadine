package main

import (
	"flag"
	"log"
	"os"

	"github.com/arborchat/muscadine/archive"
	"github.com/arborchat/muscadine/tui"
	"github.com/arborchat/muscadine/types"
)

func main() {
	var (
		ui       types.UI
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
