package controllers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/theamniel/spotify-server/services/spotify"
)

func GetNowPlaying(client *spotify.SpotifyClient) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		payload, err := client.GetNowPlaying()
		if err != nil {
			return ctx.Status(500).JSON(err)
		}

		if strings.Contains(ctx.OriginalURL(), "?open") && payload.Item != nil {
			return ctx.Redirect(payload.Item.ExternalUrls["spotify"])
		}
		return ctx.Status(200).JSON(payload)
	}
}

// TODO: handle limit
func GetRecentlyPlayed(client *spotify.SpotifyClient) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		payload, err := client.GetRecentlyPlayed()
		if err != nil {
			return ctx.Status(500).JSON(err)
		}

		if strings.Contains(ctx.OriginalURL(), "?open") {
			return ctx.Redirect(payload.Items[0].Track.ExternalUrls["spotify"])
		}
		return ctx.Status(200).JSON(payload)
	}
}
