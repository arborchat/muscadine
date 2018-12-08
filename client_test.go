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

const testMsg = "{\"Type\":2,\"UUID\":\"92d24e9d-12cc-4742-6aaf-ea781a6b09ec\",\"Parent\":\"f4ae0b74-4025-4810-41d6-5148a513c580\",\"Content\":\"A riveting example message.\",\"Username\":\"Examplius_Caesar\",\"Timestamp\":1537738224}\n"

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
