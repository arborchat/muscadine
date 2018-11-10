package tui

import (
	"io"
	"sort"

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
	current                   string
	currentIndex              int
	changeFuncs               chan func()
}

const (
	defaultHistoryCapacity = 1000
	defaultHistoryLength   = 0
	// CurrentColor is the ANSI escape sequence for the color that is used to highlight
	// the currently-selected mesage
	CurrentColor = "\x1b[0;31m"
	// ClearColor is the ANSI escape sequence to return to the default color
	ClearColor = "\x1b[0;0m"
)

// NewHistoryState creates an empty HistoryState ready to be updated.
func NewHistoryState() (*HistoryState, error) {
	h := &HistoryState{
		History:     make([]*arbor.ChatMessage, defaultHistoryLength, defaultHistoryCapacity),
		changeFuncs: make(chan func()),
	}
	// launch a goroutine to serially execute all state modifications
	go func(h *HistoryState) {
		for f := range h.changeFuncs {
			f()
		}
	}(h)
	return h, nil
}

// lastNElems returns the final `n` elements of the provided slice of messages
func lastNElems(slice []*arbor.ChatMessage, n int) []*arbor.ChatMessage {
	if n >= len(slice) {
		return slice
	}
	return slice[len(slice)-n:]
}

// lastNElems returns the final `n` elements of the provided slice of messages
func lastNElemsBytes(slice [][]byte, n int) [][]byte {
	if n >= len(slice) {
		return slice
	}
	return slice[len(slice)-n:]
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
func (h HistoryState) RenderMessage(message *arbor.ChatMessage) []byte {
	const separator = ": "

	text := message.Content
	if len(text) > 0 && text[len(text)-1] != '\n' {
		text += "\n"
	}
	if h.Current() == message.UUID {
		text = CurrentColor + text + ClearColor
	}
	text = message.Username + separator + text
	return []byte(text)
}

// Render writes the correct contents of the history to the provided
// writer. Each time it is invoked, it will render the entire history, so the
// writer should be empty when it is invoked.
func (h HistoryState) Render(target io.Writer) error {
	// ensure we're only working with the maximum number of messages to fill the screen
	//	renderableHist := lastNElems(h.History, h.renderHeight)
	renderableHist := h.History
	rendered := make([][]byte, h.renderHeight)
	// render each message onto however many lines it needs and capture them all.
	for _, message := range renderableHist {
		rendered = append(rendered, h.RenderMessage(message))
	}
	// draw the lines that are visible to the screen
	for _, line := range rendered {
		_, err := target.Write(line)
		if err != nil {
			return err
		}
	}
	return nil
}

// New alerts the HistoryState of a newly received message.
func (h *HistoryState) New(message *arbor.ChatMessage) error {
	done := make(chan error)
	h.changeFuncs <- func() {
		defer close(done)
		h.History = append(h.History, message)
		// ensure the new message is in the proper place
		sort.SliceStable(h.History, func(i, j int) bool {
			return h.History[i].Timestamp < h.History[j].Timestamp
		})
		if h.current == "" {
			h.current = message.UUID
		}
		for index, curMsg := range h.History {
			if h.current == curMsg.UUID {
				h.currentIndex = index
			}
		}
	}

	return <-done
}

// SetDimensions notifes the HistoryState that the renderable display area has changed
// so that its next render can avoid rendering offscreen.
func (h *HistoryState) SetDimensions(height, width int) {
	done := make(chan error)
	h.changeFuncs <- func() {
		defer close(done)
		h.renderHeight = height
		h.renderWidth = width
	}
	<-done
}

// Current returns the id of the currently-selected message, if there is one. The first message
// added to a HistoryState is marked as current automatically. After that, Current can only
// be changed by scrolling.
func (h HistoryState) Current() string {
	return h.current
}

// CursorDown moves the current message downward within the history, if it is possible to do
// so. If there are no messages in the history, it does nothing. If the current message is
// at the bottom of the history, it does nothing.
func (h *HistoryState) CursorDown() {
	done := make(chan error)
	h.changeFuncs <- func() {
		defer close(done)
		if len(h.History) < 2 {
			// couldn't possibly scroll the cursor, 0 or 1 messages available
			return
		}
		if h.currentIndex+1 >= len(h.History) {
			// current message is at bottom of history, can't scroll down
			return
		}
		h.current = h.History[h.currentIndex+1].UUID
		h.currentIndex++
	}
	<-done
}

// CursorUp moves the current message upward within the history, if it is possible to do
// so. If there are no messages in the history, it does nothing. If the current message is
// at the top of the history, it does nothing.
func (h *HistoryState) CursorUp() {
	done := make(chan error)
	h.changeFuncs <- func() {
		defer close(done)
		if len(h.History) < 2 {
			// couldn't possibly scroll the cursor, 0 or 1 messages available
			return
		}
		if h.currentIndex-1 < 0 {
			// current message is at top of history, can't scroll up
			return
		}
		h.current = h.History[h.currentIndex-1].UUID
		h.currentIndex--
	}
	<-done
}
