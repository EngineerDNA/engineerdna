package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

// Widget query configuration constants
const (
	DefaultTimeseriesPeriodCount   = 12
	MaxComparisonEntities          = 100
	DefaultPerformanceLimit        = 50
	HealthScoreWarningThreshold    = 70.0
	HealthScoreCriticalThreshold   = 50.0
	AlertCountWarningThreshold     = 2
	AlertCountCriticalThreshold    = 5
	GoalsOffTrackWarningThreshold  = 1
	GoalsOffTrackCriticalThreshold = 3
	MaxAlertsTimelineEvents        = 100

	// Calculation constants
	PercentageMultiplier = 100.0
	DaysInWeek           = 7
)

// Widget Query Endpoints for Dashboard Widgets

// handleMetricsAggregate returns a single KPI value with comparison to previous period
// Used by: Number Card widget
// Query params: metric (team_score, pr_volume, etc), entity_type, entity_id, period (day, week, month)
func (s *Server) handleMetricsAggregate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metric := r.URL.Query().Get("metric")
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")
	period := r.URL.Query().Get("period")

	if metric == "" {
		respondError(w, http.StatusBadRequest, "metric parameter is required", nil)
		return
	}

	if period == "" {
		period = "week"
	}

	// Get current period snapshot
	currentStart, currentEnd := getPeriodDates(period, 0)
	current, err := s.dashboardStore.GetMetricSnapshot(metric, entityType, entityID, currentStart)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get current metric", err)
		return
	}

	// Get previous period snapshot for comparison
	previousStart, _ := getPeriodDates(period, 1)
	previous, err := s.dashboardStore.GetMetricSnapshot(metric, entityType, entityID, previousStart)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get previous metric", err)
		return
	}

	response := map[string]interface{}{
		"metric":       metric,
		"entity_type":  entityType,
		"entity_id":    entityID,
		"period":       period,
		"period_start": currentStart,
		"period_end":   currentEnd,
		"value":        0.0,
		"previous":     0.0,
		"change":       0.0,
		"change_pct":   0.0,
	}

	if current != nil {
		response["value"] = current.Value
	}

	if previous != nil {
		response["previous"] = previous.Value
		if current != nil && previous.Value > 0 {
			change := current.Value - previous.Value
			changePct := (change / previous.Value) * PercentageMultiplier
			response["change"] = change
			response["change_pct"] = changePct
		}
	}

	respondJSON(w, http.StatusOK, response)
}

// handleMetricsTimeseries returns metric values over time
// Used by: Timeseries Chart widget
// Query params: metric, entity_type, entity_id, period (day, week), count (number of periods)
func (s *Server) handleMetricsTimeseries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metric := r.URL.Query().Get("metric")
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")
	period := r.URL.Query().Get("period")
	count := DefaultTimeseriesPeriodCount

	if metric == "" {
		respondError(w, http.StatusBadRequest, "metric parameter is required", nil)
		return
	}

	if period == "" {
		period = "week"
	}

	// Calculate full date range (first period to last period)
	firstPeriodStart, _ := getPeriodDates(period, count-1)
	lastPeriodStart, _ := getPeriodDates(period, 0)

	// Fetch all snapshots in range with single query (eliminates N+1)
	snapshots, err := s.dashboardStore.GetMetricSnapshotsInRange(
		metric, entityType, entityID,
		firstPeriodStart, lastPeriodStart,
		count,
	)
	if err != nil {
		s.logger.Printf("Failed to get metric snapshots for %s: %v", metric, err)
		respondError(w, http.StatusInternalServerError, "Failed to get metric snapshots", err)
		return
	}

	// Build lookup map for fast access by period_start
	snapshotMap := make(map[string]*models.MetricSnapshot)
	for _, snapshot := range snapshots {
		snapshotMap[snapshot.PeriodStart] = snapshot
	}

	// Build response with all periods, filling gaps with zero values
	dataPoints := []map[string]interface{}{}
	for i := count - 1; i >= 0; i-- {
		periodStart, periodEnd := getPeriodDates(period, i)

		value := 0.0
		if snapshot, exists := snapshotMap[periodStart]; exists {
			value = snapshot.Value
		}

		dataPoints = append(dataPoints, map[string]interface{}{
			"period_start": periodStart,
			"period_end":   periodEnd,
			"value":        value,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"metric":      metric,
		"entity_type": entityType,
		"entity_id":   entityID,
		"period":      period,
		"count":       count,
		"data_points": dataPoints,
	})
}

// handleMetricsCompare returns metric values for multiple entities
// Used by: Bar Chart widget
// Query params: metric, group_by (team, engineer), period
func (s *Server) handleMetricsCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metric := r.URL.Query().Get("metric")
	groupBy := r.URL.Query().Get("group_by")
	period := r.URL.Query().Get("period")

	if metric == "" {
		respondError(w, http.StatusBadRequest, "metric parameter is required", nil)
		return
	}

	if groupBy == "" {
		respondError(w, http.StatusBadRequest, "group_by parameter is required", nil)
		return
	}

	if period == "" {
		period = "week"
	}

	periodStart, _ := getPeriodDates(period, 0)

	// Get metric snapshots for all entities of the specified type
	snapshots, err := s.dashboardStore.GetMetricSnapshots(metric, groupBy, periodStart, periodStart, MaxComparisonEntities)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get metric snapshots", err)
		return
	}

	// Format response - entity names already included from JOIN query
	entities := []map[string]interface{}{}
	for _, snapshot := range snapshots {
		// EntityName is populated by the JOIN query in GetMetricSnapshots
		entities = append(entities, map[string]interface{}{
			"entity_id":   snapshot.EntityID,
			"entity_name": snapshot.EntityName,
			"value":       snapshot.Value,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"metric":   metric,
		"group_by": groupBy,
		"period":   period,
		"entities": entities,
	})
}

// handleEngineersPerformance returns engineer performance table data
// Used by: Table widget
// Query params: period, sort_by, limit, offset
func (s *Server) handleEngineersPerformance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	period := r.URL.Query().Get("period")
	sortBy := r.URL.Query().Get("sort_by")

	if period == "" {
		period = "week"
	}

	if sortBy == "" {
		sortBy = "score"
	}

	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = DefaultPerformanceLimit
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	// Validate sort_by parameter to prevent SQL injection
	allowedSortColumns := map[string]bool{
		"score":    true,
		"pr_count": true,
		"name":     true,
	}

	if sortBy == "" {
		sortBy = "score"
	}

	if !allowedSortColumns[sortBy] {
		respondError(w, http.StatusBadRequest, "Invalid sort_by parameter", nil)
		return
	}

	// Query engineers with performance metrics using repository
	engineersData, err := s.engineerStore.GetPerformanceTable(limit, offset, sortBy)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to query engineer performance", err)
		return
	}

	engineers := []map[string]interface{}{}
	for _, eng := range engineersData {
		engineers = append(engineers, map[string]interface{}{
			"id":       eng.ID,
			"name":     eng.Name,
			"email":    eng.Email,
			"score":    eng.Score,
			"pr_count": eng.PRCount,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"engineers": engineers,
		"count":     len(engineers),
		"period":    period,
		"sort_by":   sortBy,
	})
}

// HealthCheck represents a single health indicator
type HealthCheck struct {
	Name      string   `json:"name"`
	Status    string   `json:"status"` // 'ok', 'warning', 'critical'
	Message   string   `json:"message,omitempty"`
	Value     *float64 `json:"value,omitempty"`
	Threshold *float64 `json:"threshold,omitempty"`
}

// handleMetricsHealthStatus returns overall system health status
// Used by: Status Indicator widget
func (s *Server) handleMetricsHealthStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check various health indicators using repository
	avgScore, activeAlerts, goalsOffTrack, err := s.healthStore.GetSystemHealth()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get system health", err)
		return
	}

	// Create health checks
	checks := []HealthCheck{
		{
			Name:      "Average Score",
			Status:    getHealthStatus(avgScore, HealthScoreWarningThreshold, HealthScoreCriticalThreshold, false),
			Value:     &avgScore,
			Threshold: float64Ptr(HealthScoreWarningThreshold),
		},
		{
			Name:      "Active Alerts",
			Status:    getHealthStatus(float64(activeAlerts), AlertCountWarningThreshold, AlertCountCriticalThreshold, true),
			Value:     float64Ptr(float64(activeAlerts)),
			Threshold: float64Ptr(AlertCountCriticalThreshold),
		},
		{
			Name:      "Goals Off Track",
			Status:    getHealthStatus(float64(goalsOffTrack), GoalsOffTrackWarningThreshold, GoalsOffTrackCriticalThreshold, true),
			Value:     float64Ptr(float64(goalsOffTrack)),
			Threshold: float64Ptr(GoalsOffTrackCriticalThreshold),
		},
	}

	// Determine overall status from checks
	overallStatus := "ok"
	for _, check := range checks {
		if check.Status == "critical" {
			overallStatus = "critical"
			break
		} else if check.Status == "warning" && overallStatus != "critical" {
			overallStatus = "warning"
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    overallStatus,
		"checks":    checks,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// getHealthStatus determines status based on value and thresholds
// higherIsBad: true if higher values indicate problems (e.g., alerts count)
func getHealthStatus(value, warningThreshold, criticalThreshold float64, higherIsBad bool) string {
	if higherIsBad {
		if value >= criticalThreshold {
			return "critical"
		} else if value >= warningThreshold {
			return "warning"
		}
		return "ok"
	}
	// Lower is bad (e.g., score)
	if value < criticalThreshold {
		return "critical"
	} else if value < warningThreshold {
		return "warning"
	}
	return "ok"
}

// float64Ptr returns a pointer to a float64
func float64Ptr(f float64) *float64 {
	return &f
}

// handleAlertsTimeline returns alerts for timeline overlay on charts
// Used by: Timeseries Chart widget (alert overlays)
// Query params: start, end
func (s *Server) handleAlertsTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	if start == "" || end == "" {
		respondError(w, http.StatusBadRequest, "start and end parameters are required", nil)
		return
	}

	filters := map[string]string{
		"start": start,
		"end":   end,
	}
	alerts, _, err := s.alertsStore.ListAlertInstances(filters, MaxAlertsTimelineEvents, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get alerts", err)
		return
	}

	timelineEvents := []map[string]interface{}{}
	for _, alert := range alerts {
		timelineEvents = append(timelineEvents, map[string]interface{}{
			"id":        alert.ID,
			"timestamp": alert.FiredAt.Format("2006-01-02"),
			"title":     alert.Title,
			"message":   alert.Message,
			"severity":  alert.Severity,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"events": timelineEvents,
		"count":  len(timelineEvents),
	})
}

// handleWidgetRegistry lists all available widgets from plugin registry
// GET /api/widgets/registry
func (s *Server) handleWidgetRegistry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if widget registry is available
	if s.widgetRegistry == nil {
		respondError(w, http.StatusInternalServerError, "Widget registry not available", nil)
		return
	}

	// Get optional plugin filter
	pluginName := r.URL.Query().Get("plugin")
	widgetType := r.URL.Query().Get("type")

	var widgets interface{}
	if pluginName != "" {
		widgets = s.widgetRegistry.ListWidgetsByPlugin(pluginName)
	} else if widgetType != "" {
		widgets = s.widgetRegistry.GetWidgetsByType(widgetType)
	} else {
		widgets = s.widgetRegistry.ListWidgets()
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"widgets": widgets,
	})
}

// handleWidgetByID retrieves a specific widget definition
// GET /api/widgets/registry/{widget_id}
func (s *Server) handleWidgetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract widget ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/widgets/registry/")
	widgetID := strings.TrimSpace(path)

	if widgetID == "" {
		respondError(w, http.StatusBadRequest, "Widget ID is required", nil)
		return
	}

	// Check if widget registry is available
	if s.widgetRegistry == nil {
		respondError(w, http.StatusInternalServerError, "Widget registry not available", nil)
		return
	}

	widget, ok := s.widgetRegistry.GetWidget(widgetID)
	if !ok {
		respondError(w, http.StatusNotFound, "Widget not found", nil)
		return
	}

	respondJSON(w, http.StatusOK, widget)
}

// Helper function to calculate period date ranges
func getPeriodDates(period string, offset int) (string, string) {
	now := time.Now().UTC()

	switch period {
	case "day":
		start := now.AddDate(0, 0, -offset)
		return start.Format("2006-01-02"), start.Format("2006-01-02")
	case "week":
		start := now.AddDate(0, 0, -DaysInWeek*offset)
		end := start.AddDate(0, 0, DaysInWeek-1)
		return start.Format("2006-01-02"), end.Format("2006-01-02")
	case "month":
		start := now.AddDate(0, -offset, 0)
		end := start.AddDate(0, 1, -1)
		return start.Format("2006-01-02"), end.Format("2006-01-02")
	default:
		return now.Format("2006-01-02"), now.Format("2006-01-02")
	}
}
