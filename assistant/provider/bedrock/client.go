package bedrock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/sburchfield/go-assistant-api/assistant"
)

// Client wraps the AWS Bedrock Runtime client for chat completions.
type Client struct {
	client      *bedrockruntime.Client
	modelID     string
	temperature float32
}

// NewClient creates a new Bedrock client with the given configuration.
func NewClient(ctx context.Context, region, modelID string, temperature float32) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	return &Client{
		client:      client,
		modelID:     modelID,
		temperature: temperature,
	}, nil
}

// NewClientWithConfig creates a new Bedrock client with a pre-configured AWS config.
func NewClientWithConfig(cfg aws.Config, modelID string, temperature float32) *Client {
	return &Client{
		client:      bedrockruntime.NewFromConfig(cfg),
		modelID:     modelID,
		temperature: temperature,
	}
}

// ChatStream streams chat completions without tools.
func (c *Client) ChatStream(ctx context.Context, messages []assistant.Message) (<-chan string, error) {
	return c.ChatStreamWithTools(ctx, messages, nil, assistant.ToolChoiceAuto)
}

// ChatStreamWithTools streams chat completions with optional tool support.
func (c *Client) ChatStreamWithTools(
	ctx context.Context,
	messages []assistant.Message,
	tools []assistant.Tool,
	toolChoice assistant.ToolChoice,
) (<-chan string, error) {
	if len(messages) == 0 {
		return nil, errors.New("ChatStream: no messages provided")
	}

	converseMessages, systemPrompts := c.convertMessages(messages)

	input := &bedrockruntime.ConverseStreamInput{
		ModelId:  aws.String(c.modelID),
		Messages: converseMessages,
		InferenceConfig: &types.InferenceConfiguration{
			Temperature: aws.Float32(c.temperature),
		},
	}

	// Add system prompts if present
	if len(systemPrompts) > 0 {
		input.System = systemPrompts
	}

	// Add tools if provided
	if len(tools) > 0 {
		input.ToolConfig = c.convertToolConfig(tools, toolChoice)
	}

	output, err := c.client.ConverseStream(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to start converse stream: %w", err)
	}

	out := make(chan string)
	go func() {
		defer close(out)
		c.processStream(output, out)
	}()

	return out, nil
}

// ChatStreamWithUsage streams chat completions and provides usage metadata.
func (c *Client) ChatStreamWithUsage(ctx context.Context, messages []assistant.Message) (*assistant.StreamResult, error) {
	return c.ChatStreamWithToolsAndUsage(ctx, messages, nil, assistant.ToolChoiceAuto)
}

// ChatStreamWithToolsAndUsage streams chat completions with tools and provides usage metadata.
func (c *Client) ChatStreamWithToolsAndUsage(
	ctx context.Context,
	messages []assistant.Message,
	tools []assistant.Tool,
	toolChoice assistant.ToolChoice,
) (*assistant.StreamResult, error) {
	if len(messages) == 0 {
		return nil, errors.New("ChatStream: no messages provided")
	}

	converseMessages, systemPrompts := c.convertMessages(messages)

	input := &bedrockruntime.ConverseStreamInput{
		ModelId:  aws.String(c.modelID),
		Messages: converseMessages,
		InferenceConfig: &types.InferenceConfiguration{
			Temperature: aws.Float32(c.temperature),
		},
	}

	// Add system prompts if present
	if len(systemPrompts) > 0 {
		input.System = systemPrompts
	}

	// Add tools if provided
	if len(tools) > 0 {
		input.ToolConfig = c.convertToolConfig(tools, toolChoice)
	}

	output, err := c.client.ConverseStream(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to start converse stream: %w", err)
	}

	out := make(chan string)
	var usageMetadata *assistant.UsageMetadata
	var usageMu sync.Mutex

	go func() {
		defer close(out)
		c.processStreamWithUsage(output, out, &usageMetadata, &usageMu)
	}()

	return &assistant.StreamResult{
		TextChannel: out,
		GetUsage: func() *assistant.UsageMetadata {
			usageMu.Lock()
			defer usageMu.Unlock()
			return usageMetadata
		},
	}, nil
}

// convertMessages converts assistant.Message to Bedrock Converse format.
// Returns the conversation messages and any system prompts separately.
func (c *Client) convertMessages(messages []assistant.Message) ([]types.Message, []types.SystemContentBlock) {
	var converseMessages []types.Message
	var systemPrompts []types.SystemContentBlock

	for _, msg := range messages {
		switch msg.Role {
		case assistant.RoleSystem:
			systemPrompts = append(systemPrompts, &types.SystemContentBlockMemberText{
				Value: msg.Content,
			})
		case assistant.RoleUser:
			converseMessages = append(converseMessages, types.Message{
				Role:    types.ConversationRoleUser,
				Content: []types.ContentBlock{&types.ContentBlockMemberText{Value: msg.Content}},
			})
		case assistant.RoleAssistant:
			content := []types.ContentBlock{}
			if msg.Content != "" {
				content = append(content, &types.ContentBlockMemberText{Value: msg.Content})
			}
			// Handle tool use in assistant messages
			for _, tc := range msg.ToolCalls {
				var inputDoc map[string]interface{}
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &inputDoc); err != nil {
					inputDoc = map[string]interface{}{}
				}
				content = append(content, &types.ContentBlockMemberToolUse{
					Value: types.ToolUseBlock{
						ToolUseId: aws.String(tc.ID),
						Name:      aws.String(tc.Function.Name),
						Input:     document.NewLazyDocument(inputDoc),
					},
				})
			}
			if len(content) > 0 {
				converseMessages = append(converseMessages, types.Message{
					Role:    types.ConversationRoleAssistant,
					Content: content,
				})
			}
		case assistant.RoleTool:
			// Tool results go as user messages with tool result content
			converseMessages = append(converseMessages, types.Message{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberToolResult{
						Value: types.ToolResultBlock{
							ToolUseId: aws.String(msg.ToolCallID),
							Content: []types.ToolResultContentBlock{
								&types.ToolResultContentBlockMemberText{Value: msg.Content},
							},
						},
					},
				},
			})
		}
	}

	return converseMessages, systemPrompts
}

// convertToolConfig converts assistant tools to Bedrock tool configuration.
func (c *Client) convertToolConfig(tools []assistant.Tool, toolChoice assistant.ToolChoice) *types.ToolConfiguration {
	bedrockTools := make([]types.Tool, len(tools))
	for i, t := range tools {
		bedrockTools[i] = &types.ToolMemberToolSpec{
			Value: types.ToolSpecification{
				Name:        aws.String(t.Function.Name),
				Description: aws.String(t.Function.Description),
				InputSchema: &types.ToolInputSchemaMemberJson{
					Value: document.NewLazyDocument(t.Function.Parameters),
				},
			},
		}
	}

	config := &types.ToolConfiguration{
		Tools: bedrockTools,
	}

	// Set tool choice
	switch toolChoice {
	case assistant.ToolChoiceAuto:
		config.ToolChoice = &types.ToolChoiceMemberAuto{Value: types.AutoToolChoice{}}
	case assistant.ToolChoiceRequired:
		config.ToolChoice = &types.ToolChoiceMemberAny{Value: types.AnyToolChoice{}}
	case assistant.ToolChoiceNone:
		// Don't set ToolChoice, let the model decide
	}

	return config
}

// processStream handles the streaming response and sends text to the output channel.
func (c *Client) processStream(output *bedrockruntime.ConverseStreamOutput, out chan<- string) {
	stream := output.GetStream()
	defer stream.Close()

	for event := range stream.Events() {
		switch v := event.(type) {
		case *types.ConverseStreamOutputMemberContentBlockDelta:
			if delta, ok := v.Value.Delta.(*types.ContentBlockDeltaMemberText); ok {
				out <- delta.Value
			}
			// Handle tool use deltas
			if toolDelta, ok := v.Value.Delta.(*types.ContentBlockDeltaMemberToolUse); ok {
				// Send partial tool use input as it streams
				if toolDelta.Value.Input != nil {
					out <- *toolDelta.Value.Input
				}
			}
		case *types.ConverseStreamOutputMemberContentBlockStart:
			// Handle tool use start
			if toolStart, ok := v.Value.Start.(*types.ContentBlockStartMemberToolUse); ok {
				toolCallJSON, _ := json.Marshal(map[string]interface{}{
					"type": "tool_call_start",
					"id":   aws.ToString(toolStart.Value.ToolUseId),
					"function": map[string]string{
						"name": aws.ToString(toolStart.Value.Name),
					},
				})
				out <- string(toolCallJSON)
			}
		case *types.ConverseStreamOutputMemberMessageStop:
			// Message complete
		}
	}

	if err := stream.Err(); err != nil {
		log.Printf("stream error: %v", err)
	}
}

// processStreamWithUsage handles streaming and captures usage metadata.
func (c *Client) processStreamWithUsage(
	output *bedrockruntime.ConverseStreamOutput,
	out chan<- string,
	usageMetadata **assistant.UsageMetadata,
	usageMu *sync.Mutex,
) {
	stream := output.GetStream()
	defer stream.Close()

	for event := range stream.Events() {
		switch v := event.(type) {
		case *types.ConverseStreamOutputMemberContentBlockDelta:
			if delta, ok := v.Value.Delta.(*types.ContentBlockDeltaMemberText); ok {
				out <- delta.Value
			}
			if toolDelta, ok := v.Value.Delta.(*types.ContentBlockDeltaMemberToolUse); ok {
				if toolDelta.Value.Input != nil {
					out <- *toolDelta.Value.Input
				}
			}
		case *types.ConverseStreamOutputMemberContentBlockStart:
			if toolStart, ok := v.Value.Start.(*types.ContentBlockStartMemberToolUse); ok {
				toolCallJSON, _ := json.Marshal(map[string]interface{}{
					"type": "tool_call_start",
					"id":   aws.ToString(toolStart.Value.ToolUseId),
					"function": map[string]string{
						"name": aws.ToString(toolStart.Value.Name),
					},
				})
				out <- string(toolCallJSON)
			}
		case *types.ConverseStreamOutputMemberMetadata:
			// Capture usage metadata
			if v.Value.Usage != nil {
				usageMu.Lock()
				*usageMetadata = &assistant.UsageMetadata{
					PromptTokenCount:     aws.ToInt32(v.Value.Usage.InputTokens),
					CandidatesTokenCount: aws.ToInt32(v.Value.Usage.OutputTokens),
					TotalTokenCount:      aws.ToInt32(v.Value.Usage.TotalTokens),
				}
				usageMu.Unlock()
			}
		case *types.ConverseStreamOutputMemberMessageStop:
			// Message complete
		}
	}

	if err := stream.Err(); err != nil {
		log.Printf("stream error: %v", err)
	}
}
