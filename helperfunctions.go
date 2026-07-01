package main


import (
	"github.com/fatih/color"
	"log"
	"errors"
	"strings"
	_ "github.com/lib/pq"
	"regexp"
	"time"
)

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

func helperUserNotification() {
		noticeTicker := time.NewTicker(1 * time.Minute)
		for ; ; <- noticeTicker.C {
			log.Println(color.YellowString("RSS -- Collecting feeds..."))
		}
	return
}

func cleanData(s string) string {
	re := regexp.MustCompile("<[^>]*>")
	rs := strings.TrimSpace(s)
    return re.ReplaceAllString(rs, "")
}