package middlewares

import (
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

func WebsocketCheck() fiber.Handler {
	return func(c fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return c.SendStatus(fiber.StatusUpgradeRequired)
		}
		return c.Next()
	}
}
