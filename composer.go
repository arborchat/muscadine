package main

import (
	"fmt"
	"time"

	arbor "github.com/arborchat/arbor-go"
)

// Composer writes arbor protocol messages
type Composer struct {
	username string
	sendChan chan *arbor.ProtocolMessage
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

// AnnounceHere sends a "presence/here" META message.
func (c *Composer) AnnounceHere(sessionID string) {
	c.sendChan <- &arbor.ProtocolMessage{
		Type: arbor.MetaType,
		Meta: map[string]string{
			"presence/here": c.username + "\n" + sessionID + "\n" + fmt.Sprintf("%d", time.Now().Unix()),
		},
	}
}

// AnnounceLeaving sends a "presence/leave" META message.
func (c *Composer) AnnounceLeaving(sessionID string) {
	c.sendChan <- &arbor.ProtocolMessage{
		Type: arbor.MetaType,
		Meta: map[string]string{
			"presence/leave": c.username + "\n" + sessionID + "\n" + fmt.Sprintf("%d", time.Now().Unix()),
		},
	}
}

// AskWho sends a "presence/who" META message.
func (c *Composer) AskWho() {
	c.sendChan <- &arbor.ProtocolMessage{
		Type: arbor.MetaType,
		Meta: map[string]string{
			"presence/who": "",
		},
	}
}
