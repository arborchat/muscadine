package main

import (
	"bytes"
	"io"
	"testing"

	arbor "github.com/arborchat/arbor-go"
	"github.com/arborchat/muscadine/archive"
	"github.com/jordwest/mock-conn"
	"github.com/onsi/gomega"
)

const testMsg = "{\"Type\":2,\"UUID\":\"92d24e9d-12cc-4742-6aaf-ea781a6b09ec\",\"Parent\":\"f4ae0b74-4025-4810-41d6-5148a513c580\",\"Content\":\"A riveting example message.\",\"Username\":\"Examplius_Caesar\",\"Timestamp\":1537738224}\n"

// TestClient ensures that a client can be constructed from an
// io.ReadWriter and that it invokes its ReceiveHandler when it
// reads data from the provided io.ReadWriter.
func TestClient(t *testing.T) {
	conn := mock_conn.NewConn()
	arch := archive.New()
	defer conn.Close()
	c, err := Connect(conn.Client, arch)
	if err != nil {
		t.Error("Connect errored with a valid io.ReadWriter", err)
	}
	received := make(chan struct{})
	c.RecieveHandler(func(msg *arbor.ChatMessage) {
		received <- struct{}{}
	})
	_, err = conn.Server.Write([]byte(testMsg))
	if err != nil {
		t.Skip("Unable to write to mock connection", err)
	}
	select {
	case <-received:
		// pass
		/*	case <-time.NewTicker(time.Second).C:
			t.Error("Never invoked receive handler")*/
	}
}

// TestNetClient checks that the basic operations of a NetClient function as expected.
func TestNetClient(t *testing.T) {
	address := "localhost:7777"
	history := archive.New()
	times := 0
	timesDisconnected := make(chan int)
	nc, err := NewNetClient(address, history)
	if err != nil {
		t.Error("New Client with valid address should succeed construction", err)
	}
	nc.SetConnector(func(address string) (io.ReadWriteCloser, error) {
		return arbor.NoopRWCloser(new(bytes.Buffer)), nil
	})
	nc.OnDisconnect(func(client *NetClient) {
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
