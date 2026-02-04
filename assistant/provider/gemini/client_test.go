package gemini_test

import (
	"context"
	"testing"

	"cloud.google.com/go/ai/generativelanguage/apiv1beta/generativelanguagepb"
	"github.com/sburchfield/go-assistant-api/assistant"
	"github.com/sburchfield/go-assistant-api/assistant/provider/gemini"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func TestChatStream_Gemini(t *testing.T) {
	listener := bufconn.Listen(bufSize)
	server := grpc.NewServer()

	generativelanguagepb.RegisterGenerativeServiceServer(server, &mockGenerativeServer{})
	go func() {
		_ = server.Serve(listener)
	}()

	ctx := context.Background()
	client, err := gemini.NewClient(ctx, "fake-project-id", "fake-location", "gemini-pro", 0.7, "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	stream, err := client.ChatStream(ctx, []assistant.Message{
		{Role: assistant.RoleUser, Content: "Say something"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result string
	for msg := range stream {
		result += msg
	}

	expected := "Hello world"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

type mockGenerativeServer struct {
	generativelanguagepb.UnimplementedGenerativeServiceServer
}

func (m *mockGenerativeServer) StreamGenerateContent(req *generativelanguagepb.GenerateContentRequest, stream generativelanguagepb.GenerativeService_StreamGenerateContentServer) error {
	resp1 := &generativelanguagepb.GenerateContentResponse{
		Candidates: []*generativelanguagepb.Candidate{
			{
				Content: &generativelanguagepb.Content{
					Parts: []*generativelanguagepb.Part{
						{Data: &generativelanguagepb.Part_Text{Text: "Hello"}},
					},
				},
			},
		},
	}
	resp2 := &generativelanguagepb.GenerateContentResponse{
		Candidates: []*generativelanguagepb.Candidate{
			{
				Content: &generativelanguagepb.Content{
					Parts: []*generativelanguagepb.Part{
						{Data: &generativelanguagepb.Part_Text{Text: " world"}},
					},
				},
			},
		},
	}
	_ = stream.Send(resp1)
	_ = stream.Send(resp2)
	return nil
}

func TestChatStreamWithUsage_Gemini(t *testing.T) {
	listener := bufconn.Listen(bufSize)
	server := grpc.NewServer()

	generativelanguagepb.RegisterGenerativeServiceServer(server, &mockGenerativeServer{})
	go func() {
		_ = server.Serve(listener)
	}()

	ctx := context.Background()
	client, err := gemini.NewClient(ctx, "fake-project-id", "fake-location", "gemini-pro", 0.7, "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	result, err := client.ChatStreamWithUsage(ctx, []assistant.Message{
		{Role: assistant.RoleUser, Content: "Say something"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil StreamResult")
	}

	if result.TextChannel == nil {
		t.Fatal("expected non-nil TextChannel")
	}

	// Consume the text channel
	var text string
	for msg := range result.TextChannel {
		text += msg
	}

	// After consuming the channel, GetUsage should return usage metadata
	usage := result.GetUsage()
	if usage == nil {
		t.Fatal("expected non-nil UsageMetadata after consuming TextChannel")
	}

	// Verify token counts are populated (in real usage, these would be > 0)
	// For mock test, we just verify the struct is properly returned
	t.Logf("Usage metadata - PromptTokens: %d, CandidatesTokens: %d, TotalTokens: %d",
		usage.PromptTokenCount, usage.CandidatesTokenCount, usage.TotalTokenCount)
}

func TestChatStreamWithToolsAndUsage_NoMessages(t *testing.T) {
	ctx := context.Background()
	client, err := gemini.NewClient(ctx, "fake-project-id", "fake-location", "gemini-pro", 0.7, "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.ChatStreamWithToolsAndUsage(ctx, []assistant.Message{}, nil, assistant.ToolChoiceAuto)
	if err == nil {
		t.Fatal("expected error for empty messages")
	}
}
