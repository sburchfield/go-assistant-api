package provider

import (
	"context"

	"github.com/sburchfield/go-assistant-api/assistant"
)

type ChatProvider interface {
	ChatStream(ctx context.Context, messages []assistant.Message) (<-chan string, error)
}
