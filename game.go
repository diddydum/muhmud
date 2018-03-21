// Contains all information about state/game playing.
package main

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// The current hash cost to use for bcrypt - up this as needed.
var hashCost = 4

// GameState is a container for all state related to the game
type GameState struct {
	Players map[string]Player
}

// Player describes a user that plays on the system. The player is distinct from
// a character in the game.
type Player struct {
	Username string
	Email    string
	PassHash []byte
}

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

func player(username, email, password string) (*Player, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	if err != nil {
		return nil, err
	}

	return &Player{Username: username, Email: email, PassHash: hash}, nil
}

// InitialState sets up a new initial state for the game
func InitialState() (*GameState, error) {
	state := GameState{}
	state.Players = make(map[string]Player)

	me, err := player("diddydum", "diddydum@gmail.com", "foobar")
	if err != nil {
		return nil, err
	}
	state.Players[me.Username] = *me

	return &state, nil
}
