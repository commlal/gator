package main

import (
	"github.com/commlal/gator/internal/database"
	"github.com/fatih/color"
	"log"
	"errors"
	"context"
)

type loggedInHandler func(s *state, cmd command, user database.User) error

func middlewareLoggedIn(newfunc loggedInHandler) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.cfg.CurrentUserName == "" {
			log.Print(color.RedString("ERROR -- No user logged in to this restricted function"))
			return errors.New("No user logged in to this restricted function")
		}
		user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			log.Printf(color.RedString("ERROR -- User not in database. Unable to use %s", cmd))
			return errors.New("User not in database.")
		}
		return newfunc(s, cmd, user)
	}
}