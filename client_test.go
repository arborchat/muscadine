package main

import (
	"bytes"
	"io"
	"testing"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/archive"
	"github.com/arborchat/muscadine/tui"
	"github.com/onsi/gomega"
)

func bufConnector(address string) (io.ReadWriteCloser, error) {
	return arbor.NoopRWCloser(new(bytes.Buffer)), nil
}

// TestNetClient checks that the basic operations of a NetClient function as expected.
func TestNetClient(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	address := "localhost:7777"
	history := archive.New()
	times := 0
	timesDisconnected := make(chan int)
	nc, err := NewNetClient(address, "username", history)
	g.Expect(err).To(gomega.BeNil())

	nc.SetConnector(bufConnector)
	nc.OnDisconnect(func(client tui.Connection) {
		times++
		timesDisconnected <- times
	})
	nc.OnReceive(func(m *arbor.ChatMessage) {
	})

	err = nc.Connect()
	g.Expect(err).To(gomega.BeNil())

	err = nc.Disconnect()
	g.Expect(err).To(gomega.BeNil())

	err = nc.Connect()
	g.Expect(err).To(gomega.BeNil())

	g.Eventually(func() int { return <-timesDisconnected }).Should(gomega.Equal(1))
}
