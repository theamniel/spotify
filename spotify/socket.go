package spotify

import (
	// "errors"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

const (
	pollRate = 1

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

	spotifyState        *SpotifyStatus
	pollRate            int
	hasSentInitialState bool
	hasNotifiedTrackEnd bool
	lastSentError       string
	poll                func()
}

func defaultSocketClient(c *SpotifyClient, conn *websocket.Conn) *SocketClient {
	return &SocketClient{
		spotify:             c,
		conn:                conn,
		send:                make(chan *SocketEvent),
		message:             make(chan *SocketEvent),
		done:                make(chan struct{}),
		spotifyState:        nil,
		pollRate:            pollRate,
		hasSentInitialState: false,
		hasNotifiedTrackEnd: false,
		lastSentError:       "nil",
		poll:                func() {},
	}
}

func Socket(client *SpotifyClient) fiber.Handler {
	go client.UpdateAccessTokenAfter()
	return websocket.New(func(ws *websocket.Conn) {
		socket := defaultSocketClient(client, ws)
		socket.poll = func() {
			for {
				time.Sleep(time.Second * time.Duration(socket.pollRate))
				spotifyState, err := client.GetSpotifyStatus()
				if err != nil {
					socket.handleError(err)
					continue
				}

				if !socket.hasSentInitialState {
					socket.send <- &SocketEvent{"INITIAL_STATE", spotifyState}
					socket.spotifyState = spotifyState
					socket.hasSentInitialState = true
					continue
				}
				// reset poll rate if no errors were encountered
				socket.pollRate = pollRate

				// Track change
				if spotifyState.ID != socket.spotifyState.ID {
					socket.send <- &SocketEvent{"TRACK_CHANGE", spotifyState}
					socket.hasNotifiedTrackEnd = false
				}

				// Playing state change
				if spotifyState.IsPlaying != socket.spotifyState.IsPlaying {
					socket.send <- &SocketEvent{"TRACK_STATE", JSON{"is_playing": spotifyState.IsPlaying}}
				}
				socket.spotifyState = spotifyState
			}
		}

		socket.run()
	}, websocket.Config{
		Origins:         []string{"*"},
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
	})
}

func (socket *SocketClient) run() {
	defer func() {
		close(socket.done)
		socket.conn.Close()
	}()
	go socket.poll()
	go socket.reader()
	socket.writer()
}

func (c *SocketClient) reader() {
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(closeGracePeriod))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v\n", err)
			}
			break
		}
	}
}

func (c *SocketClient) writer() {
	ticker := time.NewTicker(pingWait)
	defer ticker.Stop()
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
				fmt.Println(err)
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

func (socket *SocketClient) handleError(err *ErrorResponse) {
	receivedErr := err.Error.Message
	if receivedErr != socket.lastSentError {
		socket.lastSentError = receivedErr
		socket.send <- &SocketEvent{"ERROR", JSON{"message": receivedErr}}
	} else {
		if socket.pollRate < 5 {
			socket.pollRate = socket.pollRate + 1
		} else {
			socket.pollRate = 5
		}
	}
}
