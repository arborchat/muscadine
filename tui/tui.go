package tui

import (
	"log"

	arbor "github.com/arborchat/arbor-go"
	"github.com/jroimartin/gocui"
)

const historyView = "history"
const editView = "edit"
const preEditViewTitle = "Arrows to select, hit enter to reply"
const midEditViewTitle = "Type your reply, hit enter to send"

// TUI is the default terminal user interface implementation for this client
type TUI struct {
	*gocui.Gui
	done      chan struct{}
	messages  chan *arbor.ChatMessage
	sendChan  chan<- *arbor.ProtocolMessage
	histState *HistoryState
	editMode  bool
}

// NewTUI creates a new terminal user interface. The provided channel will be
// used to relay any protocol messages initiated by the TUI.
func NewTUI(sendChan chan<- *arbor.ProtocolMessage) (*TUI, error) {
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
		sendChan:  sendChan,
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

// send relays a ProtocolMessage to a server.
func (t *TUI) send(proto *arbor.ProtocolMessage) {
	// ensure we don't block on this
	go func() {
		t.sendChan <- proto
	}()
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

// composeReply starts replying to the current message.
func (t *TUI) composeReply(c *gocui.Gui, v *gocui.View) error {
	v, err := t.Gui.SetCurrentView(editView)
	if err != nil {
		return err
	}
	v.Title = midEditViewTitle

	return nil
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
	v.Title = preEditViewTitle
	chat, err := arbor.NewChatMessage(content[:len(content)-1])
	if err != nil {
		return err
	}
	chat.Parent = t.histState.Current()
	chat.Username = "muscadine"
	proto := &arbor.ProtocolMessage{ChatMessage: chat, Type: arbor.NewMessageType}
	t.send(proto)
	t.Gui.SetCurrentView(historyView)
	return err
}

// layout places views in the UI.
func (t *TUI) layout(gui *gocui.Gui) error {
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
		histView.Wrap = true
		// Ensure that keybindings are only registered once.
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

		if err := t.SetKeybinding(historyView, gocui.KeyEnter, gocui.ModNone, t.composeReply); err != nil {
			log.Println("Failed registering cursorUp keystroke handler", err)
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
		v.Wrap = true
		v.Editor = gocui.DefaultEditor
		v.Title = preEditViewTitle

		if err := t.SetKeybinding(editView, gocui.KeyEnter, gocui.ModNone, t.sendReply); err != nil {
			log.Println("Failed registering cursorUp keystroke handler", err)
		}

	}

	return nil
}
