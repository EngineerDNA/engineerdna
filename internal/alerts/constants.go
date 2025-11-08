package alerts

// Default alert threshold constants for standard alert rules.
// These thresholds define when alerts should fire for common conditions.
const (
	// DefaultSprintBehindThreshold is the default percentage behind expected sprint progress
	// that triggers an alert. When a sprint is more than 30% behind pace, teams should
	// be notified to take corrective action.
	DefaultSprintBehindThreshold = 30.0
)
