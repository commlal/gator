package main

import (
	"github.com/commlal/gator/internal/config"
	"github.com/commlal/gator/internal/database"
	"github.com/fatih/color"
	"log"
	"fmt"
	"errors"
	"os"
	"database/sql"
	_ "github.com/lib/pq"
	"context"

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

func main() {

	//Read in config file
	c, err := config.Read()
	if err != nil {
		log.Fatalf(color.RedString("ERROR -- Error reading config: %v", err))
	}
	log.Printf(color.GreenString("AUTHENTICATION -- Loggedin User: %+v", c.CurrentUserName))
	
	//Connect to database
	db, err := sql.Open("postgres", c.DBURL)
	if err != nil {
		log.Fatalf(color.RedString("ERROR -- Error connecting to Postgres database: %v", err))
	}
	dbQueries := database.New(db)
	s := state{cfg: &c, db: dbQueries}

	//Registering available handler functions
	commands := commands{cmdMap: make(map[string]func(*state, command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", handlerAddFeed)

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

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		log.Print(color.RedString("ERROR -- No Username submitted"))
		return errors.New("Username can not be empty")
	}
	userName := cmd.args[0]
	//Check user is in database

	_, err := s.db.GetUserByName(context.Background(), userName)
	if err != nil {
		log.Printf(color.RedString("ERROR -- User not in database: %v"), userName)
		return errors.New(fmt.Sprintf("User not in database: %v"))
	}

	err = s.cfg.SetUser(userName)
	if err != nil {
		log.Printf(color.RedString("ERROR -- Unable to set %v as active user", userName))
		return errors.New(fmt.Sprintf("Unable to set %v as active user", userName))
	}
	fmt.Printf(color.GreenString("AUTHENTICATION -- User set to %v\n", userName))
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		log.Print(color.RedString("ERROR -- No Username submitted"))
		return errors.New("Username can not be empty")
	}
	userName := cmd.args[0]
	//Check if user already exists in database

	_, err := s.db.GetUserByName(context.Background(), userName)
	if err == nil {
		log.Printf(color.RedString("ERROR -- User already exists: %v"), userName)
		return errors.New(fmt.Sprintf("User already exists: %v"))
	}

	//Register new user
	_, err = s.db.CreateUser(context.Background(), userName)
	if err != nil {
		log.Printf(color.RedString("ERROR -- Could not create user %v: %v"), userName, err)
		return errors.New("Error creating new user")
	}
	err = s.cfg.SetUser(userName)
	if err != nil {
		log.Printf(color.RedString("ERROR -- Unable to set newly registered %v as active user", userName))
		return errors.New(fmt.Sprintf("Unable to set newly registered %v as active user", userName))
	}
	log.Print(color.CyanString("DATABASE -- User Created: %v", userName))
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.PurgeUsers(context.Background())
	if err != nil {
		log.Print(color.RedString("ERROR -- Could not reset user database: %v"))
		return errors.New("Error resetting database")
	}
	log.Print(color.CyanString("DATABASE -- Database Reset"))
	return nil
}

func handlerUsers(s *state, cmd command) error {
	log.Print(color.CyanString("DATABASE -- Listing all users"))
	userList, err := s.db.GetAllUsers(context.Background())
	if err != nil {
		log.Print(color.RedString("ERROR -- Could not access full user list: %v"))
		return errors.New("Error retrieving user database")
	}

	for _, name := range userList {
		if name == s.cfg.CurrentUserName {
			fmt.Printf("%s (current)\n", name)
		} else {
			fmt.Println(name)
		}
		
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	feedURL := "https://www.wagslane.dev/index.xml"
	log.Print(color.YellowString("RSS -- Attepting to pull RSS Feeds"))

	
	rssFeed, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		log.Print(color.RedString("ERROR -- Unable to retrieve RSS"))
		return errors.New("Error retrieving RSS Feed")
	}
	fmt.Printf(color.YellowString("RSS FEED: %v \n", rssFeed))
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) == 0 || len(cmd.args) == 1 {
		log.Print(color.RedString("ERROR -- Missing either feed name or URL"))
		return errors.New("Missing either feed name or URL")
	}
	feedName := cmd.args[0]
	feedURL := cmd.args[1]
	currentUser := s.cfg.CurrentUserName
	userID, err := s.db.GetUserByName(context.Background(), currentUser)
	if err != nil {
		log.Print(color.RedString("ERROR -- User %v not found in database", currentUser))
		log.Print(color.RedString("ERROR -- Error information: %v", err))
		return errors.New(fmt.Sprintf("User %v not found in database", currentUser))
	}

	log.Printf(color.GreenString("AUTHENTICATION -- %s User ID: %v"), currentUser, userID)

	CreateFeedParams := database.CreateFeedParams{
		Name:   feedName,
		Url:    feedURL,
		UserID: userID,
		}
	
	createdFeed, err := s.db.CreateFeed(context.Background(), CreateFeedParams)
	if err != nil {
		log.Print(color.RedString("ERROR -- Feed already in database:", err))
		return errors.New(fmt.Sprintf("Feed for URL %v already in database", feedURL))
	}
	log.Printf(color.YellowString("FEED -- %s added %s(URL:%s) feed to database"), currentUser, feedName, feedURL)
	fmt.Printf("%v", createdFeed)
	return nil
}