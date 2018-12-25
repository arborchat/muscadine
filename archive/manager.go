package archive

import (
	"fmt"
	"io"
	"os"
)

// Manager facilitates populating an Archive from a persistent data store and
// persisting an archive to a persistent data store.
type Manager struct {
	*Archive
	path   string
	opener Opener
}

// Opener transforms a path into a readable, writable, closable file-like entity.
type Opener func(string) (io.ReadWriteCloser, error)

// OpenFile is an Opener that reads a file from disk.
func OpenFile(path string) (io.ReadWriteCloser, error) {
	return os.Open(path)
}

// NewManager creates a Manager that will use the provided path as persistent
// storage for its history.
func NewManager(path string) (*Manager, error) {
	if path == "" {
		return nil, fmt.Errorf("Path may not be the empty string")
	}
	return &Manager{
		Archive: New(),
		path:    path,
		opener:  OpenFile,
	}, nil
}

// SetOpener configures the Manager to open its path with the given Opener function
func (m *Manager) SetOpener(o Opener) error {
	if o == nil {
		return fmt.Errorf("Cannot set nil opener")
	}
	m.opener = o
	return nil
}

// Populate loads the managed archive with content from the manager's configured
// persistent storage.
func (m *Manager) Populate() error {
	_, err := m.opener(m.path)
	if err != nil {
		return err
	}
	return nil
}
