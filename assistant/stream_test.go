package assistant_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sburchfield/go-assistant-api/assistant"
)

func TestToSSE(t *testing.T) {
	msgs := make(chan string, 2)
	msgs <- "Hello"
	msgs <- "world"
	close(msgs)

	recorder := httptest.NewRecorder()
	assistant.ToSSE(context.TODO(), recorder, msgs)

	res := recorder.Result()
	defer res.Body.Close()

	if ct := res.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected content-type 'text/event-stream', got '%s'", ct)
	}

	body := recorder.Body.String()
	events := strings.Split(strings.TrimSpace(body), "\n\n")

	if len(events) != 4 {
		t.Fatalf("expected 4 SSE events (f, 0, 0, e/d), got %d:\n%s", len(events), body)
	}

	// Check first event (f: messageId)
	var f map[string]any
	if err := json.Unmarshal([]byte(strings.TrimPrefix(events[0], "data: ")), &f); err != nil || f["f"] == nil {
		t.Errorf("expected first chunk to have 'f', got: %v", f)
	}

	// Check token chunks
	for i, expected := range []string{"Hello", "world"} {
		var token map[string]string
		err := json.Unmarshal([]byte(strings.TrimPrefix(events[i+1], "data: ")), &token)
		if err != nil {
			t.Fatalf("error decoding chunk %d: %v", i+1, err)
		}
		if token["0"] != expected {
			t.Errorf("expected token '%s', got '%s'", expected, token["0"])
		}
	}

	// Check final event (e and d)
	var end map[string]any
	if err := json.Unmarshal([]byte(strings.TrimPrefix(events[3], "data: ")), &end); err != nil {
		t.Errorf("error decoding final chunk: %v", err)
	}
	if _, ok := end["e"]; !ok {
		t.Errorf("expected final chunk to contain 'e'")
	}
	if _, ok := end["d"]; !ok {
		t.Errorf("expected final chunk to contain 'd'")
	}
}
