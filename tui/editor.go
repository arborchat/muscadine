package tui

import (
	arbor "github.com/arborchat/arbor-go"
	"github.com/whereswaldon/gocui"
)

const borderHeight = 2
const preEditViewTitle = "Arrows to select, hit enter to reply"
const midEditViewTitle = "Type your reply, hit enter to send"

// Editor acts as a controller for an editable gocui.View
// This editor provides a layout function that will manage its own size,
// but not its position within the TUI.
// Its size when empty will be 1 internal row, but will expand as content
// is added.
//
// Insofar as it is possible, Editor is decoupled from the *gocui.View.
// Their only interactions are in the Layout function, which synchronizes
// the state of this controller with the state of the view, and the
// Action* methods defined on the editor (these are meant as keystroke
// handlers)
type Editor struct {
	name    string
	h       int
	Title   string
	ReplyTo *arbor.ChatMessage
	Content string
	// Each of these booleans represents whether or not a state change is requested next time
	// the Layout method is invoked. This decouples the Editor type from the view that it manages
	// except for Layout and the Action functions
	focus, unfocus, clear bool
}

// NewEditor creates a new controller for an Editor view.
func NewEditor() *Editor {
	return &Editor{name: editView, h: borderHeight, Title: preEditViewTitle}
}

// Focus lets the Editor perform any changes needed when it gains focus. It should
// always be called from within a gocui.Update function (or similar) so that the
// changes are rendered immediately
func (e *Editor) Focus(replyTo *arbor.ChatMessage) error {
	e.Title = midEditViewTitle
	e.ReplyTo = replyTo
	e.focus = true
	return nil
}

// Unfocus lets the Editor perform any changes needed when it loses focus. It should be
// called under the same conditions as `Focus`.
func (e *Editor) Unfocus() error {
	e.Title = preEditViewTitle
	e.ReplyTo = nil
	e.unfocus = true
	return nil
}

// Clear erases the current contents of the editor. This should be performed within a gocui.Update
// context.
func (e *Editor) Clear() error {
	e.clear = true
	e.Content = ""
	e.ReplyTo = nil
	return nil
}

// ActionInsertNewline adds a newline character into the editor at the current cursor position.
func (e *Editor) ActionInsertNewline(g *gocui.Gui, v *gocui.View) error {
	v.EditNewLine()
	return nil
}

// ActionInsertTab adds a tab character into the editor at the current cursor position.
func (e *Editor) ActionInsertTab(g *gocui.Gui, v *gocui.View) error {
	v.EditWrite(' ')
	v.EditWrite(' ')
	v.EditWrite(' ')
	v.EditWrite(' ')
	return nil
}

// Layout is responsible for setting the desired view dimensions for the
// Editor, but *not* for setting its position. That is handled by a higher-order
// layout function.
// Layout also executes any state changes the editor requests, such as gaining focus
// or clearing the Editor's contents.
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
		v.Editor = &EditCore{}
		v.Frame = true
		v.Wrap = false
	}
	// update the title regardless of whether this is the first-time initialization
	if e.ReplyTo == nil {
		v.Title = e.Title
	} else {
		v.Title = e.Title + " | replying to " + e.ReplyTo.Username
	}

	if e.focus {
		g.Cursor = true
		g.SetCurrentView(e.name)
		e.focus = false
	} else if e.unfocus {
		g.Cursor = false
		e.unfocus = false
	}
	if e.clear {
		v.Clear()
		e.clear = false
	}
	return nil
}

type EditCore struct {
}

// Edit handles a single keypress in the editor
// This is a modification of gocui's simpleEditor function.
func (e *EditCore) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	case key == gocui.KeyEnter:
		v.EditNewLine()
	case key == gocui.KeyArrowDown:
		v.MoveCursor(0, 1, false)
	case key == gocui.KeyArrowUp:
		v.MoveCursor(0, -1, false)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	}
}
