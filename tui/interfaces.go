package tui

import (
	"io"

	arbor "github.com/arborchat/arbor-go"
)

// Composer writes and sends protocol messages
type Composer interface {
	Reply(string, string) error
	Query(string)
}

// Archive stores and retrieves messages
type Archive interface {
	Last(n int) []*arbor.ChatMessage
	Has(id string) bool
	Get(id string) *arbor.ChatMessage
	Add(message *arbor.ChatMessage) error
	Persist(storage io.Writer) error
	Load(storage io.Reader) error
}
