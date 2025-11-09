package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/engineerdna/engineerdna/internal/constants"
	common "github.com/engineerdna/engineerdna/plugins/insights-common"
	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

type ClaudeInsightsPlugin struct {
	apiKey string
	model  string
}

func (p *ClaudeInsightsPlugin) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		Name:        "claude-insights",
		Version:     "0.1.0",
		Type:        "processor",
		Description: "AI-powered insights using Anthropic Claude",
		Author:      "EngineerDNA",
		ConfigFields: []sdk.ConfigField{
			{
				Name:        "api_key",
				Type:        "password",
				Required:    true,
				Description: "Your Anthropic API key (BYOK)",
				Secret:      true,
			},
			{
				Name:        "model",
				Type:        "string",
				Required:    false,
				Default:     "claude-sonnet-4-20250514",
				Description: "Model name",
			},
		},
		Anonymization: sdk.AnonymizationSpec{
			Required: true,
			Reason:   "Data sent to Anthropic API",
			Strategy: "sequential",
			Fields:   []string{"actor", "author", "assignee", "email", "name"},
		},
		Capabilities: []string{"weekly_summary", "anomaly_detection", "recommendations"},
	}
}

func (p *ClaudeInsightsPlugin) Configure(config map[string]string) error {
	apiKey, ok := config["api_key"]
	if !ok || apiKey == "" {
		return fmt.Errorf("api_key is required")
	}
	p.apiKey = apiKey

	p.model = config["model"]
	if p.model == "" {
		p.model = "claude-sonnet-4-20250514"
	}

	return nil
}

func (p *ClaudeInsightsPlugin) Health() sdk.HealthResult {
	if p.apiKey == "" {
		return sdk.HealthResult{
			Healthy: false,
			Message: "Not configured",
		}
	}

	return sdk.HealthResult{
		Healthy: true,
		Message: "Configured to use Anthropic Claude",
	}
}

func (p *ClaudeInsightsPlugin) Capabilities() []string {
	return []string{"weekly_summary", "anomaly_detection", "recommendations"}
}

func (p *ClaudeInsightsPlugin) Analyze(request sdk.AnalyzeRequest) (sdk.AnalyzeResult, error) {
	if p.apiKey == "" {
		return sdk.AnalyzeResult{}, fmt.Errorf("plugin not configured")
	}

	// Validate analysisType against whitelist to prevent prompt injection
	if err := common.ValidateAnalysisType(request.AnalysisType); err != nil {
		return sdk.AnalyzeResult{}, err
	}

	// Build prompt (system and user messages)
	prompt := common.BuildPrompt(request.Events, request.AnalysisType, common.ProviderAnthropic)

	// Call Anthropic API
	response, err := p.callAnthropic(prompt.SystemPrompt, prompt.UserMessage)
	if err != nil {
		return sdk.AnalyzeResult{}, fmt.Errorf("Anthropic API error: %w", err)
	}

	// Parse response
	result, err := common.ParseResponse(response)
	if err != nil {
		return sdk.AnalyzeResult{}, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return result, nil
}

func (p *ClaudeInsightsPlugin) callAnthropic(systemPrompt, userMessage string) (string, error) {
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
	fmt.Fprintf(os.Stderr, "Anthropic API call - estimated tokens: %d input + %d max output (est. cost: $%.4f)\n",
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

func main() {
	sdk.Serve(&ClaudeInsightsPlugin{})
}
