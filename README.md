# go-assistant-api

A clean, modular Go package for streaming OpenAI assistant-style chat responses with Server-Sent Events (SSE). Inspired by [Assistant UI](https://www.assistant-ui.com), built with clean architecture principles in mind.

---

## ✨ Features

- 🔁 Chat message struct & role helpers
- 📡 Streaming OpenAI completions via channels
- 🌐 SSE writer for browser/server compatibility
- 🧪 Fully testable with mock clients
- 🧩 Easy to integrate into any Go server (`net/http`, `gin`, `chi`, etc.)

---

## 🚀 Getting Started

### 1. Install

```bash
go get github.com/sburchfield/go-assistant-api
```

### 2. Example Usage

```go
client := assistant.NewClient(os.Getenv("OPENAI_API_KEY"), "gpt-3.5-turbo")
stream, err := client.ChatStream(ctx, []assistant.Message{
	{Role: assistant.RoleUser, Content: "What's the capital of France?"},
})

if err != nil {
	log.Fatal(err)
}

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
assistant/          # Core functionality
  ├── client.go     # OpenAI streaming wrapper
  ├── message.go    # Message roles and struct
  ├── stream.go     # SSE formatter
examples/           # Example HTTP server
```

---

## 🧠 Inspiration

This project was inspired by [Assistant UI](https://www.assistant-ui.com) but refactored for Go backends, fast performance, and clean API boundaries.

---

## 📄 License

MIT
