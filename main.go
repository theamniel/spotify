package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"spotify.amniel/config"
	"spotify.amniel/middlewares"
	"spotify.amniel/spotify"
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
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		Prefork:               cfg.Server.Prefork,
	})

	/* --- MIDDLEWARES ---*/
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		TimeZone: loc.String(),
	}))

	/* --- ROUTES --- */
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
		raw, open, url := c.QueryBool("raw"), c.QueryBool("open"), ""
		payload, err := client.GetLastPlayed(raw)
		if err != nil {
			return c.Status(500).JSON(err)
		}

		if raw {
			payload = payload.([]spotify.RecentlyPlayedItem)[0]
			if open {
				url = payload.(spotify.RecentlyPlayedItem).Track.ExternalURLs["spotify"]
			}
		} else if open {
			url = payload.(*spotify.Track).URL
		}

		if open {
			return c.Redirect(url, 308)
		}
		return c.Status(200).JSON(payload)
	})

	/* Websocket service */
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
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if !fiber.IsChild() {
		log.Println("Shutting down Socket server...")
	}

	if err := app.Shutdown(); err != nil {
		log.Println("There was an error while closing the Socket server")
		log.Printf("%T: %v\n", err, err)
	}
}
