package socket

import "github.com/goccy/go-json"

const (
	// Default Opcode when receiving core events [RECEIVE ONLY]
	SocketDispatch int = iota

	// Sends this when clients initially connect [RECEIVE ONLY]
	SocketHello

	// This is what the client sends when receiving opcode 1 [SEND ONLY]
	SocketInitialize

	// Clients should send Opcode 3 [SEND/RECEIVE]
	SocketHeartbeat

	// Sends this when clients sends heartbeat [RECEIVE ONLY]
	SocketHeartbeatACK
)

const (
	// [4001] Invalid/unknown Opcode
	CloseInvalidOpcode int = iota + 4001

	// [4002] Invalid message/payload
	CloseInvalidMessage

	// [4003] Not Authenticated
	CloseNotAuthenticated

	// [4004] Close by server request
	CloseByServerRequest

	// [4005] already authenticated
	CloseAlreadyAuthenticated
)

type JSON map[string]any

type Message struct {
	// Operation code
	OP int `json:"op"`

	// Event payload
	T string `json:"t,omitempty"`

	// Data payload
	D any `json:"d,omitempty"`
}

// Format Message to struct
func FormatMessage(op int, t string, d any) *Message {
	return &Message{op, t, d}
}

// Convert struct to []bytes
func (sm *Message) ToBytes() []byte {
	if bytes, err := json.Marshal(sm); err != nil {
		return nil
	} else {
		return bytes
	}
}
