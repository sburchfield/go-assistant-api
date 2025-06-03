package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/xid"
)

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

	// Emit initial messageId for ai-sdk clients
	startChunk := map[string]any{
		"f": map[string]string{"messageId": fmt.Sprintf("msg-%s", xid.New().String())},
	}
	if data, err := json.Marshal(startChunk); err == nil {
		fmt.Fprintf(w, "%s\n\n", data)
		flusher.Flush()
	}

	keepAliveTicker := time.NewTicker(30 * time.Second)
	defer keepAliveTicker.Stop()

	for {
		select {
		case msg, ok := <-stream:
			if !ok {
				// Final end-of-stream message
				endChunk := map[string]any{
					"e": map[string]string{},
					"d": map[string]string{},
				}
				if data, err := json.Marshal(endChunk); err == nil {
					fmt.Fprintf(w, "%s\n\n", data)
					flusher.Flush()
				}
				return
			}

			if msg == "" {
				continue
			}

			chunk := map[string]string{"0": msg}
			if data, err := json.Marshal(chunk); err == nil {
				if _, err := fmt.Fprintf(w, "%s\n\n", data); err != nil {
					http.Error(w, "Error writing to stream", http.StatusInternalServerError)
					return
				}
				flusher.Flush()
			}

		case <-keepAliveTicker.C:
			if _, err := fmt.Fprintf(w, ":keepalive\n\n"); err != nil {
				http.Error(w, "Error writing keepalive", http.StatusInternalServerError)
				return
			}
			flusher.Flush()

		case <-ctx.Done():
			return
		}
	}
}
