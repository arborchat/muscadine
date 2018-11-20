package arbor

import (
	"encoding/json"
	"io"
	"log"
)

// MakeMessageWriter wraps the io.Writer and returns a channel of
// ProtocolMessage pointers. Any ProtocolMessage sent over that channel will be
// written onto the io.Writer as JSON. This function handles all
// marshalling. If a message fails to marshal for any reason, or if a write error
// occurs, the returned channel will be closed and no further messages will be
// written to the io.Writer.
func MakeMessageWriter(conn io.Writer) chan<- *ProtocolMessage {
	input := make(chan *ProtocolMessage)
	go func() {
		defer close(input)
		encoder := json.NewEncoder(conn)
		for message := range input {
			err := encoder.Encode(message)
			if err != nil {
				if err == io.EOF {
					log.Println("Writer connection closed", err)
					return
				}
				log.Println("Error encoding message", err)
				return
			}
		}
	}()
	return input
}

// MakeMessageReader wraps the io.ReadCloser and returns a channel of
// ProtocolMessage pointers. Any JSON received over the io.ReadCloser will
// be unmarshalled into an ProtocolMessage struct and sent over the returned
// channel. If invalid JSON is received, the ReadCloser will close the io.ReadCloser
// and the returned channel.
func MakeMessageReader(conn io.ReadCloser) <-chan *ProtocolMessage {
	output := make(chan *ProtocolMessage)
	go func() {
		defer close(output)
		decoder := json.NewDecoder(conn)
		for {
			a := &ProtocolMessage{}
			err := decoder.Decode(a)
			if err != nil {
				if err == io.EOF {
					log.Println("Reader connection closed", err)
					return
				}
				// if we received unparsable JSON, just hang up.
				defer func() {
					if closeErr := conn.Close(); closeErr != nil {
						log.Println("Error closing connection:", closeErr)
					}
				}()

				log.Println("Error decoding json, hanging up:", err)
				return
			}
			output <- a
		}
	}()
	return output
}
