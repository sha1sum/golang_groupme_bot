/*
Package groupmebot handles posting a message to a GroupMe bot.

To use the bot functionality, you will need to first set BotID to the ID of the bot you wish to use.
*/
package bot

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// BotID is the ID of the GroupMe bot as found on GroupMe's developer site.
var BotID string

// IncomingMessage is used to indicate the message properties from the POST sent from a GroupMe bot callback.
type IncomingMessage struct {
	Text       string `json:"text"`
	Name       string `json:"name"`
	UserID     string `json:"user_id"`
	SenderType string `json:"sender_type"`
}

// Link is used to house a string URL along with any error that may have resulted from fetching the URL from
// the resource.
type OutgoingMessage struct {
	Message string
	Err     error
}

// Handler will be used to perform actions and output an OutgoingMessage result to a channel.
type Handler interface {
	Handle(term string, c chan []*OutgoingMessage, message IncomingMessage)
}

// PostMessage posts a string to a GroupMe bot as long as the BotID is present.
func PostMessage(message string) (*http.Response, error) {
	if len(BotID) < 1 {
		return nil, errors.New("BotID cannot be blank.")
	}
	m := map[string]string{"bot_id": BotID, "text": message}
	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return http.Post("https://api.groupme.com/v3/bots/post", "application/json", strings.NewReader(string(j)))
}
