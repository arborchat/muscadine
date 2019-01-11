package session

import "time"

// Session represents known login sessions of other users. The Id
// field is unique per session, and LastSeen is the most recent
// time at which the session has been active.
type Session struct {
	ID       string
	LastSeen time.Time
}

// List represents the known active user sessions on an arbor server.
type List struct {
	Active map[string][]*Session
}

// NewList creates an empty list of sessions.
func NewList() *List {
	return &List{make(map[string][]*Session)}
}

// Add updates the List with the given session information for the given
// user. If the user has a session with the same ID already, the LastSeen
// time is updated to reflect the LastSeen time in sess.
func (l *List) Add(username string, sess Session) error {
	return nil
}

// Remove takes the session with ID sessID out of the List for the user
// with username.
func (l *List) Remove(username, sessID string) error {
	return nil
}

// ActiveSessions returns a map from usernames to the most active session
// for each user.
func (l *List) ActiveSessions() map[string]*Session {
	return nil
}
