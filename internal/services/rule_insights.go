package services

import (
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
)

// RuleBasedInsights provides AI-optional threshold alerts and pattern detection
type RuleBasedInsights struct {
	metricStore *db.MetricStore
	eventStore  *db.EventStore
	alertsStore *db.AlertsStore
	goalsStore  *db.GoalsStore
}

// NewRuleBasedInsights creates a new rule-based insights engine
func NewRuleBasedInsights(
	metricStore *db.MetricStore,
	eventStore *db.EventStore,
	alertsStore *db.AlertsStore,
	goalsStore *db.GoalsStore,
) *RuleBasedInsights {
	return &RuleBasedInsights{
		metricStore: metricStore,
		eventStore:  eventStore,
		alertsStore: alertsStore,
		goalsStore:  goalsStore,
	}
}

// ThresholdBreach represents a metric that has exceeded its threshold
type ThresholdBreach struct {
	MetricName    string    `json:"metric_name"`
	CurrentValue  float64   `json:"current_value"`
	Threshold     float64   `json:"threshold"`
	Severity      string    `json:"severity"` // "warning", "critical"
	Message       string    `json:"message"`
	DetectedAt    time.Time `json:"detected_at"`
}

// SignificantChange represents a metric with significant week-over-week change
type SignificantChange struct {
	MetricName      string    `json:"metric_name"`
	CurrentValue    float64   `json:"current_value"`
	PreviousValue   float64   `json:"previous_value"`
	ChangePercent   float64   `json:"change_percent"`
	Direction       string    `json:"direction"` // "increase", "decrease"
	Message         string    `json:"message"`
	DetectedAt      time.Time `json:"detected_at"`
}

// WeeklySummary represents insights for a team for the week
type WeeklySummary struct {
	TeamID               string               `json:"team_id"`
	WeekStart            time.Time            `json:"week_start"`
	WeekEnd              time.Time            `json:"week_end"`
	ThresholdBreaches    []ThresholdBreach    `json:"threshold_breaches"`
	SignificantChanges   []SignificantChange  `json:"significant_changes"`
	StaleReviews         int                  `json:"stale_reviews"`
	GeneratedAt          time.Time            `json:"generated_at"`
}

// Thresholds define acceptable ranges for metrics
var DefaultThresholds = map[string]struct {
	Warning  float64
	Critical float64
	HigherIsBad bool
}{
	"cycle_time_days": {Warning: 3.0, Critical: 5.0, HigherIsBad: true},
	"review_time_hours": {Warning: 24.0, Critical: 48.0, HigherIsBad: true},
	"pr_size_lines": {Warning: 500.0, Critical: 1000.0, HigherIsBad: true},
	"deployment_frequency": {Warning: 5.0, Critical: 3.0, HigherIsBad: false},
	"test_coverage_percent": {Warning: 70.0, Critical: 50.0, HigherIsBad: false},
}

// DetectThresholdBreaches checks if metrics exceed thresholds
func (r *RuleBasedInsights) DetectThresholdBreaches(teamID string) ([]ThresholdBreach, error) {
	var breaches []ThresholdBreach
	now := time.Now().UTC()

	// Check each metric with defined thresholds
	for metricName, threshold := range DefaultThresholds {
		// Get recent metric value (last week)
		weekStart := now.AddDate(0, 0, -7)
		metrics, err := r.metricStore.GetByMetricName(metricName, weekStart, now, 1)
		if err != nil {
			continue // Skip if metric not found
		}

		if len(metrics) == 0 {
			continue
		}

		currentValue := metrics[0].Value

		// Check threshold breach
		var severity string
		var breached bool

		if threshold.HigherIsBad {
			// Higher values are bad (e.g., cycle time)
			if currentValue >= threshold.Critical {
				severity = "critical"
				breached = true
			} else if currentValue >= threshold.Warning {
				severity = "warning"
				breached = true
			}
		} else {
			// Lower values are bad (e.g., deployment frequency)
			if currentValue <= threshold.Critical {
				severity = "critical"
				breached = true
			} else if currentValue <= threshold.Warning {
				severity = "warning"
				breached = true
			}
		}

		if breached {
			var thresholdValue float64
			if severity == "critical" {
				thresholdValue = threshold.Critical
			} else {
				thresholdValue = threshold.Warning
			}

			breaches = append(breaches, ThresholdBreach{
				MetricName:   metricName,
				CurrentValue: currentValue,
				Threshold:    thresholdValue,
				Severity:     severity,
				Message:      fmt.Sprintf("%s is %.2f (threshold: %.2f)", metricName, currentValue, thresholdValue),
				DetectedAt:   now,
			})
		}
	}

	return breaches, nil
}

// DetectSignificantChanges checks for >20% changes week-over-week
func (r *RuleBasedInsights) DetectSignificantChanges(teamID string) ([]SignificantChange, error) {
	var changes []SignificantChange
	now := time.Now().UTC()

	// Define metrics to check for changes
	metricsToCheck := []string{
		"pr_volume",
		"deployment_frequency",
		"cycle_time_days",
		"test_coverage_percent",
	}

	for _, metricName := range metricsToCheck {
		// Get current week value
		thisWeekStart := now.AddDate(0, 0, -7)
		thisWeek, err := r.metricStore.GetByMetricName(metricName, thisWeekStart, now, 1)
		if err != nil || len(thisWeek) == 0 {
			continue
		}

		// Get previous week value
		lastWeekStart := now.AddDate(0, 0, -14)
		lastWeekEnd := now.AddDate(0, 0, -7)
		lastWeek, err := r.metricStore.GetByMetricName(metricName, lastWeekStart, lastWeekEnd, 1)
		if err != nil || len(lastWeek) == 0 {
			continue
		}

		currentValue := thisWeek[0].Value
		previousValue := lastWeek[0].Value

		// Skip if previous value is zero (avoid division by zero)
		if previousValue == 0 {
			continue
		}

		// Calculate percent change
		changePercent := ((currentValue - previousValue) / previousValue) * 100

		// Check if change is significant (>20%)
		if changePercent > 20 || changePercent < -20 {
			direction := "increase"
			if changePercent < 0 {
				direction = "decrease"
			}

			changes = append(changes, SignificantChange{
				MetricName:    metricName,
				CurrentValue:  currentValue,
				PreviousValue: previousValue,
				ChangePercent: changePercent,
				Direction:     direction,
				Message:       fmt.Sprintf("%s %s by %.1f%% this week", metricName, direction, abs(changePercent)),
				DetectedAt:    now,
			})
		}
	}

	return changes, nil
}

// DetectStaleReviews counts PRs waiting >48 hours for review
func (r *RuleBasedInsights) DetectStaleReviews() (int, error) {
	now := time.Now().UTC()
	staleThreshold := now.Add(-48 * time.Hour)

	// Query open PRs (reduced limit to prevent memory exhaustion)
	filters := map[string]interface{}{
		"types": []string{"pull_request"},
		"end_time": now,
		"start_time": staleThreshold.AddDate(0, 0, -30), // Last 30 days
	}

	events, err := r.eventStore.List(filters, db.MaxQueryLimit, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to query events: %w", err)
	}

	// Count stale PRs (open and created before threshold)
	staleCount := 0
	for _, event := range events {
		// Check if PR is still open
		if status, ok := event.Data["state"].(string); ok {
			if status == "open" {
				// Check if created before stale threshold
				if event.Timestamp.Before(staleThreshold) {
					staleCount++
				}
			}
		}
	}

	return staleCount, nil
}

// GenerateWeeklySummary combines insights into a weekly report
func (r *RuleBasedInsights) GenerateWeeklySummary(teamID string) (*WeeklySummary, error) {
	now := time.Now().UTC()
	weekStart := now.AddDate(0, 0, -7)

	// Detect threshold breaches
	breaches, err := r.DetectThresholdBreaches(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to detect threshold breaches: %w", err)
	}

	// Detect significant changes
	changes, err := r.DetectSignificantChanges(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to detect significant changes: %w", err)
	}

	// Detect stale reviews
	staleReviews, err := r.DetectStaleReviews()
	if err != nil {
		staleReviews = 0 // Don't fail if stale review detection fails
	}

	return &WeeklySummary{
		TeamID:             teamID,
		WeekStart:          weekStart,
		WeekEnd:            now,
		ThresholdBreaches:  breaches,
		SignificantChanges: changes,
		StaleReviews:       staleReviews,
		GeneratedAt:        now,
	}, nil
}

// abs returns absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// CheckMetricThreshold checks a single metric against a threshold
func (r *RuleBasedInsights) CheckMetricThreshold(metricName string, currentValue, threshold float64, higherIsBad bool) *ThresholdBreach {
	breached := false
	severity := "warning"

	if higherIsBad && currentValue >= threshold {
		breached = true
	} else if !higherIsBad && currentValue <= threshold {
		breached = true
	}

	if !breached {
		return nil
	}

	return &ThresholdBreach{
		MetricName:   metricName,
		CurrentValue: currentValue,
		Threshold:    threshold,
		Severity:     severity,
		Message:      fmt.Sprintf("%s is %.2f (threshold: %.2f)", metricName, currentValue, threshold),
		DetectedAt:   time.Now().UTC(),
	}
}
