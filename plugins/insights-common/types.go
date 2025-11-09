package common

// Provider represents the AI provider type
type Provider string

const (
	ProviderAnthropic Provider = "anthropic"
	ProviderOpenAI    Provider = "openai"
	ProviderOllama    Provider = "ollama"
)

// SafeEvent represents a sanitized event for AI analysis
type SafeEvent struct {
	Type      string `json:"type"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Timestamp string `json:"timestamp"`
}

// PromptResult contains the prompt(s) for AI providers
// SystemPrompt is used by providers that support system prompts (Anthropic)
// UserMessage contains the user message or combined prompt
type PromptResult struct {
	SystemPrompt string
	UserMessage  string
}
