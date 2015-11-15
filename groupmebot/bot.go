package groupmebot

import (
	"encoding/json"
	"net/http"
	"strings"
)

var BotID string

// All we care about in the incoming bot messages are the text of the message
type IncomingMessage struct {
	Text string `json:"text"`
}

// Post a message using a GroupMe bot
func PostMessage(message string) (*http.Response, error) {
	messageMap := map[string]string{"bot_id": BotID, "text": message}
	jsonMap, _ := json.Marshal(messageMap)
	return http.Post("https://api.groupme.com/v3/bots/post", "application/json", strings.NewReader(string(jsonMap)))
}