// Package types contains the interface and concrete types used for interoperability
// within all of the modules of muscadine. Wherever possible, we do not rely
// on a specific implementation of functionality, but rather an interface.
// Having a separate package that defines those interfaces makes importing
// them anywhere within the codebase simpler (avoids weird circular imports).
package types

import (
	"io"
	"time"

	arbor "github.com/arborchat/arbor-go"
)

// UI is all of the operations that an Arbor client front-end needs to support
// in order to be a drop-in replacement for the default.
type UI interface {
	Display(*arbor.ChatMessage) // adds a chat message to the UI
	AwaitExit()                 // blocks until UI exit
}

// Client manages the connection between a TUI and a specific server
type Client interface {
	Composer
	Archive
	Connection
	SessionList
}

// SessionList tracks the sessions of other users.
type SessionList interface {
	ActiveSessions() map[string]time.Time
}

// Composer writes and sends protocol messages
type Composer interface {
	Reply(string, string) error
	Query(string)
	AskWho()
	AnnounceHere(string)
	AnnounceLeaving(string)
}

// Connection models a live connection to a server
type Connection interface {
	SessionID() string
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
	ChildrenOf(string) []string
	Add(message *arbor.ChatMessage) error
	Persist(storage io.Writer) error
	Populate(storage io.Reader) error
}
