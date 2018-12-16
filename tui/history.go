package tui

import (
	"time"

	"github.com/jroimartin/gocui"
)

// History acts as a controller for a gocui.View that holds chat history.
type History struct {
	name string
	// Each of these booleans represents whether or not a state change is requested next time
	// the Layout method is invoked. This decouples the History type from the view that it manages
	// except for Layout and the Action functions
	focus, unfocus, render bool
	*HistoryState
}

// NewHistory creates a new controller for an History view.
func NewHistory(state *HistoryState) *History {
	return &History{name: historyView, HistoryState: state}
}

// Focus lets the History perform any changes needed when it gains focus.
func (h *History) Focus() error {
	h.focus = true
	return nil
}

// Unfocus lets the History perform any changes needed when it loses focus.
func (h *History) Unfocus() error {
	h.unfocus = true
	return nil
}

// ActionCursorUp attempt to move the selected message upward through the message
// history.
func (h *History) ActionCursorUp(c *gocui.Gui, v *gocui.View) error {
	h.CursorUp()
	cursorStart, _ := h.CursorLines()
	_, currentY := v.Origin()
	for currentY > 0 && currentY+1 > cursorStart {
		h.ActionScrollUp(c, v)
		_, currentY = v.Origin()
	}
	h.reRender()
	return nil
}

// ActionCursorDown attempt to move the selected message downward through the message
// history.
func (h *History) ActionCursorDown(c *gocui.Gui, v *gocui.View) error {
	h.CursorDown()
	_, cursorEnd := h.CursorLines()
	_, currentY := v.Origin()
	_, viewHeight := v.Size()
	for currentY+viewHeight-1 <= cursorEnd {
		h.ActionScrollDown(c, v)
		_, currentY = v.Origin()
	}
	h.reRender()
	return nil
}

// ActionScrollTop jumps to the top of the history.
func (h *History) ActionScrollTop(c *gocui.Gui, v *gocui.View) error {
	currentX, _ := v.Origin()
	h.CursorBeginning()
	h.renderView(v)
	return v.SetOrigin(currentX, 0)
}

// ActionScrollBottom attempts to move the view to the end of the history.
func (h *History) ActionScrollBottom(c *gocui.Gui, v *gocui.View) error {
	currentX, currentY := v.Origin()
	_, viewHeight := v.Size()
	maxY := h.Height()
	h.CursorEnd()
	// ensure that we are both not already at the end *and* that the
	// end is off-screen
	if currentY < (maxY-1) && maxY > viewHeight {
		return v.SetOrigin(currentX, maxY-viewHeight)
	}
	h.renderView(v)
	return nil
}

// ActionScrollUp attempts to move the view upwards through the history.
func (h *History) ActionScrollUp(c *gocui.Gui, v *gocui.View) error {
	currentX, currentY := v.Origin()
	if currentY > 0 {
		return v.SetOrigin(currentX, currentY-1)
	}
	return nil
}

// ActionScrollDown attempts to move the view downwards through the history.
func (h *History) ActionScrollDown(c *gocui.Gui, v *gocui.View) error {
	currentX, currentY := v.Origin()
	maxY := h.Height()
	if currentY < (maxY - 1) {
		return v.SetOrigin(currentX, currentY+1)
	}
	return nil
}

// reRender requests a redraw of the history at the next invocation of Layout()
func (h *History) reRender() {
	h.render = true
}

// renderView actually performs the work of redrawing the chat history into the given view
func (h *History) renderView(v *gocui.View) error {
	v.Clear()
	if msg := h.Get(h.Current()); msg != nil {
		timestamp := time.Unix(msg.Timestamp, 0).Local().Format(time.UnixDate)
		v.Title = histViewTitlePrefix + " | Selected: " + timestamp
	}
	return h.Render(v)
}

// Layout is responsible for setting the desired view dimensions for the
// History, but *not* for setting its position. That is handled by a higher-order
// layout function.
// Layout also executes any state changes requested by other methods, such as gaining focus
// or clearing the History's contents.
func (h *History) Layout(g *gocui.Gui) error {
	// Set the view's dimensions
	width, height := g.Size()
	h.SetDimensions(height-1, width-1)
	v, err := g.SetView(h.name, 0, 0, width-1, height-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		// If we are creating the view for the first time, configure its settings
		v.Frame = true
		v.Title = "Chat History"

	}
	if h.focus {
		g.SetCurrentView(h.name)
		h.focus = false
	} else if h.unfocus {
		h.unfocus = false
	}
	if h.render {
		h.render = false
		err = h.renderView(v)
		if err != nil {
			return err
		}
	}
	return nil
}
