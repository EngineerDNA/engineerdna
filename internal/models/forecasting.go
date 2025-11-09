package models

import "time"

// Forecast represents a predicted future outcome
type Forecast struct {
	ID                     string     `json:"id"`
	ForecastType           string     `json:"forecast_type"` // 'sprint_completion', 'goal_completion', 'team_velocity', etc.
	EntityType             string     `json:"entity_type"`   // 'engineer', 'team', 'sprint', 'goal', 'org'
	EntityID               *string    `json:"entity_id,omitempty"`
	TimeHorizon            string     `json:"time_horizon"` // 'next_week', 'next_sprint', 'next_quarter', specific date
	PredictedValue         float64    `json:"predicted_value"`
	PredictedDate          *time.Time `json:"predicted_date,omitempty"` // when outcome expected
	ConfidencePercentage   float64    `json:"confidence_percentage"`    // 0-100
	ConfidenceIntervalLow  *float64   `json:"confidence_interval_low,omitempty"`
	ConfidenceIntervalHigh *float64   `json:"confidence_interval_high,omitempty"`
	ModelType              string     `json:"model_type"`            // 'linear_regression', 'moving_average', 'monte_carlo'
	InputData              string     `json:"input_data,omitempty"`  // JSON: what data was used
	Assumptions            string     `json:"assumptions,omitempty"` // JSON: assumptions made
	CreatedAt              time.Time  `json:"created_at"`
	ExpiresAt              *time.Time `json:"expires_at,omitempty"` // when forecast is no longer relevant
}

// Scenario represents a what-if hypothetical situation
type Scenario struct {
	ID                 string    `json:"id"`
	ScenarioName       string    `json:"scenario_name"`
	ScenarioType       string    `json:"scenario_type"` // 'capacity_change', 'timeline_adjustment', 'cost_optimization'
	EntityType         string    `json:"entity_type"`
	EntityID           *string   `json:"entity_id,omitempty"`
	Description        string    `json:"description,omitempty"`
	Parameters         string    `json:"parameters"`                     // JSON: scenario parameters
	BaselineForecastID *string   `json:"baseline_forecast_id,omitempty"` // original forecast being modified
	CreatedBy          string    `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
}

// ScenarioResult represents the outcome of a what-if scenario
type ScenarioResult struct {
	ID                     string    `json:"id"`
	ScenarioID             string    `json:"scenario_id"`
	MetricName             string    `json:"metric_name"` // 'completion_date', 'total_cost', 'velocity', etc.
	PredictedValue         float64   `json:"predicted_value"`
	DifferenceFromBaseline *float64  `json:"difference_from_baseline,omitempty"` // delta vs baseline
	Impact                 string    `json:"impact"`                             // 'positive', 'negative', 'neutral'
	ComputedAt             time.Time `json:"computed_at"`
}

// RiskPrediction represents a predicted risk
type RiskPrediction struct {
	ID                    string     `json:"id"`
	RiskType              string     `json:"risk_type"` // 'attrition', 'timeline_miss', 'quality_degradation', 'burnout'
	EntityType            string     `json:"entity_type"`
	EntityID              *string    `json:"entity_id,omitempty"`
	RiskScore             float64    `json:"risk_score"`                       // 0.0-1.0
	ProbabilityPercentage float64    `json:"probability_percentage"`           // 0-100
	ImpactSeverity        string     `json:"impact_severity"`                  // 'low', 'medium', 'high', 'critical'
	ContributingFactors   string     `json:"contributing_factors,omitempty"`   // JSON: array of factors
	MitigationSuggestions string     `json:"mitigation_suggestions,omitempty"` // JSON: array of suggestions
	PredictedAt           time.Time  `json:"predicted_at"`
	ValidUntil            *time.Time `json:"valid_until,omitempty"`
}

// ForecastAccuracy tracks how accurate a forecast was
type ForecastAccuracy struct {
	ID                          string    `json:"id"`
	ForecastID                  string    `json:"forecast_id"`
	ActualValue                 float64   `json:"actual_value"`
	ActualDate                  time.Time `json:"actual_date"`
	ErrorAmount                 float64   `json:"error_amount"` // actual - predicted
	ErrorPercentage             float64   `json:"error_percentage"`
	WasWithinConfidenceInterval bool      `json:"was_within_confidence_interval"`
	MeasuredAt                  time.Time `json:"measured_at"`
}

// ForecastSummary provides aggregated forecast statistics
type ForecastSummary struct {
	EntityType         string  `json:"entity_type"`
	EntityID           *string `json:"entity_id,omitempty"`
	TotalForecasts     int     `json:"total_forecasts"`
	ActiveForecasts    int     `json:"active_forecasts"`
	ExpiredForecasts   int     `json:"expired_forecasts"`
	AverageConfidence  float64 `json:"average_confidence"`
	AccuracyPercentage float64 `json:"accuracy_percentage,omitempty"` // based on historical accuracy
}

// ScenarioWithResults includes scenario and its computed results
type ScenarioWithResults struct {
	Scenario         *Scenario         `json:"scenario"`
	Results          []*ScenarioResult `json:"results,omitempty"`
	BaselineForecast *Forecast         `json:"baseline_forecast,omitempty"`
}

// RiskFactor represents a contributing factor to a risk prediction
type RiskFactor struct {
	Name     string  `json:"name"`
	Weight   float64 `json:"weight"`
	Evidence string  `json:"evidence"`
}

// Mitigation represents a suggested action to mitigate a risk
type Mitigation struct {
	Action      string `json:"action"`
	Priority    string `json:"priority"` // 'low', 'medium', 'high'
	Description string `json:"description"`
}
