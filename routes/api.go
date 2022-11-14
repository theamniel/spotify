package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/theamniel/spotify-server/controllers"
	"github.com/theamniel/spotify-server/services/spotify"
)

func SetupAPI(app *fiber.App, client *spotify.Client) {
	app.Get("/now-playing", controllers.GetNowPlaying(client))
	app.Get("/recently-played", controllers.GetRecentlyPlayed(client))
}
