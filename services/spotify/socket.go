package spotify

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"spotify/config"
	"spotify/protocols"
	"spotify/services/socket"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func Socket(client *SpotifyClient, cfg *config.SocketConfig, grpc protocols.SpotifyClient) fiber.Handler {
	client.Socket = socket.New[Track]()
	// start poll data
	track, err := grpc.GetTrack(context.Background(), &protocols.Request{ID: fmt.Sprintf("%d", os.Getpid())})
	if err != nil {
		log.Fatal(err)
	}
	client.Socket.SetState(FromProtoToTrack(track))

	go poll(client, grpc)
	return websocket.New(client.Socket.Handle, websocket.Config{
		Origins:         cfg.Origins,
		ReadBufferSize:  cfg.ReadBufferSize,
		WriteBufferSize: cfg.WriteBufferSize,
	})
}

func poll(client *SpotifyClient, grpc protocols.SpotifyClient) {
	stream, err := grpc.OnListen(context.Background(), &protocols.Request{ID: fmt.Sprintf("%d", os.Getpid())})
	if err != nil {
		return
	}
	for {
		res, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("error while reading stream: %v", err)
		}

		if res.E == "CHANGE" {
			client.Socket.Broadcast(socket.Dispatch("TRACK_CHANGE", FromProtoToTrack(res.Track)))
		}

		if res.E == "PROGRESS" {
			client.Socket.Broadcast(socket.Dispatch("TRACK_PROGRESS", res.Progress))
		}

		if res.Track != nil {
			oldTrack := client.Socket.GetState()
			newTrack := FromProtoToTrack(res.Track)
			if oldTrack.ID != newTrack.ID {
				client.Socket.SetState(newTrack)
			}
		}
	}
}

func (client *SpotifyClient) OnError() {
	if client.PollRate > (DefaultPollRate + 3) {
		client.PollRate = (DefaultPollRate + 3)
	} else {
		client.PollRate += 1
	}
}
