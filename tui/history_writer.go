package tui

import (
	"io"

	arbor "github.com/arborchat/arbor-go"
)

// HistoryState maintains the state of what is visible in the client and
// can render it to any io.Writer.
type HistoryState struct{}

// NewHistoryState creates an empty HistoryState ready to be updated.
func NewHistoryState() (*HistoryState, error) {
	return nil, nil
}

// Render writes the correct contents of the history to the provided
// writer.
func (h *HistoryState) Render(io.Writer) error {
	return nil
}

// New alerts the HistoryState of a newly received message.
func (h *HistoryState) New(message *arbor.ChatMessage) error {
	return nil
}
