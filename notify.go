package main

import (
	"time"

	arbor "github.com/arborchat/arbor-go"
	"github.com/gen2brain/beeep"
)

// Notifier manages sending notifications about new messages
type Notifier struct {
	// a function to decide whether to send notifications
	ShouldNotify func(*NetClient, *arbor.ChatMessage) bool
}

// Handle processes a message and sends any notifications based on the
// current notification policy.
func (n *Notifier) Handle(cli *NetClient, msg *arbor.ChatMessage) {
	if n.ShouldNotify(cli, msg) {
		beeep.Notify("Muscadine", msg.Username+": "+msg.Content, "")
	}
}

// Recent sends a notification for every incoming message within the recent
// past that wasn't authored by the current user.
func Recent(cli *NetClient, msg *arbor.ChatMessage) bool {
	// is the message new?
	if msg.Timestamp > (time.Now().Unix() - int64(5)) {
		// do not reply to self
		if cli.username != msg.Username {
			return true
		}
	}
	return false
}
