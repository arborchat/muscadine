package tui

import (
	"log"
	"sync"

	arbor "github.com/arborchat/arbor-go"
	"github.com/jroimartin/gocui"
)

const historyView = "history"
const editView = "edit"
const preEditViewTitle = "Arrows to select, hit enter to reply"
const midEditViewTitle = "Type your reply, hit enter to send"

// Composer writes and sends protocol messages
type Composer interface {
	Reply(string, string) error
	Query(string)
}

// TUI is the default terminal user interface implementation for this client
type TUI struct {
	*gocui.Gui
	done     chan struct{}
	messages chan *arbor.ChatMessage
	Composer
	histState *HistoryState
	init      sync.Once
	editMode  bool
}

// NewTUI creates a new terminal user interface. The provided channel will be
// used to relay any protocol messages initiated by the TUI.
func NewTUI(composer Composer) (*TUI, error) {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, err
	}
	gui.InputEsc = true
	hs, err := NewHistoryState()
	if err != nil {
		return nil, err
	}

	t := &TUI{
		Gui:       gui,
		messages:  make(chan *arbor.ChatMessage),
		histState: hs,
		Composer:  composer,
	}
	t.done = t.mainLoop()

	go t.update()
	return t, err
}

// mainLoop sets up the TUI and runs its event loop in a goroutine
// until it tries to exit. The channel that it returns will close
// when the TUI event loop ends, which can be used to block until
// the TUI exits.
func (t *TUI) mainLoop() chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer t.Close()

		t.SetManagerFunc(t.layout)

		if err := t.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
			log.Println("Failed registering exit keystroke handler", err)
		}

		if err := t.MainLoop(); err != nil && err != gocui.ErrQuit {
			log.Println("Error during UI redraw", err)
		}
	}()
	return done
}

// AwaitExit unconditionally blocks until the TUI exits.
func (t *TUI) AwaitExit() {
	<-t.done
}

// update listens for new messages to display and redraws the screen.
func (t *TUI) update() {
	for message := range t.messages {
		// can't do this inside the loop or it will bind the wrong value of
		// `message` and will be prone to race conditions on whether the
		// `New()` method is invoked before the value of `message` is
		// reassigned.
		err := t.histState.New(message)
		if err != nil {
			log.Println(err)
		}

		t.reRender()
	}
}

// Display adds the provided message to the visible interface.
func (t *TUI) Display(message *arbor.ChatMessage) {
	t.messages <- message
}

// quit asks the TUI to stop running. Should only be called as
// a keystroke or mouse input handler.
func quit(c *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// cursorDown attempt to move the selected message downward through the message
// history.
func (t *TUI) cursorDown(c *gocui.Gui, v *gocui.View) error {
	t.histState.CursorDown()
	t.reRender()
	return nil
}

// cursorUp attempt to move the selected message upward through the message
// history.
func (t *TUI) cursorUp(c *gocui.Gui, v *gocui.View) error {
	t.histState.CursorUp()
	t.reRender()
	return nil
}

// scrollDown attempts to move the view downwards through the history.
func (t *TUI) scrollDown(c *gocui.Gui, v *gocui.View) error {
	currentX, currentY := v.Origin()
	maxY := t.histState.Height()
	_, viewHeight := v.Size()
	if currentY < (maxY - viewHeight - 2) {
		return v.SetOrigin(currentX, currentY+1)
	}
	return nil
}

// scrollUp attempts to move the view upwards through the history.
func (t *TUI) scrollUp(c *gocui.Gui, v *gocui.View) error {
	currentX, currentY := v.Origin()
	if currentY > 0 {
		return v.SetOrigin(currentX, currentY-1)
	}
	return nil
}

// reRender forces a redraw of the historyView
func (t *TUI) reRender() {
	t.Update(func(g *gocui.Gui) error {
		v, err := g.View(historyView)
		if err != nil {
			return err
		}
		v.Clear()
		return t.histState.Render(v)
	})
}

// historyMode transitions the TUI to interactively scroll the history.
// All state change related to that transition should be defined here.
func (t *TUI) historyMode() error {
	t.Gui.Cursor = false
	v, err := t.Gui.View(editView)
	if err != nil {
		return err
	}
	v.Title = preEditViewTitle
	_, err = t.Gui.SetCurrentView(historyView)
	if err != nil {
		return err
	}
	return nil
}

// composeMode transitions the TUI to interactively editing messages.
// All state change related to that transition should be defined here.
func (t *TUI) composeMode() error {
	t.Gui.Cursor = true
	v, err := t.Gui.SetCurrentView(editView)
	if err != nil {
		return err
	}
	v.Title = midEditViewTitle
	return nil

}

// composeReply starts replying to the current message.
func (t *TUI) composeReply(c *gocui.Gui, v *gocui.View) error {
	return t.composeMode()
}

// cancelReply exits compose mode and returns to history mode.
func (t *TUI) cancelReply(c *gocui.Gui, v *gocui.View) error {
	return t.historyMode()
}

// sendReply starts replying to the current message.
func (t *TUI) sendReply(c *gocui.Gui, v *gocui.View) error {
	content := v.Buffer()
	if len(content) < 2 {
		// don't allow messages shorter than one character
		return nil
	}
	v.Clear()
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)
	t.Composer.Reply(t.histState.Current(), content[:len(content)-1])
	return t.historyMode()
}

// layout places views in the UI.
func (t *TUI) layout(gui *gocui.Gui) error {
	keybindings := []struct {
		View        string
		Key         interface{}
		Modifier    gocui.Modifier
		Handler     func(*gocui.Gui, *gocui.View) error
		HandlerName string
	}{
		{historyView, gocui.KeyArrowDown, gocui.ModNone, t.cursorDown, "cursorDown"},
		{historyView, 'j', gocui.ModNone, t.cursorDown, "cursorDown"},
		{historyView, gocui.KeyArrowRight, gocui.ModNone, t.scrollDown, "scrollDown"},
		{historyView, 'l', gocui.ModNone, t.scrollDown, "scrollDown"},
		{historyView, gocui.KeyArrowUp, gocui.ModNone, t.cursorUp, "cursorUp"},
		{historyView, 'k', gocui.ModNone, t.cursorUp, "cursorUp"},
		{historyView, gocui.KeyArrowLeft, gocui.ModNone, t.scrollUp, "scrollUp"},
		{historyView, 'h', gocui.ModNone, t.scrollUp, "scrollUp"},
		{historyView, gocui.KeyEnter, gocui.ModNone, t.composeReply, "composeReply"},
		{editView, gocui.KeyEnter, gocui.ModNone, t.sendReply, "sendReply"},
		{editView, gocui.KeyEsc, gocui.ModNone, t.cancelReply, "cancelReply"},
	}
	mX, mY := gui.Size()
	histMaxX := mX - 1
	histMaxY := mY - 1
	histMaxY -= 3
	t.histState.SetDimensions(histMaxY-1, histMaxX-1)
	histView, err := gui.SetView(historyView, 0, 0, histMaxX, histMaxY)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		histView.Title = "Chat History"

		for _, binding := range keybindings {
			if binding.View == historyView {
				if err := t.SetKeybinding(binding.View, binding.Key, binding.Modifier, binding.Handler); err != nil {
					log.Printf("Failed registering %s keystroke handler: %v\n", binding.HandlerName, err)
				}
			}
		}
		if _, err := t.SetCurrentView(historyView); err != nil {
			log.Println("Failed to set historyView focus", err)
		}
	}

	if v, err := gui.SetView(editView, 0, histMaxY+1, histMaxX, mY-1); err != nil {
		if err != gocui.ErrUnknownView {
			log.Println("Error creating editView", err)
		}
		v.Editable = true
		v.Editor = gocui.DefaultEditor
		v.Title = preEditViewTitle

		for _, binding := range keybindings {
			if binding.View == editView {
				if err := t.SetKeybinding(binding.View, binding.Key, binding.Modifier, binding.Handler); err != nil {
					log.Printf("Failed registering %s keystroke handler: %v\n", binding.HandlerName, err)
				}
				log.Printf("registered binding for %s", binding.HandlerName)
			}
		}
	}

	return nil
}
