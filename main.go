package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/theamniel/spotify-server/controllers"
	"github.com/theamniel/spotify-server/middlewares"
	"github.com/theamniel/spotify-server/services/config"
	"github.com/theamniel/spotify-server/services/spotify"
)

func init() {
	log.SetFlags(log.Ltime)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Panic(err)
	}

	client := spotify.New(cfg.Spotify)

	app := fiber.New(fiber.Config{
		AppName:               "Spotify Server",
		DisableStartupMessage: true,
		StrictRouting:         cfg.App.StrictRouting,
		CaseSensitive:         cfg.App.CaseSensitive,
		UnescapePath:          cfg.App.UnescapePath,
		Prefork:               cfg.App.Prefork,
		BodyLimit:             cfg.App.Limit << 20,
	})

	/* --- MIDDLEWARES ---*/
	if cfg.Middleware.Recover {
		app.Use(recover.New())
	}

	if cfg.Middleware.Logger {
		app.Use(logger.New(logger.Config{
			TimeZone: "America/Caracas",
		}))
	}

	/* --- ROUTES --- */
	app.Get("/now-playing", controllers.GetNowPlaying(client))
	app.Get("/recently-played", controllers.GetRecentlyPlayed(client))
	app.Get("/socket", middlewares.WebsocketCheck(), spotify.Socket(client, cfg.Socket))
	/* 404 */
	app.Use(func(c *fiber.Ctx) error {
		return c.Redirect("https://github.com/TheAmniel", 308)
	})

	if !fiber.IsChild() {
		log.Printf("Running Socket server on \"%s:%s\"\n", cfg.Server.Host, cfg.Server.Port)
	}
	go func() {
		if err := app.Listen(fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)); err != nil {
			log.Fatal(err)
		}
	}()
	if !fiber.IsChild() {
		log.Println("Press CTRL-C to stop the application")
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	if !fiber.IsChild() {
		log.Println("Shutting down Socket server...")
	}

	if err := app.Shutdown(); err != nil {
		log.Println("There was an error while closing the Socket server")
		log.Printf("%T: %v\n", err, err)
	}
}
