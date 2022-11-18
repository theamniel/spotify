package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/theamniel/spotify-server/spotify"
)

func GetNowPlaying(client *spotify.SpotifyClient) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		payload, err := client.GetNowPlaying()
		if err != nil {
			return ctx.Status(500).JSON(err.Error)
		}
		return ctx.Status(200).JSON(payload)
	}
}

func GetRecentlyPlayed(client *spotify.SpotifyClient) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		payload, err := client.GetRecentlyPlayed()
		if err != nil {
			return ctx.Status(500).JSON(err.Error)
		}
		return ctx.Status(200).JSON(payload)
	}
}
