package assistant

import (
	"fmt"
	"net/http"
)

func ToSSE(w http.ResponseWriter, stream <-chan string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	for msg := range stream {
		fmt.Fprintf(w, "data: %s\n\n", msg)
		flusher.Flush()
	}
}
