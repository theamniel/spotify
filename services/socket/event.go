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

	// Sent to the client when an error occurs [RECEIVE ONLY]
	SocketError
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

	// retries for send a message
	retries int `json:"-"`
}

// Dispatch event to struct
func Dispatch(event string, data any) *Message {
	return &Message{OP: SocketDispatch, T: event, D: data, retries: 0}
}

// Hello event to struct
func Hello(d any) *Message {
	return &Message{OP: SocketHello, T: "", D: d, retries: 0}
}

// Initialize event to struct
func Initialize(t string, d any) *Message {
	return &Message{OP: SocketInitialize, T: t, D: d, retries: 0}
}

// HeartbeatACK event to struct
func HeartbeatACK() *Message {
	return &Message{OP: SocketHeartbeatACK, T: "", D: nil, retries: 0}
}

// Heartbeat event to struct
func Heartbeat() *Message {
	return &Message{OP: SocketHeartbeat, T: "", D: nil, retries: 0}
}

// Error event to struct
func Error(msg any) *Message {
	return &Message{OP: SocketError, T: "", D: msg, retries: 0}
}

// Convert struct to []bytes
func (sm *Message) ToBytes() []byte {
	if bytes, err := json.Marshal(sm); err != nil {
		return nil
	} else {
		return bytes
	}
}
