package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/theamniel/spotify-server/routes"
	"github.com/theamniel/spotify-server/services/spotify"
	"github.com/theamniel/spotify-server/services/websocket"
)

func main() {
	app := fiber.New()

	spotifyService := spotify.New(spotify.Config{
		ClientID:     "",
		ClientSecret: "",
		RefreshToken: "",
	})

	wsService := websocket.New()

	routes.SetupWS(app, wsService)
	routes.SetupAPI(app, spotifyService)

	// ws://localhost:5050/ws
	log.Fatal(app.Listen(":5050"))
}
