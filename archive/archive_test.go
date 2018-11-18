package archive_test

import (
	"bytes"
	"io"
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

func addOrSkip(t *testing.T, a *archive.Archive, message *arbor.ChatMessage) {
	if err := a.Add(message); err != nil {
		t.Skip("Failed adding message to archive:", err)
	}
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

// TestLast checks that the Last() method returns a sorted slice with a length less
// than the requested length.
func TestLast(t *testing.T) {
	a := newOrSkip(t)
	message := arbor.ChatMessage{
		UUID:      "whatever",
		Parent:    "something",
		Content:   "a lame test",
		Timestamp: 500000,
		Username:  "Socrates",
	}
	message2 := message
	message2.UUID += "2"
	message2.Timestamp += 6
	message3 := message
	message3.UUID += "3"
	message3.Timestamp -= 30
	const (
		histLen      = 5
		shortLen     = 2
		maliciousLen = -1
	)
	if last := a.Last(histLen); len(last) != 0 {
		t.Errorf("Empty archive returned non-empty slice when length was %d", histLen)
	}
	if last := a.Last(maliciousLen); len(last) != 0 {
		t.Errorf("Empty archive returned non-empty slice when length was %d", maliciousLen)
	}
	for index, m := range []arbor.ChatMessage{message, message2, message3} {
		addOrSkip(t, a, &m)
		if last := a.Last(histLen); len(last) != index+1 {
			t.Errorf("archive with %d element(s) returned slice with length %d", index+1, len(last))
		} else {
			ordered := true
			for i, k := range last {
				if i == 0 {
					continue
				}
				if k.Timestamp < last[i-1].Timestamp {
					ordered = false
				}
			}
			if !ordered {
				t.Error("Slice from Last() not sorted", last)
			}
		}
	}
	if last := a.Last(shortLen); len(last) != shortLen {
		t.Errorf("Requested Last of len %d, got %d", shortLen, len(last))
	}
	if last := a.Last(maliciousLen); len(last) != 0 {
		t.Errorf("non-Empty archive returned non-empty slice when length was %d", maliciousLen)
	}
}

// TestLoadPersist ensures that an Archive can load and store messages reliably.
func TestLoadPersist(t *testing.T) {
	a := newOrSkip(t)
	if err := a.Persist(nil); err == nil {
		t.Error("Archive failed to return error when asked to persist to nil")
	}
	if err := a.Load(nil); err == nil {
		t.Error("Archive failed to return error when asked to load to nil")
	}
	message := arbor.ChatMessage{
		UUID:      "whatever",
		Parent:    "something",
		Content:   "a lame test",
		Timestamp: 500000,
		Username:  "Socrates",
	}
	message2 := message
	message2.UUID += "2"
	message2.Timestamp += 6
	message3 := message
	message3.UUID += "3"
	message3.Timestamp -= 30
	addOrSkip(t, a, &message)
	addOrSkip(t, a, &message2)
	addOrSkip(t, a, &message3)
	buf := new(bytes.Buffer)
	if err := a.Persist(buf); err != nil {
		t.Error("Unable to persist messages to in-memory buffer", err)
	}
	a2 := newOrSkip(t)
	if err := a2.Load(buf); err != nil {
		t.Error("Unable to load message from in-memory buffer", err)
	}
	hist1 := a.Last(10)
	hist2 := a2.Last(10)
	if len(hist1) != len(hist2) {
		t.Error("Loaded and persisted history have different lengths")
	}
	for i, m1 := range hist1 {
		if len(hist2) > i && !m1.Equals(hist2[i]) {
			t.Error("Persisted and loaded history have different contents")
		}
	}
}

func persistOrSkip(t *testing.T, m *arbor.ChatMessage) io.Reader {
	a := newOrSkip(t)
	buf := new(bytes.Buffer)
	if err := a.Add(m); err != nil {
		t.Skip("Unable to add message to archive", err)
	}
	if err := a.Persist(buf); err != nil {
		t.Skip("Unable to persist into buffer", err)
	}
	return buf
}

// TestLoadPersistMultiple ensures that an Archive can load messages from
// multiple sources. It checks that the data is unioned together unless
// there is a conflict (single ID for multiple different messages), in
// which case it rejects the source that generated the conflict.
func TestLoadPersistMultiple(t *testing.T) {
	message := arbor.ChatMessage{
		UUID:      "whatever",
		Parent:    "something",
		Content:   "a lame test",
		Timestamp: 500000,
		Username:  "Socrates",
	}
	message2 := message
	message2.UUID += "2"
	message2.Timestamp += 6
	message3 := message
	message3.UUID += "3"
	message3.Timestamp -= 30
	messageBad := message
	messageBad.Content = "I'm the wrong one"
	// load three buffers with three different messages
	b1 := persistOrSkip(t, &message)
	b2 := persistOrSkip(t, &message2)
	b3 := persistOrSkip(t, &message3)
	a := newOrSkip(t)
	for _, buf := range []io.Reader{b1, b2, b3} {
		if err := a.Load(buf); err != nil {
			t.Error("Error loading from buffer", err)
		}
	}
	for _, msg := range []arbor.ChatMessage{message, message2, message3} {
		if !a.Has(msg.UUID) {
			t.Errorf("After loading all messages, \"%s\" is missing", msg.UUID)
		}
	}
	bBad := persistOrSkip(t, &messageBad)
	if err := a.Load(bBad); err == nil {
		t.Error("Failed to generate error when loading buffer with ID conflict")
	}
	if a.Get(messageBad.UUID).Content == messageBad.Content {
		t.Error("Message with ID conflict should not have replaced original message")
	}
}
