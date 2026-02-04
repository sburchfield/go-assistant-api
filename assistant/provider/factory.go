// assistant/provider/factory.go
package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/sburchfield/go-assistant-api/assistant/provider/bedrock"
	"github.com/sburchfield/go-assistant-api/assistant/provider/gemini"
	"github.com/sburchfield/go-assistant-api/assistant/provider/openai"
)

// getSecretFromAWS retrieves a secret from AWS Secrets Manager
func getSecretFromAWS(ctx context.Context, secretName string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	if result.SecretString != nil {
		return *result.SecretString, nil
	}

	return "", fmt.Errorf("secret is not a string")
}

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
		ctx := context.Background()
		projectID := os.Getenv("GEMINI_PROJECT_ID")
		location := os.Getenv("GEMINI_LOCATION")
		model := os.Getenv("GEMINI_MODEL")
		temperatureStr := os.Getenv("TEMPERATURE")
		if temperatureStr != "" {
			if t, err := strconv.ParseFloat(temperatureStr, 32); err == nil {
				temperature = float32(t)
			}
		}

		if projectID == "" || location == "" || model == "" {
			return nil, fmt.Errorf("missing GEMINI_PROJECT_ID, GEMINI_LOCATION, or GEMINI_MODEL")
		}

		// Check if we should use AWS Secrets Manager for credentials
		secretName := os.Getenv("GEMINI_SECRET_NAME")
		var credentialsJSON string

		if secretName != "" {
			// Fetch from AWS Secrets Manager
			secret, err := getSecretFromAWS(ctx, secretName)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch GCP credentials from AWS Secrets Manager: %w", err)
			}
			credentialsJSON = secret
		}

		return gemini.NewClient(ctx, projectID, location, model, temperature, credentialsJSON)
	case "bedrock":
		ctx := context.Background()
		region := os.Getenv("AWS_REGION")
		model := os.Getenv("BEDROCK_MODEL")
		temperatureStr := os.Getenv("TEMPERATURE")
		if temperatureStr != "" {
			if t, err := strconv.ParseFloat(temperatureStr, 32); err == nil {
				temperature = float32(t)
			}
		}

		if region == "" {
			region = "us-east-1" // Default region
		}
		if model == "" {
			return nil, fmt.Errorf("missing BEDROCK_MODEL")
		}

		return bedrock.NewClient(ctx, region, model, temperature)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
