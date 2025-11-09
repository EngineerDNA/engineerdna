package common

import "fmt"

// ValidAnalysisTypes is the whitelist of allowed analysis types
var ValidAnalysisTypes = map[string]bool{
	"weekly_summary":    true,
	"anomaly_detection": true,
	"recommendations":   true,
}

// ValidateAnalysisType validates that the analysis type is in the whitelist
// to prevent prompt injection attacks
func ValidateAnalysisType(analysisType string) error {
	if !ValidAnalysisTypes[analysisType] {
		return fmt.Errorf("invalid analysis type: %s (must be one of: weekly_summary, anomaly_detection, recommendations)", analysisType)
	}
	return nil
}
