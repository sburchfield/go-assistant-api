// assistant/client_test.go
package assistant_test

import (
	"context"
	"errors"
	"testing"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sburchfield/go-assistant-api/assistant"
)

type mockStream struct {
	responses []openai.ChatCompletionStreamResponse
	index     int
	closed    bool
}

func (m *mockStream) Recv() (openai.ChatCompletionStreamResponse, error) {
	if m.index >= len(m.responses) {
		return openai.ChatCompletionStreamResponse{}, errors.New("EOF")
	}
	resp := m.responses[m.index]
	m.index++
	return resp, nil
}

func (m *mockStream) Close() error {
	m.closed = true
	return nil
}

type mockOpenAIClient struct {
	stream assistant.ChatStream
}

func (m *mockOpenAIClient) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (assistant.ChatStream, error) {
	return m.stream, nil
}

func TestChatStream(t *testing.T) {
	mockResp := []openai.ChatCompletionStreamResponse{
		{Choices: []openai.ChatCompletionStreamChoice{{Delta: openai.ChatCompletionStreamChoiceDelta{Content: "Hello"}}}},
		{Choices: []openai.ChatCompletionStreamChoice{{Delta: openai.ChatCompletionStreamChoiceDelta{Content: " world"}}}},
	}

	mockClient := &mockOpenAIClient{
		stream: &mockStream{responses: mockResp},
	}

	client := assistant.NewClientWithSDK(mockClient, "gpt-3.5-turbo")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	messages := []assistant.Message{
		{Role: assistant.RoleUser, Content: "Say something"},
	}

	stream, err := client.ChatStream(ctx, messages)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var result string
	for msg := range stream {
		result += msg
	}

	expected := "Hello world"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}
