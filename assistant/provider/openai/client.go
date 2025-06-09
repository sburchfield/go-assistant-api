package openai

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sburchfield/go-assistant-api/assistant"
)

// ChatStream defines the interface for streaming OpenAI chat completions.
type ChatStream interface {
	Recv() (openai.ChatCompletionStreamResponse, error)
	Close() error
}

// OpenAIClient defines the subset of the OpenAI SDK used by our code.
type OpenAIClient interface {
	CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (ChatStream, error)
}

// sdkWrapper wraps the actual OpenAI client to match our OpenAIClient interface.
type sdkWrapper struct {
	inner *openai.Client
}

func (s *sdkWrapper) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (ChatStream, error) {
	return s.inner.CreateChatCompletionStream(ctx, req)
}

type Client struct {
	sdk         OpenAIClient
	model       string
	temperature float32
}

func NewClientWithSDK(sdk OpenAIClient, model string, temperature float32) *Client {
	return &Client{
		sdk:         sdk,
		model:       model,
		temperature: temperature,
	}
}

func NewClient(apiKey string, model string, temperature float32) *Client {
	return NewClientWithSDK(&sdkWrapper{inner: openai.NewClient(apiKey)}, model, temperature)
}

func (c *Client) ChatStream(ctx context.Context, messages []assistant.Message) (<-chan string, error) {
	input := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		input[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    input,
		Stream:      true,
		Temperature: c.temperature,
	}

	stream, err := c.sdk.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	out := make(chan string)
	go func() {
		defer close(out)
		defer stream.Close()
		for {
			resp, err := stream.Recv()
			if err != nil {
				break
			}
			out <- resp.Choices[0].Delta.Content
		}
	}()

	return out, nil
}
