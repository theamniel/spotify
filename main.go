package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/theamniel/spotify-server/routes"
	"github.com/theamniel/spotify-server/spotify"
)

func main() {
	app := fiber.New()

	client := spotify.New(spotify.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RefreshToken: os.Getenv("REFRESH_TOKEN"),
	})

	routes.Setup(app, client)

	// ws://localhost:5050/ws
	log.Fatal(app.Listen(":5050"))
}
