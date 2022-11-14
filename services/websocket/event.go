package websocket

import "encoding/json"

type Event struct {
	T string `json:"t"`
	D JSON   `json:"d"`
}

type JSON map[string]interface{}

type EventHandler func(event *Event)

func NewEvent(e string, d JSON) *Event {
	return &Event{e, d}
}

func NewEventFromBytes(raw []byte) (*Event, error) {
	ev := Event{}
	if err := json.Unmarshal(raw, &ev); err != nil {
		return nil, err
	}
	return &ev, nil
}

func (e *Event) ToBytes() []byte {
	ev, err := json.Marshal(*e)
	if err != nil {
		return []byte{}
	}
	return ev
}
