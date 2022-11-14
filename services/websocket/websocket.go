package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

const (
	closeGracePeriod = 30 * time.Second

	pingWait = 20 * time.Second
)

type Websocket struct {
	Send      chan *Event
	On        <-chan []byte
	writeLock sync.Mutex
}

func New() *Websocket {
	return &Websocket{
		Send: make(chan *Event),
		On:   make(<-chan []byte),
	}
}

func (w *Websocket) Setup() func(conn *websocket.Conn) {
	return func(conn *websocket.Conn) {
		go w.reader(conn)
		w.writer(conn)
	}
}

func (w *Websocket) reader(ws *websocket.Conn) {
	ws.SetPongHandler(func(string) error {
		fmt.Println("pong received")
		ws.SetReadDeadline(time.Now().Add(closeGracePeriod))
		return nil
	})

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Printf("Error while trying to read message: %v\n", err)
			break
		}

		ev, err := NewEventFromBytes(msg)
		if err != nil {
			fmt.Printf("Error while trying to convert message from binary to struct: %v\n", err)
			break
		}
		fmt.Println("Received from client: ", ev.T)
	}
}

func (w *Websocket) writer(ws *websocket.Conn) {
	ticker := time.NewTicker(pingWait)
	defer ticker.Stop()
	defer ws.Close()

	for {
		select {
		case ev, ok := <-w.Send:
			if !ok {
				ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w.writeLock.Lock()
			err := ws.WriteJSON(ev)
			w.writeLock.Unlock()
			if err != nil {
				fmt.Println("Error while trying to send msg: ", err)
				return
			}

		case <-ticker.C:
			ws.SetWriteDeadline(time.Now().Add(closeGracePeriod))
			fmt.Println("Send ping request...")
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
