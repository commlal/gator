package main

import (
	"github.com/commlal/gator/internal/config"
	"github.com/fatih/color"
	"log"
	"fmt"
	"errors"
	"os"
)

type state struct {
	cfg	*config.Config
}

type command struct {
	name	string
	args	[]string
}

type commands struct {
	cmdMap 	map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	//runs a given command with the provided state if it exists
	cmdFunc, found := c.cmdMap[cmd.name]
	if !found {
		return errors.New(color.RedString("ERROR -- Command not found: %s", cmd.name))
	}
	err := cmdFunc(s, cmd)
	if err != nil {
		return errors.New(color.RedString("ERROR -- Command could not be run: %s", cmd.name))
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmdMap[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		log.Print(color.RedString("ERROR -- No Username submitted"))
		return errors.New("Username can not be empty")
	}
	err := s.cfg.SetUser(cmd.args[0])
	if err != nil {
		log.Printf(color.RedString("ERROR -- Unable to set %v as active user", cmd.args[0]))
		return errors.New(fmt.Sprintf("Unable to set %v as active user", cmd.args[0]))
	}
	fmt.Printf(color.GreenString("User set to %v\n", cmd.args[0]))
	return nil
}


func main() {

	//Read in config file
	c, err := config.Read()
	if err != nil {
		log.Fatalf(color.RedString("ERROR -- Error reading config: %v", err))
	}
	fmt.Printf(color.GreenString("Config Initiated: %+v\n", c))
	s := state{cfg: &c}

	//Registering available handler functions
	commands := commands{cmdMap: make(map[string]func(*state, command) error)}
	commands.register("login", handlerLogin)

	//Running Commands
	CLIargs := os.Args
	if len(CLIargs) < 2 {
		log.Fatal(color.RedString("ERROR -- Command or arguments missing"))
	}
	
	cmdName := CLIargs[1]
	args := CLIargs[2:]
	cmd := command{
		name: cmdName, 
		args: args,
	}

	err = commands.run(&s, cmd)
	if err != nil {
		log.Fatalf(color.RedString("%s", err))
	}
}
