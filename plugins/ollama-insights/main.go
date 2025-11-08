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

type OllamaInsightsPlugin struct {
	endpoint string
	model    string
}

func (p *OllamaInsightsPlugin) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		Name:        "ollama-insights",
		Version:     "0.1.0",
		Type:        "processor",
		Description: "AI-powered insights using local Ollama models",
		Author:      "EngineerDNA",
		ConfigFields: []sdk.ConfigField{
			{
				Name:        "endpoint",
				Type:        "string",
				Required:    false,
				Default:     "http://localhost:11434",
				Description: "Ollama API endpoint (e.g., http://localhost:11434)",
			},
			{
				Name:        "model",
				Type:        "string",
				Required:    false,
				Default:     "llama3",
				Description: "Model name (e.g., llama3, mistral, codellama)",
			},
		},
		Anonymization: sdk.AnonymizationSpec{
			Required: true,
			Reason:   "Data sent to AI model for analysis",
			Strategy: "sequential",
			Fields:   []string{"actor", "author", "assignee", "email", "name"},
		},
		Capabilities: []string{"weekly_summary", "anomaly_detection", "recommendations"},
	}
}

func (p *OllamaInsightsPlugin) Configure(config map[string]string) error {
	p.endpoint = config["endpoint"]
	if p.endpoint == "" {
		p.endpoint = "http://localhost:11434"
	}

	// Validate Ollama endpoint to prevent SSRF
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

	p.model = config["model"]
	if p.model == "" {
		p.model = "llama3"
	}

	return nil
}

func (p *OllamaInsightsPlugin) Health() sdk.HealthResult {
	if p.endpoint == "" {
		return sdk.HealthResult{
			Healthy: false,
			Message: "Not configured",
		}
	}

	return sdk.HealthResult{
		Healthy: true,
		Message: fmt.Sprintf("Configured to use Ollama at %s", p.endpoint),
	}
}

func (p *OllamaInsightsPlugin) Capabilities() []string {
	return []string{"weekly_summary", "anomaly_detection", "recommendations"}
}

func (p *OllamaInsightsPlugin) Analyze(request sdk.AnalyzeRequest) (sdk.AnalyzeResult, error) {
	if p.endpoint == "" {
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

	// Call Ollama API
	response, err := p.callOllama(prompt)
	if err != nil {
		return sdk.AnalyzeResult{}, fmt.Errorf("Ollama API error: %w", err)
	}

	// Parse response
	result, err := p.parseResponse(response)
	if err != nil {
		return sdk.AnalyzeResult{}, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return result, nil
}

func (p *OllamaInsightsPlugin) buildPrompt(events []sdk.Event, analysisType string) string {
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

func (p *OllamaInsightsPlugin) callOllama(prompt string) (string, error) {
	// Ollama is local, no cost concerns but still limit prompt size for performance
	const maxPromptChars = 50000 // More generous for local models

	if len(prompt) > maxPromptChars {
		return "", fmt.Errorf("prompt too large (%d chars, max %d)", len(prompt), maxPromptChars)
	}

	payload := map[string]interface{}{
		"model":  p.model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.7,
			"num_predict": 2000, // Max tokens to generate
		},
	}

	// Log request for transparency
	fmt.Fprintf(os.Stderr, "Ollama API call - model: %s, endpoint: %s\n", p.model, p.endpoint)

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	apiURL := p.endpoint + "/api/generate"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	// Use custom client with longer timeout for local models (they can be slow)
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
		// Parse error response to extract safe error message
		var errResp map[string]interface{}
		if json.Unmarshal(bodyBytes, &errResp) == nil {
			if msg, ok := errResp["error"].(string); ok {
				return "", fmt.Errorf("API error: %s", msg)
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

	response, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format: no 'response' field")
	}

	return response, nil
}

func (p *OllamaInsightsPlugin) parseResponse(response string) (sdk.AnalyzeResult, error) {
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
	sdk.Serve(&OllamaInsightsPlugin{})
}
