// examples/main.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/sburchfield/go-assistant-api/assistant"
)

type chatRequest struct {
	Messages []assistant.Message `json:"messages"`
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing OPENAI_API_KEY environment variable")
	}

	client := assistant.NewClient(apiKey, "gpt-3.5-turbo")

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

		stream, err := client.ChatStream(r.Context(), req.Messages)
		if err != nil {
			http.Error(w, "Failed to stream from OpenAI", http.StatusInternalServerError)
			log.Println("stream error:", err)
			return
		}

		assistant.ToSSE(w, stream)
	})

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
