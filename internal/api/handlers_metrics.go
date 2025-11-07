package api

import (
	"net/http"
	"strconv"
	"time"
)

// handleMetricsThroughput calculates and returns throughput metrics
// Uses SQL aggregation to prevent memory exhaustion
func (s *Server) handleMetricsThroughput(w http.ResponseWriter, r *http.Request) {
	// Parse days parameter
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if parsed, err := strconv.Atoi(daysStr); err == nil && parsed > 0 {
			days = parsed
		}
	}

	// Calculate time range
	endTime := time.Now().UTC()
	startTime := endTime.AddDate(0, 0, -days)

	// Fetch throughput metrics using SQL aggregation
	throughputDays, err := s.eventStore.GetThroughputByDay(startTime, endTime)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get throughput metrics", err)
		return
	}

	// Convert to map for easier lookup
	throughputMap := make(map[string]*struct {
		PullRequests int
		Issues       int
	})
	for _, day := range throughputDays {
		throughputMap[day.Date] = &struct {
			PullRequests int
			Issues       int
		}{
			PullRequests: day.PullRequests,
			Issues:       day.Issues,
		}
	}

	// Generate data array with all days (including zeros for days with no events)
	type DayMetric struct {
		Date         string `json:"date"`
		PullRequests int    `json:"pull_requests"`
		Issues       int    `json:"issues"`
	}

	var data []DayMetric
	for d := 0; d < days; d++ {
		date := startTime.AddDate(0, 0, d)
		dateKey := date.Format("2006-01-02")

		var prs, issues int
		if counts, exists := throughputMap[dateKey]; exists {
			prs = counts.PullRequests
			issues = counts.Issues
		}

		data = append(data, DayMetric{
			Date:         dateKey,
			PullRequests: prs,
			Issues:       issues,
		})
	}

	// Return response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"period": map[string]string{
			"start": startTime.Format(time.RFC3339),
			"end":   endTime.Format(time.RFC3339),
		},
		"data": data,
	})
}

// handleMetricsCycleTime returns cycle time metrics
func (s *Server) handleMetricsCycleTime(w http.ResponseWriter, r *http.Request) {
	// Cycle time metrics not yet implemented - returns empty list
	respondJSON(w, http.StatusOK, map[string]interface{}{"metrics": []interface{}{}})
}

// handleMetricsToday returns real-time metrics for today's engineering activity
// Uses SQL aggregation to prevent memory exhaustion
func (s *Server) handleMetricsToday(w http.ResponseWriter, r *http.Request) {
	// Get today's start in UTC
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Fetch today's metrics using SQL aggregation
	metrics, err := s.eventStore.GetTodayMetricsAggregated(startOfDay)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get today's metrics", err)
		return
	}

	// Count PRs in review (open PRs)
	prsInReview, err := s.eventStore.CountOpenPRs()
	if err != nil {
		// Log error but don't fail the request
		prsInReview = 0
	}

	// Count active alerts
	activeAlerts := 0
	if s.alertsStore != nil {
		alerts, _, err := s.alertsStore.ListAlertInstances(map[string]string{"status": "active"}, 100, 0)
		if err == nil {
			activeAlerts = len(alerts)
		}
	}

	// Calculate sprint progress (average across active sprints)
	sprintProgress := 0.0
	if s.planningStore != nil {
		sprints, err := s.planningStore.ListSprints(100)
		if err == nil {
			activeCount := 0
			totalProgress := 0.0
			for _, sprint := range sprints {
				if sprint.Status == "active" && sprint.CommittedPoints > 0 {
					progress := (float64(sprint.CompletedPoints) / float64(sprint.CommittedPoints)) * 100
					totalProgress += progress
					activeCount++
				}
			}
			if activeCount > 0 {
				sprintProgress = totalProgress / float64(activeCount)
			}
		}
	}

	// Return response matching frontend LiveMetrics interface
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"prs_merged_today":      metrics.PRsMerged,
		"prs_opened_today":      metrics.PRsCreated,
		"prs_in_review":         prsInReview,
		"avg_review_time_today": metrics.AvgReviewTime,
		"active_alerts":         activeAlerts,
		"sprint_progress":       sprintProgress,
	})
}
