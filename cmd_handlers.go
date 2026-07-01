package main

import (
	"github.com/commlal/gator/internal/database"
	"github.com/fatih/color"
	"log"
	"fmt"
	"errors"
	_ "github.com/lib/pq"
	"context"
	"time"
	"strconv"
)
/*Log Messages

Yellow 	FEED
Magenta	DEBUG
Green	AUTHENTICATION
Red		ERROR
Cyan	DATABASE
Blue	----

*/

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
		log.Printf(color.RedString("ERROR -- Could not access full user list: %v", err))
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
	//Getting time from argument and parsing it
	if len(cmd.args) == 0 {
		log.Print(color.RedString("ERROR -- Missing duration value"))
		return errors.New("Missing duration value")
	}
	timeString := cmd.args[0]
	timebetweenreqs, err := time.ParseDuration(timeString)
	
	if err != nil {
		log.Print(color.RedString("ERROR -- Invalid time submitted: %v", err))
		return errors.New("Invalid time submitted")
	}
	log.Print(color.YellowString("RSS -- Aggregator fetch cycle set to %v", timebetweenreqs))


	//User notification ticker
	go helperUserNotification()

	//Start aggrigation loop
	ticker := time.NewTicker(timebetweenreqs)
	for ; ; <-ticker.C {
		err = scrapeFeed(s, context.Background())
		if err != nil {
			log.Print(color.RedString("ERROR -- Error in scrapeFeed function: %v", err))
			return errors.New("Error in scrapeFeed function")
		}
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 || len(cmd.args) == 1 {
		log.Print(color.RedString("ERROR -- Missing either feed name or URL"))
		return errors.New("Missing either feed name or URL")
	}
	feedName := cmd.args[0]
	feedURL := cmd.args[1]
	log.Printf(color.GreenString("AUTHENTICATION -- %s User ID: %v"), user.Name, user.ID)

	CreateFeedParams := database.CreateFeedParams{
		Name:   feedName,
		Url:    feedURL,
		UserID: user.ID,
		}
	
	createdFeed, err := s.db.CreateFeed(context.Background(), CreateFeedParams)
	if err != nil {
		log.Print(color.RedString("ERROR -- Feed already in database:", err))
		return errors.New(fmt.Sprintf("Feed for URL %v already in database", feedURL))
	}

	CreateFeedFollowParams :=  database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: createdFeed.ID,
	}
	followResults, err := s.db.CreateFeedFollow(context.Background(), CreateFeedFollowParams)
	if err != nil {
		log.Printf(color.RedString("ERROR -- User %s already follows this blog"), s.cfg.CurrentUserName)
		return errors.New(fmt.Sprintf("User %s already follows this blog", s.cfg.CurrentUserName))
	}
	log.Printf(color.YellowString("FEED -- %s added %s(URL:%s) feed to database"), followResults.UserName, followResults.FeedName, feedURL)
	return nil
}

func handlerListFeeds(s *state, cmd command) error {
	log.Print(color.CyanString("DATABASE -- Listing all feeds"))

	feedList, err := s.db.ListAllFeeds(context.Background())
	if err != nil {
		log.Print(color.RedString("ERROR -- Unable to pull list of feeds from database"))
		return errors.New("Unable to pull list of feeds from database")
	}

	for i, feed := range feedList {
		userName, err := s.db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
		log.Print(color.RedString("ERROR -- Invalid User ID:", err))
		}
		fmt.Printf("FEED %v\n", i)
		fmt.Printf("BLOG TITLE: %s\n", feed.Name)
		fmt.Printf("BLOG URL: %s\n", feed.Url)
		fmt.Printf("ADDED BY: %s\n\n", userName)
	}
	return nil
}

func handlerFollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		log.Print(color.RedString("ERROR -- Missing feed URL"))
		return errors.New("Missing feed URL")
	}

	feedURL := cmd.args[0]
	feedID, err := s.db.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		log.Print(color.RedString("ERROR -- Feed not found in feed database - use addfeed to add blog to gator!"))
		return errors.New("Feed not found in feed database - use addfeed to add blog to gator!")
	}

	CreateFeedFollowParams :=  database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feedID.ID,
	}

	followResults, err := s.db.CreateFeedFollow(context.Background(), CreateFeedFollowParams)
	if err != nil {
		log.Printf(color.RedString("ERROR -- User %s already follows this blog"), user.Name)
		return errors.New(fmt.Sprintf("User %s already follows this blog", user.Name))
	}
	fmt.Printf(color.YellowString("NEW FEED FOLLOW CREATED\n Blog: %s \n User: %s\n", followResults.FeedName, followResults.UserName))
	log.Printf(color.YellowString("FEED -- New Feed Follow Created: User %s, Blog %s"), followResults.UserName, followResults.FeedName)
	return nil
}

func handlerUserFollows(s *state, cmd command, user database.User) error {
	feedList, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		log.Printf(color.RedString("ERROR -- Issue retrieving follows for user: %v"), user.Name)
		return errors.New(fmt.Sprintf("Issue retrieving follows for user: %v", user.Name))
	}
	fmt.Printf(color.CyanString("DATABASE -- User %s follows:\n", user.Name))
	for _, feed := range feedList {
		fmt.Printf("%s\n", feed.FeedName)
	}
	return nil
}

func handlerUserUnfollows(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		log.Print(color.RedString("ERROR -- Missing feed URL"))
		return errors.New("Missing feed URL")
	}

	feedURL := cmd.args[0]

	feedData, err := s.db.GetFeedByURL(context.Background(), feedURL) //returns Feed struct
	if err != nil {
		log.Print(color.RedString("ERROR -- Feed not found in feed database"))
		return errors.New("Feed not found in feed database")
	}
	deleteFeed := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feedData.ID,
	}
	err = s.db.DeleteFeedFollow(context.Background(), deleteFeed)
	if err != nil {
		log.Print(color.RedString("ERROR -- User is not following feed"))
		return errors.New("User is not following feed")
	}
	log.Printf(color.YellowString("User %s has stopped following %s", user.Name,  feedData.Name))
	return nil
}

func handlerUserBrowse(s *state, cmd command, user database.User) error {
	//Set limit for posts
	GetPostsForUserParams := database.GetPostsForUserParams{
		UserID:	user.ID,
		Limit:  2,
	}

	if len(cmd.args) == 0 {
		log.Print(color.YellowString("RSS -- Limit set to default of 2 entries"))
	} else {
		limit, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			log.Print(color.YellowString("RSS -- Invalid number! Limit set to the default of 2"))
		} else {
			GetPostsForUserParams.Limit = int32(limit)
			log.Print(color.YellowString("RSS -- Limit set to default to %v entries", GetPostsForUserParams.Limit))
		}
	}
	userFeed, err := s.db.GetPostsForUser(context.Background(), GetPostsForUserParams)
	if err != nil {
		log.Print(color.RedString("ERROR -- Unable to pull posts for user"))
		return errors.New(("Unable to pull posts for user"))
	}
	//Print posts to terminal

	for i, post := range userFeed {
		fmt.Printf("Post %v\n", i+1)
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("%s\n \n", post.Body)
	}
	return nil
}