package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/theamniel/spotify-server/controllers"
	"github.com/theamniel/spotify-server/middlewares"
	"github.com/theamniel/spotify-server/spotify"
)

func Setup(app *fiber.App, client *spotify.SpotifyClient) {
	app.Get("/now-playing", controllers.GetNowPlaying(client))
	app.Get("/recently-played", controllers.GetRecentlyPlayed(client))
	app.Get("/ws", middlewares.WebsocketCheck(), spotify.Socket(client))
}
