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