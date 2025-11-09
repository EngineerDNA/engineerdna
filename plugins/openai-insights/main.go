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
	if err := common.ValidateAnalysisType(request.AnalysisType); err != nil {
		return sdk.AnalyzeResult{}, err
	}

	// Build prompt
	promptResult := common.BuildPrompt(request.Events, request.AnalysisType, common.ProviderOpenAI)

	// Call OpenAI API (uses combined prompt in UserMessage)
	response, err := p.callOpenAI(promptResult.UserMessage)
	if err != nil {
		return sdk.AnalyzeResult{}, fmt.Errorf("OpenAI API error: %w", err)
	}

	// Parse response
	result, err := common.ParseResponse(response)
	if err != nil {
		return sdk.AnalyzeResult{}, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return result, nil
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

func main() {
	sdk.Serve(&OpenAIInsightsPlugin{})
}
