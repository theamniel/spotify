package controllers

import (
	"github.com/gofiber/fiber/v2"
	gws "github.com/gofiber/websocket/v2"
	"github.com/theamniel/spotify-server/services/websocket"
)

func SetupWS(ws *websocket.Websocket) fiber.Handler {
	cfg := gws.Config{
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
	}
	return gws.New(ws.Setup(), cfg)
}
