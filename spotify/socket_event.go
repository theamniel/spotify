package spotify

type SocketEvent struct {
	E string      `json:"e"`
	D interface{} `json:"d,omitempty"`
}

type JSON map[string]interface{}
