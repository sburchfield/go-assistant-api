// examples/main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/sburchfield/go-assistant-api/assistant"
	"github.com/sburchfield/go-assistant-api/assistant/provider"
)

type chatRequest struct {
	Messages []assistant.Message `json:"messages"`
}

func main() {
	providerClient, err := provider.NewProviderFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize provider: %v", err)
	}

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		stream, err := providerClient.ChatStream(r.Context(), req.Messages)
		if err != nil {
			http.Error(w, "Failed to stream from provider", http.StatusInternalServerError)
			log.Println("stream error:", err)
			return
		}

		assistant.ToSSE(context.TODO(), w, stream)
	})

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
