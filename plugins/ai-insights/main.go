package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/constants"
	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

// sanitizeInput prevents prompt injection by escaping special characters
// and removing potential instruction-like patterns
func sanitizeInput(input string) string {
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

type AIInsightsPlugin struct {
	provider string
	apiKey   string
	model    string
	endpoint string
}

func (p *AIInsightsPlugin) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		Name:        "ai-insights",
		Version:     "1.0.0",
		Type:        "processor",
		Description: "AI-powered metrics analysis",
		Author:      "EngineerDNA",
		ConfigFields: []sdk.ConfigField{
			{
				Name:        "provider",
				Type:        "select",
				Options:     []string{"anthropic"},
				Required:    true,
				Description: "AI provider (OpenAI and Ollama coming soon)",
			},
			{
				Name:        "api_key",
				Type:        "password",
				Required:    true,
				Description: "Your API key (BYOK)",
				Secret:      true,
			},
			{
				Name:        "model",
				Type:        "string",
				Required:    false,
				Default:     "claude-sonnet-4-20250514",
				Description: "Model name",
			},
			{
				Name:        "endpoint",
				Type:        "string",
				Required:    false,
				Description: "For Ollama: http://localhost:11434",
			},
		},
		Anonymization: sdk.AnonymizationSpec{
			Required: true,
			Reason:   "Data sent to external AI API",
			Strategy: "sequential",
			Fields:   []string{"actor", "author", "assignee", "email", "name"},
		},
		Capabilities: []string{"weekly_summary", "anomaly_detection", "recommendations"},
	}
}

func (p *AIInsightsPlugin) Configure(config map[string]string) error {
	provider, ok := config["provider"]
	if !ok || provider == "" {
		return fmt.Errorf("provider is required")
	}
	p.provider = provider

	apiKey, ok := config["api_key"]
	if !ok || apiKey == "" {
		return fmt.Errorf("api_key is required")
	}
	p.apiKey = apiKey

	p.model = config["model"]
	if p.model == "" {
		p.model = "claude-sonnet-4-20250514"
	}

	p.endpoint = config["endpoint"]
	if p.endpoint == "" && p.provider == "ollama" {
		p.endpoint = "http://localhost:11434"
	}

	// Validate Ollama endpoint to prevent SSRF
	if p.provider == "ollama" && p.endpoint != "" {
		parsedURL, err := url.Parse(p.endpoint)
		if err != nil {
			return fmt.Errorf("invalid endpoint URL: %w", err)
		}

		// Only allow http/https schemes
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return fmt.Errorf("invalid endpoint scheme: %s (must be http or https)", parsedURL.Scheme)
		}

		// For V1 (localhost-only): restrict to localhost to prevent SSRF
		// This prevents user from pointing to arbitrary internal services
		hostname := parsedURL.Hostname()
		if hostname != "localhost" && hostname != "127.0.0.1" && hostname != "::1" {
			return fmt.Errorf("invalid endpoint host: %s (must be localhost, 127.0.0.1, or ::1 for security)", hostname)
		}
	}

	return nil
}

func (p *AIInsightsPlugin) Health() sdk.HealthResult {
	if p.provider == "" || p.apiKey == "" {
		return sdk.HealthResult{
			Healthy: false,
			Message: "Not configured",
		}
	}

	return sdk.HealthResult{
		Healthy: true,
		Message: fmt.Sprintf("Configured to use %s", p.provider),
	}
}

func (p *AIInsightsPlugin) Capabilities() []string {
	return []string{"weekly_summary", "anomaly_detection", "recommendations"}
}

func (p *AIInsightsPlugin) Analyze(request sdk.AnalyzeRequest) (sdk.AnalyzeResult, error) {
	if p.provider == "" {
		return sdk.AnalyzeResult{}, fmt.Errorf("plugin not configured")
	}

	// Validate analysisType against whitelist to prevent prompt injection
	validTypes := map[string]bool{
		"weekly_summary":    true,
		"anomaly_detection": true,
		"recommendations":   true,
	}
	if !validTypes[request.AnalysisType] {
		return sdk.AnalyzeResult{}, fmt.Errorf("invalid analysis type: %s (must be one of: weekly_summary, anomaly_detection, recommendations)", request.AnalysisType)
	}

	// Build prompt (system and user messages)
	systemPrompt, userMessage := p.buildPrompt(request.Events, request.AnalysisType)

	// Call AI provider
	var response string
	var err error

	switch p.provider {
	case "anthropic":
		response, err = p.callAnthropic(systemPrompt, userMessage)
	case "openai":
		response, err = p.callOpenAI(systemPrompt + "\n\n" + userMessage)
	case "ollama":
		response, err = p.callOllama(systemPrompt + "\n\n" + userMessage)
	default:
		return sdk.AnalyzeResult{}, fmt.Errorf("unsupported provider: %s", p.provider)
	}

	if err != nil {
		return sdk.AnalyzeResult{}, fmt.Errorf("AI API error: %w", err)
	}

	// Parse response
	result, err := p.parseResponse(response)
	if err != nil {
		return sdk.AnalyzeResult{}, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return result, nil
}

func (p *AIInsightsPlugin) buildPrompt(events []sdk.Event, analysisType string) (string, string) {
	// Build structured event data (no string concatenation of user input)
	type SafeEvent struct {
		Type      string `json:"type"`
		Action    string `json:"action"`
		Actor     string `json:"actor"`
		Timestamp string `json:"timestamp"`
	}

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
		sanitizeInput(analysisType),
		prCount,
		issueCount,
		len(actors),
		string(eventsJSON))

	// Return system and user messages separately for proper API usage
	return systemPrompt, userMessage
}

func (p *AIInsightsPlugin) callAnthropic(systemPrompt, userMessage string) (string, error) {
	// Enforce conservative token limits to prevent API cost abuse
	const maxOutputTokens = 2000
	// Conservative input limit: 20k chars = ~5k tokens worst-case (4 chars/token)
	// At Claude 4 Sonnet pricing (~$3/MTok input, ~$15/MTok output):
	// Max cost per call: $0.015 input + $0.03 output = $0.045 per call
	const maxPromptChars = 20000

	totalLength := len(systemPrompt) + len(userMessage)
	if totalLength > maxPromptChars {
		return "", fmt.Errorf("prompt too large (%d chars, max %d / ~5k tokens)", totalLength, maxPromptChars)
	}

	// Conservative token estimation (worst case: 1 token per 3 chars for non-English)
	// Use 3.5 chars/token as conservative estimate instead of 4
	estimatedInputTokens := int(float64(totalLength) / 3.5)

	// Hard limit: reject if estimated cost exceeds $0.05 per call
	const maxEstimatedCostCents = 5 // $0.05
	// Anthropic Claude Sonnet pricing (as of 2025): $3/MTok input, $15/MTok output
	// Convert to cents: ($3/1M tokens) * tokens * 100 cents/dollar / 1M
	estimatedInputCostCents := float64(estimatedInputTokens) * 3.0 * 100.0 / 1000000.0
	estimatedOutputCostCents := float64(maxOutputTokens) * 15.0 * 100.0 / 1000000.0
	estimatedTotalCostCents := estimatedInputCostCents + estimatedOutputCostCents

	if estimatedTotalCostCents > float64(maxEstimatedCostCents) {
		return "", fmt.Errorf("estimated API cost too high ($%.3f, max $%.2f)", estimatedTotalCostCents/100, float64(maxEstimatedCostCents)/100)
	}

	payload := map[string]interface{}{
		"model":      p.model,
		"max_tokens": maxOutputTokens,
		"system":     systemPrompt,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": userMessage,
			},
		},
	}

	// Log token estimate and cost for transparency
	fmt.Fprintf(os.Stderr, "AI API call - estimated tokens: %d input + %d max output (est. cost: $%.4f)\n",
		estimatedInputTokens, maxOutputTokens, estimatedTotalCostCents/100)

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Use custom client with timeout to prevent hanging on slow API
	client := &http.Client{
		Timeout: constants.LongAPITimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Limit error response size to prevent memory issues (1MB max)
		limitedReader := io.LimitReader(resp.Body, 1024*1024)
		bodyBytes, _ := io.ReadAll(limitedReader)
		// Parse error response to extract safe error message (don't leak full response)
		var errResp map[string]interface{}
		if json.Unmarshal(bodyBytes, &errResp) == nil {
			if errMsg, ok := errResp["error"].(map[string]interface{}); ok {
				if msg, ok := errMsg["message"].(string); ok {
					return "", fmt.Errorf("API error: %s", msg)
				}
			}
		}
		// Fallback to generic error if parsing fails
		// Note: Not logging raw error response to prevent potential information disclosure
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Limit response size to prevent unbounded memory usage (5MB max)
	limitedReader := io.LimitReader(resp.Body, 5*1024*1024)
	var result map[string]interface{}
	if err := json.NewDecoder(limitedReader).Decode(&result); err != nil {
		return "", err
	}

	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "", fmt.Errorf("unexpected response format")
	}

	textBlock, ok := content[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected content format")
	}

	text, ok := textBlock["text"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected text format")
	}

	return text, nil
}

func (p *AIInsightsPlugin) callOpenAI(prompt string) (string, error) {
	// OpenAI provider not yet implemented
	return "", fmt.Errorf("OpenAI provider not yet implemented")
}

func (p *AIInsightsPlugin) callOllama(prompt string) (string, error) {
	// Ollama provider not yet implemented
	return "", fmt.Errorf("Ollama provider not yet implemented")
}

func (p *AIInsightsPlugin) parseResponse(response string) (sdk.AnalyzeResult, error) {
	// Strip markdown if present
	response = strings.TrimPrefix(response, "```json\n")
	response = strings.TrimPrefix(response, "```\n")
	response = strings.TrimSuffix(response, "\n```")
	response = strings.TrimSpace(response)

	var result sdk.AnalyzeResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Truncate response in error to prevent leaking large/sensitive data
		preview := response
		if len(preview) > 200 {
			preview = preview[:200] + "... (truncated)"
		}
		return sdk.AnalyzeResult{}, fmt.Errorf("failed to parse JSON: %w (response preview: %s)", err, preview)
	}

	return result, nil
}

func main() {
	sdk.Serve(&AIInsightsPlugin{})
}
