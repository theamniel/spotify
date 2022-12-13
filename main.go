package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/theamniel/spotify-server/config"
	"github.com/theamniel/spotify-server/middlewares"
	"github.com/theamniel/spotify-server/spotify"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Panic(err)
	}

	loc, locErr := time.LoadLocation(cfg.Server.TimeZone)
	if locErr != nil {
		loc = time.Local
	}
	log.SetFlags(0)
	log.SetPrefix("[" + time.Now().In(loc).Format("15:04:05") + "] ")

	client := spotify.New(cfg.Spotify)

	app := fiber.New(fiber.Config{
		AppName:               "Spotify Server",
		DisableStartupMessage: true,
		StrictRouting:         false,
		CaseSensitive:         false,
		UnescapePath:          true,
		Prefork:               cfg.Server.Prefork,
		BodyLimit:             5 << 20,
	})

	/* --- MIDDLEWARES ---*/
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		TimeZone: loc.String(),
	}))

	/* --- ROUTES --- */
	app.Get("/now-playing", func(c *fiber.Ctx) error {
		payload, err := client.GetNowPlaying()
		if err != nil {
			return c.Status(500).JSON(err)
		}
		if strings.Contains(c.OriginalURL(), "?open") && payload.Item != nil {
			return c.Redirect(payload.Item.ExternalUrls["spotify"], 308)
		}
		return c.Status(200).JSON(payload)
	})
	app.Get("/recently-played", func(c *fiber.Ctx) error {
		payload, err := client.GetRecentlyPlayed()
		if err != nil {
			return c.Status(500).JSON(err)
		}

		if strings.Contains(c.OriginalURL(), "?open") {
			return c.Redirect(payload.Items[0].Track.ExternalUrls["spotify"], 308)
		}
		return c.Status(200).JSON(payload)
	})
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
