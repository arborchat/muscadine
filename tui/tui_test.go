package tui_test

import (
	"bytes"
	"strings"
	"testing"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/tui"
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
	hist, err := tui.NewHistoryState()
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
	if numFound > height {
		t.Errorf("With height=%d, Render should only display %d messages, got %d", height, height, numFound)
	}
}

func historyStateOrSkip(t *testing.T) *tui.HistoryState {
	hist, err := tui.NewHistoryState()
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
	hist := historyStateOrSkip(t)
	message := testMsg
	message.Content = "let's use a ลาญฤๅเข่นฆ่าบีฑ much longer string いろはにほへとちりぬるを so that line wrapping アサキユメミシhappens"
	separator := ": "
	rendered := hist.RenderMessage(&message)
	if !strings.HasPrefix(string(rendered), message.Username+separator) || !strings.Contains(string(rendered), message.Content) {
		t.Error("Render produced malformed message", string(rendered))
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

// TestRenderSelectMessage ensures that the current message in a HistoryState is rendered in a different
// color.
func TestRenderSelectMessage(t *testing.T) {
	hist := historyStateOrSkip(t)
	message := testMsg
	newOrSkip(t, hist, &message)
	rendered := hist.RenderMessage(&message)
	if !strings.Contains(string(rendered), tui.CurrentColor) && !strings.Contains(string(rendered), tui.ClearColor) {
		t.Error("Expected current message to be rendered in color", string(rendered))
	}
	second := testMsg
	second.UUID = "different"
	rendered = hist.RenderMessage(&second)
	if strings.Contains(string(rendered), tui.CurrentColor) || strings.Contains(string(rendered), tui.ClearColor) {
		t.Error("Did not expect non-current message to be rendered in color", string(rendered))
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
	hist.RenderMessage(&message)

	// now check that we correctly render messages with a trailing newline
	message.Content = "\n"
	rendered := hist.RenderMessage(&message)
	if !strings.HasSuffix(string(rendered), "\n") {
		t.Errorf("Should have inserted newline at end of empty rendered message line, found %v", string(rendered))
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
	message.Content = "one"
	message.Timestamp = 10
	newOrSkip(t, hist, &message)
	message3 := testMsg
	message3.Content = "three"
	message3.Timestamp = 30
	newOrSkip(t, hist, &message3)
	message2 := testMsg
	message2.Content = "two"
	message2.Timestamp = 20
	newOrSkip(t, hist, &message2)
	message0 := testMsg
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
