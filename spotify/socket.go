package spotify

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

const (
	closeGracePeriod = 30 * time.Second

	pingWait = 20 * time.Second
)

var ticker = time.NewTicker(pingWait)

type ClientSocket struct {
	send chan *SocketEvent
	on   <-chan []byte
	mu   sync.Mutex
}

func (c *Client) UpdateStatuLoop() {
	for {
		c.socket.send <- &SocketEvent{"UPDATE_STATUS", JSON{"isPlaying": true}}
		time.Sleep(1 * time.Second)
	}
}

func (c *Client) Socket() fiber.Handler {
	return websocket.New(
		func(ws *websocket.Conn) {
			go c.SocketReader(ws)
			c.SocketWriter(ws)
		},
		websocket.Config{
			ReadBufferSize:  2048,
			WriteBufferSize: 2048,
		})
}

func (c *Client) SocketReader(ws *websocket.Conn) {
	ws.SetPongHandler(func(string) error {
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

func (c *Client) CloseSocket(ws *websocket.Conn) {
	ticker.Stop()
	ws.Close()
}

func (c *Client) SocketWriter(ws *websocket.Conn) {
	defer c.CloseSocket(ws)
	for {
		select {
		case ev, ok := <-c.socket.send:
			if !ok {
				ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.socket.mu.Lock()
			err := ws.WriteJSON(ev)
			c.socket.mu.Unlock()
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
