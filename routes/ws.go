package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/theamniel/spotify-server/controllers"
	"github.com/theamniel/spotify-server/middlewares"
	"github.com/theamniel/spotify-server/services/websocket"
)

func SetupWS(app *fiber.App, ws *websocket.Websocket) {
	app.Use("/ws", middlewares.WebsocketAuth("localhost:3000"))
	app.Get("/ws", controllers.SetupWS(ws))
}
