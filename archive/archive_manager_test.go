package archive_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/archive"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/onsi/gomega"
)

// TestNewManager ensures that the ArchiveManager constructor
// rejects invalid parameters and accepts valid ones.
func TestNewManager(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	path := "path"
	mgr, err := archive.NewManager(path)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(mgr).ToNot(gomega.BeNil())
	mgr, err = archive.NewManager("")
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(mgr).To(gomega.BeNil())
}

func mgrOrSkip(t *testing.T, path string) *archive.Manager {
	m, err := archive.NewManager(path)
	if err != nil {
		t.Skip(err)
	}
	return m
}

// TestSetOpener ensures that the SetOpener method accepts input
// calls the Opener when Populate() is invoked.
func TestSetOpener(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	mgr := mgrOrSkip(t, "path")
	called := make(chan struct{})
	err := mgr.SetOpener(func(path string) (io.ReadWriteCloser, error) {
		go func() {
			called <- struct{}{}
		}()
		return memoryArchiveOrSkip(t, "foo"), nil
	})
	g.Expect(err).To(gomega.BeNil())
	err = mgr.Populate()
	g.Expect(err).To(gomega.BeNil())
	g.Eventually(func() bool {
		select {
		case <-called:
			return true
		default:
			return false
		}
	}).Should(gomega.BeTrue())
}

// TestSetOpenerInvalid ensures that invalid input to SetOpener is rejected
func TestSetOpenerInvalid(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	mgr := mgrOrSkip(t, "path")
	err := mgr.SetOpener(nil)
	g.Expect(err).ToNot(gomega.BeNil())
}

// TestSetOpenerError ensures that openers that generate errors cause Populate
// to return an error.
func TestSetOpenerError(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	mgr := mgrOrSkip(t, "path")
	err := mgr.SetOpener(func(string) (io.ReadWriteCloser, error) {
		return nil, fmt.Errorf("always error")
	})
	g.Expect(err).To(gomega.BeNil())
	err = mgr.Populate()
	g.Expect(err).ToNot(gomega.BeNil())
}

// TestSetOpenerNil ensures that an opener function that returns two nil results
// simply makes Populate() error.
func TestSetOpenerNil(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	mgr := mgrOrSkip(t, "path")
	err := mgr.SetOpener(func(string) (io.ReadWriteCloser, error) {
		return nil, nil
	})
	g.Expect(err).To(gomega.BeNil())
	err = mgr.Populate()
	g.Expect(err).ToNot(gomega.BeNil())
}

// memoryArchiveOrSkip creates an in-memory file-like object to use as a test history
// file. It accepts the UUID of the message that it inserts as a parameter for testing
// whether the message is later present in an Archive
func memoryArchiveOrSkip(t *testing.T, id string) io.ReadWriteCloser {
	a := archive.New()
	err := a.Add(&arbor.ChatMessage{UUID: id, Parent: "bar", Username: "baz", Content: "bin", Timestamp: time.Now().Unix()})
	if err != nil {
		t.Skip(err)
	}
	buf := new(bytes.Buffer)
	err = a.Persist(buf)
	if err != nil {
		t.Skip(err)
	}
	return arbor.NoopRWCloser(buf)
}

// TestPopulate ensures that data is in the archive after Populating from a non-empty
// persistent storage.
func TestPopulate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	mgr := mgrOrSkip(t, "path")
	id := "foo"
	g.Expect(mgr.Has(id)).ToNot(gomega.BeTrue())
	err := mgr.SetOpener(func(string) (io.ReadWriteCloser, error) {
		return memoryArchiveOrSkip(t, id), nil
	})
	if err != nil {
		t.Skip(err)
	}
	err = mgr.Populate()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(mgr.Has(id)).To(gomega.BeTrue())
}

// TestOpenFile checks that the OpenFile function returns a valid file when given valid input.
func TestOpenFile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	id, err := uuid.NewV4()
	if err != nil {
		t.Skip(err)
	}
	idString := id.String()
	data := []byte(idString)
	err = ioutil.WriteFile(idString, data, 0600)
	if err != nil {
		t.Skip(err)
	}
	file, err := archive.OpenFile(idString)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(file).ToNot(gomega.BeNil())
	err = file.Close()
	g.Expect(err).To(gomega.BeNil())
	err = os.Remove(idString)
	g.Expect(err).To(gomega.BeNil())
}
