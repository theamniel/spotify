package grpc

import (
	"fmt"
	"log"

	"spotify/protocols"

	"github.com/knadh/koanf/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SpotifyClient = protocols.SpotifyClient

func Connect(k *koanf.Koanf) (protocols.SpotifyClient, error) {
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", k.String("grpc.host"), k.Int("grpc.port")), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	log.Printf("Connect to GRPC server at \"%s\"", conn.CanonicalTarget())
	return protocols.NewSpotifyClient(conn), nil
}
