package main

import (
	"encoding/xml"
	"net/http"
	"log"
	"github.com/fatih/color"
	"context"
	"fmt"
	"errors"
	"io"
	"html"
	"github.com/commlal/gator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	//Make GET request to feedURL to retrieve XML
	compiledRSS := RSSFeed{}
	
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		log.Printf(color.RedString("ERROR -- Unable to generate request to: %s", feedURL))
		return nil, errors.New(fmt.Sprintf("Unable to generate request to: %s", feedURL))
	}
	req.Header.Set("User-Agent", "gator")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Printf(color.RedString("ERROR -- Unable to retrieve XML from: %s", feedURL))
		log.Printf(color.RedString("----- -- Error Information: %v", err))
		return nil, errors.New(fmt.Sprintf("Unable to retrieve XML from: %s", feedURL))
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf(color.RedString("ERROR -- Unable to read response: %v", err))
		return nil, errors.New(fmt.Sprintf("ERROR -- Unable to read response: %v", err))
	}

	err = xml.Unmarshal(data, &compiledRSS)
	if err != nil {
		log.Printf(color.RedString("ERROR -- Unable to parse XML data: %v", err))
		return nil, errors.New(fmt.Sprintf("ERROR -- Unable to parse XML data: %v", err))
	}

	compiledRSS.Channel.Title = html.UnescapeString(compiledRSS.Channel.Title)
	compiledRSS.Channel.Description = html.UnescapeString(compiledRSS.Channel.Description)
	
	for i, item := range compiledRSS.Channel.Item {
		compiledRSS.Channel.Item[i].Title = html.UnescapeString(item.Title)
		compiledRSS.Channel.Item[i].Description = html.UnescapeString(item.Description)
	}

	return &compiledRSS, nil
}

func scrapeFeed (s *state, ctx context.Context) error {
	//Get the next feed to fetch from database
	nextFeedData , err := s.db.GetNextFeedToFetch(context.Background()) //URL is nextFeedData.Url, ID is nextFeedData.ID
	if err != nil {
		log.Printf(color.RedString("ERROR -- Unable to obtain next feed from database: %v", err))
		return errors.New(fmt.Sprintf("ERROR -- Unable to obtain next feed from database: %v", err))
	}
	log.Printf(color.YellowString("RSS -- Next blog to fetch is %s from %s", nextFeedData.Name, nextFeedData.Url))

	//Mark it as fetched
	err = s.db.MarkFeedFetched(context.Background(), nextFeedData.ID)
	if err != nil {
		log.Printf(color.RedString("ERROR -- Unable to mark the feed (URL:%s) as 'fetched': %v", nextFeedData.Name, err))
		return errors.New(fmt.Sprintf("ERROR -- Unable to mark the feed (URL:%s) as 'fetched': %v", nextFeedData.Name, err))
	}
	log.Print(color.YellowString("RSS -- Fetch time updated"))
	
	//Fetch the feed
	rssFeed, err := fetchFeed(context.Background(), nextFeedData.Url)
	if err != nil {
		log.Printf(color.RedString("ERROR -- Unable to obtain next feed from database: %v", err))
		return errors.New(fmt.Sprintf("ERROR -- Unable to obtain next feed from database: %v", err))
	}
	log.Printf(color.YellowString("RSS -- %s has been fetched. Processing data...", nextFeedData.Name))

	//Save feed to database
	for _, entry := range rssFeed.Channel.Item {
		_, err := s.db.PostInDatabase(context.Background(), entry.Link)
		if err == nil {
			log.Printf(color.CyanString("DATABASE -- Post already in database: %v", cleanData(entry.Title)))
		} else {
			CreatePostParams := database.CreatePostParams{
				Title:			cleanData(entry.Title),
				Url:			entry.Link,
				Description:	cleanData(entry.Description),
				ToTimestamp:	entry.PubDate,
				FeedID:       	nextFeedData.ID,
			}
			_, err = s.db.CreatePost(context.Background(), CreatePostParams)
			fmt.Println(color.YellowString("RSS -- Added %s to database", cleanData(entry.Title)))
		}
	}
		return nil
}