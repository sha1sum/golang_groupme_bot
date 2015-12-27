package bot

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
