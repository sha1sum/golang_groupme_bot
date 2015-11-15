package main

import (
	"github.com/sha1sum/groupme_news_bot/groupmebot"
	"github.com/sha1sum/groupme_news_bot/matchers"
	"net/url"
	"net/http"
	"encoding/json"
	"strings"
	"os"
	"fmt"
)

func handler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Handling request...")
	decoder := json.NewDecoder(request.Body)
	var post groupmebot.IncomingMessage
	decoder.Decode(&post)
	if strings.ToLower(post.Text[0:6]) != "!news " && strings.ToLower(post.Text[0:7]) != "! news " { return }
	queryString := strings.Replace(strings.ToLower(post.Text), "!news ", "", 1)
	queryString = strings.Replace(queryString, "! news ", "", 1)
	go search(queryString)
}

// Take the given search string submitted on GroupMe and get the latest story from Google, posting it with the bot.
func search(queryString string) {
	fmt.Println("Searching for \"" + queryString + "\"")
	// Get the "NEWS_BOT_ID" environment variable to use for the BOT ID (we don't want this committed).
	groupmebot.BotID = os.Getenv("NEWS_BOT_ID")
	fmt.Println("Using bot ID", groupmebot.BotID)
	// Fetch the Google news search results for the search term as an RSS feed
	document, _ := matchers.Retrieve("http://news.google.com/news?q=" + url.QueryEscape(queryString) + "&output=rss")
	// Get the <item>'s from the feed
	items := document.Channel.Item
	// If there are no items, return a "No results" message
	if len(items) < 1 {
		groupmebot.PostMessage("No results for \"" + queryString + "\".")
		return
	}
	fmt.Println("Link retrieved.")
	// Get the link with all the Googley stuff in it
	firstStoryLink := document.Channel.Item[0].Link
	link, _ := url.Parse(firstStoryLink)
	// Get the query string values so we can just get the normal URL instead of the Googley one
	queryValues, _ := url.ParseQuery(link.RawQuery)
	fmt.Println("Posting", queryValues["url"][0])
	// Post the link to GroupMe
	groupmebot.PostMessage(queryValues["url"][0])
}

// Get the port that Heroku is running on for the app
func getPort() string {
	var port = os.Getenv("PORT")
	if port == "" { port = "4747" }
	fmt.Println("Using port", port)
	return ":" + port
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("HTTP handler set. Listening.")
	http.ListenAndServe(getPort(), nil)
}