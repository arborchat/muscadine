package archive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	arbor "github.com/arborchat/arbor-go"
)

// Archive stores the chat history of conversations had over Arbor.
// It provides mechanisms to persist history to and load history from disk.
type Archive struct {
	chronological []*arbor.ChatMessage
	root          string
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

// Needed returns at most `n` message IDs that are referenced as parents within
// the archive but are not present within the archive. These IDs should be sorted
// such that the most-recently-referenced parents are returned first. If an empty
// slice is returned, all messages within the archive have a complete ancestry.
func (a *Archive) Needed(n int) []string {
	if n <= 0 {
		return make([]string, 0)
	}
	needed := make([]string, 0)
	for _, m := range a.chronological {
		if m.Parent != "" && !a.Has(m.Parent) {
			needed = append(needed, m.Parent)
		}
	}
	if n >= len(needed) {
		return needed
	}
	return needed[len(needed)-n:]
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
	if a.Has(message.UUID) {
		// don't attempt to add messages that are already present.
		// has the nice side-effect of preventing collisions.
		return nil
	}
	messageCopy := *message
	a.chronological = append(a.chronological, &messageCopy)
	a.sort()
	if a.root == "" && message.Parent == "" {
		a.root = message.UUID
	}
	return nil
}

// Root returns the root message within the archive. If no root message is known,
// it instead returns the oldest message within the archive.
func (a *Archive) Root() (string, error) {
	if a.root != "" {
		return a.root, nil
	} else if len(a.chronological) > 0 {
		return a.chronological[0].UUID, nil
	}
	return "", fmt.Errorf("No known messages")
}

// Persist stores the contents of the archive into the provided io.Writer.
// This will always persist the entire contents of the archive, even if
// the archive contents were loaded from more than one source.
func (a *Archive) Persist(storage io.Writer) error {
	if storage == nil {
		return fmt.Errorf("Unable to persist to nil")
	}
	encoder := json.NewEncoder(storage)
	return encoder.Encode(a.chronological)
}

// OldArchivePrefix is the sequence of bytes that go-multicodec used to
// indicate JSON-encoded data at the beginning of our history files. That
// library has been deprecated, and we're switching away from that archive
// format. For the time being, we have this defined so that we can read old
// archives correctly. At a future time, this byte sequence and the logic that
// handles it should be removed.
var OldArchivePrefix = []byte{0x06, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x0a}

// Populate reads messages from the io.Reader. It expects those messages to be
// in the format written by Archive.Persist(), and should only be used on
// io.Readers that were populated with data from a call to Persist(). It
// is legal to call Populate() more than once to load the contents of more than
// one io.Reader into the Archive. This will always be performed nondestructively
// when possible. Conflicting data (multiple different messages with the same
// ID) will cause an error. All data from a source containing a conflict will
// be rejected, so io.Readers loaded first take precedence.
func (a *Archive) Populate(storage io.Reader) error {
	if storage == nil {
		return fmt.Errorf("Unable to load from nil")
	}
	// check for old history format
	prefix := make([]byte, len(OldArchivePrefix))
	// read the prefix to prevent it from causing a misparse. If it's not there, we'll put the
	// bytes that we read back.
	n, err := storage.Read(prefix)
	if err != nil {
		return fmt.Errorf("Unable to complete history prefix check: %s", err)
	}
	// if the file didn't have the prefix, put the read bytes back
	if !(n == len(OldArchivePrefix) && bytes.Equal(prefix, OldArchivePrefix)) {
		prefixBuf := bytes.NewBuffer(prefix)
		// make a reader that will read like the original reader by returning the bytes that
		// we already processed first
		storage = io.MultiReader(prefixBuf, storage)
	}
	decoder := json.NewDecoder(storage)
	if len(a.chronological) < 1 {
		return decoder.Decode(&a.chronological)
	}
	// otherwise, we need to merge the contents with what's already in a.chronological
	newMessages := make([]*arbor.ChatMessage, 0, defaultCapacity)
	if err := decoder.Decode(&newMessages); err != nil {
		return err
	}
	// ensure no bad data
	for _, message := range newMessages {
		if msg := a.Get(message.UUID); msg != nil {
			if !msg.Equals(message) {
				// we have discovered two messages with the same ID but
				// different contents. Reject all messages from the current
				// Reader.
				return fmt.Errorf("ID Collision for UUID \"%s\"", message.UUID)
			}
		}
	}
	// if we get here, no ID conflicts were discovered
	for _, message := range newMessages {
		if err := a.Add(message); err != nil {
			return err
		}
	}
	return nil
}
