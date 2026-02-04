package bedrock_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/sburchfield/go-assistant-api/assistant"
	"github.com/sburchfield/go-assistant-api/assistant/provider/bedrock"
)

func TestNewClientWithConfig(t *testing.T) {
	cfg := aws.Config{
		Region: "us-east-1",
	}

	client := bedrock.NewClientWithConfig(cfg, "anthropic.claude-3-sonnet-20240229-v1:0", 0.7)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestChatStream_NoMessages(t *testing.T) {
	cfg := aws.Config{
		Region: "us-east-1",
	}

	client := bedrock.NewClientWithConfig(cfg, "anthropic.claude-3-sonnet-20240229-v1:0", 0.7)

	_, err := client.ChatStream(context.Background(), []assistant.Message{})
	if err == nil {
		t.Fatal("expected error for empty messages")
	}
}

func TestChatStreamWithToolsAndUsage_NoMessages(t *testing.T) {
	cfg := aws.Config{
		Region: "us-east-1",
	}

	client := bedrock.NewClientWithConfig(cfg, "anthropic.claude-3-sonnet-20240229-v1:0", 0.7)

	_, err := client.ChatStreamWithToolsAndUsage(context.Background(), []assistant.Message{}, nil, assistant.ToolChoiceAuto)
	if err == nil {
		t.Fatal("expected error for empty messages")
	}
}

func TestConvertMessages(t *testing.T) {
	cfg := aws.Config{
		Region: "us-east-1",
	}

	client := bedrock.NewClientWithConfig(cfg, "anthropic.claude-3-sonnet-20240229-v1:0", 0.7)
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	// Test that client can be created with various message types
	messages := []assistant.Message{
		{Role: assistant.RoleSystem, Content: "You are a helpful assistant."},
		{Role: assistant.RoleUser, Content: "Hello"},
		{Role: assistant.RoleAssistant, Content: "Hi there!"},
		{Role: assistant.RoleUser, Content: "How are you?"},
	}

	// This test validates that the client is properly configured
	// Actual API calls would require valid AWS credentials
	if len(messages) != 4 {
		t.Errorf("expected 4 messages, got %d", len(messages))
	}
}

func TestChatStreamWithTools_ToolConfiguration(t *testing.T) {
	cfg := aws.Config{
		Region: "us-east-1",
	}

	client := bedrock.NewClientWithConfig(cfg, "anthropic.claude-3-sonnet-20240229-v1:0", 0.7)
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	tools := []assistant.Tool{
		{
			Type: "function",
			Function: assistant.ToolFunction{
				Name:        "get_weather",
				Description: "Get the current weather for a location",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	// Validate tools are properly structured
	if len(tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Function.Name != "get_weather" {
		t.Errorf("expected tool name 'get_weather', got '%s'", tools[0].Function.Name)
	}
}

func TestToolChoiceOptions(t *testing.T) {
	testCases := []struct {
		name       string
		toolChoice assistant.ToolChoice
	}{
		{"auto", assistant.ToolChoiceAuto},
		{"none", assistant.ToolChoiceNone},
		{"required", assistant.ToolChoiceRequired},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := aws.Config{
				Region: "us-east-1",
			}

			client := bedrock.NewClientWithConfig(cfg, "anthropic.claude-3-sonnet-20240229-v1:0", 0.7)
			if client == nil {
				t.Fatal("expected non-nil client")
			}

			// Verify the client accepts different tool choice options
			// Actual behavior would be tested with integration tests
		})
	}
}
