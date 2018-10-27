package tui_test

import (
	"bytes"
	"strings"
	"testing"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/tui"
)

// TestHistoryState ensures that HistoryStates can be created and that they
// write the correct output whenever instructed to Render.
func TestHistoryState(t *testing.T) {
	hist, err := tui.NewHistoryState()
	if err != nil {
		t.Error("Should have been able to construct HistoryState with valid params", err)
	}
	b := new(bytes.Buffer)
	err = hist.Render(b)
	if err != nil {
		t.Error("Should have been able to write empty history to buffer", err)
	}
	if len(b.String()) > 0 {
		t.Error("Wrote data when no messages to render")
	}
	a := &arbor.ChatMessage{
		UUID:      "foo",
		Parent:    "bar",
		Content:   "what",
		Username:  "test",
		Timestamp: 10000,
	}
	const iterations = 10
	for i := 1; i <= iterations; i++ {
		hist.New(a)
		b = new(bytes.Buffer)
		err = hist.Render(b)
		if err != nil {
			t.Error("Should have been able to write messages to buffer", err)
		}
		if len(b.String()) == 0 {
			t.Error("After rendering to buffer, buffer len should not be zero")
		}
		numFound := strings.Count(b.String(), "what")
		if numFound != i {
			t.Errorf("Have added %d copies of message, but render only displays %d", i, numFound)
		}
	}
}
