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
	History                   []*arbor.ChatMessage
	renderWidth, renderHeight int
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

// lastNElems returns the final `n` elements of the provided slice.
func lastNElems(slice []*arbor.ChatMessage, n int) []*arbor.ChatMessage {
	if n >= len(slice) {
		return slice
	}
	return slice[len(slice)-n : len(slice)]
}

// RenderMessage creates a text format of a message that wraps its contents to fit
// within the provided width. If a user "foo" sent a long message, the result should
// look like:
//
//`foo: jsdkfljsdfkljsfkljsdkfj
//      jskfldjfkdjsflsdkfjsldf
//      jksdfljskdfjslkfjsldkfj`
//
// The important thing to note is that lines are broken at the same place and that
// subsequent lines are padded with runewidth(username)+2 spaces. Each row of output is returned
// as a byte slice.
func RenderMessage(message *arbor.ChatMessage, width int) [][]byte {
	return nil
}

// Render writes the correct contents of the history to the provided
// writer. Each time it is invoked, it will render the entire history, so the
// writer should be empty when it is invoked.
func (h *HistoryState) Render(target io.Writer) error {
	renderableHist := lastNElems(h.History, h.renderHeight)
	for _, message := range renderableHist {
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

// SetDimensions notifes the HistoryState that the renderable display area has changed
// so that its next render can avoid rendering offscreen.
func (h *HistoryState) SetDimensions(height, width int) {
	h.renderHeight = height
	h.renderWidth = width
}
