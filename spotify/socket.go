package spotify

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

const (
	pollRate             = 1000
	hasFinishedThreshold = 2000
	hasScrubbedThreshold = 1500

	closeGracePeriod = 30 * time.Second

	pingWait = 20 * time.Second
)

type SocketClient struct {
	spotify *SpotifyClient
	conn    *websocket.Conn
	send    chan *SocketEvent
	message chan *SocketEvent
	done    chan struct{}
	mu      sync.Mutex

	playerState         *JSON
	pollRate            int
	hasSentInitialState bool
	hasNotifiedTrackEnd bool
	poll                func()
}

func defaultSocketClient(c *SpotifyClient, conn *websocket.Conn) *SocketClient {
	return &SocketClient{
		spotify:             c,
		conn:                conn,
		send:                make(chan *SocketEvent),
		message:             make(chan *SocketEvent),
		done:                make(chan struct{}),
		playerState:         nil,
		pollRate:            pollRate,
		hasSentInitialState: false,
		hasNotifiedTrackEnd: false,
		poll:                func() {},
	}
}

func Socket(client *SpotifyClient) fiber.Handler {
	return websocket.New(func(ws *websocket.Conn) {
		socket := defaultSocketClient(client, ws)
		socket.poll = func() {
			for {
				if !socket.hasSentInitialState {
					socket.send <- &SocketEvent{"INITIAL_STATE", JSON{}}
					socket.hasSentInitialState = true
				}
				time.Sleep(time.Millisecond * time.Duration(socket.pollRate))
				socket.UpdateStatus()
			}
		}

		go socket.poll()
		go socket.reader()
		socket.writer()
	}, websocket.Config{
		Origins:         []string{"*"},
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
	})
}

func (c *SocketClient) reader() {
	defer close(c.done)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(closeGracePeriod))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v\n", err)
			}
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

func (c *SocketClient) writer() {
	ticker := time.NewTicker(pingWait)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case <-c.done:
			return
		case ev, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.mu.Lock()
			err := c.conn.WriteJSON(ev)
			c.mu.Unlock()
			if err != nil {
				fmt.Println("Error while trying to send msg: ", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(closeGracePeriod))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *SocketClient) UpdateStatus() {
	// track, err := c.GetNowPlaying()
	// if err != nil {
	// 	fmt.Println(err)
	// 	c.send <- &SocketEvent{"UPDATE_STATUS", JSON{"isPlaying": false}}
	// 	return
	// }
	c.send <- &SocketEvent{"UPDATE_STATUS", JSON{"isPlaying": false}}
	// c.send <- &SocketEvent{"UPDATE_STATUS", track}
}
