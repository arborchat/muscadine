package archive

import (
	"fmt"
	"io"
	"sort"

	arbor "github.com/arborchat/arbor-go"
)

// Archive stores the chat history of conversations had over Arbor.
// It provides mechanisms to persist history to and load history from disk.
type Archive struct {
	chronological []*arbor.ChatMessage
}

const defaultCapacity = 1024

// New creates an empty archive. Use Load() or Add() to populate with data.
func New() *Archive {
	return &Archive{chronological: make([]*arbor.ChatMessage, 0, defaultCapacity)}
}

// Last returns the most chronologically "recent" `n` messages known to the
// archive. The length of the returned slice may be shorter than `n` if `n`
// is greater than the number of known messages.
func (a *Archive) Last(n int) []*arbor.ChatMessage {
	if n <= 0 {
		return make([]*arbor.ChatMessage, 0)
	}
	if n >= len(a.chronological) {
		return a.chronological
	}
	return a.chronological[len(a.chronological)-n:]
}

// Has returns whether the archive contains a message with the given ID.
func (a *Archive) Has(id string) bool {
	for _, message := range a.chronological {
		if message.UUID == id {
			return true
		}
	}
	return false
}

// Get returns the message with the given id, or nil if the message is
// not in the archive.
func (a *Archive) Get(id string) *arbor.ChatMessage {
	for _, message := range a.chronological {
		if message.UUID == id {
			return message
		}
	}
	return nil
}

// sort updates the internal representation to ensure that messages are ordered
// correctly.
func (a *Archive) sort() {
	sort.SliceStable(a.chronological, func(i, j int) bool {
		return a.chronological[i].Timestamp < a.chronological[j].Timestamp
	})
}

// Add adds the provided message to the archive.
func (a *Archive) Add(message *arbor.ChatMessage) error {
	if message == nil {
		return fmt.Errorf("Unable to add nil message")
	}
	a.chronological = append(a.chronological, message)
	a.sort()
	return nil
}

// Persist stores the contents of the archive into the provided io.Writer.
// This will always persist the entire contents of the archive, even if
// the archive contents were loaded from more than one source.
func (a *Archive) Persist(storage io.Writer) error {
	return nil
}

// Load reads messages from the io.Reader. It expects those messages to be
// in the format written by Archive.Persist(), and should only be used on
// io.Readers that were populated with data from a call to Persist(). It
// is legal to call Load() more than once to load the contents of more than
// one io.Reader into the Archive. This will always be performed nondestructively
// when possible. Conflicting data (multiple different messages with the same
// ID) will cause an error.
func (a *Archive) Load(storage io.Reader) error {
	return nil
}
