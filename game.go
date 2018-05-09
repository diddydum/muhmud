// Contains all information about state/game playing.
package main

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// The current hash cost to use for bcrypt - up this as needed.
var hashCost = 4

// The default buffer size for all channels.
var defaultBufferSize = 32

// GameState is a container for all state related to the game
type GameState struct {
	// Mapping from email to player
	Players          map[string]*Player
	Connections      map[ConnectionID]chan<- string
	NextConnectionID ConnectionID
	Characters       map[ID]*Character
	NextID           ID
	mux              sync.Mutex
}

// InitialState sets up a new initial state for the game
func InitialState() (*GameState, error) {
	state := GameState{}
	state.Players = make(map[string]*Player)
	state.Connections = make(map[ConnectionID]chan<- string)
	state.NextConnectionID = 0
	state.Characters = make(map[ID]*Character)
	state.NextID = 0
	state.mux = sync.Mutex{}

	me, err := player("diddydum@gmail.com", "foobar")
	if err != nil {
		return nil, err
	}
	state.Players[me.Email] = me
	you, err := player("bazbam@gmail.com", "bazbam")
	if err != nil {
		return nil, err
	}
	state.Players[you.Email] = you

	return &state, nil
}

// Player describes a user that plays on the system. The player is distinct from
// a character in the game.
type Player struct {
	Email       string
	PassHash    []byte
	Connections map[ConnectionID]bool
	// MessageChan represents a channel for incoming messages - if the player is
	// connected, a goroutine will read messages from here and handle them.
	MessageChan chan string
	Attributes  map[string]interface{}
}

// MessageHandler handles messages and returns an error
type MessageHandler func(pName string, s *GameState) error

func player(email, password string) (*Player, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	if err != nil {
		return nil, err
	}

	return &Player{
		Email:       email,
		PassHash:    hash,
		Connections: make(map[ConnectionID]bool),
		Attributes:  make(map[string]interface{})}, nil
}

// ConnectionID represents a "handle" to a connection.
type ConnectionID int

// ID type refers to, well, any entity in the game (not including connection).
type ID int

// CheckPassword checks if the provided password is valid for the given
// username. Returns nil on success, error on failure.
func (s *GameState) CheckPassword(username, password string) error {
	user, ok := s.Players[username]

	if !ok {
		// TDO: Make sure to burn some cpu time so as not to discover unknown users
		return errors.New("Unknown username " + username)
	}

	return bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
}

// ConnectionHandle has all the needed details for an ongoing connection. Note
// that both channels are handled by the server and should not be closed by
// consumers.
type ConnectionHandle struct {
	ConnectionID ConnectionID
	MessageChan  chan<- string
	MBox         <-chan string
}

// ConnectPlayer connects a player to the game.
func (s *GameState) ConnectPlayer(email string) (*ConnectionHandle, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	player, ok := s.Players[email]
	if !ok {
		return nil, fmt.Errorf(
			"tried to connect player %s but doesn't exist in our list", email)
	}
	// Check to see if this is the first connection. If it is, start a goroutine
	// reader and message channel
	if len(player.Connections) == 0 {
		player.MessageChan = make(chan string, defaultBufferSize)
		// TODO here is where we start a goroutine for reading and handling messages
		go func() {
			for {
				msg, ok := <-player.MessageChan
				if !ok {
					break
				}
				log.Printf("Got message: %s", msg)
			}
		}()
	}

	// Mbox is the outgoing channel for writing directly to this connection
	mbox := make(chan string, defaultBufferSize)

	connID := s.NextConnectionID
	s.NextConnectionID = s.NextConnectionID + 1

	s.Connections[connID] = mbox
	player.Connections[connID] = true

	return &ConnectionHandle{
		MessageChan:  player.MessageChan,
		ConnectionID: connID,
		MBox:         mbox}, nil
}

// DisconnectPlayer removes a connection from the game.
func (s *GameState) DisconnectPlayer(connID ConnectionID) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	// TODO consider adding map connection->player
	var player *Player
	for _, p := range s.Players {
		if _, ok := p.Connections[connID]; ok {
			player = p
			break
		}
	}
	// check if we didn't find anything and complain
	if player.Email == "" {
		return fmt.Errorf("can't find connectionID %v", connID)
	}

	// remove the connection, close the mbox
	delete(player.Connections, connID)
	if len(player.Connections) == 0 {
		// Close the message chan, no more messages coming
		close(player.MessageChan)
		player.MessageChan = nil
	}

	mbox, ok := s.Connections[connID]
	if !ok {
		return fmt.Errorf("connection was found for player %s but not in connection map", player.Email)
	}
	delete(s.Connections, connID)
	close(mbox)

	return nil
}

// NotifyEveryone sends a raw message to everyone.
func (s *GameState) NotifyEveryone(msg string) {
	for _, mbox := range s.Connections {
		mbox <- msg
	}
}

// ToPlayer sends a message to a single player
func (s *GameState) ToPlayer(name, msg string) {
	p, ok := s.Players[name]
	if !ok {
		log.Printf("attempted to send a message to player %s but wasn't found\n", name)
		return
	}
	for connID := range p.Connections {
		mbox, ok := s.Connections[connID]
		if !ok {
			log.Printf("connID doesn't match and unable to message for name %s\n", name)
			return
		}
		mbox <- msg
	}
}
