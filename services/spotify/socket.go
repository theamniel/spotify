package spotify

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/theamniel/spotify-server/services/config"
	"github.com/theamniel/spotify-server/services/socket"
)

func Socket(client *SpotifyClient, cfg *config.SocketConfig) fiber.Handler {
	client.Socket = socket.New()
	// TODO: better handler for this.
	// update accessToken every 55 minutes
	go client.UpdateAccessTokenAfter()
	// star poll data
	go client.poll()
	go client.Socket.Run()

	return websocket.New(client.Socket.Handle, websocket.Config{
		Origins:         cfg.Origins,
		ReadBufferSize:  cfg.ReadBufferSize,
		WriteBufferSize: cfg.WriteBufferSize,
	})
}

func (client *SpotifyClient) poll() {
	for {
		select {
		case <-time.Tick(client.pollRate * time.Second):
			spotifyStatus, err := client.GetSpotifyStatus()
			if err != nil {
				if client.pollRate < 5 {
					client.pollRate += 1
				} else {
					client.pollRate = 5
				}
				continue
			}

			if !client.Socket.HasState() {
				client.Socket.SetState(spotifyStatus)
				continue
			}

			if spotifyStatus.IsPlaying && client.pollRate > 1 {
				client.pollRate = 1
			}
			state := client.Socket.GetState().(*SocketData)

			// ------ TRACK CHANGE -----
			if spotifyStatus.ID != state.ID {
				client.Socket.Broadcast <- &socket.SocketMessage{
					socket.SocketDispatch,
					"TRACK_CHANGE",
					spotifyStatus,
				}
			}

			// -------- PLAYING STATE CHANGE -------
			if spotifyStatus.IsPlaying != state.IsPlaying {
				if !spotifyStatus.IsPlaying && len(spotifyStatus.PlayedAt) > 0 && spotifyStatus.PlayedAt != state.PlayedAt {
					client.Socket.Broadcast <- &socket.SocketMessage{
						socket.SocketDispatch,
						"TRACK_STATE",
						&socket.JSON{
							"is_playing": spotifyStatus.IsPlaying,
							"played_at":  spotifyStatus.PlayedAt,
						},
					}
				} else {
					client.Socket.Broadcast <- &socket.SocketMessage{
						socket.SocketDispatch,
						"TRACK_STATE",
						&socket.JSON{
							"is_playing": spotifyStatus.IsPlaying,
						},
					}
				}
			}
			client.Socket.SetState(spotifyStatus)
		}
	}
}
