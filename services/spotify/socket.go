package spotify

import (
	"time"

	"github.com/theamniel/spotify-server/services/websocket"
)

// TEST ONLY
func (c *Client) SocketConnection(w *websocket.Websocket) {
	for {
		payload := websocket.JSON{"isPlaying": true}
		w.Send <- &websocket.Event{"UPDATE_STATUS", payload}
		time.Sleep(time.Second * 1)
	}
}
