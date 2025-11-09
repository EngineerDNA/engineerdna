package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// CorrelationEngine calculates cross-data-type correlations
type CorrelationEngine struct {
	metricEngine   *MetricEngine
	eventStore     *db.EventStore
	metricStore    *db.MetricStore
	attributeStore *db.AttributeStore
	corrStore      *db.CorrelationStore
}

// NewCorrelationEngine creates a new correlation calculation engine
func NewCorrelationEngine(
	metricEngine *MetricEngine,
	eventStore *db.EventStore,
	metricStore *db.MetricStore,
	attributeStore *db.AttributeStore,
	corrStore *db.CorrelationStore,
) *CorrelationEngine {
	return &CorrelationEngine{
		metricEngine:   metricEngine,
		eventStore:     eventStore,
		metricStore:    metricStore,
		attributeStore: attributeStore,
		corrStore:      corrStore,
	}
}

// CorrelationDefinition describes how to calculate a correlation
type CorrelationDefinition struct {
	Name        string                      `json:"name"`
	Description string                      `json:"description,omitempty"`
	Inputs      map[string]CorrelationInput `json:"inputs"`
	Formula     string                      `json:"formula"`
	TimeWindow  string                      `json:"time_window,omitempty"` // day, week, month, quarter
	Metadata    map[string]interface{}      `json:"metadata,omitempty"`
}

// CorrelationInput defines a single input to a correlation
type CorrelationInput struct {
	Source    string                 `json:"source"` // "metric", "events", "attribute"
	Metric    string                 `json:"metric,omitempty"`
	EventType string                 `json:"event_type,omitempty"`
	Attribute string                 `json:"attribute,omitempty"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
}

// CorrelationResult represents the result of a correlation calculation
type CorrelationResult struct {
	CorrelationName string                 `json:"correlation_name"`
	Value           float64                `json:"value"`
	TimeWindow      string                 `json:"time_window"`
	Timestamp       time.Time              `json:"timestamp"`
	Breakdown       map[string]interface{} `json:"breakdown,omitempty"`
	Inputs          map[string]float64     `json:"inputs"` // Input values used
	Interpretation  string                 `json:"interpretation,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Calculate executes a correlation calculation
func (e *CorrelationEngine) Calculate(definition *CorrelationDefinition, params map[string]interface{}) (*CorrelationResult, error) {
	// Validate definition
	if definition.Name == "" {
		return nil, fmt.Errorf("correlation name is required")
	}
	if len(definition.Inputs) == 0 {
		return nil, fmt.Errorf("at least one input is required")
	}
	if definition.Formula == "" {
		return nil, fmt.Errorf("formula is required")
	}

	// Extract time range
	periodStart, periodEnd, err := e.extractTimeRange(params, definition.TimeWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to extract time range: %w", err)
	}

	// Fetch input values
	inputValues := make(map[string]float64)
	for inputName, inputSpec := range definition.Inputs {
		value, err := e.fetchInputValue(inputSpec, periodStart, periodEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch input %s: %w", inputName, err)
		}
		inputValues[inputName] = value
	}

	// Execute formula
	result, err := e.executeFormula(definition.Formula, inputValues)
	if err != nil {
		return nil, fmt.Errorf("failed to execute formula: %w", err)
	}

	// Generate interpretation
	interpretation := e.generateInterpretation(definition, result, inputValues)

	return &CorrelationResult{
		CorrelationName: definition.Name,
		Value:           result,
		TimeWindow:      definition.TimeWindow,
		Timestamp:       periodStart,
		Inputs:          inputValues,
		Interpretation:  interpretation,
		Metadata:        definition.Metadata,
	}, nil
}

// extractTimeRange gets time range from params or defaults
func (e *CorrelationEngine) extractTimeRange(params map[string]interface{}, timeWindow string) (time.Time, time.Time, error) {
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
	case "day":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 0, 1)
	case "week":
		weekday := int(now.Weekday())
		start = now.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
		end = start.AddDate(0, 0, 7)
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0)
	case "quarter":
		month := int(now.Month())
		quarterStart := ((month-1)/3)*3 + 1
		start = time.Date(now.Year(), time.Month(quarterStart), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 3, 0)
	default:
		end = now
		start = now.AddDate(0, 0, -7) // Default to last week
	}

	return start, end, nil
}

// fetchInputValue retrieves the value for a correlation input
func (e *CorrelationEngine) fetchInputValue(input CorrelationInput, start, end time.Time) (float64, error) {
	switch input.Source {
	case "metric":
		return e.fetchMetricValue(input.Metric, start, end)
	case "events":
		return e.fetchEventCount(input.EventType, input.Filters, start, end)
	case "attribute":
		return e.fetchAttributeValue(input.Attribute, input.Filters)
	default:
		return 0, fmt.Errorf("unsupported input source: %s", input.Source)
	}
}

// fetchMetricValue retrieves a metric value from the store
func (e *CorrelationEngine) fetchMetricValue(metricName string, start, end time.Time) (float64, error) {
	if metricName == "" {
		return 0, fmt.Errorf("metric name is required")
	}

	metrics, err := e.metricStore.GetByMetricName(metricName, start, end, 1)
	if err != nil {
		return 0, fmt.Errorf("failed to query metric: %w", err)
	}

	if len(metrics) == 0 {
		return 0, nil
	}

	// Use the most recent metric value
	return metrics[0].Value, nil
}

// fetchEventCount counts events matching criteria using SQL aggregation
func (e *CorrelationEngine) fetchEventCount(eventType string, filters map[string]interface{}, start, end time.Time) (float64, error) {
	queryFilters := map[string]interface{}{
		"start_time": start,
		"end_time":   end,
	}

	if eventType != "" {
		queryFilters["types"] = []string{eventType}
	}

	// Add basic filters that are supported by SQL
	for k, v := range filters {
		switch k {
		case "source", "actor", "engineer_id", "type":
			queryFilters[k] = v
		}
	}

	// Count events using SQL aggregation (no memory loading)
	count, err := e.eventStore.CountEvents(queryFilters)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return float64(count), nil
}

// fetchAttributeValue retrieves an attribute value
func (e *CorrelationEngine) fetchAttributeValue(attributeName string, filters map[string]interface{}) (float64, error) {
	if attributeName == "" {
		return 0, fmt.Errorf("attribute name is required")
	}

	queryFilters := map[string]interface{}{
		"attribute_name": attributeName,
		"as_of":          time.Now().UTC(),
	}

	// Add entity filters
	if entityType, ok := filters["entity_type"].(string); ok {
		queryFilters["entity_type"] = entityType
	}
	if entityID, ok := filters["entity_id"].(string); ok {
		queryFilters["entity_id"] = entityID
	}

	attributes, err := e.attributeStore.List(queryFilters, 1, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to query attribute: %w", err)
	}

	if len(attributes) == 0 {
		return 0, nil
	}

	// Parse attribute value as float64
	attr := attributes[0]
	if attr.ValueType == "number" {
		var value float64
		if err := json.Unmarshal([]byte(attr.Value), &value); err == nil {
			return value, nil
		}
	}

	return 0, fmt.Errorf("attribute is not a number")
}

// executeFormula evaluates a simple formula with basic math operations
func (e *CorrelationEngine) executeFormula(formula string, inputs map[string]float64) (float64, error) {
	// Simple formula parser supporting: +, -, *, /
	// Format: "input1 / input2" or "input1 * 100"

	// For now, support only two-operand formulas
	// This is a simplified implementation - a full parser would be more complex

	// Parse formula: "numerator / denominator"
	var operand1, operand2 string
	var operator rune

	// Find operator
	for i, ch := range formula {
		if ch == '+' || ch == '-' || ch == '*' || ch == '/' {
			operand1 = formula[:i]
			operator = ch
			operand2 = formula[i+1:]
			break
		}
	}

	if operator == 0 {
		// No operator found - try to parse as single value or input name
		if val, ok := inputs[formula]; ok {
			return val, nil
		}
		return 0, fmt.Errorf("invalid formula: %s", formula)
	}

	// Trim spaces
	operand1 = trimSpace(operand1)
	operand2 = trimSpace(operand2)

	// Get values
	val1, ok := inputs[operand1]
	if !ok {
		return 0, fmt.Errorf("unknown input: %s", operand1)
	}

	val2, ok := inputs[operand2]
	if !ok {
		return 0, fmt.Errorf("unknown input: %s", operand2)
	}

	// Execute operation
	switch operator {
	case '+':
		return val1 + val2, nil
	case '-':
		return val1 - val2, nil
	case '*':
		return val1 * val2, nil
	case '/':
		if val2 == 0 {
			return 0, nil // Avoid division by zero
		}
		return val1 / val2, nil
	default:
		return 0, fmt.Errorf("unsupported operator: %c", operator)
	}
}

// trimSpace removes leading and trailing spaces
func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}

	return s[start:end]
}

// generateInterpretation creates a human-readable interpretation
func (e *CorrelationEngine) generateInterpretation(def *CorrelationDefinition, result float64, inputs map[string]float64) string {
	// Simple interpretation generation
	// For more advanced interpretation, could use templates or AI

	if def.Name == "cost_per_feature" {
		return fmt.Sprintf("Cost per feature: $%.2f based on inputs: %v", result, inputs)
	}

	return fmt.Sprintf("%s: %.2f", def.Name, result)
}

// SaveCorrelationValue saves a correlation result to the database
func (e *CorrelationEngine) SaveCorrelationValue(correlationID string, result *CorrelationResult) error {
	value := &models.CorrelationValue{
		ID:            uuid.New().String(),
		CorrelationID: correlationID,
		Timestamp:     result.Timestamp,
		TimeWindow:    result.TimeWindow,
		Value:         result.Value,
		Breakdown:     result.Breakdown,
		CreatedAt:     time.Now().UTC(),
	}

	return e.corrStore.SaveCorrelationValue(value)
}

// CalculateAndSave calculates a correlation and saves the result
func (e *CorrelationEngine) CalculateAndSave(correlationID string, definition *CorrelationDefinition, params map[string]interface{}) (*CorrelationResult, error) {
	result, err := e.Calculate(definition, params)
	if err != nil {
		return nil, err
	}

	if err := e.SaveCorrelationValue(correlationID, result); err != nil {
		return nil, fmt.Errorf("failed to save correlation value: %w", err)
	}

	return result, nil
}

// ParseCorrelationDefinition parses a JSON correlation definition
func ParseCorrelationDefinition(data []byte) (*CorrelationDefinition, error) {
	var def CorrelationDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse correlation definition: %w", err)
	}
	return &def, nil
}
