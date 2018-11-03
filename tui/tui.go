package tui

import (
	"log"
	"sync"

	arbor "github.com/arborchat/arbor-go"
	"github.com/jroimartin/gocui"
)

const historyView = "history"

// TUI is the default terminal user interface implementation for this client
type TUI struct {
	*gocui.Gui
	done      chan struct{}
	messages  chan *arbor.ChatMessage
	histState *HistoryState
	init      sync.Once
}

// NewTUI creates a new terminal user interface.
func NewTUI() (*TUI, error) {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, err
	}
	hs, err := NewHistoryState()
	if err != nil {
		return nil, err
	}

	t := &TUI{
		Gui:       gui,
		messages:  make(chan *arbor.ChatMessage),
		histState: hs,
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

		t.Update(func(g *gocui.Gui) error {
			v, err := g.View(historyView)
			if err != nil {
				return err
			}
			v.Clear()
			return t.histState.Render(v)
		})
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
	t.Update(func(g *gocui.Gui) error {
		v, err := g.View(historyView)
		if err != nil {
			return err
		}
		v.Clear()
		return t.histState.Render(v)
	})
	return nil
}

// cursorUp attempt to move the selected message upward through the message
// history.
func (t *TUI) cursorUp(c *gocui.Gui, v *gocui.View) error {
	t.histState.CursorUp()
	t.Update(func(g *gocui.Gui) error {
		v, err := g.View(historyView)
		if err != nil {
			return err
		}
		v.Clear()
		return t.histState.Render(v)
	})
	return nil
}

// layout places views in the UI.
func (t *TUI) layout(gui *gocui.Gui) error {
	mX, mY := gui.Size()
	_, err := gui.SetView(historyView, 0, 0, mX-1, mY-1)
	t.histState.SetDimensions(mY-2, mX-2)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	// Ensure that keybindings are only registered once.
	t.init.Do(func() {
		if err := t.SetKeybinding(historyView, gocui.KeyArrowDown, gocui.ModNone, t.cursorDown); err != nil {
			log.Println("Failed registering cursorDown keystroke handler", err)
		}

		if err := t.SetKeybinding(historyView, 'j', gocui.ModNone, t.cursorDown); err != nil {
			log.Println("Failed registering cursorDown keystroke handler", err)
		}

		if err := t.SetKeybinding(historyView, gocui.KeyArrowUp, gocui.ModNone, t.cursorUp); err != nil {
			log.Println("Failed registering cursorUp keystroke handler", err)
		}

		if err := t.SetKeybinding(historyView, 'k', gocui.ModNone, t.cursorUp); err != nil {
			log.Println("Failed registering cursorUp keystroke handler", err)
		}

		if _, err := t.SetCurrentView(historyView); err != nil {
			log.Println("Failed to set historyView focus", err)
		}
	})

	return nil
}
