package tui

import (
	"io"
	"strings"

	arbor "github.com/arborchat/arbor-go"
	runewidth "github.com/mattn/go-runewidth"
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
}

const (
	defaultHistoryCapacity = 1000
	defaultHistoryLength   = 0
	CurrentColor           = "\x1b[0;31m"
	ClearColor             = "\x1b[0;0m"
)

// NewHistoryState creates an empty HistoryState ready to be updated.
func NewHistoryState() (*HistoryState, error) {
	h := &HistoryState{
		History: make([]*arbor.ChatMessage, defaultHistoryLength, defaultHistoryCapacity),
	}
	return h, nil
}

// lastNElems returns the final `n` elements of the provided slice of messages
func lastNElems(slice []*arbor.ChatMessage, n int) []*arbor.ChatMessage {
	if n >= len(slice) {
		return slice
	}
	return slice[len(slice)-n : len(slice)]
}

// lastNElems returns the final `n` elements of the provided slice of messages
func lastNElemsBytes(slice [][]byte, n int) [][]byte {
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
func (h *HistoryState) RenderMessage(message *arbor.ChatMessage, width int) [][]byte {
	const separator = ": "
	usernameWidth := runewidth.StringWidth(message.Username)
	separatorWidth := runewidth.StringWidth(separator)
	firstLinePrefix := message.Username + separator
	otherLinePrefix := strings.Repeat(" ", usernameWidth+separatorWidth)
	messageRenderWidth := width - (usernameWidth + separatorWidth)
	outputLines := make([][]byte, 1)
	wrapped := runewidth.Wrap(message.Content, messageRenderWidth)
	wrappedLines := strings.SplitAfter(wrapped, "\n")
	//ensure last line ends with newline
	lastLine := wrappedLines[len(wrappedLines)-1]
	if lastLine[len(lastLine)-1] != '\n' {
		wrappedLines[len(wrappedLines)-1] = lastLine + "\n"
	}
	outputLines[0] = []byte(firstLinePrefix + wrappedLines[0])
	for i := 1; i < len(wrappedLines); i++ {
		outputLines = append(outputLines, []byte(otherLinePrefix+wrappedLines[i]))
	}
	return outputLines
}

// Render writes the correct contents of the history to the provided
// writer. Each time it is invoked, it will render the entire history, so the
// writer should be empty when it is invoked.
func (h *HistoryState) Render(target io.Writer) error {
	// ensure we're only working with the maximum number of messages to fill the screen
	renderableHist := lastNElems(h.History, h.renderHeight)
	renderedHistLines := make([][]byte, h.renderHeight)
	// render each message onto however many lines it needs and capture them all.
	for _, message := range renderableHist {
		lines := h.RenderMessage(message, h.renderWidth)
		renderedHistLines = append(renderedHistLines, lines...)
	}
	// find the lines that will actually be visible in the rendered area
	renderedHistLines = lastNElemsBytes(renderedHistLines, h.renderHeight)
	// draw the lines that are visible to the screen
	for _, line := range renderedHistLines {
		_, err := target.Write(line)
		if err != nil {
			return err
		}
	}
	return nil
}

// New alerts the HistoryState of a newly received message.
func (h *HistoryState) New(message *arbor.ChatMessage) error {
	h.History = append(h.History, message)
	if h.current == "" {
		h.current = message.UUID
	}
	return nil
}

// SetDimensions notifes the HistoryState that the renderable display area has changed
// so that its next render can avoid rendering offscreen.
func (h *HistoryState) SetDimensions(height, width int) {
	h.renderHeight = height
	h.renderWidth = width
}

// Current returns the id of the currently-selected message, if there is one. The first message
// added to a HistoryState is marked as current automatically. After that, Current can only
// be changed by scrolling.
func (h *HistoryState) Current() string {
	return h.current
}
