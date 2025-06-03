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

	if len(events) != 5 { // f + 2 tokens + d + e
		t.Fatalf("expected 5 SSE events (f, 0, 0, d, e), got %d:\n%s", len(events), body)
	}

	// Check first event: f:{"messageId":"..."}
	if !strings.HasPrefix(events[0], "f:{") {
		t.Errorf("expected first line to start with 'f:', got: %s", events[0])
	}

	var f map[string]string
	if err := json.Unmarshal([]byte(strings.TrimPrefix(events[0], "f:")), &f); err != nil {
		t.Errorf("error decoding 'f' event: %v", err)
	}
	if f["messageId"] == "" {
		t.Errorf("expected 'messageId' in 'f' event, got: %v", f)
	}

	// Token chunks: 0:"Hello", 0:"world"
	for i, expected := range []string{"Hello", "world"} {
		if !strings.HasPrefix(events[i+1], "0:") {
			t.Errorf("expected event %d to start with '0:', got: %s", i+1, events[i+1])
			continue
		}
		var token string
		if err := json.Unmarshal([]byte(strings.TrimPrefix(events[i+1], "0:")), &token); err != nil {
			t.Errorf("error decoding token %d: %v", i+1, err)
		}
		if token != expected {
			t.Errorf("expected token '%s', got '%s'", expected, token)
		}
	}

	// Check d: and e: chunks
	for _, prefix := range []string{"d:", "e:"} {
		found := false
		for _, line := range events {
			if strings.HasPrefix(line, prefix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected stream to contain %q event", prefix)
		}
	}
}
