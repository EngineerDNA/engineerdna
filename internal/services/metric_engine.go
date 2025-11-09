package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/constants"
	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// MetricEngine calculates metrics based on definitions from plugins
type MetricEngine struct {
	eventStore     *db.EventStore
	metricStore    *db.MetricStore
	attributeStore *db.AttributeStore
}

// NewMetricEngine creates a new metric calculation engine
func NewMetricEngine(
	eventStore *db.EventStore,
	metricStore *db.MetricStore,
	attributeStore *db.AttributeStore,
) *MetricEngine {
	return &MetricEngine{
		eventStore:     eventStore,
		metricStore:    metricStore,
		attributeStore: attributeStore,
	}
}

// MetricDefinition describes how to calculate a metric
type MetricDefinition struct {
	Name        string                 `json:"name"`
	Calculation CalculationSpec        `json:"calculation"`
	Unit        string                 `json:"unit,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CalculationSpec defines calculation parameters
type CalculationSpec struct {
	Type        string                 `json:"type"` // count, duration_avg, ratio, formula
	EventTypes  []string               `json:"event_types,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	TimeWindow  string                 `json:"time_window,omitempty"` // hour, day, week, month
	Formula     string                 `json:"formula,omitempty"`
	Numerator   string                 `json:"numerator,omitempty"`
	Denominator string                 `json:"denominator,omitempty"`
	Field       string                 `json:"field,omitempty"` // Field to calculate on
}

// MetricResult represents the result of a metric calculation
type MetricResult struct {
	MetricName  string                 `json:"metric_name"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit,omitempty"`
	Period      string                 `json:"period"`
	PeriodStart time.Time              `json:"period_start"`
	PeriodEnd   time.Time              `json:"period_end"`
	Breakdown   map[string]interface{} `json:"breakdown,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Calculate executes a metric calculation based on the definition
func (e *MetricEngine) Calculate(definition *MetricDefinition, params map[string]interface{}) (*MetricResult, error) {
	// Validate definition
	if definition.Name == "" {
		return nil, fmt.Errorf("metric name is required")
	}
	if definition.Calculation.Type == "" {
		return nil, fmt.Errorf("calculation type is required")
	}

	// Extract time range from params or use defaults
	periodStart, periodEnd, err := e.extractTimeRange(params, definition.Calculation.TimeWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to extract time range: %w", err)
	}

	// Execute calculation based on type
	var value float64
	var breakdown map[string]interface{}

	switch definition.Calculation.Type {
	case "count":
		value, breakdown, err = e.calculateCount(definition.Calculation, periodStart, periodEnd)
	case "duration_avg":
		value, breakdown, err = e.calculateDurationAvg(definition.Calculation, periodStart, periodEnd)
	case "ratio":
		value, breakdown, err = e.calculateRatio(definition.Calculation, periodStart, periodEnd)
	case "formula":
		value, breakdown, err = e.calculateFormula(definition.Calculation, periodStart, periodEnd)
	default:
		return nil, fmt.Errorf("unsupported calculation type: %s", definition.Calculation.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("calculation failed: %w", err)
	}

	return &MetricResult{
		MetricName:  definition.Name,
		Value:       value,
		Unit:        definition.Unit,
		Period:      definition.Calculation.TimeWindow,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Breakdown:   breakdown,
		Metadata:    definition.Metadata,
	}, nil
}

// extractTimeRange gets time range from params or defaults to current period
func (e *MetricEngine) extractTimeRange(params map[string]interface{}, timeWindow string) (time.Time, time.Time, error) {
	// Check if explicit time range provided
	if start, ok := params["start"].(time.Time); ok {
		if end, ok := params["end"].(time.Time); ok {
			return start, end, nil
		}
	}

	// Default to current period based on time window
	now := time.Now().UTC()
	var start, end time.Time

	switch timeWindow {
	case "hour":
		start = now.Truncate(time.Hour)
		end = start.Add(time.Hour)
	case "day":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 0, 1)
	case "week":
		// Start of week (Sunday)
		weekday := int(now.Weekday())
		start = now.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
		end = start.AddDate(0, 0, 7)
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0)
	default:
		// Default to last 24 hours
		end = now
		start = now.Add(-constants.DefaultMetricsWindow)
	}

	return start, end, nil
}

// calculateCount counts events matching the filters using SQL aggregation
func (e *MetricEngine) calculateCount(spec CalculationSpec, start, end time.Time) (float64, map[string]interface{}, error) {
	// Build filters
	filters := map[string]interface{}{
		"start_time": start,
		"end_time":   end,
	}

	// Add event type filter
	if len(spec.EventTypes) > 0 {
		filters["types"] = spec.EventTypes
	}

	// Add basic filters that are supported by SQL
	for k, v := range spec.Filters {
		// Only add filters that can be handled by SQL
		switch k {
		case "source", "actor", "engineer_id", "type":
			filters[k] = v
		}
	}

	// Count events using SQL aggregation (no memory loading)
	count, err := e.eventStore.CountEvents(filters)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to count events: %w", err)
	}

	return float64(count), nil, nil
}

// calculateDurationAvg calculates average duration from events using SQL aggregation
func (e *MetricEngine) calculateDurationAvg(spec CalculationSpec, start, end time.Time) (float64, map[string]interface{}, error) {
	// Validate field is specified
	if spec.Field == "" {
		return 0, nil, fmt.Errorf("field is required for duration_avg calculation")
	}

	// Build filters
	filters := map[string]interface{}{
		"start_time": start,
		"end_time":   end,
	}

	if len(spec.EventTypes) > 0 {
		filters["types"] = spec.EventTypes
	}

	// Add basic filters that are supported by SQL
	for k, v := range spec.Filters {
		switch k {
		case "source", "actor", "engineer_id", "type":
			filters[k] = v
		}
	}

	// Calculate average using SQL aggregation (no memory loading)
	avgDuration, err := e.eventStore.AvgFieldValue(spec.Field, filters)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to calculate average duration: %w", err)
	}

	return avgDuration, nil, nil
}

// calculateRatio calculates ratio of numerator to denominator
func (e *MetricEngine) calculateRatio(spec CalculationSpec, start, end time.Time) (float64, map[string]interface{}, error) {
	if spec.Numerator == "" || spec.Denominator == "" {
		return 0, nil, fmt.Errorf("numerator and denominator are required for ratio calculation")
	}

	// Get numerator count
	numeratorSpec := CalculationSpec{
		Type:       "count",
		EventTypes: spec.EventTypes,
		Filters:    map[string]interface{}{spec.Numerator: true},
	}
	numerator, _, err := e.calculateCount(numeratorSpec, start, end)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to calculate numerator: %w", err)
	}

	// Get denominator count
	denominatorSpec := CalculationSpec{
		Type:       "count",
		EventTypes: spec.EventTypes,
		Filters:    map[string]interface{}{},
	}
	denominator, _, err := e.calculateCount(denominatorSpec, start, end)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to calculate denominator: %w", err)
	}

	if denominator == 0 {
		return 0, nil, nil
	}

	ratio := numerator / denominator
	return ratio, nil, nil
}

// calculateFormula evaluates a simple formula
func (e *MetricEngine) calculateFormula(spec CalculationSpec, start, end time.Time) (float64, map[string]interface{}, error) {
	// For now, just return 0 - formula parsing would require a full expression evaluator
	// This is a placeholder for future enhancement
	return 0, nil, fmt.Errorf("formula calculation is planned - use count, duration_avg, or ratio types")
}

// applyFilters applies custom filters to events
func (e *MetricEngine) applyFilters(events []*models.Event, filters map[string]interface{}) []*models.Event {
	if len(filters) == 0 {
		return events
	}

	filtered := make([]*models.Event, 0, len(events))
	for _, event := range events {
		if e.eventMatchesFilters(event, filters) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// eventMatchesFilters checks if an event matches the filter criteria
func (e *MetricEngine) eventMatchesFilters(event *models.Event, filters map[string]interface{}) bool {
	for key, expectedValue := range filters {
		// Skip special filter keys
		if key == "start_time" || key == "end_time" || key == "types" {
			continue
		}

		// Check if the field exists and matches in event data
		actualValue, ok := e.extractFieldValue(event.Data, key)
		if !ok {
			return false
		}

		// Compare values
		if !e.valuesMatch(actualValue, expectedValue) {
			return false
		}
	}
	return true
}

// extractFieldValue extracts a field value from event data
func (e *MetricEngine) extractFieldValue(data map[string]interface{}, field string) (interface{}, bool) {
	if data == nil || field == "" {
		return nil, false
	}

	// Simple field lookup (no nested path support for now)
	value, ok := data[field]
	return value, ok
}

// valuesMatch compares two values for equality
func (e *MetricEngine) valuesMatch(actual, expected interface{}) bool {
	// Handle nil cases
	if actual == nil && expected == nil {
		return true
	}
	if actual == nil || expected == nil {
		return false
	}

	// Simple equality check
	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}

// SaveMetricValue saves a calculated metric value to the database
func (e *MetricEngine) SaveMetricValue(source string, result *MetricResult) error {
	metric := &models.MetricValue{
		ID:          uuid.New().String(),
		MetricName:  result.MetricName,
		Source:      source,
		Timestamp:   result.PeriodStart,
		Granularity: result.Period,
		Value:       result.Value,
		Unit:        result.Unit,
		Dimensions:  result.Breakdown,
		CreatedAt:   time.Now().UTC(),
	}

	return e.metricStore.Create(metric)
}

// CalculateAndSave calculates a metric and saves the result
func (e *MetricEngine) CalculateAndSave(source string, definition *MetricDefinition, params map[string]interface{}) (*MetricResult, error) {
	result, err := e.Calculate(definition, params)
	if err != nil {
		return nil, err
	}

	if err := e.SaveMetricValue(source, result); err != nil {
		return nil, fmt.Errorf("failed to save metric value: %w", err)
	}

	return result, nil
}

// ParseMetricDefinition parses a JSON metric definition
func ParseMetricDefinition(data []byte) (*MetricDefinition, error) {
	var def MetricDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse metric definition: %w", err)
	}
	return &def, nil
}

// ValidateMetricName validates a metric name for security
func ValidateMetricName(name string) error {
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("metric name too long (max 100 characters)")
	}
	// Add more validation as needed
	return nil
}
