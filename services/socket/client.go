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

const (
	ReadTimeout = 10 * time.Millisecond

	RetrySendMessage = 20 * time.Millisecond
	MaxSendRetry     = 5
)

type Client struct {
	ID                string
	Conn              *websocket.Conn
	Message           chan *Message
	Done              chan struct{}
	isConnectionAlive bool
	mu                sync.RWMutex
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		ID:                utils.UUID(),
		Conn:              conn,
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

func (socket *Client) Close(code int, msg string) {
	if socket.IsAlive() {
		close(socket.Done)
		socket.Conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, msg), time.Now().Add(time.Second))
	}
	socket.SetAlive(false)
}

func (socket *Client) Send(event *Message) {
	if !socket.IsAlive() {
		if event.retries <= MaxSendRetry {
			go func() {
				time.Sleep(RetrySendMessage)
				event.retries += 1
				socket.Send(event)
			}()
		}
		return
	}

	socket.mu.RLock()
	err := socket.Conn.WriteMessage(websocket.TextMessage, event.ToBytes())
	socket.mu.RUnlock()

	if err != nil {
		socket.Close(websocket.CloseInternalServerErr, err.Error())
	}
}

func (socket *Client) Run() {
	ctx, cancel := context.WithCancel(context.Background())

	go reader(ctx, socket)
	// go writer(ctx, socket)

	<-socket.Done
	cancel()
}

func reader(ctx context.Context, socket *Client) {
	timer := time.NewTicker(ReadTimeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			if !socket.IsAlive() {
				continue
			}

			socket.mu.RLock()
			mt, message, err := socket.Conn.ReadMessage()
			socket.mu.RUnlock()

			if mt == websocket.PingMessage {
				// todo
				continue
			}

			if mt == websocket.PongMessage {
				// todo
				continue
			}

			if mt == websocket.CloseMessage {
				// todo
				socket.Close(websocket.CloseNormalClosure, "")
				continue
			}

			if err != nil {
				// todo
				socket.Close(websocket.CloseAbnormalClosure, "")
				continue
			}

			// We have a message and we fire the message event
			var event Message
			if err := json.Unmarshal(message, &event); err != nil {
				socket.Message <- Error(fmt.Sprintf("Invalid message body: %x", err))
				continue
			}
			socket.Message <- &event
		case <-ctx.Done():
			return
		}
	}
}
