package tui

import "github.com/jroimartin/gocui"

// Binding represents a binding between a keypress and a handler function
type Binding struct {
	View        string
	Key         interface{}
	Modifier    gocui.Modifier
	Handler     func(*gocui.Gui, *gocui.View) error
	HandlerName string
}

// Keybindings returns the default keybindings
func (t *TUI) Keybindings() []Binding {
	return []Binding{
		{globalView, gocui.KeyCtrlC, gocui.ModNone, t.quit, "quit"},
		{historyView, gocui.KeyArrowDown, gocui.ModNone, t.History.ActionCursorDown, "CursorDown"},
		{historyView, 'j', gocui.ModNone, t.History.ActionCursorDown, "CursorDown"},
		{historyView, gocui.KeyArrowRight, gocui.ModNone, t.History.ActionScrollDown, "ScrollDown"},
		{historyView, 'l', gocui.ModNone, t.History.ActionScrollDown, "ScrollDown"},
		{historyView, gocui.KeyArrowUp, gocui.ModNone, t.ActionCursorUp, "CursorUp"},
		{historyView, 'k', gocui.ModNone, t.ActionCursorUp, "CursorUp"},
		{historyView, gocui.KeyArrowLeft, gocui.ModNone, t.ActionScrollUp, "ScrollUp"},
		{historyView, 'h', gocui.ModNone, t.ActionScrollUp, "ScrollUp"},
		{historyView, gocui.KeyEnter, gocui.ModNone, t.composeReply, "composeReply"},
		{historyView, 'i', gocui.ModNone, t.composeReply, "composeReply"},
		{historyView, 'r', gocui.ModNone, t.composeReplyToRoot, "composeReplyToRoot"},
		{historyView, gocui.KeyHome, gocui.ModNone, t.ActionScrollTop, "ScrollTop"},
		{historyView, 'g', gocui.ModNone, t.ActionScrollTop, "ScrollTop"},
		{historyView, gocui.KeyEnd, gocui.ModNone, t.ActionScrollBottom, "ScrollBottom"},
		{historyView, 'G', gocui.ModNone, t.ActionScrollBottom, "ScrollBottom"},
		{historyView, 'q', gocui.ModNone, t.queryNeeded, "queryNeeded"},
		{editView, gocui.KeyEnter, gocui.ModAlt, t.Editor.ActionInsertNewline, "InsertNewline"},
		{editView, gocui.KeyTab, gocui.ModNone, t.Editor.ActionInsertTab, "InsertTab"},
		{editView, gocui.KeyEnter, gocui.ModNone, t.sendReply, "sendReply"},
		{editView, gocui.KeyCtrlBackslash, gocui.ModNone, t.cancelReply, "cancelReply"},
	}
}
