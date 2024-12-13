package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"spotify/config"
	"spotify/services/grpc/proto"
	"spotify/services/spotify"
)

func main() {
	log.SetFlags(log.Ltime)

	fx.New(
		fx.Provide(
			config.Load[config.Config],
			spotify.New,
			ConfigureApp,
		),
		fx.Invoke(Server),
		fx.NopLogger,
	).Run()
}

func Server(lc fx.Lifecycle, srv *grpc.Server, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			list, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Grpc.Host, cfg.Grpc.Port))
			if err != nil {
				return err
			}
			go func() {
				log.Printf("Running Grpc on \"%s:%s\"\n", cfg.Grpc.Host, cfg.Grpc.Port)
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
	proto.RegisterSpotifyServer(srv, &server{spotify: client})
	reflection.Register(srv)
	return srv
}
