package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/theamniel/spotify-server/services/spotify"
)

func GetNowPlaying(client *spotify.Client) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(200)
	}
}

func GetRecentlyPlayed(client *spotify.Client) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(200)
	}
}
