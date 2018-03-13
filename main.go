package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// Websocket echo server
	r.GET("ws", func(c *gin.Context) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     checkOrigin,
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
			return
		}

		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			if err := conn.WriteMessage(messageType, p); err != nil {
				log.Println(err)
				return
			}
		}
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header["Origin"]
	if len(origin) == 0 {
		return true
	}
	u, err := url.Parse(origin[0])
	if err != nil {
		return false
	}
	// be permissive if running locally
	reqHost := strings.Split(r.Host, ":")[0]
	if reqHost == "localhost" && strings.Split(u.Host, ":")[0] == "localhost" {
		return true
	}

	// TODO: need to perform case folding?
	return u.Host == r.Host
}
