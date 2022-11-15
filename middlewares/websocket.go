package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func WebsocketAuth(origin string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if ctx.Get("Host") != origin {
			return ctx.Status(403).SendString("Request origin not allowed.")
		}
		if !websocket.IsWebSocketUpgrade(ctx) {
			return ctx.Status(426).SendString("Upgrade required.")
		}
		return ctx.Next()
	}
}
