package tui

import (
	"io"

	arbor "github.com/arborchat/arbor-go"
)

// Client manages the connection between a TUI and a specific server
type Client interface {
	Composer
	Archive
	Connection
}

// Composer writes and sends protocol messages
type Composer interface {
	Reply(string, string) error
	Query(string)
}

// Connection models a live connection to a server
type Connection interface {
	OnDisconnect(handler func(Connection))
	OnReceive(handler func(*arbor.ChatMessage))
	Connect() error
	Disconnect() error
}

// Archive stores and retrieves messages
type Archive interface {
	Last(n int) []*arbor.ChatMessage
	Needed(n int) []string
	Has(id string) bool
	Get(id string) *arbor.ChatMessage
	Root() (string, error)
	Add(message *arbor.ChatMessage) error
	Persist(storage io.Writer) error
	Load(storage io.Reader) error
}
