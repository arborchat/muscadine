package session

import (
	"fmt"
	"sort"
	"time"
)

// Session represents known login sessions of other users. The Id
// field is unique per session, and LastSeen is the most recent
// time at which the session has been active.
type Session struct {
	ID       string
	LastSeen time.Time
}

// List represents the known active user sessions on an arbor server.
type List struct {
	Active map[string]map[string]time.Time
}

// NewList creates an empty list of sessions.
func NewList() *List {
	return &List{make(map[string]map[string]time.Time)}
}

// Track updates the List with the given session information for the given
// user. If the user has a session with the same ID already, the LastSeen
// time is updated to reflect the LastSeen time in sess.
func (l *List) Track(username string, sess Session) error {
	if username == "" || sess.ID == "" {
		return fmt.Errorf("Invalid username (%s) or session ID (%s)", username, sess.ID)
	}
	userSessions, present := l.Active[username]
	if present {
		// search the existing sessions for the user
		lastSeen, exists := userSessions[sess.ID]
		if exists {
			// update the timestamp if we found the same session
			if lastSeen.Before(sess.LastSeen) {
				userSessions[sess.ID] = sess.LastSeen
			}
			return nil
		}
		// insert the session if it wasn't present
		l.Active[username][sess.ID] = sess.LastSeen
		return nil
	}
	// user has no sessions, create this one for them
	userSessions = make(map[string]time.Time)
	userSessions[sess.ID] = sess.LastSeen
	l.Active[username] = userSessions

	return nil
}

// Remove takes the session with ID sessID out of the List for the user
// with username.
func (l *List) Remove(username, sessID string) error {
	if username == "" || sessID == "" {
		return fmt.Errorf("Invalid username (%s) or session ID (%s)", username, sessID)
	}
	userSessions, present := l.Active[username]
	if present {
		// search the existing sessions for the user
		_, exists := userSessions[sessID]
		if exists {
			// delete the session if we found it
			delete(userSessions, sessID)
			// TODO: delete the whole map for the user if this was their only session
			return nil
		}
		// we were asked to delete a nonexistent session
		return fmt.Errorf("Can't delete nonexistent session (%s) for user (%s)", sessID, username)
	}
	// user has no sessions
	return fmt.Errorf("Can't delete session (%s) for user (%s), user has no sessions", sessID, username)
}

// ActiveSessions returns a map from usernames to the most active session
// for each user.
func (l *List) ActiveSessions() map[string]*Session {
	activeMap := make(map[string]*Session)
	for user, sessionMap := range l.Active {
		// make a list of all sessions for the user
		sessions := make([]*Session, 0, len(sessionMap))
		for sessionID, lastSeen := range sessionMap {
			sessions = append(sessions, &Session{
				ID:       sessionID,
				LastSeen: lastSeen,
			})
		}
		// sort in descending order by time
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].LastSeen.After(sessions[j].LastSeen)
		})
		// chose the first session (the most recently updated)
		activeMap[user] = sessions[0]
	}
	return activeMap
}
