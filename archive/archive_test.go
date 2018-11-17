package archive_test

import (
	"testing"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/archive"
)

// TestNew ensures that the Archive constructor is working.
func TestNew(t *testing.T) {
	a := archive.New()
	if a == nil {
		t.Errorf("Should not fail to create archive")
	}
}

func newOrSkip(t *testing.T) *archive.Archive {
	a := archive.New()
	if a == nil {
		t.Skipf("Skipping because archive construction failed")
	}
	return a
}

// TestAddHasGet ensures that messages can be added to the Archive, then are
// available for retrieval.
func TestAddHasGet(t *testing.T) {
	a := newOrSkip(t)
	message := arbor.ChatMessage{
		UUID:      "whatever",
		Parent:    "something",
		Content:   "a lame test",
		Timestamp: 500000,
		Username:  "Socrates",
	}
	if a.Has(message.UUID) {
		t.Errorf("An empty archive reports containing ID \"%s\"", message.UUID)
	}
	if m := a.Get(message.UUID); m != nil {
		t.Errorf("An empty archive returned a message for ID \"%s\", %v", message.UUID, m)
	}
	if err := a.Add(nil); err == nil {
		t.Error("Did not recieve error when adding nil message to archive")
	}
	if err := a.Add(&message); err != nil {
		t.Error("Received an error when adding a valid message to an empty archive", err)
	}
	if !a.Has(message.UUID) {
		t.Errorf("After adding message, Has() returned false for that message")
	}
	if m := a.Get(message.UUID); m == nil {
		t.Error("After adding message, Get() returned nil for that message")
	} else if m.UUID != message.UUID || m.Content != message.Content || m.Parent != message.Parent || m.Username != message.Username || m.Timestamp != message.Timestamp {
		t.Errorf("Added %v, but got non-equal %v from Get()", message, m)
	}
}
