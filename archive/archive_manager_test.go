package archive_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/arborchat/muscadine/archive"
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
		return nil, nil
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
