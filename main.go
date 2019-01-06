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
	history, err := archive.NewManager(histfile)
	if err != nil {
		log.Fatalln("unable to construct archive", err)
	}
	if err := history.Load(); err != nil {
		log.Println("error loading history", err)
	}
	client, err := NewNetClient(flag.Arg(0), username, history)
	if err != nil {
		log.Fatal(err)
	}
	ui, err = tui.NewTUI(client)
	if err != nil {
		log.Fatal(err)
	}
	ui.AwaitExit()
	if err := history.Save(); err != nil {
		log.Fatalln("error saving history", err)
	}
}
