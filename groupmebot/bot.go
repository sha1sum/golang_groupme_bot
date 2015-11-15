package groupmebot

import (
	"encoding/json"
	"net/http"
	"strings"
)

var BotID string

type IncomingMessage struct {
	Text string `json:"text"`
}

func PostMessage(message string) (*http.Response, error) {
	messageMap := map[string]string{"bot_id": BotID, "text": message}
	jsonMap, _ := json.Marshal(messageMap)
	return http.Post("https://api.groupme.com/v3/bots/post", "application/json", strings.NewReader(string(jsonMap)))
}