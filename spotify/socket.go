package spotify

import (
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/theamniel/spotify-server/config"
	"github.com/theamniel/spotify-server/socket"
)

func Socket(client *SpotifyClient, cfg *config.SocketConfig) fiber.Handler {
	client.Socket = socket.New[SocketData]()
	// start socket "listeners"
	go client.Socket.Run()
	// start poll data
	go client.poll()

	return websocket.New(client.Socket.Handle, websocket.Config{
		Origins:         cfg.Origins,
		ReadBufferSize:  cfg.ReadBufferSize,
		WriteBufferSize: cfg.WriteBufferSize,
	})
}

func (client *SpotifyClient) poll() {
	for {
		if client.IsConnected() {
			if !client.Socket.HasState() {
				if spotifyStatus, err := client.GetSpotifyStatus(); err != nil {
					client.onError()
					continue
				} else {
					if client.pollRate > 5 {
						client.pollRate = 5
					}
					client.Socket.SetState(spotifyStatus)
				}
				continue
			} else if client.Socket.Listeners() > 0 {
				if spotifyStatus, err := client.GetSpotifyStatus(); err != nil {
					client.onError()
					continue
				} else {
					if spotifyStatus.IsPlaying && client.pollRate > 5 {
						client.pollRate = 5
					}
					state := client.Socket.GetState()

					// ------ TRACK CHANGE -----
					if spotifyStatus.ID != state.ID {
						client.Socket.Broadcast <- socket.FormatMessage(socket.SocketDispatch, "TRACK_CHANGE", spotifyStatus)
					}

					// ----- TRACK PROGRESS -----
					if spotifyStatus.ID == state.ID && spotifyStatus.IsPlaying {
						client.Socket.Broadcast <- socket.FormatMessage(socket.SocketDispatch, "TRACK_PROGRESS", spotifyStatus.Timestamp.Progress)
					}

					// -------- TRACK STATE CHANGE -------
					if spotifyStatus.IsPlaying != state.IsPlaying {
						client.Socket.Broadcast <- socket.FormatMessage(socket.SocketDispatch, "TRACK_STATE", &socket.JSON{"is_playing": spotifyStatus.IsPlaying})
					}
					client.Socket.SetState(spotifyStatus)
				}
			}
		}
		time.Sleep(client.pollRate * time.Second)
	}
}

func (client *SpotifyClient) onError() {
	if client.pollRate < 8 {
		client.pollRate += 1
	} else {
		client.pollRate = 8
	}
}
