package tui

import (
	"fmt"
	"log"
	"sync"
	"time"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/types"
	"github.com/jroimartin/gocui"
)

const historyView = "history"
const editView = "edit"
const globalView = ""
const histViewTitlePrefix = "Chat History"

// TUI is the default terminal user interface implementation for this client
type TUI struct {
	*gocui.Gui
	done     chan struct{}
	messages chan *arbor.ChatMessage
	types.Composer
	*Editor
	histState      *HistoryState
	init           sync.Once
	editMode       bool
	connected      bool
	lastKnownWidth int
}

// NewTUI creates a new terminal user interface. The provided channel will be
// used to relay any protocol messages initiated by the TUI.
func NewTUI(client types.Client) (*TUI, error) {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, err
	}
	//gui.InputEsc = true
	hs, err := NewHistoryState(client)
	if err != nil {
		return nil, err
	}

	t := &TUI{
		Gui:       gui,
		messages:  make(chan *arbor.ChatMessage),
		histState: hs,
		Composer:  client,
		Editor:    NewEditor(),
	}
	client.OnReceive(t.Display)
	t.done = t.mainLoop()
	go t.manageConnection(client)

	go t.update()
	return t, err
}

// manageConnection handles the actual policy of when to connect from and disconnect
// from the server. The current implementation tries to connect as soon as possible,
// then tries to reconnect every 5 seconds if it is disconnected.
func (t *TUI) manageConnection(c types.Connection) {
	disconnected := make(chan struct{})
	c.OnDisconnect(func(disconn types.Connection) {
		disconnected <- struct{}{}
	})
	for {
		for {
			err := c.Connect()
			if err != nil {
				time.Sleep(time.Second * 5)
				continue
			}
			go func() {
				for i := 0; i < 5; i++ {
					if root, err := t.histState.Root(); err == nil {
						t.Composer.Reply(root, "[join]")
						return
					}
					time.Sleep(5 * time.Second)
				}
				log.Println("Gave up greeting server")
			}()
			break
		}
		t.connected = true
		t.reRender()
		<-disconnected
		t.connected = false
		t.reRender()
		// if we get here, we've been disconnected and will now loop around to a
		// connection attempt
	}
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

		makeHist := gocui.ManagerFunc(t.layout)
		layout := gocui.ManagerFunc(BottomPrimaryLayout(historyView, editView))
		t.SetManager(t.Editor, makeHist, layout)

		for _, binding := range t.Keybindings() {
			if err := t.SetKeybinding(binding.View, binding.Key, binding.Modifier, binding.Handler); err != nil {
				log.Printf("Failed registering %s keystroke handler on view %v: %s\n", binding.HandlerName, binding.View, err)
			}
		}

		if err := t.MainLoop(); err != nil && err != gocui.ErrQuit {
			log.Println("Error during UI redraw", err)
		} else if err != nil {
			log.Println(err)
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
func (t *TUI) quit(c *gocui.Gui, v *gocui.View) error {
	if root, err := t.histState.Root(); err != nil {
		log.Println("Not notifying that we quit:", err)
	} else {
		t.Composer.Reply(root, "[quit]")
		time.Sleep(time.Millisecond * 250) // wait in the hope that Quit will be sent
	}
	return gocui.ErrQuit
}

// cursorDown attempt to move the selected message downward through the message
// history.
func (t *TUI) cursorDown(c *gocui.Gui, v *gocui.View) error {
	t.histState.CursorDown()
	t.reRender()
	_, cursorEnd := t.histState.CursorLines()
	_, currentY := v.Origin()
	_, viewHeight := v.Size()
	for currentY+viewHeight-1 <= cursorEnd {
		t.scrollDown(c, v)
		_, currentY = v.Origin()
	}
	return nil
}

// cursorUp attempt to move the selected message upward through the message
// history.
func (t *TUI) cursorUp(c *gocui.Gui, v *gocui.View) error {
	t.histState.CursorUp()
	t.reRender()
	cursorStart, _ := t.histState.CursorLines()
	_, currentY := v.Origin()
	for currentY > 0 && currentY+1 > cursorStart {
		t.scrollUp(c, v)
		_, currentY = v.Origin()
	}
	return nil
}

// scrollDown attempts to move the view downwards through the history.
func (t *TUI) scrollDown(c *gocui.Gui, v *gocui.View) error {
	currentX, currentY := v.Origin()
	maxY := t.histState.Height()
	if currentY < (maxY - 1) {
		return v.SetOrigin(currentX, currentY+1)
	}
	return nil
}

// scrollBottom attempts to move the view to the end of the history.
func (t *TUI) scrollBottom(c *gocui.Gui, v *gocui.View) error {
	currentX, currentY := v.Origin()
	_, viewHeight := v.Size()
	maxY := t.histState.Height()
	t.histState.CursorEnd()
	t.reRender()
	// ensure that we are both not already at the end *and* that the
	// end is off-screen
	if currentY < (maxY-1) && maxY > viewHeight {
		return v.SetOrigin(currentX, maxY-viewHeight)
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

// scrollTop jumps to the top of the history.
func (t *TUI) scrollTop(c *gocui.Gui, v *gocui.View) error {
	currentX, _ := v.Origin()
	t.histState.CursorBeginning()
	t.reRender()
	return v.SetOrigin(currentX, 0)
}

// queryNeeded sends a batch of queries to the server to update the history.
func (t *TUI) queryNeeded(c *gocui.Gui, v *gocui.View) error {
	needed := t.histState.Needed(10)
	for _, n := range needed {
		t.Query(n)
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
		needed := t.histState.Needed(100)
		var suffix string
		if !t.connected {
			suffix += "Connecting... "
		} else {
			suffix += "Connected, "
		}
		if len(needed) == 0 {
			suffix += "all known threads complete"
		} else {
			suffix += fmt.Sprintf("%d+ broken threads, q to query", len(needed))
		}
		if msg := t.histState.Get(t.histState.Current()); msg != nil {
			timestamp := time.Unix(msg.Timestamp, 0).Local().Format(time.UnixDate)
			v.Title = histViewTitlePrefix + " | Selected: " + timestamp + " | " + suffix
		}
		return t.histState.Render(v)
	})
}

// historyMode transitions the TUI to interactively scroll the history.
// All state change related to that transition should be defined here.
func (t *TUI) historyMode() error {
	err := t.Editor.Unfocus()
	if err != nil {
		return err
	}
	_, err = t.Gui.SetCurrentView(historyView)
	if err != nil {
		return err
	}
	return nil
}

// composeMode transitions the TUI to interactively editing messages.
// All state change related to that transition should be defined here.
func (t *TUI) composeMode(replyTo *arbor.ChatMessage) error {
	return t.Editor.Focus(replyTo)
}

// composeReply starts replying to the current message.
func (t *TUI) composeReply(c *gocui.Gui, v *gocui.View) error {
	msg := t.histState.Get(t.histState.Current())
	return t.composeMode(msg)
}

// composeReplyToRoot starts replying to the earliest known message (root, unless something is very wrong).
func (t *TUI) composeReplyToRoot(c *gocui.Gui, v *gocui.View) error {
	root, err := t.histState.Root()
	if err != nil {
		return err
	}
	rootMsg := t.histState.Get(root)
	return t.composeMode(rootMsg)
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
	t.Composer.Reply(t.Editor.ReplyTo.UUID, content[:len(content)-1])
	return t.historyMode()
}

// layout places views in the UI.
func (t *TUI) layout(gui *gocui.Gui) error {
	mX, mY := gui.Size()
	histMaxX := mX - 1
	histMaxY := mY - 1
	histMaxY -= 3
	t.histState.SetDimensions(histMaxY-1, histMaxX-1)
	if t.lastKnownWidth != mX {
		t.reRender()
	}
	t.lastKnownWidth = mX
	// update view dimensions or create for the first time
	histView, err := gui.SetView(historyView, 0, 0, histMaxX, histMaxY)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		// if you reach this point, you are initializing the view for the first time
		histView.Title = "Chat History"

		if _, err := t.SetCurrentView(historyView); err != nil {
			log.Println("Failed to set historyView focus", err)
		}
	}

	return nil
}
