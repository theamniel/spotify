package middlewares

import "github.com/gofiber/fiber/v2"

func Authenticated(secret string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if ctx.Get("Authorization") == "" {
			return ctx.Status(400).SendString("Bad request.")
		} else if ctx.Get("Authorization") != secret {
			return ctx.Status(401).SendString("Unauthorized.")
		}
		return ctx.Next()
	}
}
