package socket

const (
	// Default Opcode when receiving core events [RECEIVE ONLY]
	SocketDispatch = 0

	// Sends this when clients initially connect [RECEIVE ONLY]
	SocketHello = 1

	// This is what the client sends when receiving opcode 1 [SEND ONLY]
	SocketInitialize = 2

	// Clients should send Opcode 3 [SEND ONLY]
	SocketHeartBeat = 3
)

type SocketMessage struct {
	OP int         `json:"op"`
	T  string      `json:"t,omitempty"`
	D  interface{} `json:"d,omitempty"`
}

type JSON map[string]interface{}
