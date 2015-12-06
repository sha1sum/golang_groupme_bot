/*
Golang GroupMe Bot will listen to a callback from a GroupMe bot (https://dev.groupme.com/tutorials/bots) and if the text
of the message starts with a recognized trigger then the string following will be used as a search term to fetch the
results and a message will be output back to the GroupMe group via the bot.

This application also works with Heroku by listening on the port indicated by the "PORT" environment variable. If there is
no "PORT" environment variable set, then port 80 is used by default to listen for incoming requests.

To use this application, it's necessary to set the "GROUPME_BOT_ID" environment variable to the bot ID of the GroupMe bot.
*/
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sha1sum/golang_groupme_bot/handlers/googlenews"
	"github.com/sha1sum/golang_groupme_bot/bot"
	"github.com/sha1sum/golang_groupme_bot/handlers/adultpoints"
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
	if post.Text[0:1] != "!" { return }
	var term string
	if strings.ToLower(post.Text[0:6]) == "!news " || strings.ToLower(post.Text[0:7]) == "! news " {
		term = strings.Replace(strings.ToLower(post.Text), "!news ", "", 1)
		term = strings.Replace(term, "! news ", "", 1)
		go search(term, new(googlenews.Handler), post)
		return
	}
	term = strings.Trim(post.Text[1:], " ")
	go search(term, new(adultpoints.Handler), post)
}

// search takes a given search term and queries uses the searcher to find the term, and then
// posts the message returned from the searcher using bot.PostMessage.
func search(term string, searcher bot.Handler, message bot.IncomingMessage) {
	fmt.Println("Searching for \"" + term + "\".")
	// Get the "NEWS_BOT_ID" environment variable to use for the BOT ID (we don't want this committed).
	bot.BotID = os.Getenv("GROUPME_BOT_ID")
	fmt.Println("Using bot ID", bot.BotID+".")

	c := make(chan *bot.OutgoingMessage, 1)
	// Fetch the Google news search results for the search term as an RSS feed
	go searcher.Handle(term, c, message)
	m := <-c
	if m.Err != nil {
		_, err := bot.PostMessage(fmt.Sprint(m.Err))
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	_, err := bot.PostMessage(m.Message)
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
