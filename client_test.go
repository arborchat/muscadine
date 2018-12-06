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

// TestNetClient checks that the basic operations of a NetClient function as expected.
func TestNetClient(t *testing.T) {
	address := "localhost:7777"
	history := archive.New()
	times := 0
	timesDisconnected := make(chan int)
	nc, err := NewNetClient(address, "username", history)
	if err != nil {
		t.Error("New Client with valid address should succeed construction", err)
	}
	nc.SetConnector(func(address string) (io.ReadWriteCloser, error) {
		return arbor.NoopRWCloser(new(bytes.Buffer)), nil
	})
	nc.OnDisconnect(func(client tui.Connection) {
		times++
		timesDisconnected <- times
	})
	nc.OnReceive(func(m *arbor.ChatMessage) {
	})
	err = nc.Connect()
	if err != nil {
		t.Error("Should have been able to connect", err)
	}
	err = nc.Disconnect()
	if err != nil {
		t.Error("Should have been able to disconnect", err)
	}
	err = nc.Connect()
	if err != nil {
		t.Error("Should have been able to connect", err)
	}
	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() int { return <-timesDisconnected }).Should(gomega.Equal(1))
}
