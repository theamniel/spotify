package spotify

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/theamniel/spotify-server/services/config"
	"github.com/theamniel/spotify-server/services/socket"
)

type SpotifySocket struct {
	sync.RWMutex
	clients    map[*socket.SocketClient]bool
	broadcast  chan *socket.SocketMessage
	register   chan *socket.SocketClient
	unregister chan *socket.SocketClient

	pollRate      int
	spotifyStatus *SocketData
}

func Socket(client *SpotifyClient, cfg *config.SocketConfig) fiber.Handler {
	client.Socket = &SpotifySocket{
		clients:       make(map[*socket.SocketClient]bool),
		broadcast:     make(chan *socket.SocketMessage),
		register:      make(chan *socket.SocketClient),
		unregister:    make(chan *socket.SocketClient),
		spotifyStatus: nil,
		pollRate:      5,
	}
	// TODO: pause/resume
	// update accessToken every 55 minutes
	go client.UpdateAccessTokenAfter()
	// star poll data
	go client.Socket.poll(client)
	go client.Socket.start()

	return websocket.New(func(conn *websocket.Conn) {
		socketClient := socket.New(conn)

		client.Socket.register <- socketClient
		socketClient.Run()
		client.Socket.unregister <- socketClient
	}, websocket.Config{
		Origins:         cfg.Origins,
		ReadBufferSize:  cfg.ReadBufferSize,
		WriteBufferSize: cfg.WriteBufferSize,
	})
}

func (ss *SpotifySocket) start() {
	for {
		select {
		case message := <-ss.broadcast:
			for client, initialize := range ss.clients {
				if !initialize {
					continue
				}
				client.Send <- message
			}
		case client := <-ss.register:
			ss.Lock()
			ss.clients[client] = false
			ss.Unlock()
			go ss.watch(client)
		case client := <-ss.unregister:
			if _, ok := ss.clients[client]; ok {
				close(client.Send)
				delete(ss.clients, client)
			}
		}
	}
}

func (ss *SpotifySocket) watch(client *socket.SocketClient) {
	client.Send <- &socket.SocketMessage{socket.SocketHello, "", nil}
	<-client.Initialize
	client.Send <- &socket.SocketMessage{socket.SocketDispatch, "INIT_STATE", ss.spotifyStatus}
	ss.Lock()
	ss.clients[client] = true
	ss.Unlock()
}

func (ss *SpotifySocket) poll(client *SpotifyClient) {
	for {
		time.Sleep(time.Second * time.Duration(ss.pollRate))
		spotifyStatus, err := client.GetSpotifyStatus()
		if err != nil {
			fmt.Println(err.Message)
			if ss.pollRate < 5 {
				ss.pollRate = ss.pollRate + 1
			} else {
				ss.pollRate = 5
			}
			continue
		}
		if ss.spotifyStatus != nil {
			if spotifyStatus.IsPlaying && ss.pollRate > 1 {
				ss.pollRate = 1
			} else {
				ss.pollRate = 5
			}

			// Track change
			if spotifyStatus.ID != ss.spotifyStatus.ID {
				ss.broadcast <- &socket.SocketMessage{
					socket.SocketDispatch,
					"TRACK_CHANGE",
					spotifyStatus,
				}
			}

			// Playing state change
			if spotifyStatus.IsPlaying != ss.spotifyStatus.IsPlaying {
				if !spotifyStatus.IsPlaying && len(spotifyStatus.Timestamp) > 0 && spotifyStatus.Timestamp != ss.spotifyStatus.Timestamp {
					ss.broadcast <- &socket.SocketMessage{
						socket.SocketDispatch,
						"TRACK_STATE",
						&socket.JSON{
							"is_playing": spotifyStatus.IsPlaying,
							"played_at":  spotifyStatus.Timestamp,
						},
					}
				} else {
					ss.broadcast <- &socket.SocketMessage{
						socket.SocketDispatch,
						"TRACK_STATE",
						&socket.JSON{
							"is_playing": spotifyStatus.IsPlaying,
						},
					}
				}
			}
		}
		ss.Lock()
		ss.spotifyStatus = spotifyStatus
		ss.Unlock()
	}
}
