package tui

import (
	"fmt"
	"io"
	"strings"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/types"
	"github.com/bbrks/wrap"
	runewidth "github.com/mattn/go-runewidth"
)

// HistoryState maintains the state of what is visible in the client and
// can render it to any io.Writer.
type HistoryState struct {
	// history represents chat messages in the order in which they were received.
	// Index 0 holds the oldes messages, and the highest valid index holds the most
	// recent.
	History []*arbor.ChatMessage
	types.Archive
	renderWidth, renderHeight      int
	historyHeight                  int
	current                        string
	currentIndex                   int
	cursorLineStart, cursorLineEnd int
	changeFuncs                    chan func()
}

const (
	defaultHistoryCapacity = 1000
	defaultHistoryLength   = 0
	red                    = "\x1b[0;31m"
	yellow                 = "\x1b[0;33m"
	// CurrentColor is the ANSI escape sequence for the color that is used to highlight
	// the currently-selected message
	CurrentColor = red
	// AncestorColor is the ANSI escape sequence for the color that is use to highlight
	// the ancestors of the currently-selected message
	AncestorColor = yellow
	// ClearColor is the ANSI escape sequence to return to the default color
	ClearColor = "\x1b[0;0m"
)

// NewHistoryState creates an empty HistoryState ready to be updated.
func NewHistoryState(a types.Archive) (*HistoryState, error) {
	if a == nil {
		return nil, fmt.Errorf("Cannot create HistoryState will nil Archive")
	}
	h := &HistoryState{
		History:     make([]*arbor.ChatMessage, defaultHistoryLength, defaultHistoryCapacity),
		Archive:     a,
		changeFuncs: make(chan func()),
	}
	h.History = h.Archive.Last(defaultHistoryCapacity)
	if len(h.History) > 0 {
		h.currentIndex = 0
		h.current = h.History[0].UUID
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
func RenderMessage(message *arbor.ChatMessage, width int, colorPre string, colorPost string) [][]byte {
	const separator = ": "
	usernameWidth := runewidth.StringWidth(message.Username)
	separatorWidth := runewidth.StringWidth(separator)
	firstLinePrefix := message.Username + separator
	otherLinePrefix := strings.Repeat(" ", usernameWidth+separatorWidth)
	messageRenderWidth := width - (usernameWidth + separatorWidth)
	outputLines := make([][]byte, 1)
	wrapper := wrap.NewWrapper()
	wrapper.StripTrailingNewline = true
	wrapped := wrapper.Wrap(message.Content, messageRenderWidth)
	wrappedLines := strings.SplitAfter(wrapped, "\n")
	//ensure last line ends with newline
	lastLine := wrappedLines[len(wrappedLines)-1]
	if (len(lastLine) > 0 && lastLine[len(lastLine)-1] != '\n') || len(lastLine) == 0 {
		wrappedLines[len(wrappedLines)-1] = lastLine + "\n"
	}
	wrappedLines[0] = colorPre + wrappedLines[0]
	wrappedLines[len(wrappedLines)-1] += colorPost
	outputLines[0] = []byte(firstLinePrefix + wrappedLines[0])
	for i := 1; i < len(wrappedLines); i++ {
		outputLines = append(outputLines, []byte(otherLinePrefix+wrappedLines[i]))
	}
	return outputLines
}

// currentAncestors returns the ancestor ids for the HistoryState's currently-selected
// message.
func (h *HistoryState) currentAncestors() []string {
	ancestors := make([]string, 0)
	if len(h.History) < 2 {
		return ancestors
	}
	parent := h.History[h.currentIndex].Parent
	for i := h.currentIndex - 1; i >= 0; i-- {
		if h.History[i].UUID == parent {
			ancestors = append(ancestors, h.History[i].UUID)
			parent = h.History[i].Parent
		}
	}
	return ancestors
}

// Render writes the correct contents of the history to the provided
// writer. Each time it is invoked, it will render the entire history, so the
// writer should be empty when it is invoked.
func (h *HistoryState) Render(target io.Writer) error {
	// ensure we're only working with the maximum number of messages to fill the screen
	//	renderableHist := lastNElems(h.History, h.renderHeight)
	renderableHist := h.History
	renderedHistLines := make([][]byte, 0, h.renderHeight) // ensure starting len is zero
	ancestors := h.currentAncestors()
	var (
		colorPre, colorPost string
	)
	// render each message onto however many lines it needs and capture them all.
	for _, message := range renderableHist {
		if message.UUID == h.current {
			colorPre = CurrentColor
			colorPost = ClearColor
		} else {
			colorPre = ""
			colorPost = ""
		colorize:
			for _, id := range ancestors {
				if id == message.UUID {
					colorPre = AncestorColor
					colorPost = ClearColor
					break colorize
				}
			}
		}
		lines := RenderMessage(message, h.renderWidth, colorPre, colorPost)
		if message.UUID == h.current {
			h.cursorLineStart = len(renderedHistLines)
		}
		renderedHistLines = append(renderedHistLines, lines...)
		if message.UUID == h.current {
			h.cursorLineEnd = len(renderedHistLines) - 1
		}
	}
	// find the lines that will actually be visible in the rendered area
	//	renderedHistLines = lastNElemsBytes(renderedHistLines, h.renderHeight)
	// draw the lines that are visible to the screen
	for _, line := range renderedHistLines {
		_, err := target.Write(line)
		if err != nil {
			return err
		}
	}
	h.historyHeight = len(renderedHistLines)
	return nil
}

// Height returns the number of lines of text rendered in the last render.
func (h *HistoryState) Height() int {
	return h.historyHeight
}

// CursorLines returns the range of rendered lines that contain the selected message.
// These are expressed in 0-based indicies. If the results were (1,2), that would
// mean that the current message spans the second and third lines of the rendered
// output.
func (h *HistoryState) CursorLines() (int, int) {
	done := make(chan error)
	var start, end int
	h.changeFuncs <- func() {
		defer close(done)
		start = h.cursorLineStart
		end = h.cursorLineEnd
	}
	<-done
	return start, end
}

// New alerts the HistoryState of a newly received message.
func (h *HistoryState) New(message *arbor.ChatMessage) error {
	done := make(chan error)
	h.changeFuncs <- func() {
		defer close(done)
		h.Archive.Add(message)
		h.History = h.Archive.Last(defaultHistoryCapacity)
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

// CursorEnd moves the current message to the end of the history.
func (h *HistoryState) CursorEnd() {
	done := make(chan error)
	h.changeFuncs <- func() {
		defer close(done)
		if len(h.History) < 2 {
			// couldn't possibly scroll the cursor, 0 or 1 messages available
			return
		}
		h.current = h.History[len(h.History)-1].UUID
		h.currentIndex = len(h.History) - 1
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

// CursorBeginning moves the current message to the beginning of the history.
func (h *HistoryState) CursorBeginning() {
	done := make(chan error)
	h.changeFuncs <- func() {
		defer close(done)
		if len(h.History) < 2 {
			// couldn't possibly scroll the cursor, 0 or 1 messages available
			return
		}
		h.current = h.History[0].UUID
		h.currentIndex = 0
	}
	<-done
}
