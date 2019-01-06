package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/arborchat/muscadine/archive"
	"github.com/arborchat/muscadine/tui"
	"github.com/arborchat/muscadine/types"
)

// getDefaultLogFile returns a path to the default muscadine log file location.
func getDefaultLogFile() string {
	cwdFile := "muscadine.log"
	u, err := user.Current()
	if err != nil {
		return cwdFile
	}
	return path.Join(u.HomeDir, UserDataPath, cwdFile)
}

// configureLogging attempts to set the global logger to use the named file, and logs
// an error to stdout if it fails. It returns a teardown function that can be used to
// clean up the logging and print a status message to the user.
func configureLogging(logfile string) func() {
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Printf("Unable to begin logging to %s, %v", logfile, err)
		return func() {}
	}
	log.SetOutput(file)
	log.Println("--- New Session ---")
	return func() {
		fmt.Fprintln(os.Stderr, "Logs written to", file.Name())
		file.Close()
	}
}

func main() {
	var (
		ui                types.UI
		err               error
		username          string
		histfile, logfile string
	)
	flag.StringVar(&username, "username", "muscadine", "Set your username on the server")
	flag.StringVar(&histfile, "histfile", "", "Load history from this file")
	flag.StringVar(&logfile, "logfile", getDefaultLogFile(), "Write logs to this file")
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage: " + os.Args[0] + " <ip>:<port>")
	}
	defer configureLogging(logfile)() // defer the returned cleanup function
	history, err := archive.NewManager(histfile)
	if err != nil {
		log.Fatalln("unable to construct archive", err)
	}
	if err := history.Load(); err != nil {
		log.Println("error loading history", err)
	}
	client, err := NewNetClient(flag.Arg(0), username, history)
	if err != nil {
		log.Println("Error creating client", err)
		return
	}
	ui, err = tui.NewTUI(client)
	if err != nil {
		log.Fatal("Error creating TUI", err)
		return
	}
	ui.AwaitExit()
	if err := history.Save(); err != nil {
		log.Fatalln("error saving history", err)
	}
}
