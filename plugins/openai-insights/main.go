package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

type OpenAIInsightsPlugin struct {
	apiKey string
	model  string
}

func (p *OpenAIInsightsPlugin) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		Name:        "openai-insights",
		Version:     "0.1.0",
		Type:        "processor",
		Description: "AI-powered insights using OpenAI GPT",
		Author:      "EngineerDNA",
		ConfigFields: []sdk.ConfigField{
			{
				Name:        "api_key",
				Type:        "password",
				Required:    true,
				Description: "Your OpenAI API key (BYOK)",
				Secret:      true,
			},
			{
				Name:        "model",
				Type:        "string",
				Required:    false,
				Default:     "gpt-4",
				Description: "Model name (e.g., gpt-4, gpt-4-turbo)",
			},
		},
		Anonymization: sdk.AnonymizationSpec{
			Required: true,
			Reason:   "Data sent to OpenAI API",
			Strategy: "sequential",
			Fields:   []string{"actor", "author", "assignee", "email", "name"},
		},
		Capabilities: []string{"weekly_summary", "anomaly_detection", "recommendations"},
	}
}

func (p *OpenAIInsightsPlugin) Configure(config map[string]string) error {
	apiKey, ok := config["api_key"]
	if !ok || apiKey == "" {
		return fmt.Errorf("api_key is required")
	}
	p.apiKey = apiKey

	p.model = config["model"]
	if p.model == "" {
		p.model = "gpt-4"
	}

	return nil
}

func (p *OpenAIInsightsPlugin) Health() sdk.HealthResult {
	if p.apiKey == "" {
		return sdk.HealthResult{
			Healthy: false,
			Message: "Not configured",
		}
	}

	return sdk.HealthResult{
		Healthy: true,
		Message: "Configured to use OpenAI GPT",
	}
}

func (p *OpenAIInsightsPlugin) Capabilities() []string {
	return []string{"weekly_summary", "anomaly_detection", "recommendations"}
}

func (p *OpenAIInsightsPlugin) Analyze(request sdk.AnalyzeRequest) (sdk.AnalyzeResult, error) {
	if p.apiKey == "" {
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

	// Build prompt
	prompt := p.buildPrompt(request.Events, request.AnalysisType)

	// Call OpenAI API
	response, err := p.callOpenAI(prompt)
	if err != nil {
		return sdk.AnalyzeResult{}, fmt.Errorf("OpenAI API error: %w", err)
	}

	// Parse response
	result, err := p.parseResponse(response)
	if err != nil {
		return sdk.AnalyzeResult{}, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return result, nil
}

func (p *OpenAIInsightsPlugin) buildPrompt(events []sdk.Event, analysisType string) string {
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

	// Build combined prompt
	prompt := fmt.Sprintf(`You are analyzing engineering metrics for a software team.

Analysis type: %s

Summary statistics:
- Pull Requests: %d
- Issues: %d
- Team size: %d engineers

Events (JSON array):
%s

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

Respond ONLY with valid JSON. No markdown, no explanations, just the JSON object.`,
		sanitizeInput(analysisType),
		prCount,
		issueCount,
		len(actors),
		string(eventsJSON))

	return prompt
}

func (p *OpenAIInsightsPlugin) callOpenAI(prompt string) (string, error) {
	// Enforce conservative token limits to prevent API cost abuse
	const maxOutputTokens = 2000
	// Conservative input limit: 20k chars = ~5k tokens worst-case (4 chars/token)
	// At GPT-4 pricing (~$30/MTok input, ~$60/MTok output):
	// Max cost per call: $0.15 input + $0.12 output = $0.27 per call
	const maxPromptChars = 20000

	if len(prompt) > maxPromptChars {
		return "", fmt.Errorf("prompt too large (%d chars, max %d / ~5k tokens)", len(prompt), maxPromptChars)
	}

	// Conservative token estimation (worst case: 1 token per 3 chars for non-English)
	estimatedInputTokens := int(float64(len(prompt)) / 3.5)

	// Hard limit: reject if estimated cost exceeds $0.30 per call
	const maxEstimatedCostCents = 30 // $0.30
	// OpenAI GPT-4 pricing (as of 2025): $30/MTok input, $60/MTok output
	estimatedInputCostCents := float64(estimatedInputTokens) * 30.0 * 100.0 / 1000000.0
	estimatedOutputCostCents := float64(maxOutputTokens) * 60.0 * 100.0 / 1000000.0
	estimatedTotalCostCents := estimatedInputCostCents + estimatedOutputCostCents

	if estimatedTotalCostCents > float64(maxEstimatedCostCents) {
		return "", fmt.Errorf("estimated API cost too high ($%.3f, max $%.2f)", estimatedTotalCostCents/100, float64(maxEstimatedCostCents)/100)
	}

	payload := map[string]interface{}{
		"model":       p.model,
		"max_tokens":  maxOutputTokens,
		"temperature": 0.7,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	// Log token estimate and cost for transparency
	fmt.Fprintf(os.Stderr, "OpenAI API call - estimated tokens: %d input + %d max output (est. cost: $%.4f)\n",
		estimatedInputTokens, maxOutputTokens, estimatedTotalCostCents/100)

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Limit response size to prevent unbounded memory usage (5MB max)
	limitedReader := io.LimitReader(resp.Body, 5*1024*1024)
	var result map[string]interface{}
	if err := json.NewDecoder(limitedReader).Decode(&result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("unexpected response format: no choices")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected content format")
	}

	return content, nil
}

func (p *OpenAIInsightsPlugin) parseResponse(response string) (sdk.AnalyzeResult, error) {
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
	sdk.Serve(&OpenAIInsightsPlugin{})
}
