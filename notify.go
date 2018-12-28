package main

import (
	"strconv"
	"time"

	arbor "github.com/arborchat/arbor-go"
	"github.com/gen2brain/beeep"
)

var notificationPolicy int

// This method makes notifications and handles all notification logic
func notificationEngine(cli *NetClient, msg *arbor.ChatMessage) {
	beeep.Alert("Muscadine Test", "Your notification policy is "+strconv.Itoa(notificationPolicy), "")
	// do not make notifications if the window is focused
	if true {
		// is the message new?
		if msg.Timestamp > (time.Now().Unix() - int64(5)) {
			// do not reply to self
			if cli.username != msg.Username {
				toSend := msg.Username + ": " + msg.Content
				beeep.Notify("Muscadine", toSend, "")
			}
		}
	}
}

func setNPolicy(newPolicy int) {
	notificationPolicy = newPolicy
}

// type notificationPolicy int

// const (
// 	PolicyNever notificationPolicy = iota
// 	PolicyAlways
// 	PolicyMention
// 	PolicyReply
// 	PolicyMentionReply
// )

// type Notifier struct {
// 	policy notificationPolicy // which is just an int
// }

// func NewNotifier(policy notificationPolicy) *Notifier {
// 	return &Notifier{policy: policy}
// }

// func (n *Notifier) Notify(cli *NetClient, msg *arbor.ChatMessage) {
// 	switch n.policy {
// 	case PolicyAlways:
// 		// whatever, etc...
// 	}
// }
