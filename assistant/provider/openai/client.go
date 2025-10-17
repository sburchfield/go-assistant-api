package openai

import (
	"context"
	"encoding/json"

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
	return c.ChatStreamWithTools(ctx, messages, nil, "")
}

func (c *Client) ChatStreamWithTools(ctx context.Context, messages []assistant.Message, tools []assistant.Tool, toolChoice assistant.ToolChoice) (<-chan string, error) {
	input := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		msg := openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}

		// Handle tool calls
		if len(m.ToolCalls) > 0 {
			msg.ToolCalls = make([]openai.ToolCall, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				msg.ToolCalls[j] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolType(tc.Type),
					Function: openai.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}

		// Handle tool responses
		if m.ToolCallID != "" {
			msg.ToolCallID = m.ToolCallID
		}

		input[i] = msg
	}

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    input,
		Stream:      true,
		Temperature: c.temperature,
	}

	// Add tools if provided
	if len(tools) > 0 {
		req.Tools = make([]openai.Tool, len(tools))
		for i, t := range tools {
			req.Tools[i] = openai.Tool{
				Type: openai.ToolType(t.Type),
				Function: &openai.FunctionDefinition{
					Name:        t.Function.Name,
					Description: t.Function.Description,
					Parameters:  t.Function.Parameters,
				},
			}
		}

		if toolChoice != "" {
			req.ToolChoice = string(toolChoice)
		}
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

			// Handle tool calls
			if len(resp.Choices) > 0 && len(resp.Choices[0].Delta.ToolCalls) > 0 {
				for _, tc := range resp.Choices[0].Delta.ToolCalls {
					toolCallJSON, _ := json.Marshal(map[string]interface{}{
						"type": "tool_call",
						"id":   tc.ID,
						"function": map[string]string{
							"name":      tc.Function.Name,
							"arguments": tc.Function.Arguments,
						},
					})
					out <- string(toolCallJSON)
				}
				continue
			}

			// Handle regular content
			if len(resp.Choices) > 0 {
				out <- resp.Choices[0].Delta.Content
			}
		}
	}()

	return out, nil
}
