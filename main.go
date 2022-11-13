package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/theamniel/spotify-server/routes"
)

func main() {
	app := fiber.New()

	routes.SetupWS(app)

	// ws://localhost:5050/ws
	log.Fatal(app.Listen(":5050"))
}
