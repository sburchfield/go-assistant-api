package assistant

// Message represents a single chat message with a role and content.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Predefined roles for clarity and consistency.
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

// IsValidRole checks whether the given role is one of the expected roles.
func IsValidRole(role string) bool {
	switch role {
	case RoleUser, RoleAssistant, RoleSystem:
		return true
	default:
		return false
	}
}
