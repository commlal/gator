package main

import (
	"github.com/commlal/gator/internal/config"
	"github.com/fatih/color"
	"log"
	"fmt"

)

type State struct {
	State	*config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf(color.RedString("ERROR -- Error reading config: %v", err))
	}
	fmt.Printf(color.GreenString("Config: %+v\n", cfg))

	err = cfg.SetUser("lane")
	if err != nil {
		log.Fatalf(color.RedString("ERROR -- Could not set current user: %v", err))
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatalf(color.RedString("ERROR -- Error reading config: %v", err))
	}
	fmt.Printf(color.GreenString("Config (Again!): %+v\n", cfg))
}
