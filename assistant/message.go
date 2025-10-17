package assistant

// Message represents a single chat message with a role and content.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall represents a request from the assistant to call a tool
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
}

// FunctionCall contains the function name and arguments
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// Predefined roles for clarity and consistency.
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
	RoleTool      = "tool"
)

// IsValidRole checks whether the given role is one of the expected roles.
func IsValidRole(role string) bool {
	switch role {
	case RoleUser, RoleAssistant, RoleSystem, RoleTool:
		return true
	default:
		return false
	}
}
