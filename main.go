package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func setupRouter(s *GameState, jwtSecret []byte) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// TODO eventually make cors more permissive
	r.Use(cors.Default())

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// Login
	r.POST("/login", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")

		if s.CheckPassword(email, password) != nil {
			c.JSON(403, struct{ Error string }{"Invalid email/password"})
			return
		}
		// TODO this is an abuse of jwts - implement a better system with refresh
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iss":   "muhmud",
			"exp":   time.Now().Add(time.Hour * 24),
			"nbf":   time.Now(),
			"email": email,
		})

		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			log.Println("Got an error when signing token", err)
			c.JSON(500, nil)
			return
		}
		c.JSON(200, struct{ Token string }{tokenString})
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

		// TODO eventually factor out jwt auth into middleware
		tokenString := c.Query("token")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return jwtSecret, nil
		})
		if err != nil {
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4000, "Invalid token"))
			conn.Close()
			return
		}

		var email string
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			email = claims["email"].(string)
		} else {
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4000, "Invalid token"))
			conn.Close()
		}
		log.Printf("Email %s has connected\n", email)

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
	// Startup the game
	game, err := InitialState()
	if err != nil {
		log.Fatalln("Got error when initializing state", err)
	}
	// Pull secrets
	jwtSecret := os.Getenv("MUHMUD_SECRET")
	if jwtSecret == "" {
		log.Fatalln("Refusing to start without a jwt secret defined. Set MUHMUD_SECRET to something")
		os.Exit(1)
	}
	// Setup our router/handlers
	r := setupRouter(game, []byte(jwtSecret))
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
