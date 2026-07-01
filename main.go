package main

import (
	"github.com/commlal/gator/internal/config"
	"github.com/commlal/gator/internal/database"
	"github.com/fatih/color"
	"log"
	"os"
	"database/sql"
	_ "github.com/lib/pq"
)
/*Log Messages

Yellow 	FEED
Magenta	DEBUG
Green	AUTHENTICATION
Red		ERROR
Cyan	DATABASE
Blue	----

*/
type state struct {
	cfg	*config.Config
	db	*database.Queries
}

type command struct {
	name	string
	args	[]string
}

type commands struct {
	cmdMap 	map[string]func(*state, command) error
}

func main() {

	//Read in config file
	c, err := config.Read()
	if err != nil {
		log.Fatalf(color.RedString("ERROR -- Error reading config: %v", err))
	}
	//log.Printf(color.GreenString("AUTHENTICATION -- Loggedin User: %+v", c.CurrentUserName))
	
	//Connect to database
	db, err := sql.Open("postgres", c.DBURL)
	if err != nil {
		log.Fatalf(color.RedString("ERROR -- Error connecting to Postgres database: %v", err))
	}
	dbQueries := database.New(db)
	s := state{cfg: &c, db: dbQueries}

	cl := color.New(color.BgHiGreen).Add(color.FgBlack)
	cl.Println("Welcome to Gator - The Golang Blog Aggregator")

	//Registering available handler functions
	commands := commands{cmdMap: make(map[string]func(*state, command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("feeds", handlerListFeeds)

	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("follow", middlewareLoggedIn(handlerFollowFeed))
	commands.register("following", middlewareLoggedIn(handlerUserFollows))
	commands.register("unfollow", middlewareLoggedIn(handlerUserUnfollows))
	commands.register("browse", middlewareLoggedIn(handlerUserBrowse))

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
