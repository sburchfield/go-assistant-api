package gemini

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/sburchfield/go-assistant-api/assistant"
	"google.golang.org/api/option"

	genai "cloud.google.com/go/ai/generativelanguage/apiv1beta"
	genaipb "cloud.google.com/go/ai/generativelanguage/apiv1beta/generativelanguagepb"
)

type Client struct {
	model   *genai.GenerativeClient
	modelID string
}

func NewClient(ctx context.Context, apiKey, modelID string) (*Client, error) {
	generativeClient, err := genai.NewGenerativeClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create generative service client: %w", err)
	}

	return &Client{model: generativeClient, modelID: modelID}, nil
}

func (c *Client) ChatStream(
	ctx context.Context,
	messages []assistant.Message,
) (<-chan string, error) {
	if len(messages) == 0 {
		return nil, errors.New("ChatStream: no messages provided")
	}

	var sdkContents []*genaipb.Content
	for _, msg := range messages {
		sdkRole := "user"
		if msg.Role == "assistant" || msg.Role == "model" {
			sdkRole = "model"
		}
		sdkContents = append(sdkContents, &genaipb.Content{
			Role: sdkRole,
			Parts: []*genaipb.Part{
				{Data: &genaipb.Part_Text{Text: msg.Content}},
			},
		})
	}

	req := &genaipb.GenerateContentRequest{
		Model:    fmt.Sprintf("models/%s", c.modelID),
		Contents: sdkContents,
	}

	stream, err := c.model.StreamGenerateContent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	out := make(chan string)
	go func() {
		defer close(out)

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("stream recv error: %v", err)
				return
			}

			for _, cand := range resp.Candidates {
				for _, part := range cand.Content.Parts {
					text := part.GetText()
					if text != "" {
						out <- text
					}
				}
			}
		}
	}()
	return out, nil

}
