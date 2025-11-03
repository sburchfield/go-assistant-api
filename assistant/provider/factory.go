// assistant/provider/factory.go
package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sburchfield/go-assistant-api/assistant/provider/gemini"
	"github.com/sburchfield/go-assistant-api/assistant/provider/openai"
)

func NewProviderFromEnv() (ChatProvider, error) {
	provider := os.Getenv("LLM_PROVIDER")
	temperature := float32(0) // Default temperature
	switch provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		model := os.Getenv("OPENAI_MODEL")
		temperatureStr := os.Getenv("TEMPERATURE")
		if temperatureStr != "" {
			if t, err := strconv.ParseFloat(temperatureStr, 32); err == nil {
				temperature = float32(t)
			}
		}

		if apiKey == "" || model == "" {
			return nil, fmt.Errorf("missing OPENAI_API_KEY or OPENAI_MODEL")
		}
		return openai.NewClient(apiKey, model, temperature), nil
	case "gemini":
		apiKey := os.Getenv("GEMINI_API_KEY")
		model := os.Getenv("GEMINI_MODEL")
		temperatureStr := os.Getenv("TEMPERATURE")
		if temperatureStr != "" {
			if t, err := strconv.ParseFloat(temperatureStr, 32); err == nil {
				temperature = float32(t)
			}
		}

		if apiKey == "" || model == "" {
			return nil, fmt.Errorf("missing GEMINI_API_KEY or GEMINI_MODEL")
		}
		return gemini.NewClient(context.TODO(), apiKey, "", model, temperature)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
