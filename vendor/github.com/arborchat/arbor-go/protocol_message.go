package arbor

import (
	"encoding/json"
	"fmt"
)

const (
	// WelcomeType should be used as the `Type` field of a WELCOME ProtocolMessage
	WelcomeType = 0
	// QueryType should be used as the `Type` field of a QUERY ProtocolMessage
	QueryType = 1
	// NewMessageType should be used as the `Type` field of a NEW_MESSAGE ProtocolMessage
	NewMessageType = 2
)

// ProtocolMessage represents a message in the Arbor chat protocol. This may or
// may not contain a chat message sent between users.
type ProtocolMessage struct {
	// Root is only used in WELCOME messages and identifies the root of this server's message tree
	Root string
	// Recent is only used in WELCOME messages and provides a list of recently-sent message ids
	Recent []string
	// The type of the message, should be one of the constants defined in this
	// package.
	Type uint8
	// Major is only used in WELCOME messages and identifies the major version number of the protocol version in use
	Major uint8
	// Minor is only used in WELCOME messages and identifies the minor version number of the protocol version in use
	Minor uint8
	// Message is the actual chat message content, if any. This is currently only
	// used in NEW_MESSAGE messages
	*ChatMessage
}

// MarshalJSON transforms a ProtocolMessage into JSON
func (m *ProtocolMessage) MarshalJSON() ([]byte, error) {
	switch m.Type {
	case WelcomeType:
		return json.Marshal(struct {
			Root   string
			Recent []string
			Type   uint8
			Major  uint8
			Minor  uint8
		}{Type: m.Type, Root: m.Root, Recent: m.Recent, Major: m.Major, Minor: m.Minor})
	case QueryType:
		return json.Marshal(struct {
			UUID string
			Type uint8
		}{UUID: m.UUID, Type: m.Type})
	case NewMessageType:
		return json.Marshal(struct {
			*ChatMessage
			Type uint8
		}{ChatMessage: m.ChatMessage, Type: m.Type})
	default:
		return nil, fmt.Errorf("Unknown message type, could not marshal")
	}
}

// String returns a JSON representation of the message as a string.
func (m *ProtocolMessage) String() string {
	data, _ := json.Marshal(m) // nolint: gosec
	dataString := string(data)
	return dataString
}
