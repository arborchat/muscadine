package tui_test

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/archive"
	"github.com/arborchat/muscadine/tui"
	runewidth "github.com/mattn/go-runewidth"
)

var testMsg = arbor.ChatMessage{
	UUID:      "foo",
	Parent:    "bar",
	Content:   "what",
	Username:  "test",
	Timestamp: 10000,
}

// TestHistoryState ensures that HistoryStates can be created and that they
// write the correct output whenever instructed to Render.
func TestHistoryState(t *testing.T) {
	if _, err := tui.NewHistoryState(nil); err == nil {
		t.Error("Should fail to create HistoryState if nil archive provided")
	}
	hist, err := tui.NewHistoryState(archive.New())
	if err != nil {
		t.Error("Should have been able to construct HistoryState with valid params", err)
	}
	const iterations = 10
	const height = 5
	const width = 80
	hist.SetDimensions(iterations, width)
	b := new(bytes.Buffer)
	err = hist.Render(b)
	if err != nil {
		t.Error("Should have been able to write empty history to buffer", err)
	}
	if len(b.String()) > 0 {
		t.Error("Wrote data when no messages to render")
	}
	for i := 1; i <= iterations; i++ {
		testMsg.UUID += strconv.Itoa(i)
		newOrSkip(t, hist, &testMsg)
		b = new(bytes.Buffer)
		err = hist.Render(b)
		if err != nil {
			t.Error("Should have been able to write messages to buffer", err)
		}
		if len(b.String()) == 0 {
			t.Error("After rendering to buffer, buffer len should not be zero")
		}
		numFound := strings.Count(b.String(), testMsg.Content)
		if numFound != i {
			t.Errorf("Have added %d copies of message, but render only displays %d", i, numFound)
		}
	}
	hist.SetDimensions(height, width)
	b = new(bytes.Buffer)
	err = hist.Render(b)
	if err != nil {
		t.Error("Failed to render with dimensions set", err)
	}
	numFound := strings.Count(b.String(), testMsg.Content)
	if numFound != hist.Height() {
		t.Errorf("Found %d rendered messages, but Height() returned %d, buffer: %s", numFound, hist.Height(), b.String())
	}
}

func historyStateOrSkip(t *testing.T) *tui.HistoryState {
	hist, err := tui.NewHistoryState(archive.New())
	if err != nil {
		t.Skip("Should have been able to construct HistoryState with valid params", err)
	}
	hist.SetDimensions(24, 80)
	return hist
}

func errorIfNil(t *testing.T, rendered [][]byte, message string) {
	if rendered == nil {
		t.Fatal(message)
	}
}

func errorIfNotPrefix(t *testing.T, prefix, actual, message string) {
	if !strings.HasPrefix(actual, prefix) {
		t.Error(message)
	}
}

// TestRenderMessage ensures that the rendering function correctly handles line wrapping
// and related problems when preparing messages to be displayed.
func TestRenderMessage(t *testing.T) {
	message := testMsg
	message.Content = "let's use a ลาญฤๅเข่นฆ่าบีฑ much longer string いろはにほへとちりぬるを so that line wrapping アサキユメミシhappens"
	separator := ": "
	usernameWidth := runewidth.StringWidth(message.Username)
	separatorWidth := runewidth.StringWidth(separator)
	contentWidth := runewidth.StringWidth(message.Content)
	prefix := message.Username + separator
	prefixWidth := runewidth.StringWidth(prefix)
	startingWidth := usernameWidth + contentWidth + separatorWidth
	// test that it all fits on one line given sufficient space
	rendered := tui.RenderMessage(&message, startingWidth, tui.CurrentColor, tui.ClearColor)
	errorIfNil(t, rendered, "Render produced nil output for 1 line message")

	if len(rendered) < 1 || len(rendered) > 1 {
		t.Errorf("Expected rendered message of %d lines, got %d", 1, len(rendered))
	}
	if string(rendered[0][:len(prefix)]) != prefix {
		t.Errorf("Expected prefix \"%s\", got \"%s\"", prefix, rendered[0][:len(prefix)])
	}
	// test every message width between a single line and having no space for message
	// content to be displayed
	for i := startingWidth - 1; i > usernameWidth+separatorWidth; i-- {
		collect := ""
		rendered := tui.RenderMessage(&message, i, "", "")
		errorIfNil(t, rendered, fmt.Sprintf("Render produced nil for %d width message", i))
		for index, line := range rendered {
			if index == 0 {
				errorIfNotPrefix(t, prefix, string(line), fmt.Sprintf("Expected prefix \"%s\" on line %d, got \"%s\"", prefix, index, line))
			} else {
				errorIfNotPrefix(t, strings.Repeat(" ", prefixWidth), string(line), fmt.Sprintf("Expected line %d to start with %d spaces, found \"%s\"", index, prefixWidth, line))
			}
			if line[len(line)-1] == '\n' {
				collect += string(line[len(prefix) : len(line)-1]) // don't include trailing newline
			} else {
				collect += string(line[len(prefix):])
			}
		}
		if collect != message.Content {
			t.Errorf("Expected line contents to be \"%s\", found \"%s\"", message.Content, collect)
		}
	}
}

// TestSelectMessage ensures that the first message a historystate receives is marked as the current.
func TestSelectMessage(t *testing.T) {
	hist := historyStateOrSkip(t)
	message := testMsg
	id := hist.Current()
	if id != "" {
		t.Errorf("Expected empty history to have empty current, got \"%s\"", id)
	}
	newOrSkip(t, hist, &message)
	id = hist.Current()
	if id != message.UUID {
		t.Errorf("Expected hist to set current to first received message (\"%s\"), got \"%s\"", message.UUID, id)
	}
	second := testMsg
	second.UUID = "different"
	newOrSkip(t, hist, &second)
	id = hist.Current()
	if id != message.UUID {
		t.Errorf("Expected hist to keep first received message id as current (\"%s\"), got \"%s\"", message.UUID, id)
	}
}

func newOrSkip(t *testing.T, hist *tui.HistoryState, msg *arbor.ChatMessage) {
	if err := hist.New(msg); err != nil {
		t.Skip(err)
	}
}

// TestRenderColoredMessage ensures that the render function uses the colors it is provided to render messages.
func TestRenderColoredMessage(t *testing.T) {
	message := testMsg
	rendered := tui.RenderMessage(&message, 80, tui.CurrentColor, tui.ClearColor)
	if !strings.Contains(string(rendered[0]), tui.CurrentColor) && !strings.Contains(string(rendered[0]), tui.ClearColor) {
		t.Error("Expected current message to be rendered in color CurrentColor", string(rendered[0]))
	}
	second := testMsg
	second.UUID = "different"
	rendered = tui.RenderMessage(&second, 80, tui.AncestorColor, tui.ClearColor)
	if !strings.Contains(string(rendered[0]), tui.AncestorColor) && !strings.Contains(string(rendered[0]), tui.ClearColor) {
		t.Error("Expected message to be rendered in color AncestorColor", string(rendered[0]))
	}
}

// TestRenderEmptyMessage ensures that the HistoryState doesn't panic when trying to render
// messages whose content contains empty lines.
func TestRenderEmptyMessage(t *testing.T) {
	hist := historyStateOrSkip(t)
	hist.SetDimensions(24, 80)
	message := testMsg
	message.Content = ""
	// making sure that we don't crash
	tui.RenderMessage(&message, 80, "", "")

	// now check that we correctly render messages with a trailing newline
	message.Content = "\n"
	rendered := tui.RenderMessage(&message, 80, "", "")
	if !strings.HasSuffix(string(rendered[len(rendered)-1]), "\n") {
		t.Errorf("Should have inserted newline at end of empty rendered message line, found %v", rendered)
	}
}

// TestCursorDown checks that the current message can be scrolled downward through the history.
func TestCursorDown(t *testing.T) {
	hist := historyStateOrSkip(t)
	hist.SetDimensions(24, 80)
	message := testMsg
	newOrSkip(t, hist, &message)
	second := testMsg
	second.UUID = "second-message"
	newOrSkip(t, hist, &second)
	if id := hist.Current(); id != message.UUID {
		t.Skip("History setting current message improperly")
	}
	hist.CursorDown()
	if id := hist.Current(); id != second.UUID {
		t.Errorf("History current message should havd id \"%s\", not \"%s\"", second.UUID, id)
	}
}

// TestCursorUp checks that the current message can be scrolled upward through the history.
func TestCursorUp(t *testing.T) {
	hist := historyStateOrSkip(t)
	hist.SetDimensions(24, 80)
	message := testMsg
	newOrSkip(t, hist, &message)
	second := testMsg
	second.UUID = "second-message"
	newOrSkip(t, hist, &second)
	if id := hist.Current(); id != message.UUID {
		t.Skip("History setting current message improperly")
	}
	hist.CursorDown()
	if id := hist.Current(); id != second.UUID {
		t.Skipf("Scrolling down is broken, which makes scrolling up untestable")
	}
	hist.CursorUp()
	if id := hist.Current(); id != message.UUID {
		t.Errorf("History current message should havd id \"%s\", not \"%s\"", message.UUID, id)
	}
}

// TestMessageSort ensures that the historystate sorts messages that it displays
// according to their timestamps.
func TestMessageSort(t *testing.T) {
	hist := historyStateOrSkip(t)
	hist.SetDimensions(24, 80)
	message := testMsg
	message.UUID += "1"
	message.Content = "one"
	message.Timestamp = 10
	newOrSkip(t, hist, &message)
	message3 := testMsg
	message3.UUID += "3"
	message3.Content = "three"
	message3.Timestamp = 30
	newOrSkip(t, hist, &message3)
	message2 := testMsg
	message2.UUID += "2"
	message2.Content = "two"
	message2.Timestamp = 20
	newOrSkip(t, hist, &message2)
	message0 := testMsg
	message0.UUID += "0"
	message0.Content = "zero"
	message0.Timestamp = 0
	newOrSkip(t, hist, &message0)
	buf := new(bytes.Buffer)
	err := hist.Render(buf)
	if err != nil {
		t.Error("Unable to render buffer", err)
	}
	zeroIndex := strings.Index(buf.String(), "zero")
	oneIndex := strings.Index(buf.String(), "one")
	twoIndex := strings.Index(buf.String(), "two")
	threeIndex := strings.Index(buf.String(), "three")
	if zeroIndex >= oneIndex || oneIndex >= twoIndex || twoIndex >= threeIndex {
		t.Errorf("Messages not rendered in timestamp order, 0 at %d, 1 at %d, 2 at %d, 3 at %d", zeroIndex, oneIndex, twoIndex, threeIndex)
	}
}
