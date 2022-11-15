package spotify

import "encoding/json"

type SocketEvent struct {
	T string `json:"t"`
	D JSON   `json:"d"`
}

type JSON map[string]interface{}

func NewEvent(e string, d JSON) *SocketEvent {
	return &SocketEvent{e, d}
}

func NewEventFromBytes(raw []byte) (*SocketEvent, error) {
	ev := SocketEvent{}
	if err := json.Unmarshal(raw, &ev); err != nil {
		return nil, err
	}
	return &ev, nil
}

func (e *SocketEvent) ToBytes() []byte {
	ev, err := json.Marshal(*e)
	if err != nil {
		return []byte{}
	}
	return ev
}
