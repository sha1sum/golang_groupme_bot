package bot

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	w := httptest.NewRecorder()
	m := IncomingMessage{Text: "test"}
	j, _ := json.Marshal(m)

	r, err := http.NewRequest("POST", "/", bytes.NewReader(j))
	if err != nil {
		t.Fatalf("error constructing test HTTP request [%s]", err)
	}

	h := handler()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestParsing(t *testing.T) {
	var m IncomingMessage
	ex := IncomingMessage{
		AvatarURL:  "http://foo.com/bar.png",
		GroupID:    "1234",
		ID:         "1234",
		Name:       "Foo",
		SenderID:   "1234",
		SenderType: "user",
		SourceGUID: "1234",
		Text:       "@Bar",
		UserID:     "1234",
		Attachments: []Attachment{
			Attachment{
				Loci: [][2]int{
					[2]int{0, 3},
				},
				Type: "mentions",
				UserIDs: []int{
					1234,
				},
			},
			Attachment{
				Type: "image",
				URL:  "https://foo.com/bar.png",
			},
			Attachment{
				Type:       "video",
				PreviewURL: "https://foo.com/bar.jpg",
				URL:        "https://foo.com/bar.mp4",
			},
			Attachment{
				Type: "location",
				Name: "Current Location",
				Lat:  "1.234",
				Lng:  "-1.234",
			},
			Attachment{
				Type:    "event",
				EventID: "1234",
				View:    "full",
			},
		},
		CreatedAt: 1234,
		System:    false,
	}

	err := json.NewDecoder(strings.NewReader(`{
	"attachments":[
	{
	"loci":[
	[
	0,
	3
	]
	],
	"type":"mentions",
	"user_ids":[
	1234
	]
	},
	{
    "type":"image",
    "url":"https://foo.com/bar.png"
	},
	{
    "preview_url":"https://foo.com/bar.jpg",
    "type":"video",
    "url":"https://foo.com/bar.mp4"
	},
	{
    "lat":"1.234",
    "lng":"-1.234",
    "name":"Current Location",
    "type":"location"
	},
	{
    "event_id":"1234",
    "type":"event",
    "view":"full"
	}
	],
	"avatar_url":"http://foo.com/bar.png",
	"created_at":1234,
	"group_id":"1234",
	"id":"1234",
	"location":{
	"lat":"",
	"lng":"",
	"name":null
	},
	"name":"Foo",
	"picture_url":null,
	"sender_id":"1234",
	"sender_type":"user",
	"source_guid":"1234",
	"system":false,
	"text":"@Bar",
	"user_id":"1234"
	}`)).Decode(&m)

	if err != nil {
		t.Fatalf("error parsing valid json [%s]", err)
	}

	if !reflect.DeepEqual(m, ex) {
		t.Errorf("Parsed JSON was not expected values, got %+v", m)
	}
}
