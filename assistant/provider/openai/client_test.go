// assistant/provider/openai/client_test.go
package openai_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sdk "github.com/sashabaranov/go-openai"
	"github.com/sburchfield/go-assistant-api/assistant"
	"github.com/sburchfield/go-assistant-api/assistant/provider/openai"
)

type mockStream struct {
	responses []sdk.ChatCompletionStreamResponse
	index     int
	closed    bool
}

func (m *mockStream) Recv() (sdk.ChatCompletionStreamResponse, error) {
	if m.index >= len(m.responses) {
		return sdk.ChatCompletionStreamResponse{}, errors.New("EOF")
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
	stream openai.ChatStream
}

func (m *mockOpenAIClient) CreateChatCompletionStream(ctx context.Context, req sdk.ChatCompletionRequest) (openai.ChatStream, error) {
	return m.stream, nil
}

func TestChatStream(t *testing.T) {
	mockResp := []sdk.ChatCompletionStreamResponse{
		{Choices: []sdk.ChatCompletionStreamChoice{{Delta: sdk.ChatCompletionStreamChoiceDelta{Content: "Hello"}}}},
		{Choices: []sdk.ChatCompletionStreamChoice{{Delta: sdk.ChatCompletionStreamChoiceDelta{Content: " world"}}}},
	}

	mockClient := &mockOpenAIClient{
		stream: &mockStream{responses: mockResp},
	}

	client := openai.NewClientWithSDK(mockClient, "gpt-3.5-turbo", 0.0)
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
