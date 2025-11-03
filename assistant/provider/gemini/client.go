package gemini

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"cloud.google.com/go/vertexai/genai"
	"github.com/sburchfield/go-assistant-api/assistant"
)

type Client struct {
	model       *genai.GenerativeModel
	temperature float32
}

func NewClient(ctx context.Context, projectID, location, modelID string, temperature float32) (*Client, error) {
	vertexAIClient, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create vertex ai client: %w", err)
	}

	model := vertexAIClient.GenerativeModel(modelID)
	model.Temperature = &temperature

	return &Client{model: model, temperature: temperature}, nil
}

func (c *Client) ChatStream(ctx context.Context, messages []assistant.Message) (<-chan string, error) {
	return c.ChatStreamWithTools(ctx, messages, nil, assistant.ToolChoiceAuto)
}

func (c *Client) ChatStreamWithTools(
	ctx context.Context,
	messages []assistant.Message,
	tools []assistant.Tool,
	toolChoice assistant.ToolChoice,
) (<-chan string, error) {
	if len(messages) == 0 {
		return nil, errors.New("ChatStream: no messages provided")
	}

	var history []*genai.Content
	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" || msg.Role == "model" {
			role = "model"
		}
		history = append(history, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		})
	}

	session := c.model.StartChat()
	session.History = history[:len(history)-1]

	// Get the last message to send
	lastMessage := history[len(history)-1]

	out := make(chan string)
	go func() {
		defer close(out)

		stream := session.SendMessageStream(ctx, lastMessage.Parts...)
		for {
			resp, err := stream.Next()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("stream recv error: %v", err)
				return
			}

			for _, cand := range resp.Candidates {
				for _, part := range cand.Content.Parts {
					if txt, ok := part.(genai.Text); ok {
						out <- string(txt)
					}
				}
			}
		}
	}()

	return out, nil
}
