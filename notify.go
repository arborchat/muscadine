package main

import (
	"time"

	arbor "github.com/arborchat/arbor-go"
	"github.com/gen2brain/beeep"
)

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
