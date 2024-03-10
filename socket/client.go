package socket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2/utils"
)

type Client struct {
	ID                string
	Conn              *websocket.Conn
	Send              chan *Message
	Message           chan *Message
	Done              chan struct{}
	isConnectionAlive bool
	mu                sync.RWMutex
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		ID:                utils.UUID(),
		Conn:              conn,
		Send:              make(chan *Message),
		Message:           make(chan *Message),
		Done:              make(chan struct{}),
		isConnectionAlive: conn != nil,
	}
}

func (socket *Client) IsAlive() bool {
	socket.mu.RLock()
	defer socket.mu.RUnlock()
	return socket.isConnectionAlive
}

func (socket *Client) SetAlive(alive bool) {
	socket.mu.Lock()
	defer socket.mu.Unlock()
	socket.isConnectionAlive = alive
}

func (socket *Client) Close(code int) {
	if socket.IsAlive() {
		close(socket.Done)
		socket.Conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, ""), time.Now().Add(time.Second))
	}
	socket.SetAlive(false)
}

func (socket *Client) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	go reader(ctx, socket)
	go writer(ctx, socket)
	<-socket.Done
	cancel()
}

func reader(ctx context.Context, socket *Client) {
	defer close(socket.Message)
	timer := time.NewTicker(10 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			if !socket.IsAlive() {
				return
			}

			socket.mu.RLock()
			mt, message, err := socket.Conn.ReadMessage()
			socket.mu.RUnlock()

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("UnexpectedError: %v\n", err)
				}
				return
			}

			if mt == websocket.TextMessage { // We have a message and we fire the message event
				var event Message
				if err := json.Unmarshal(message, &event); err != nil {
					socket.Close(CloseInvalidMessage)
					return
				}
				socket.Message <- &event

			} else if mt == websocket.CloseMessage {
				socket.Close(websocket.CloseNormalClosure)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func writer(ctx context.Context, socket *Client) {
	defer close(socket.Send)

	for {
		select {
		case event, ok := <-socket.Send:
			if ok {
				socket.mu.RLock()
				err := socket.Conn.WriteMessage(websocket.TextMessage, event.ToBytes())
				socket.mu.RUnlock()

				if err == nil {
					continue
				}
			}
			socket.Close(websocket.CloseInternalServerErr)
			return
		case <-ctx.Done():
			return
		}
	}
}
