package session_test

import (
	"testing"
	"time"

	"github.com/arborchat/muscadine/session"
	"github.com/onsi/gomega"
)

const (
	username    = "username"
	sessionName = "sessionname"
)

// TestCreateUserList ensures that CreateUserList returns a valid UserList
func TestCreateUserList(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	list := session.NewList()
	g.Expect(list).ToNot(gomega.BeNil())
}

// TestTrackSession ensures that adding a valid session succeeds
func TestTrackSession(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	list := session.NewList()
	err := list.Track(username, session.Session{ID: sessionName, LastSeen: time.Now()})
	g.Expect(err).To(gomega.BeNil())
	// adding the same session twice should not err
	err = list.Track(username, session.Session{ID: sessionName, LastSeen: time.Now()})
	g.Expect(err).To(gomega.BeNil())
}

// TestTrackBadSession ensures that adding an invalid session succeeds
func TestTrackBadSession(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	list := session.NewList()
	err := list.Track("", session.Session{sessionName, time.Now()})
	g.Expect(err).ToNot(gomega.BeNil())
	err = list.Track(username, session.Session{})
	g.Expect(err).ToNot(gomega.BeNil())
}

// TestRemoveSession ensures that removing a real session succeeds
func TestRemoveSession(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	list := session.NewList()
	err := list.Track(username, session.Session{ID: sessionName, LastSeen: time.Now()})
	if err != nil {
		t.Skip("Tracking failed", err)
	}
	err = list.Remove(username, sessionName)
	g.Expect(err).To(gomega.BeNil())

	// removing it again should fail
	err = list.Remove(username, sessionName)
	g.Expect(err).ToNot(gomega.BeNil())
}

// TestRemoveFakeSession ensures that removing a nonexistent session fails
func TestRemoveFakeSession(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	list := session.NewList()
	err := list.Remove(username, sessionName)
	g.Expect(err).ToNot(gomega.BeNil())
}

// TestRemoveInvalidSession ensures that invalid parameters cause an error
func TestRemoveInvalidSession(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	list := session.NewList()
	err := list.Remove(username, "")
	g.Expect(err).ToNot(gomega.BeNil())
	err = list.Remove("", sessionName)
	g.Expect(err).ToNot(gomega.BeNil())
}

// TestActiveSessions ensures that the ActiveSessions accessor returns a properly
// structured map of users and their most-recently-active sessions
func TestActiveSessions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	list := session.NewList()
	_ = list.Track(username, session.Session{sessionName, time.Now()})
	active := list.ActiveSessions()
	g.Expect(active).ToNot(gomega.BeNil())
	g.Expect(len(active)).To(gomega.BeEquivalentTo(1))
}

// TestActiveMultiSessions ensures that the ActiveSessions accessor returns a properly
// structured map of users and their most-recently-active sessions when there
// is more than one session for the same user
func TestActiveMultiSessions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	list := session.NewList()
	secondName := sessionName + "-second"
	_ = list.Track(username, session.Session{sessionName, time.Now().Track(-1 * time.Second)})
	_ = list.Track(username, session.Session{secondName, time.Now()})
	// multiple sessions for the same user should still result in a single result
	active := list.ActiveSessions()
	g.Expect(active).ToNot(gomega.BeNil())
	g.Expect(len(active)).To(gomega.BeEquivalentTo(1))
	// the more recently-added session should come out
	for username, session := range active {
		g.Expect(username).To(gomega.BeEquivalentTo(username))
		g.Expect(session.ID).To(gomega.BeEquivalentTo(secondName))
	}
}

// TestActiveMultiUserSessions ensures that the ActiveSessions accessor returns a properly
// structured map of users and their most-recently-active sessions when there
// are multiple users
func TestActiveMultiUserSessions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	list := session.NewList()
	secondUser := username + "-second"
	_ = list.Track(username, session.Session{sessionName, time.Now()})
	_ = list.Track(secondUser, session.Session{sessionName, time.Now()})
	// multiple sessions for the same user should still result in a single result
	active := list.ActiveSessions()
	g.Expect(active).ToNot(gomega.BeNil())
	g.Expect(len(active)).To(gomega.BeEquivalentTo(2))
	usernames := make([]string, 0, 2)
	for username := range active {
		usernames = append(usernames, username)
	}
	g.Expect(usernames).To(gomega.ContainElement(username))
	g.Expect(usernames).To(gomega.ContainElement(secondUser))
}
