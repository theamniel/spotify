package grpc

import (
	"fmt"
	"log"

	"spotify/config"
	"spotify/protocols"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SpotifyClient = protocols.SpotifyClient

func Connect(cfg *config.Config) (protocols.SpotifyClient, error) {
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", cfg.Grpc.Host, cfg.Grpc.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	log.Printf("Connect to GRPC server at \"%s:%s\"", cfg.Grpc.Host, cfg.Grpc.Port)
	return protocols.NewSpotifyClient(conn), nil
}
