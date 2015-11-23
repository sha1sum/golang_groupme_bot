/*
GroupMe News Bot will listen to a callback from a GroupMe bot (https://dev.groupme.com/tutorials/bots) and if the text
of the message starts with "!news" or "! news" then the string following will be used as a search term to fetch the
most popular news story from Google News using Google News's RSS output, and the story's link will be output back to the
GroupMe group via the bot.

This application also works with Heroku by listening on the port indicated by the "PORT" environment variable. If there is
no "PORT" environment variable set, then port 80 is used by default to listen for incoming requests.

To use this application, it's necessary to set the "NEWS_BOT_ID" environment variable to the bot ID of the GroupMe bot.
*/
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sha1sum/groupme_news_bot/googlenews"
	"github.com/sha1sum/groupme_news_bot/bot"
)

// handler will take an incoming HTTP request and treat it as a POST request from a GroupMe bot and then fire off the
// search function as a goroutine.
func handler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Handling request...")
	decoder := json.NewDecoder(request.Body)
	var post bot.IncomingMessage
	err := decoder.Decode(&post)
	if err != nil {
		fmt.Println(err)
	}
	if strings.ToLower(post.Text[0:6]) != "!news " && strings.ToLower(post.Text[0:7]) != "! news " {
		return
	}
	term := strings.Replace(strings.ToLower(post.Text), "!news ", "", 1)
	term = strings.Replace(term, "! news ", "", 1)
	go search(term)
}

// search takes a given search term and queries Google News for RSS output related to the term, parses the link for the
// first story returned, then posts the link using bot.PostMessage.
func search(term string) {
	fmt.Println("Searching for \"" + term + "\".")
	// Get the "NEWS_BOT_ID" environment variable to use for the BOT ID (we don't want this committed).
	bot.BotID = os.Getenv("NEWS_BOT_ID")
	fmt.Println("Using bot ID", bot.BotID+".")

	c := make(chan googlenews.Link, 1)
	// Fetch the Google news search results for the search term as an RSS feed
	go googlenews.Search(term, c)
	link := <-c
	if link.Err != nil {
		_, err := bot.PostMessage(fmt.Sprint(link.Err))
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	_, err := bot.PostMessage(link.URL)
	if err != nil {
		fmt.Println(err)
	}
}

// port determines the port to listen on as declared by the "PORT" environment variable, or uses 80 if the environment
// variable is not defined.
func port() string {
	var port = os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	fmt.Println("Using port", port)
	return ":" + port
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("HTTP handler set. Listening.")
	err := http.ListenAndServe(port(), nil)
	if err != nil {
		fmt.Println(err)
	}
}
