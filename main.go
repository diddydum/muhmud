package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v2"
)

// Configuration represents high-level configuration for how the mud operates.
type Configuration struct {
	// JWTSecret is the string used to sign JWT tokens.
	JWTSecret string `yaml:"jwt_secret"`
	// AllowedOrigins represents valid Origins to accept requests from
	AllowedOrigins []string `yaml:"allowed_origins"`
}

func setupRouter(s *GameState, jwtSecret []byte, origins []string) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Bearer"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
			CheckOrigin:     checkOrigin(origins),
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
		connHandle, err := s.ConnectPlayer(email)
		if err != nil {
			log.Println(err)
			conn.Close()
			return
		}
		controlLoop(s, email, conn, connHandle)
	})

	return r
}

func controlLoop(s *GameState, email string, conn *websocket.Conn, connHandle *ConnectionHandle) {
	// Write forever until mbox says to close
	go func() {
		for {
			msg, ok := <-connHandle.MBox
			if !ok {
				// Go ahead and close
				conn.Close()
				break
			}
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Printf("Got an error while writing: %s", err.Error())
			}
		}
	}()

	// Read forever until error
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			// Disconnect the user and return
			s.DisconnectPlayer(connHandle.ConnectionID)
			break
		}
		if messageType != websocket.TextMessage {
			// ignore?
			log.Println("Got a non text message on socket, ignoring")
			continue
		}

		// send the message off
		connHandle.MessageChan <- string(p)
	}
}

func main() {
	// load configuration
	bs, err := ioutil.ReadFile("muhmud.conf.yaml")
	if err != nil {
		log.Fatalln("Got errror when attempting to open config file muhmud.conf.yaml:", err)
	}
	var config Configuration
	err = yaml.UnmarshalStrict(bs, &config)
	if err != nil {
		log.Fatalln("Got error when attempting to read config:", err)
	}

	// Startup the game
	game, err := InitialState()
	if err != nil {
		log.Fatalln("Got error when initializing state", err)
	}
	// Setup our router/handlers
	r := setupRouter(game, []byte(config.JWTSecret), config.AllowedOrigins)
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}

func checkOrigin(allowedOrigins []string) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		origin := r.Header["Origin"]

		if len(origin) == 0 {
			return true
		}
		for _, o := range allowedOrigins {
			if o == origin[0] {
				return true
			}
		}
		return false
	}
}
