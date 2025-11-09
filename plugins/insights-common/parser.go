package common

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

// ParseResponse parses the AI response into an AnalyzeResult
// Strips markdown formatting if present and handles errors gracefully
func ParseResponse(response string) (sdk.AnalyzeResult, error) {
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
