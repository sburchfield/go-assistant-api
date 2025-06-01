package assistant_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sburchfield/go-assistant-api/assistant"
)

func TestToSSE(t *testing.T) {
	msgs := make(chan string, 3)
	msgs <- "Hello"
	msgs <- "world"
	close(msgs)

	recorder := httptest.NewRecorder()
	assistant.ToSSE(recorder, msgs)

	res := recorder.Result()
	defer res.Body.Close()

	if ct := res.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected content-type 'text/event-stream', got '%s'", ct)
	}

	body := recorder.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n\n")
	expectedLines := []string{"data: Hello", "data: world"}

	if len(lines) != len(expectedLines) {
		t.Fatalf("expected %d SSE events, got %d: %v", len(expectedLines), len(lines), lines)
	}

	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf("event %d: expected %q, got %q", i, expected, lines[i])
		}
	}
}
