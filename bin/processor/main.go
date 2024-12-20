package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"spotify/protocols"
	"spotify/services/spotify"
)

func main() {
	k := koanf.New(".")
	if err := k.Load(file.Provider("config.toml"), toml.Parser()); err != nil {
		log.Fatal(err)
	}

	fx.New(
		fx.Supply(k),
		fx.Provide(
			spotify.New,
			ConfigureApp,
		),
		fx.Invoke(Server),
	).Run()
}

func Server(lc fx.Lifecycle, srv *grpc.Server, k *koanf.Koanf) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			list, err := net.Listen("tcp", fmt.Sprintf(":%d", k.Int("grpc.port")))
			if err != nil {
				return err
			}
			go func() {
				log.Printf("Running Grpc on \"%s\"\n", list.Addr().String())
				log.Println("Press CTRL-C to stop the application")
				if err := srv.Serve(list); err != nil {
					log.Fatal(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Shutting down Grpc...")
			srv.Stop()
			return nil
		},
	})
}

func ConfigureApp(client *spotify.SpotifyClient) *grpc.Server {
	srv := grpc.NewServer()
	protocols.RegisterSpotifyServer(srv, &server{spotify: client})
	reflection.Register(srv)
	return srv
}
