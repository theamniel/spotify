package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// _ "github.com/BurntSushi/toml"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/theamniel/spotify-server/controllers"
	"github.com/theamniel/spotify-server/middlewares"
	"github.com/theamniel/spotify-server/spotify"
)

func init() {
	log.SetFlags(log.Ltime)
}

func main() {
	client := spotify.New(spotify.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RefreshToken: os.Getenv("REFRESH_TOKEN"),
	})

	app := fiber.New(fiber.Config{
		AppName:               "Spotify Websocket Server",
		DisableStartupMessage: true,
		StrictRouting:         true,
		CaseSensitive:         true,
		UnescapePath:          true,
	})

	app.Use(logger.New(logger.Config{
		TimeZone: "America/Caracas",
	}))

	app.Use(recover.New())

	app.Use(cache.New(cache.Config{
		CacheHeader:  "X-Cache-Status",
		Expiration:   7 * 24 * time.Hour,
		CacheControl: true,
		Next: func(c *fiber.Ctx) bool {
			return strings.Contains(c.Route().Path, "/socket")
		},
	}))

	/* --- ROUTES --- */
	app.Get("/now-playing", controllers.GetNowPlaying(client))
	app.Get("/recently-played", controllers.GetRecentlyPlayed(client))
	app.Get("/socket", middlewares.WebsocketCheck(), spotify.Socket(client))
	/* 404 */
	app.Use(func(c *fiber.Ctx) error {
		return c.Redirect("https://github.com/TheAmniel", 308)
	})

	if !fiber.IsChild() {
		log.Println("Running Socket server on \":5050\"")
	}
	go func() {
		// ws://localhost:5050/socket
		if err := app.Listen(fmt.Sprintf("%s:%s", "", "5050")); err != nil {
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
