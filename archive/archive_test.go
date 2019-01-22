package archive_test

import (
	"bytes"
	"io"
	"sort"
	"testing"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/archive"
	"github.com/onsi/gomega"
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
	} else if !m.Equals(&message) {
		t.Errorf("Added %v, but got non-equal %v from Get()", message, m)
	}
	messageConflict := message
	messageConflict.Content = "contentious"
	if err := a.Add(&messageConflict); err != nil {
		t.Skipf("Skipped adding conflicting message")
	}
	if a.Get(messageConflict.UUID).Content == messageConflict.Content {
		t.Error("Adding a conflicting message overwrote the existing message")
	}
}

// TestRoot ensures that the Root() method returns either the UUID of the root
// message or that of the oldest known message.
func TestRoot(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
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
	message2.Parent = ""
	message2.Timestamp += 6

	root, err := a.Root()
	g.Expect(root).To(gomega.Equal(""))
	g.Expect(err).ToNot(gomega.BeNil())

	g.Expect(a.Add(&message)).To(gomega.BeNil())
	root, err = a.Root()
	g.Expect(root).To(gomega.Equal(message.UUID))
	g.Expect(err).To(gomega.BeNil())

	g.Expect(a.Add(&message2)).To(gomega.BeNil())
	root, err = a.Root()
	g.Expect(root).To(gomega.Equal(message2.UUID))
	g.Expect(err).To(gomega.BeNil())
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
	for index, m := range []*arbor.ChatMessage{&message, &message2, &message3} {
		addOrSkip(t, a, m)
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
	if messagePointer := a.Get(message.UUID); messagePointer == &message {
		t.Errorf("Add should copy provided data to prevent internal structures from unexpected modification")
	}
}

// TestPopulatePersist ensures that an Archive can load and store messages reliably.
func TestPopulatePersist(t *testing.T) {
	a := newOrSkip(t)
	if err := a.Persist(nil); err == nil {
		t.Error("Archive failed to return error when asked to persist to nil")
	}
	if err := a.Populate(nil); err == nil {
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
	if err := a2.Populate(buf); err != nil {
		t.Error("Unable to load message from in-memory buffer", err)
	}
	hist1 := a.Last(10)
	hist2 := a2.Last(10)
	if len(hist1) != len(hist2) {
		t.Error("Populateed and persisted history have different lengths")
	}
	for i, m1 := range hist1 {
		if len(hist2) > i && !m1.Equals(hist2[i]) {
			t.Error("Persisted and loaded history have different contents")
		}
	}
}

// TestPopulatePersistOld ensures that an Archive can load messages from the old
// archive format (pre multicodec-go deprecation).
func TestPopulatePersistOld(t *testing.T) {
	a := newOrSkip(t)
	message := arbor.ChatMessage{
		UUID:      "whatever",
		Parent:    "something",
		Content:   "a lame test",
		Timestamp: 500000,
		Username:  "Socrates",
	}
	addOrSkip(t, a, &message)
	buf := new(bytes.Buffer)
	n, err := buf.Write(archive.OldArchivePrefix)
	if err != nil || n != len(archive.OldArchivePrefix) {
		t.Skip("Unable to write prefix into buffer", err)
	}
	if err := a.Persist(buf); err != nil {
		t.Error("Unable to persist messages to in-memory buffer", err)
	}
	a2 := newOrSkip(t)
	if err := a2.Populate(buf); err != nil {
		t.Error("Loading old archive format failed", err)
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

// TestPopulatePersistMultiple ensures that an Archive can load messages from
// multiple sources. It checks that the data is unioned together unless
// there is a conflict (single ID for multiple different messages), in
// which case it rejects the source that generated the conflict.
func TestPopulatePersistMultiple(t *testing.T) {
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
		if err := a.Populate(buf); err != nil {
			t.Error("Error loading from buffer", err)
		}
	}
	for _, msg := range []arbor.ChatMessage{message, message2, message3} {
		if !a.Has(msg.UUID) {
			t.Errorf("After loading all messages, \"%s\" is missing", msg.UUID)
		}
	}
	bBad := persistOrSkip(t, &messageBad)
	if err := a.Populate(bBad); err == nil {
		t.Error("Failed to generate error when loading buffer with ID conflict")
	}
	if a.Get(messageBad.UUID).Content == messageBad.Content {
		t.Error("Message with ID conflict should not have replaced original message")
	}
}

// TestNeeded checks that the Needed() method returns a sorted slice with a length less
// than the requested length.
func TestNeeded(t *testing.T) {
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
	message2.Parent += "2"
	message2.Timestamp += 6
	message3 := message
	message3.UUID += "3"
	message3.Parent += "3"
	message3.Timestamp -= 30
	correctOrder := []*arbor.ChatMessage{}
	const (
		histLen      = 5
		shortLen     = 2
		maliciousLen = -1
	)
	if needed := a.Needed(histLen); len(needed) != 0 {
		t.Errorf("Empty archive returned non-empty slice when length was %d", histLen)
	}
	if needed := a.Needed(maliciousLen); len(needed) != 0 {
		t.Errorf("Empty archive returned non-empty slice when length was %d", maliciousLen)
	}
	for index, m := range []*arbor.ChatMessage{&message, &message2, &message3} {
		correctOrder = append(correctOrder, m)
		sort.Slice(correctOrder, func(i, j int) bool {
			return correctOrder[i].Timestamp < correctOrder[j].Timestamp
		})
		addOrSkip(t, a, m)
		if needed := a.Needed(histLen); len(needed) != index+1 {
			t.Errorf("archive with %d element(s) with unknown parents returned slice with length %d", index+1, len(needed))
		} else {
			for i, k := range needed {
				if k != correctOrder[i].Parent {
					t.Errorf("Expected needed[%d]=\"%s\", found \"%s\"", i, correctOrder[i].Parent, k)
				}
			}
		}
	}
	if needed := a.Needed(shortLen); len(needed) != shortLen {
		t.Errorf("Requested Needed of len %d, got %d", shortLen, len(needed))
	}
	if needed := a.Needed(maliciousLen); len(needed) != 0 {
		t.Errorf("non-Empty archive returned non-empty slice when length was %d", maliciousLen)
	}
	root := message
	root.Parent = ""
	addOrSkip(t, a, &root)
	for _, parent := range a.Needed(10) {
		if parent == "" {
			t.Errorf("Should not express the empty string as a needed parent")
		}
	}
}

// TestLongHistNeeded is a regression test that ensures that a very long message history with many unknown
// parents doesn't crash the client. (github.com/arborchat/muscadine/issues/61)
func TestLongHistNeeded(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	a := newOrSkip(t)
	message := arbor.ChatMessage{
		UUID:      "whatever",
		Parent:    "something",
		Content:   "a lame test",
		Timestamp: 500000,
		Username:  "Socrates",
	}
	// add lots of messages with known parents
	for i := 0; i < 100; i++ {
		message.Parent = message.UUID   // make a child of the previous iteration's message
		message.UUID += "a"             // ensure new id each iteration
		added := new(arbor.ChatMessage) // allocate new memory for message so all pointers don't go to same address
		*added = message
		addOrSkip(t, a, added)
	}
	// add ten messages with unknown parents
	for i := 0; i < 10; i++ {
		message.Parent += "b"           // make a child of the previous iteration's message
		message.UUID += "a"             // ensure new id each iteration
		added := new(arbor.ChatMessage) // allocate new memory for message so all pointers don't go to same address
		*added = message
		addOrSkip(t, a, added)
	}
	needed := a.Needed(5)
	g.Expect(len(needed)).To(gomega.Equal(5))
}
