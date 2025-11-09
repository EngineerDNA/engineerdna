package common

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

// SanitizeInput prevents prompt injection by escaping special characters
// and removing potential instruction-like patterns
func SanitizeInput(input string) string {
	// Convert to string if needed
	s := fmt.Sprintf("%v", input)

	// Remove or escape characters that could be used for prompt injection
	s = strings.ReplaceAll(s, "\n", " ") // Remove newlines
	s = strings.ReplaceAll(s, "\r", "")  // Remove carriage returns
	s = strings.ReplaceAll(s, "\t", " ") // Replace tabs with spaces
	s = strings.ReplaceAll(s, "```", "") // Remove code blocks
	s = strings.ReplaceAll(s, "<|", "")  // Remove special tokens
	s = strings.ReplaceAll(s, "|>", "")  // Remove special tokens
	s = strings.ReplaceAll(s, "{{", "")  // Remove template markers
	s = strings.ReplaceAll(s, "}}", "")  // Remove template markers

	// Limit length to prevent overwhelming the prompt
	if len(s) > 500 {
		s = s[:500] + "..."
	}

	// Trim whitespace
	s = strings.TrimSpace(s)

	return s
}

// BuildPrompt constructs the AI prompt from events and analysis type
// Returns separate system and user prompts for providers that support it (Anthropic),
// or a combined prompt for providers that don't (OpenAI, Ollama)
func BuildPrompt(events []sdk.Event, analysisType string, provider Provider) PromptResult {
	// Build structured event data (no string concatenation of user input)
	var safeEvents []SafeEvent
	prCount := 0
	issueCount := 0
	actors := make(map[string]bool)

	// Limit to 50 events for token efficiency
	limit := 50
	if len(events) < limit {
		limit = len(events)
	}

	for i := 0; i < limit; i++ {
		event := events[i]

		// Count types
		switch event.Type {
		case "pull_request":
			prCount++
		case "issue":
			issueCount++
		}
		actors[event.Actor] = true

		// Extract action safely
		action := ""
		if val, ok := event.Data["action"]; ok {
			if str, ok := val.(string); ok {
				action = str
			}
		}

		// Build safe structured event (all fields properly typed)
		safeEvents = append(safeEvents, SafeEvent{
			Type:      event.Type,
			Action:    action,
			Actor:     event.Actor,
			Timestamp: event.Timestamp.Format("2006-01-02"),
		})
	}

	// Marshal events to JSON (safe from injection)
	eventsJSON, err := json.Marshal(safeEvents)
	if err != nil {
		eventsJSON = []byte("[]")
	}

	// Build system prompt (contains no user input)
	systemPrompt := `You are analyzing engineering metrics for a software team.

You will receive:
1. Summary statistics
2. A JSON array of events with type, action, actor (anonymized), and timestamp

Generate insights in this exact JSON format:
{
  "insights": [
    {
      "severity": "info|warning|error",
      "title": "Brief title",
      "description": "Detailed description",
      "recommendation": "Actionable suggestion",
      "metrics": {}
    }
  ],
  "summary": "Overall 2-3 sentence summary"
}

Focus on:
1. Unusual patterns (anomalies)
2. Productivity trends
3. Potential bottlenecks
4. Team health signals

Respond ONLY with valid JSON. No markdown, no explanations, just the JSON object.`

	// Build user message with structured data only
	userMessage := fmt.Sprintf(`Analysis type: %s

Summary statistics:
- Pull Requests: %d
- Issues: %d
- Team size: %d engineers

Events (JSON array):
%s`,
		SanitizeInput(analysisType),
		prCount,
		issueCount,
		len(actors),
		string(eventsJSON))

	// For Anthropic, return system and user prompts separately
	// For OpenAI/Ollama, combine them into user message
	if provider == ProviderAnthropic {
		return PromptResult{
			SystemPrompt: systemPrompt,
			UserMessage:  userMessage,
		}
	}

	// For other providers, combine into single prompt
	combinedPrompt := systemPrompt + "\n\n" + userMessage
	return PromptResult{
		SystemPrompt: "",
		UserMessage:  combinedPrompt,
	}
}
