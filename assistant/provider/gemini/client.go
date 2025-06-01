// assistant/provider/gemini/client.go
package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sburchfield/go-assistant-api/assistant"
)

type Client struct {
	apiKey string
	model  string
	client *http.Client
}

func NewClient(apiKey, model string) *Client {
	return &Client{apiKey: apiKey, model: model, client: http.DefaultClient}
}

// for testing
func NewClientWithHTTP(apiKey, model string, client *http.Client) *Client {
	return &Client{apiKey: apiKey, model: model, client: client}
}

type geminiMessage struct {
	Role    string `json:"role"`
	Content string `json:"text"`
}

type geminiRequest struct {
	Contents []struct {
		Parts []geminiMessage `json:"parts"`
		Role  string          `json:"role"`
	} `json:"contents"`
}

func (c *Client) ChatStream(ctx context.Context, messages []assistant.Message) (<-chan string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?key=%s", c.model, c.apiKey)

	var contents []struct {
		Parts []geminiMessage `json:"parts"`
		Role  string          `json:"role"`
	}
	for _, msg := range messages {
		contents = append(contents, struct {
			Parts []geminiMessage `json:"parts"`
			Role  string          `json:"role"`
		}{
			Parts: []geminiMessage{{Role: msg.Role, Content: msg.Content}},
			Role:  msg.Role,
		})
	}

	body, _ := json.Marshal(geminiRequest{Contents: contents})
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	out := make(chan string)
	go func() {
		defer close(out)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data:") {
				var chunk struct {
					Candidates []struct {
						Content struct {
							Parts []struct {
								Text string `json:"text"`
							} `json:"parts"`
						} `json:"content"`
					} `json:"candidates"`
				}
				if err := json.Unmarshal([]byte(line[5:]), &chunk); err == nil {
					for _, part := range chunk.Candidates {
						for _, p := range part.Content.Parts {
							out <- p.Text
						}
					}
				}
			}
		}
	}()

	return out, nil
}
