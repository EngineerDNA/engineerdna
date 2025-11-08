package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/internal/services"
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

// handleMetricValues handles GET and POST for metric values
func (s *Server) handleMetricValues(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleMetricValuesGet(w, r)
	case http.MethodPost:
		s.handleMetricValuesPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMetricValuesGet queries metric values with filters
// Query params: metric_name, source, since, until, granularity, limit, offset
func (s *Server) handleMetricValuesGet(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filters := make(map[string]interface{})

	if metricName := r.URL.Query().Get("metric_name"); metricName != "" {
		filters["metric_name"] = metricName
	}

	if source := r.URL.Query().Get("source"); source != "" {
		filters["source"] = source
	}

	if granularity := r.URL.Query().Get("granularity"); granularity != "" {
		filters["granularity"] = granularity
	}

	// Parse time range
	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		since, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid since timestamp", err)
			return
		}
		filters["since"] = since
	}

	if untilStr := r.URL.Query().Get("until"); untilStr != "" {
		until, err := time.Parse(time.RFC3339, untilStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid until timestamp", err)
			return
		}
		filters["until"] = until
	}

	// Parse pagination
	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = db.DefaultQueryLimit
	}
	if limit > db.MaxQueryLimit {
		limit = db.MaxQueryLimit
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	// Query metrics
	metrics, err := s.metricStore.List(filters, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to query metrics", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"metrics": metrics,
		"count":   len(metrics),
		"limit":   limit,
		"offset":  offset,
	})
}

// handleMetricValuesPost creates a new metric value (for metric_source plugins)
func (s *Server) handleMetricValuesPost(w http.ResponseWriter, r *http.Request) {
	var metric models.MetricValue
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if metric.MetricName == "" {
		respondError(w, http.StatusBadRequest, "metric_name is required", nil)
		return
	}

	// Validate metric name for security
	if err := services.ValidateMetricName(metric.MetricName); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid metric name", err)
		return
	}

	if metric.Source == "" {
		respondError(w, http.StatusBadRequest, "source is required", nil)
		return
	}
	if metric.Granularity == "" {
		respondError(w, http.StatusBadRequest, "granularity is required", nil)
		return
	}

	// Set timestamp to now if not provided
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now().UTC()
	}

	// Validate granularity
	validGranularities := map[string]bool{
		"hourly":  true,
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}
	if !validGranularities[metric.Granularity] {
		respondError(w, http.StatusBadRequest, "invalid granularity (must be: hourly, daily, weekly, monthly)", nil)
		return
	}

	// Create metric
	if err := s.metricStore.Create(&metric); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create metric", err)
		return
	}

	respondJSON(w, http.StatusCreated, metric)
}

// handleMetricCalculate calculates a metric on-demand using the metric engine
// POST /api/metrics/calculate
// Body: { "definition": {...}, "params": {...} }
func (s *Server) handleMetricCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Definition json.RawMessage        `json:"definition"`
		Params     map[string]interface{} `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Parse metric definition
	definition, err := services.ParseMetricDefinition(req.Definition)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid metric definition", err)
		return
	}

	// Check if metric engine is available
	if s.metricEngine == nil {
		respondError(w, http.StatusInternalServerError, "Metric engine not available", nil)
		return
	}

	// Calculate metric
	result, err := s.metricEngine.Calculate(definition, req.Params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to calculate metric", err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}
