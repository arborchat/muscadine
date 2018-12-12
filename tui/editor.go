package tui

import (
	arbor "github.com/arborchat/arbor-go"
	"github.com/jroimartin/gocui"
)

const borderHeight = 2
const preEditViewTitle = "Arrows to select, hit enter to reply"
const midEditViewTitle = "Type your reply, hit enter to send"

// Editor acts as a controller for an editable gocui.View
// This editor provides a layout function that will manage its own size,
// but not its position within the TUI.
// Its size when empty will be 1 internal row, but will expand as content
// is added.
type Editor struct {
	name    string
	h       int
	Title   string
	ReplyTo *arbor.ChatMessage
	Content string
}

// NewEditor creates a new controller for an Editor view.
func NewEditor() *Editor {
	return &Editor{name: editView, h: borderHeight, Title: preEditViewTitle}
}

// Focus lets the Editor perform any changes needed when it gains focus. It should
// always be called from within a gocui.Update function (or similar) so that the
// changes are rendered immediately
func (e *Editor) Focus(g *gocui.Gui, replyTo *arbor.ChatMessage) error {
	e.Title = midEditViewTitle
	e.ReplyTo = replyTo
	g.Cursor = true
	_, err := g.SetCurrentView(e.name)
	if err != nil {
		return err
	}
	return nil

}

// Unfocus lets the Editor perform any changes needed when it loses focus. It should be
// called under the same conditions as `Focus`.
func (e *Editor) Unfocus(g *gocui.Gui) error {
	e.Title = preEditViewTitle
	e.ReplyTo = nil
	g.Cursor = false
	return nil
}

// Clear erases the current contents of the editor. This should be performed within a gocui.Update
// context.
func (e *Editor) Clear(g *gocui.Gui) error {
	v, err := g.View(e.name)
	if err != nil {
		return err
	}
	v.Clear()
	e.Content = ""
	e.ReplyTo = nil
	return nil
}

// insertNewline adds a newline character into the editor at the current cursor position.
func (e *Editor) insertNewline(g *gocui.Gui, v *gocui.View) error {
	v.EditNewLine()
	return nil
}

// insertTab adds a tab character into the editor at the current cursor position.
func (e *Editor) insertTab(g *gocui.Gui, v *gocui.View) error {
	v.EditWrite(' ')
	v.EditWrite(' ')
	v.EditWrite(' ')
	v.EditWrite(' ')
	return nil
}

// Layout is responsible for setting the desired view dimensions for the
// Editor, but *not* for setting its position. That is handled by a higher-order
// layout function.
func (e *Editor) Layout(g *gocui.Gui) error {
	// If the new has already been initialized, update its height to reflect its
	// current contents
	if v, err := g.View(e.name); err == nil {
		lines := len(v.BufferLines())
		e.h = lines + borderHeight
		e.Content = v.Buffer()
	}
	// Set the view's dimensions
	width, _ := g.Size()
	v, err := g.SetView(e.name, 0, 0, width-1, e.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		// If we are creating the view for the first time, configure its settings
		v.Editable = true
		v.Editor = gocui.DefaultEditor
		v.Frame = true
	}
	// update the title regardless of whether this is the first-time initialization
	if e.ReplyTo == nil {
		v.Title = e.Title
	} else {
		v.Title = e.Title + " | replying to " + e.ReplyTo.Username
	}
	return nil
}
