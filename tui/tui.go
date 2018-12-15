package tui

import (
	"log"
	"sync"
	"time"

	arbor "github.com/arborchat/arbor-go"
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
	Composer
	*Editor
	*History
	init           sync.Once
	editMode       bool
	connected      bool
	lastKnownWidth int
}

// NewTUI creates a new terminal user interface. The provided channel will be
// used to relay any protocol messages initiated by the TUI.
func NewTUI(client Client) (*TUI, error) {
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
		Gui:      gui,
		messages: make(chan *arbor.ChatMessage),
		History:  NewHistory(hs),
		Composer: client,
		Editor:   NewEditor(),
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
func (t *TUI) manageConnection(c Connection) {
	disconnected := make(chan struct{})
	c.OnDisconnect(func(disconn Connection) {
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
					if root, err := t.Root(); err == nil {
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

		err := t.History.Focus()
		if err != nil {
			log.Println("Error focusing chat history", err)
		}
		t.History.reRender()
		layout := gocui.ManagerFunc(BottomPrimaryLayout(historyView, editView))
		t.SetManager(t.Editor, t.History, layout)

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
		err := t.New(message)
		if err != nil {
			log.Println(err)
		}
		t.reRender()
		// trigger a layout update
		t.Gui.Update(func(g *gocui.Gui) error {
			return nil
		})

	}
}

// Display adds the provided message to the visible interface.
func (t *TUI) Display(message *arbor.ChatMessage) {
	t.messages <- message
}

// quit asks the TUI to stop running. Should only be called as
// a keystroke or mouse input handler.
func (t *TUI) quit(c *gocui.Gui, v *gocui.View) error {
	if root, err := t.Root(); err != nil {
		log.Println("Not notifying that we quit:", err)
	} else {
		t.Composer.Reply(root, "[quit]")
		time.Sleep(time.Millisecond * 250) // wait in the hope that Quit will be sent
	}
	return gocui.ErrQuit
}

// queryNeeded sends a batch of queries to the server to update the history.
func (t *TUI) queryNeeded(c *gocui.Gui, v *gocui.View) error {
	needed := t.Needed(10)
	for _, n := range needed {
		t.Query(n)
	}
	return nil
}

// historyMode transitions the TUI to interactively scroll the history.
// All state change related to that transition should be defined here.
func (t *TUI) historyMode() error {
	err := t.Editor.Unfocus()
	if err != nil {
		return err
	}
	return t.History.Focus()
}

// composeMode transitions the TUI to interactively editing messages.
// All state change related to that transition should be defined here.
func (t *TUI) composeMode(replyTo *arbor.ChatMessage) error {
	err := t.History.Unfocus()
	if err != nil {
		return err
	}
	return t.Editor.Focus(replyTo)
}

// composeReply starts replying to the current message.
func (t *TUI) composeReply(c *gocui.Gui, v *gocui.View) error {
	msg := t.Get(t.Current())
	return t.composeMode(msg)
}

// composeReplyToRoot starts replying to the earliest known message (root, unless something is very wrong).
func (t *TUI) composeReplyToRoot(c *gocui.Gui, v *gocui.View) error {
	root, err := t.Root()
	if err != nil {
		return err
	}
	rootMsg := t.Get(root)
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
