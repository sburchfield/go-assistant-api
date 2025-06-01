package gemini_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sburchfield/go-assistant-api/assistant"
	"github.com/sburchfield/go-assistant-api/assistant/provider/gemini"
)

func TestChatStream_Gemini(t *testing.T) {
	body := "" +
		"data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"Hello\"}]}}]}\n" +
		"data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\" world\"}]}}]}\n"

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}
		}),
	}

	sut := gemini.NewClientWithHTTP("fake-api-key", "gemini-pro", client)
	ctx := context.Background()

	stream, err := sut.ChatStream(ctx, []assistant.Message{
		{Role: assistant.RoleUser, Content: "Say something"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result string
	for msg := range stream {
		result += msg
	}

	expected := "Hello world"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}
