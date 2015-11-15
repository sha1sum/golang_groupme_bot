package main

import (
	"github.com/sha1sum/groupme_news_bot/groupmebot"
	"github.com/sha1sum/groupme_news_bot/matchers"
	"net/url"
	"net/http"
	"encoding/json"
	"strings"
	"os"
)

func handler(writer http.ResponseWriter, request *http.Request) {
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
	groupmebot.BotID = os.Getenv("NEWS_BOT_ID")
	document, _ := matchers.Retrieve("http://news.google.com/news?q=" + url.QueryEscape(queryString) + "&output=rss")
	items := document.Channel.Item
	if len(items) < 1 {
		groupmebot.PostMessage("No results for \"" + queryString + "\".")
		return
	}
	firstStoryLink := document.Channel.Item[0].Link
	link, _ := url.Parse(firstStoryLink)
	queryValues, _ := url.ParseQuery(link.RawQuery)
	groupmebot.PostMessage(queryValues["url"][0])
}

// Get the port that Heroku is running on for the app
func getPort() string {
	var port = os.Getenv("PORT")
	if port == "" { port = "4747" }
	return ":" + port
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(getPort(), nil)
}