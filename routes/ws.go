package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/theamniel/spotify-server/controllers"
)

func SetupWS(app *fiber.App) {
	app.Use("/ws", func(ctx *fiber.Ctx) error {
		if ctx.Get("Host") != "localhost:3000" {
			return ctx.Status(403).SendString("Request origin not allowed")
		}
		if websocket.IsWebSocketUpgrade(ctx) {
			ctx.Locals("allowed", true)
			return ctx.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", controllers.SetupWS())
}
