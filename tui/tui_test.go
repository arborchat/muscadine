package tui_test

import (
	"bytes"
	"strings"
	"testing"

	arbor "github.com/arborchat/arbor-go"
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
		hist.New(&testMsg)
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

// TestRenderMessage ensures that the rendering function correctly handles line wrapping
// and related problems when preparing messages to be displayed.
func TestRenderMessage(t *testing.T) {
	hist, err := tui.NewHistoryState()
	if err != nil {
		t.Skip("Should have been able to construct HistoryState with valid params", err)
	}
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
	rendered := hist.RenderMessage(&message, startingWidth)
	if rendered == nil {
		t.Fatal("Render produced nil output for 1 line message")
	}
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
		rendered := hist.RenderMessage(&message, i)
		if rendered == nil {
			t.Fatalf("Render produced nil for %d width message", i)
		}
		for index, line := range rendered {
			if index == 0 {
				if string(line[:len(prefix)]) != prefix {
					t.Errorf("Expected prefix \"%s\" on line %d, got \"%s\"", prefix, index, line)
				}
			} else {
				if string(line[:len(prefix)]) != strings.Repeat(" ", prefixWidth) {
					t.Errorf("Expected line %d to start with %d spaces, found \"%s\"", index, prefixWidth, line)
				}
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
	hist, err := tui.NewHistoryState()
	if err != nil {
		t.Skip("Should have been able to construct HistoryState with valid params", err)
	}
	message := testMsg
	id := hist.Current()
	if id != "" {
		t.Errorf("Expected empty history to have empty current, got \"%s\"", id)
	}
	hist.New(&message)
	id = hist.Current()
	if id != message.UUID {
		t.Errorf("Expected hist to set current to first received message (\"%s\"), got \"%s\"", message.UUID, id)
	}
	second := testMsg
	second.UUID = "different"
	hist.New(&second)
	id = hist.Current()
	if id != message.UUID {
		t.Errorf("Expected hist to keep first received message id as current (\"%s\"), got \"%s\"", message.UUID, id)
	}
}

// TestRenderSelectMessage ensures that the current message in a HistoryState is rendered in a different
// color.
func TestRenderSelectMessage(t *testing.T) {
	hist, err := tui.NewHistoryState()
	if err != nil {
		t.Skip("Should have been able to construct HistoryState with valid params", err)
	}
	hist.SetDimensions(24, 80)
	message := testMsg
	hist.New(&message)
	rendered := hist.RenderMessage(&message, 80)
	if !strings.Contains(string(rendered[0]), tui.CurrentColor) && !strings.Contains(string(rendered[0]), tui.CurrentColor) {
		t.Error("Expected current message to be rendered in color", string(rendered[0]))
	}
	second := testMsg
	second.UUID = "different"
	rendered = hist.RenderMessage(&second, 80)
	if strings.Contains(string(rendered[0]), tui.CurrentColor) || strings.Contains(string(rendered[0]), tui.CurrentColor) {
		t.Error("Did not expect non-current message to be rendered in color", string(rendered[0]))
	}
}

// TestRenderEmptyMessage ensures that the HistoryState doesn't panic when trying to render
// messages whose content contains empty lines.
func TestRenderEmptyMessage(t *testing.T) {
	hist, err := tui.NewHistoryState()
	if err != nil {
		t.Skip("Should have been able to construct HistoryState with valid params", err)
	}
	hist.SetDimensions(24, 80)
	message := testMsg
	message.Content = ""
	// making sure that we don't crash
	hist.RenderMessage(&message, 80)

	// now check that we correctly render messages with a trailing newline
	message.Content = "\n"
	rendered := hist.RenderMessage(&message, 80)
	if !strings.HasSuffix(string(rendered[len(rendered)-1]), "\n") {
		t.Errorf("Should have inserted newline at end of empty rendered message line, found %v", rendered)
	}
}
