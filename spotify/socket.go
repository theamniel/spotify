package spotify

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/theamniel/spotify-server/config"
	"github.com/theamniel/spotify-server/socket"
)

func Socket(client *SpotifyClient, cfg *config.SocketConfig) fiber.Handler {
	client.Socket = socket.New[SocketData]()
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
			if !client.Socket.HasState() {
				spotifyStatus, err := client.GetSpotifyStatus()
				if err != nil {
					if client.pollRate < 5 {
						client.pollRate += 1
					} else {
						client.pollRate = 5
					}
					continue
				}
				if client.pollRate > 1 {
					client.pollRate = 1
				}
				client.Socket.SetState(spotifyStatus)
				continue
			} else if client.Socket.Pool.Len() > 0 {
				spotifyStatus, err := client.GetSpotifyStatus()
				if err != nil {
					if client.pollRate < 5 {
						client.pollRate += 1
					} else {
						client.pollRate = 5
					}
					continue
				}

				if spotifyStatus.IsPlaying && client.pollRate > 1 {
					client.pollRate = 1
				}
				state := client.Socket.GetState()

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
}
