package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2/utils"
)

type SocketClient struct {
	sync.RWMutex
	ID                string
	Conn              *websocket.Conn
	Send              chan *SocketMessage
	Message           chan *SocketMessage
	Done              chan struct{}
	IsConnectionAlive bool
}

func NewClient(conn *websocket.Conn) *SocketClient {
	return &SocketClient{
		ID:                utils.UUID(),
		Conn:              conn,
		Send:              make(chan *SocketMessage),
		Message:           make(chan *SocketMessage),
		Done:              make(chan struct{}),
		IsConnectionAlive: conn != nil,
	}
}

func (socket *SocketClient) HasConn() bool {
	socket.RLock()
	defer socket.RUnlock()
	return socket.Conn != nil
}

func (socket *SocketClient) IsAlive() bool {
	socket.RLock()
	defer socket.RUnlock()
	return socket.IsConnectionAlive
}

func (socket *SocketClient) SetAlive(alive bool) {
	socket.Lock()
	defer socket.Unlock()
	socket.IsConnectionAlive = alive
}

func (socket *SocketClient) Close(code int) {
	if socket.IsAlive() {
		close(socket.Done)
		socket.Conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, ""), time.Now().Add(time.Second))
	}
	socket.SetAlive(false)
}

func (socket *SocketClient) reader(ctx context.Context) {
	defer close(socket.Message)
	timer := time.NewTicker(10 * time.Millisecond)

	for {
		select {
		case <-timer.C:
			if !socket.HasConn() {
				return
			}

			socket.RLock()
			mt, message, err := socket.Conn.ReadMessage()
			socket.RUnlock()

			if mt == websocket.PingMessage || mt == websocket.PongMessage {
				if err != nil {
					return
				}
				continue
			}

			if mt == websocket.CloseMessage {
				socket.Close(1000)
				return
			}

			if err != nil { // TODO handle error
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("UnexpectedError: %v\n", err)
				}
				return
			}

			// We have a message and we fire the message event
			var event SocketMessage
			if err := json.Unmarshal(message, &event); err != nil {
				socket.Close(CloseInvalidMessage)
				return
			}
			socket.Message <- &event
		case <-ctx.Done():
			timer.Stop()
			return
		}
	}
}

func (socket *SocketClient) writer(ctx context.Context) {
	defer close(socket.Send)
	for {
		select {
		case event, ok := <-socket.Send:
			if ok {
				socket.RLock()
				err := socket.Conn.WriteJSON(event)
				socket.RUnlock()

				if err != nil {
					socket.Close(1011)
					return
				}
			} else {
				socket.Close(1011)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (socket *SocketClient) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	go socket.reader(ctx)
	go socket.writer(ctx)
	<-socket.Done
	cancel()
}
