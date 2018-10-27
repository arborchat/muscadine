package main

import (
	"testing"

	arbor "github.com/arborchat/arbor-go"
	"github.com/jordwest/mock-conn"
)

const testMsg = "{\"Type\":2,\"UUID\":\"92d24e9d-12cc-4742-6aaf-ea781a6b09ec\",\"Parent\":\"f4ae0b74-4025-4810-41d6-5148a513c580\",\"Content\":\"A riveting example message.\",\"Username\":\"Examplius_Caesar\",\"Timestamp\":1537738224}\n"

// TestClient ensures that a client can be constructed from an
// io.ReadWriter and that it invokes its ReceiveHandler when it
// reads data from the provided io.ReadWriter.
func TestClient(t *testing.T) {
	conn := mock_conn.NewConn()
	defer conn.Close()
	c, err := Connect(conn.Client)
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
