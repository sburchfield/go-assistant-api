# go-assistant-api

A clean, modular Go package for streaming OpenAI, Gemini, and AWS Bedrock assistant-style chat responses with Server-Sent Events (SSE). Inspired by [Assistant UI](https://www.assistant-ui.com), built with clean architecture principles in mind.

---

## ✨ Features

- 🔁 Chat message struct & role helpers
- 📡 Streaming OpenAI, Gemini, and AWS Bedrock completions via channels
- 🛠️ Tool/function calling support across providers
- 📊 Token usage metadata tracking
- 🌐 SSE writer for browser/server compatibility
- 🧪 Fully testable with mock clients
- 🧩 Easy to integrate into any Go server (`net/http`, `gin`, `chi`, etc.)

---

## 🚀 Getting Started

### 1. Install

```bash
go get github.com/sburchfield/go-assistant-api
```

### 2. Choose Your Provider

Set environment variables before running:

#### For OpenAI:
```bash
export LLM_PROVIDER=openai
export OPENAI_API_KEY=your-openai-key
export OPENAI_MODEL=gpt-3.5-turbo
```

#### For Gemini:
```bash
export LLM_PROVIDER=gemini
export GEMINI_API_KEY=your-gemini-key
export GEMINI_MODEL=gemini-pro
```

#### For AWS Bedrock:
```bash
export LLM_PROVIDER=bedrock
export AWS_REGION=us-east-1  # Optional, defaults to us-east-1
export BEDROCK_MODEL=anthropic.claude-3-sonnet-20240229-v1:0
export TEMPERATURE=0.7  # Optional
```

> **Note:** AWS Bedrock uses your default AWS credentials. Ensure you have configured `aws configure` or set `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.

### 3. Example Usage

```go
providerClient, _ := provider.NewProviderFromEnv()
stream, err := providerClient.ChatStream(ctx, []assistant.Message{
	{Role: assistant.RoleUser, Content: "What's the capital of France?"},
})

assistant.ToSSE(w, stream)
```

---

## 💬 Message Format

```go
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"` // message text
}
```

---

## 🧪 Running Tests

```bash
go test ./...
```

---

## 📁 Project Structure

```
assistant/                  # Core functionality
  ├── message.go            # Message roles and struct
  ├── stream.go             # SSE formatter & StreamResult
  ├── tool.go               # Tool/function definitions
  ├── usage.go              # Token usage metadata
  └── provider/             # Multi-provider LLM support
      ├── openai/           # OpenAI implementation
      ├── gemini/           # Gemini implementation
      ├── bedrock/          # AWS Bedrock implementation
      └── factory.go        # Provider selector (env-based)
examples/                   # Example HTTP server
```

---

## 🧠 Inspiration

This project was inspired by [Assistant UI](https://www.assistant-ui.com) but refactored for Go backends, fast performance, and clean API boundaries.

---

## 📄 License

MIT
