-- Predictive Analytics & Forecasting
-- Migration 023: Forecasts, scenarios, risk predictions, and accuracy tracking

-- Forecasts: Predicted future outcomes
CREATE TABLE IF NOT EXISTS forecasts (
    id TEXT PRIMARY KEY,
    forecast_type TEXT NOT NULL, -- 'sprint_completion', 'goal_completion', 'attrition_risk', 'team_velocity', 'cost_projection'
    entity_type TEXT NOT NULL, -- 'engineer', 'team', 'sprint', 'goal', 'org'
    entity_id TEXT,
    time_horizon TEXT NOT NULL, -- 'next_week', 'next_sprint', 'next_quarter', specific date
    predicted_value REAL NOT NULL,
    predicted_date TEXT, -- when outcome expected (RFC3339)
    confidence_percentage REAL NOT NULL, -- 0-100
    confidence_interval_low REAL, -- lower bound
    confidence_interval_high REAL, -- upper bound
    model_type TEXT NOT NULL, -- 'linear_regression', 'moving_average', 'monte_carlo'
    input_data TEXT, -- JSON: what data was used
    assumptions TEXT, -- JSON: assumptions made
    created_at TEXT NOT NULL,
    expires_at TEXT -- when forecast is no longer relevant (RFC3339)
);

-- What-If Scenarios: Hypothetical situations
CREATE TABLE IF NOT EXISTS scenarios (
    id TEXT PRIMARY KEY,
    scenario_name TEXT NOT NULL,
    scenario_type TEXT NOT NULL, -- 'capacity_change', 'timeline_adjustment', 'cost_optimization'
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    description TEXT,
    parameters TEXT NOT NULL, -- JSON: scenario parameters
    baseline_forecast_id TEXT, -- original forecast being modified
    created_by TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (baseline_forecast_id) REFERENCES forecasts(id) ON DELETE SET NULL
);

-- Scenario Results: Outcomes of what-if scenarios
CREATE TABLE IF NOT EXISTS scenario_results (
    id TEXT PRIMARY KEY,
    scenario_id TEXT NOT NULL,
    metric_name TEXT NOT NULL, -- 'completion_date', 'total_cost', 'velocity', etc.
    predicted_value REAL NOT NULL,
    difference_from_baseline REAL, -- delta vs baseline
    impact TEXT, -- 'positive', 'negative', 'neutral'
    computed_at TEXT NOT NULL,
    FOREIGN KEY (scenario_id) REFERENCES scenarios(id) ON DELETE CASCADE
);

-- Risk Predictions: Predicted risks
CREATE TABLE IF NOT EXISTS risk_predictions (
    id TEXT PRIMARY KEY,
    risk_type TEXT NOT NULL, -- 'attrition', 'timeline_miss', 'quality_degradation', 'burnout'
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    risk_score REAL NOT NULL, -- 0.0-1.0
    probability_percentage REAL NOT NULL, -- 0-100
    impact_severity TEXT NOT NULL, -- 'low', 'medium', 'high', 'critical'
    contributing_factors TEXT, -- JSON: array of factors
    mitigation_suggestions TEXT, -- JSON: array of suggestions
    predicted_at TEXT NOT NULL,
    valid_until TEXT
);

-- Historical Accuracy: Track forecast accuracy
CREATE TABLE IF NOT EXISTS forecast_accuracy (
    id TEXT PRIMARY KEY,
    forecast_id TEXT NOT NULL,
    actual_value REAL NOT NULL,
    actual_date TEXT NOT NULL,
    error_amount REAL NOT NULL, -- actual - predicted
    error_percentage REAL NOT NULL,
    was_within_confidence_interval BOOLEAN NOT NULL,
    measured_at TEXT NOT NULL,
    FOREIGN KEY (forecast_id) REFERENCES forecasts(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_forecasts_entity ON forecasts(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_forecasts_type ON forecasts(forecast_type);
CREATE INDEX IF NOT EXISTS idx_forecasts_expires ON forecasts(expires_at);
CREATE INDEX IF NOT EXISTS idx_scenarios_entity ON scenarios(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_scenario_results_scenario ON scenario_results(scenario_id);
CREATE INDEX IF NOT EXISTS idx_risk_predictions_entity ON risk_predictions(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_risk_predictions_score ON risk_predictions(risk_score DESC);
CREATE INDEX IF NOT EXISTS idx_forecast_accuracy_forecast ON forecast_accuracy(forecast_id);
