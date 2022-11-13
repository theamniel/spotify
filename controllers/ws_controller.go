package controllers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func SetupWS() fiber.Handler {
	return websocket.New(func(con *websocket.Conn) {
		fmt.Println(con.Locals("allowed"))
		for {
			mt, msg, err := con.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)
			err = con.WriteMessage(mt, msg)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	})
}
