package main

import (
	"github.com/sha1sum/groupme_news_bot/groupmebot"
	"github.com/sha1sum/groupme_news_bot/matchers"
	"net/url"
	"net/http"
	"encoding/json"
	"strings"
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

func search(queryString string) {
	groupmebot.BotID = "754cc09eb0c0bf4f48ad01eba7"
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

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":443", nil)
}