package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/xid"
)

// StreamResult contains both the text channel and a way to get usage metadata after streaming completes
type StreamResult struct {
	TextChannel <-chan string
	// GetUsage returns the usage metadata. Must be called after TextChannel is closed.
	// Returns nil if usage data is not available.
	GetUsage func() *UsageMetadata
}

func ToSSE(ctx context.Context, w http.ResponseWriter, stream <-chan string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	flusher.Flush()

	messageID := fmt.Sprintf("msg-%s", xid.New().String())
	fmt.Fprintf(w, "f:{\"messageId\":\"%s\"}\n\n", messageID)
	flusher.Flush()

	keepAliveTicker := time.NewTicker(30 * time.Second)
	defer keepAliveTicker.Stop()

	for {
		select {
		case msg, ok := <-stream:
			if !ok {
				// Emit final usage and finishReason metadata
				d := `d:{"finishReason":"stop","usage":{"promptTokens":0,"completionTokens":0}}`
				e := `e:{"finishReason":"stop","usage":{"promptTokens":0,"completionTokens":0},"isContinued":false}`
				fmt.Fprintf(w, "%s\n\n%s\n\n", d, e)
				flusher.Flush()
				return
			}

			if msg == "" {
				continue
			}

			escaped, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			fmt.Fprintf(w, "0:%s\n\n", escaped)
			flusher.Flush()

		case <-keepAliveTicker.C:
			fmt.Fprintf(w, ":keepalive\n\n")
			flusher.Flush()

		case <-ctx.Done():
			return
		}
	}
}
