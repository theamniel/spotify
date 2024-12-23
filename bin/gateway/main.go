package main

import (
	"context"
	"fmt"
	"log"

	"spotify/middlewares"
	"spotify/services/grpc"
	"spotify/services/spotify"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	log.SetFlags(log.Ltime)

	k := koanf.New(".")
	if err := k.Load(file.Provider("config.toml"), toml.Parser()); err != nil {
		log.Fatal(err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	fx.New(
		fx.Supply(k, logger.Sugar()),
		fx.Provide(
			grpc.Connect,
			spotify.New,
			ConfigureApp,
		),
		fx.Invoke(
			ConfigureMiddlewares,
			ConfigureRoutes,
			Server,
		),
	).Run()
}

func Server(lc fx.Lifecycle, app *fiber.App, k *koanf.Koanf, client *spotify.SpotifyClient) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			app.Hooks().OnListen(func(ld fiber.ListenData) error {
				if !fiber.IsChild() {
					log.Printf("Running Socket server on \"%s:%s\"\n", ld.Host, ld.Port)
					log.Println("Press CTRL-C to stop the application")
				}
				return nil
			})

			go func() {
				if err := app.Listen(fmt.Sprintf(":%d", k.Int("server.port"))); err != nil {
					log.Fatal(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if !fiber.IsChild() {
				log.Println("Shutting down Socket server...")
				if client.Socket != nil {
					client.Socket.Close()
				}
			}
			return app.Shutdown()
		},
	})
}

func ConfigureApp(k *koanf.Koanf) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:               "Spotify Server",
		DisableStartupMessage: true,
		StrictRouting:         false,
		CaseSensitive:         false,
		UnescapePath:          true,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		Prefork:               k.Bool("server.prefork"),
	})
}

func ConfigureMiddlewares(app *fiber.App, k *koanf.Koanf) {
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		TimeZone: k.String("server.timezone"),
	}))
}

func ConfigureRoutes(app *fiber.App, client *spotify.SpotifyClient, k *koanf.Koanf, grpc grpc.SpotifyClient) {
	app.Get("/now-playing", func(c *fiber.Ctx) error {
		raw, open, url := c.QueryBool("raw"), c.QueryBool("open"), ""
		payload, err := client.GetNowPlaying(raw)
		if err != nil {
			return c.Status(500).JSON(err)
		}

		if open {
			if raw {
				url = payload.(*spotify.CurrentlyPlaying).Item.ExternalURLs["spotify"]
			} else {
				url = payload.(*spotify.Track).URL
			}
			return c.Redirect(url, 308)
		}
		return c.Status(200).JSON(payload)
	})

	app.Get("/recently-played", func(c *fiber.Ctx) error {
		raw, open, limit, url := c.QueryBool("raw"), c.QueryBool("open"), c.QueryInt("limit"), ""
		payload, err := client.GetLastPlayed(raw, limit)
		if err != nil {
			return c.Status(500).JSON(err)
		}

		if open {
			if raw {
				url = payload.([]*spotify.RecentlyPlayedItem)[0].Track.ExternalURLs["spotify"]
			} else {
				url = payload.([]*spotify.Track)[0].URL
			}
			return c.Redirect(url, 308)
		}
		return c.Status(200).JSON(payload)
	})

	/* Websocket service */
	app.Get("/socket", middlewares.WebsocketCheck(), spotify.Socket(client, k, grpc))
	/* 404 */
	app.Use(func(c *fiber.Ctx) error {
		return c.Redirect("https://github.com/TheAmniel", 308)
	})

}
