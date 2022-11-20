package socket

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/websocket/v2"
)

type SocketClient struct {
	sync.RWMutex
	ID                string
	Conn              *websocket.Conn
	Send              chan *SocketMessage
	Message           chan *SocketMessage
	Done              chan struct{}
	Initialize        chan bool
	IsConnectionAlive bool
}

func New(conn *websocket.Conn) *SocketClient {
	return &SocketClient{
		ID:                utils.UUID(),
		Conn:              conn,
		Send:              make(chan *SocketMessage),
		Message:           make(chan *SocketMessage),
		Done:              make(chan struct{}),
		Initialize:        make(chan bool, 1),
		IsConnectionAlive: conn != nil,
	}
}

func (socket *SocketClient) IsAlive() bool {
	socket.RLock()
	defer socket.RUnlock()
	return socket.IsConnectionAlive
}

func (socket *SocketClient) SetAlive(alive bool) {
	socket.Lock()
	socket.IsConnectionAlive = alive
	socket.Unlock()
}

func (socket *SocketClient) Close() {
	if socket.IsAlive() {
		close(socket.Done)
		socket.Conn.WriteMessage(websocket.CloseMessage, []byte("Connection closed"))
	}
	socket.SetAlive(false)
	socket.Conn.Close()
}

func (socket *SocketClient) reader() {
	for {
		socket.RLock()
		mt, message, err := socket.Conn.ReadMessage()
		socket.RUnlock()
		if err != nil { // TODO handle error
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v\n", err)
			}
			break
		}

		if mt == websocket.TextMessage {
			var event SocketMessage
			if err := json.Unmarshal(message, &event); err != nil {
				fmt.Printf("error unmarshal: %v\n", err)
				socket.Close()
				return
			}

			if event.OP == SocketInitialize {
				socket.Initialize <- true
				continue
			} else if event.OP == SocketHeartBeat {
				continue
			}
		}

		if mt == websocket.CloseMessage {
			socket.Close()
			return
		}
		select {
		case <-socket.Done:
			return
		}
	}
}

func (socket *SocketClient) writer() {
	for {
		select {
		case event, ok := <-socket.Send:
			if ok {
				socket.RLock()
				err := socket.Conn.WriteJSON(event)
				socket.RUnlock()
				if err != nil {
					socket.Close()
					break
				}
			} else {
				socket.Close()
			}
		case <-socket.Done:
			return
		}
	}
}

func (socket *SocketClient) Run() {
	go socket.reader()
	go socket.writer()
	<-socket.Done
}
