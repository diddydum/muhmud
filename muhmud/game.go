// Contains all information about state/game playing.
package main

import (
	"errors"
	"fmt"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// The current hash cost to use for bcrypt - up this as needed.
var hashCost = 4

// GameState is a container for all state related to the game
type GameState struct {
	// Mapping from email to player
	Players          map[string]Player
	Connections      map[ConnectionID]chan<- string
	NextConnectionID ConnectionID
	mux              sync.Mutex
}

// InitialState sets up a new initial state for the game
func InitialState() (*GameState, error) {
	state := GameState{}
	state.Players = make(map[string]Player)
	state.Connections = make(map[ConnectionID]chan<- string)
	state.NextConnectionID = 0
	state.mux = sync.Mutex{}

	me, err := player("diddydum@gmail.com", "foobar")
	if err != nil {
		return nil, err
	}
	state.Players[me.Email] = *me
	you, err := player("bazbam@gmail.com", "bazbam")
	if err != nil {
		return nil, err
	}
	state.Players[you.Email] = *you

	return &state, nil
}

// Player describes a user that plays on the system. The player is distinct from
// a character in the game.
type Player struct {
	Email       string
	PassHash    []byte
	Connections map[ConnectionID]bool
}

func player(email, password string) (*Player, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	if err != nil {
		return nil, err
	}

	return &Player{Email: email, PassHash: hash, Connections: make(map[ConnectionID]bool)}, nil
}

// ConnectionID represents a "handle" to a connection.
type ConnectionID int

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

// ConnectPlayer connects a player to the game.
func (s *GameState) ConnectPlayer(email string) (ConnectionID, <-chan string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	player, ok := s.Players[email]
	if !ok {
		return -1, nil, fmt.Errorf("tried to connect player %s but doesn't exist in our list", email)
	}
	// for now, message everyone
	if len(player.Connections) == 0 {
		s.NotifyEveryone(fmt.Sprintf("%s has connected.", email))
	}
	mbox := make(chan string, 100)
	mbox <- WelcomeMsg()

	connID := s.NextConnectionID
	s.NextConnectionID = s.NextConnectionID + 1

	s.Connections[connID] = mbox
	player.Connections[connID] = true

	return connID, mbox, nil
}

// DisconnectPlayer removes a connection from the game.
func (s *GameState) DisconnectPlayer(connID ConnectionID) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	// TODO consider adding map connection->player
	var player Player
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

	// remove the connection, close the mbox, then notify everyone.
	delete(player.Connections, connID)
	mbox, ok := s.Connections[connID]
	if !ok {
		return fmt.Errorf("connection was found for player %s but not in connection map", player.Email)
	}
	delete(s.Connections, connID)
	close(mbox)

	if len(player.Connections) == 0 {
		s.NotifyEveryone(fmt.Sprintf("%s has disconnected.", player.Email))
	}
	return nil
}

// NotifyEveryone sends a raw message to everyone.
func (s *GameState) NotifyEveryone(msg string) {
	for _, mbox := range s.Connections {
		mbox <- msg
	}
}

// WelcomeMsg is a welcome message
func WelcomeMsg() string {
	return "Welcome to muhmud; the mud with %rfantastic%c colors."
}
