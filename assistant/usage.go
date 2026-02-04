package assistant

// UsageMetadata contains token usage information from AI model responses
type UsageMetadata struct {
	PromptTokenCount     int32 `json:"prompt_token_count"`     // Input tokens
	CandidatesTokenCount int32 `json:"candidates_token_count"` // Output tokens
	TotalTokenCount      int32 `json:"total_token_count"`      // Total tokens
}
