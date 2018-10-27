package tui

import (
	"fmt"
	"io"

	arbor "github.com/arborchat/arbor-go"
)

// HistoryState maintains the state of what is visible in the client and
// can render it to any io.Writer.
type HistoryState struct {
	// history represents chat messages in the order in which they were received.
	// Index 0 holds the oldes messages, and the highest valid index holds the most
	// recent.
	History []*arbor.ChatMessage
}

const (
	defaultHistoryCapacity = 1000
	defaultHistoryLength   = 0
)

// NewHistoryState creates an empty HistoryState ready to be updated.
func NewHistoryState() (*HistoryState, error) {
	h := &HistoryState{
		History: make([]*arbor.ChatMessage, defaultHistoryLength, defaultHistoryCapacity),
	}
	return h, nil
}

// Render writes the correct contents of the history to the provided
// writer. Each time it is invoked, it will render the entire history, so the
// writer should be empty when it is invoked.
func (h *HistoryState) Render(target io.Writer) error {
	for _, message := range h.History {
		_, err := fmt.Fprintf(target, "%s: %s\n", message.Username, message.Content)
		if err != nil {
			return err
		}
	}
	return nil
}

// New alerts the HistoryState of a newly received message.
func (h *HistoryState) New(message *arbor.ChatMessage) error {
	h.History = append(h.History, message)
	return nil
}
