package tui

import "github.com/whereswaldon/gocui"

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
		{historyView, gocui.KeyArrowDown, gocui.ModNone, t.cursorDown, "cursorDown"},
		{historyView, 'j', gocui.ModNone, t.cursorDown, "cursorDown"},
		{historyView, gocui.KeyArrowRight, gocui.ModNone, t.scrollDown, "scrollDown"},
		{historyView, 'l', gocui.ModNone, t.scrollDown, "scrollDown"},
		{historyView, gocui.KeyArrowUp, gocui.ModNone, t.cursorUp, "cursorUp"},
		{historyView, 'k', gocui.ModNone, t.cursorUp, "cursorUp"},
		{historyView, gocui.KeyArrowLeft, gocui.ModNone, t.scrollUp, "scrollUp"},
		{historyView, 'h', gocui.ModNone, t.scrollUp, "scrollUp"},
		{historyView, gocui.KeyEnter, gocui.ModNone, t.composeReply, "composeReply"},
		{historyView, 'i', gocui.ModNone, t.composeReply, "composeReply"},
		{historyView, 'r', gocui.ModNone, t.composeReply, "composeReply"},
		{historyView, 'n', gocui.ModNone, t.composeReplyToRoot, "composeReplyToRoot"},
		{historyView, gocui.KeyHome, gocui.ModNone, t.scrollTop, "scrollTop"},
		{historyView, 'g', gocui.ModNone, t.scrollTop, "scrollTop"},
		{historyView, gocui.KeyEnd, gocui.ModNone, t.scrollBottom, "scrollBottom"},
		{historyView, 'G', gocui.ModNone, t.scrollBottom, "scrollBottom"},
		{historyView, 'q', gocui.ModNone, t.queryNeeded, "queryNeeded"},
		{editView, gocui.KeyTab, gocui.ModNone, t.Editor.ActionInsertTab, "InsertTab"},
		{editView, gocui.KeyEnter, gocui.ModNone, t.handleEnter, "handleEnter"},
		{editView, gocui.KeyCtrlBackslash, gocui.ModNone, t.cancelReply, "cancelReply"},
		{editView, gocui.KeyCtrlP, gocui.ModNone, t.Editor.ActionTogglePasteMode, "TogglePasteMode"},
	}
}
