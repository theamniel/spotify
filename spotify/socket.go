package spotify

import (
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"spotify.amniel/config"
	"spotify.amniel/socket"
)

func Socket(client *SpotifyClient, cfg *config.SocketConfig) fiber.Handler {
	client.Socket = socket.New[Track]()
	// start socket "listeners"
	go client.Socket.Run()
	// start poll data
	go poll(client)

	return websocket.New(client.Socket.Handle, websocket.Config{
		Origins:         cfg.Origins,
		ReadBufferSize:  cfg.ReadBufferSize,
		WriteBufferSize: cfg.WriteBufferSize,
	})
}

func poll(client *SpotifyClient) {
	for {
		if client.IsConnected() {
			if !client.Socket.HasState() {
				if track, err := client.GetSpotifyStatus(); err != nil {
					client.onError()
					continue
				} else {
					if client.pollRate > DefaultPollRate {
						client.pollRate = DefaultPollRate
					}
					client.Socket.SetState(track)
				}
				continue
			} else if client.Socket.Listeners() > 0 {
				if track, err := client.GetSpotifyStatus(); err != nil {
					client.onError()
					continue
				} else {
					if track.IsPlaying && client.pollRate > DefaultPollRate {
						client.pollRate = DefaultPollRate
					}
					oldTrack := client.Socket.GetState()

					// ------ TRACK CHANGE -----
					if track.ID != oldTrack.ID {
						client.Socket.Broadcast <- socket.Dispatch("TRACK_CHANGE", track)
					}

					// ----- TRACK PROGRESS -----
					if track.ID == oldTrack.ID && track.IsPlaying {
						client.Socket.Broadcast <- socket.Dispatch("TRACK_PROGRESS", track.Timestamp.Progress)
					}

					// -------- TRACK STATE CHANGE -------
					if track.IsPlaying != oldTrack.IsPlaying {
						client.Socket.Broadcast <- socket.Dispatch("TRACK_STATE", &socket.JSON{"is_playing": track.IsPlaying})
					}
					client.Socket.SetState(track)
				}
			}
		}
		time.Sleep(client.pollRate * time.Second)
	}
}

func (client *SpotifyClient) onError() {
	if client.pollRate > (DefaultPollRate + 3) {
		client.pollRate = (DefaultPollRate + 3)
	} else {
		client.pollRate += 1
	}
}
