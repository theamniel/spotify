package grpc

import (
	"fmt"
	"log"

	"spotify/config"
	"spotify/services/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Connect(cfg *config.Config) (proto.SpotifyClient, error) {
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", cfg.Grpc.Host, cfg.Grpc.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	log.Printf("Connect to GRPC server at \"%s:%s\"", cfg.Grpc.Host, cfg.Grpc.Port)
	return proto.NewSpotifyClient(conn), nil
}
