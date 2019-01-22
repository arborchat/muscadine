package archive

import (
	"fmt"
	"io"
	"os"
	"path"
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
func OpenFile(histPath string) (io.ReadWriteCloser, error) {
	const perms = 0700
	file, err := os.OpenFile(histPath, os.O_RDWR, perms)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path.Dir(histPath), perms); err != nil {
				return nil, err
			}
			return os.Create(histPath)
		}
		return nil, err
	}
	return file, err
}

// NewManager creates a Manager that will use the provided path as persistent
// storage for its history. It defaults to the OpenFile implementation of Opener,
// and you only need to call SetOpener for testing.
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

// Load loads the managed archive with content from the manager's configured
// persistent storage.
func (m *Manager) Load() error {
	file, err := m.opener(m.path)
	if err != nil {
		return err
	}
	if file == nil {
		return fmt.Errorf("Opener returned no error but nil file")
	}
	defer file.Close()
	return m.Archive.Populate(file)
}

// Save stores the managed archive's state into the configured persistent storage.
func (m *Manager) Save() error {
	file, err := m.opener(m.path)
	if err != nil {
		return err
	}
	if file == nil {
		return fmt.Errorf("Opener returned no error but nil file")
	}
	defer file.Close()
	return m.Archive.Persist(file)
}
