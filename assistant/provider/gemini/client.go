package gemini

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/sburchfield/go-assistant-api/assistant"
	"google.golang.org/genai"
)

type Client struct {
	client      *genai.Client
	modelID     string
	temperature float32
}

func NewClient(ctx context.Context, projectID, location, modelID string, temperature float32, credentialsJSON string) (*Client, error) {
	// If credentials JSON is provided (from AWS Secrets Manager), write to temp file
	// and set GOOGLE_APPLICATION_CREDENTIALS env var
	var cleanupFunc func()

	if credentialsJSON != "" {
		tmpFile, err := os.CreateTemp("", "gcp-credentials-*.json")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file for credentials: %w", err)
		}

		if _, err := tmpFile.WriteString(credentialsJSON); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return nil, fmt.Errorf("failed to write credentials to temp file: %w", err)
		}
		tmpFile.Close()

		// Set environment variable temporarily
		oldCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpFile.Name())

		// Setup cleanup function
		cleanupFunc = func() {
			os.Remove(tmpFile.Name())
			if oldCreds != "" {
				os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", oldCreds)
			} else {
				os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
			}
		}
	}

	// Build client config
	clientConfig := &genai.ClientConfig{
		Project:  projectID,
		Location: location,
		Backend:  genai.BackendVertexAI,
	}

	// Create the client (will use GOOGLE_APPLICATION_CREDENTIALS if set)
	client, err := genai.NewClient(ctx, clientConfig)

	// Clean up credentials file if we created one
	if cleanupFunc != nil {
		cleanupFunc()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create vertex ai client: %w", err)
	}

	return &Client{
		client:      client,
		modelID:     modelID,
		temperature: temperature,
	}, nil
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

	var contents []*genai.Content
	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" || msg.Role == "model" {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{{Text: msg.Content}},
		})
	}

	out := make(chan string)
	go func() {
		defer close(out)

		// Create generation config
		config := &genai.GenerateContentConfig{
			Temperature: &c.temperature,
		}

		// Generate content (non-streaming for now)
		resp, err := c.client.Models.GenerateContent(ctx, c.modelID, contents, config)
		if err != nil {
			log.Printf("failed to generate content: %v", err)
			return
		}

		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					if part.Text != "" {
						out <- part.Text
					}
				}
			}
		}
	}()

	return out, nil
}
