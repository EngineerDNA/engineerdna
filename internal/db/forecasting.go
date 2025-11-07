package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

type ForecastingStore struct {
	db *sql.DB
}

func NewForecastingStore(db *sql.DB) *ForecastingStore {
	return &ForecastingStore{db: db}
}

// Forecast CRUD operations

func (s *ForecastingStore) CreateForecast(forecast *models.Forecast) error {
	if forecast.ID == "" {
		forecast.ID = uuid.New().String()
	}
	if forecast.CreatedAt.IsZero() {
		forecast.CreatedAt = time.Now().UTC()
	}

	var predictedDate, expiresAt *string
	if forecast.PredictedDate != nil {
		pd := forecast.PredictedDate.UTC().Format(time.RFC3339)
		predictedDate = &pd
	}
	if forecast.ExpiresAt != nil {
		ea := forecast.ExpiresAt.UTC().Format(time.RFC3339)
		expiresAt = &ea
	}

	_, err := s.db.Exec(`
		INSERT INTO forecasts (
			id, forecast_type, entity_type, entity_id, time_horizon,
			predicted_value, predicted_date, confidence_percentage,
			confidence_interval_low, confidence_interval_high, model_type,
			input_data, assumptions, created_at, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, forecast.ID, forecast.ForecastType, forecast.EntityType, forecast.EntityID,
		forecast.TimeHorizon, forecast.PredictedValue, predictedDate,
		forecast.ConfidencePercentage, forecast.ConfidenceIntervalLow,
		forecast.ConfidenceIntervalHigh, forecast.ModelType, forecast.InputData,
		forecast.Assumptions, forecast.CreatedAt.UTC().Format(time.RFC3339), expiresAt)

	if err != nil {
		return fmt.Errorf("failed to create forecast: %w", err)
	}

	return nil
}

func (s *ForecastingStore) GetForecast(id string) (*models.Forecast, error) {
	var forecast models.Forecast
	var predictedDate, expiresAt sql.NullString

	err := s.db.QueryRow(`
		SELECT id, forecast_type, entity_type, entity_id, time_horizon,
		       predicted_value, predicted_date, confidence_percentage,
		       confidence_interval_low, confidence_interval_high, model_type,
		       input_data, assumptions, created_at, expires_at
		FROM forecasts WHERE id = ?
	`, id).Scan(
		&forecast.ID, &forecast.ForecastType, &forecast.EntityType, &forecast.EntityID,
		&forecast.TimeHorizon, &forecast.PredictedValue, &predictedDate,
		&forecast.ConfidencePercentage, &forecast.ConfidenceIntervalLow,
		&forecast.ConfidenceIntervalHigh, &forecast.ModelType, &forecast.InputData,
		&forecast.Assumptions, &forecast.CreatedAt, &expiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get forecast: %w", err)
	}

	if predictedDate.Valid {
		pd, _ := time.Parse(time.RFC3339, predictedDate.String)
		forecast.PredictedDate = &pd
	}
	if expiresAt.Valid {
		ea, _ := time.Parse(time.RFC3339, expiresAt.String)
		forecast.ExpiresAt = &ea
	}

	return &forecast, nil
}

func (s *ForecastingStore) ListForecasts(entityType, entityID, forecastType string, includeExpired bool, limit, offset int) ([]*models.Forecast, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Build WHERE clause
	whereClause := " WHERE 1=1"
	args := []interface{}{}

	if entityType != "" {
		whereClause += " AND entity_type = ?"
		args = append(args, entityType)
	}
	if entityID != "" {
		whereClause += " AND entity_id = ?"
		args = append(args, entityID)
	}
	if forecastType != "" {
		whereClause += " AND forecast_type = ?"
		args = append(args, forecastType)
	}
	if !includeExpired {
		whereClause += " AND (expires_at IS NULL OR expires_at > ?)"
		args = append(args, time.Now().UTC().Format(time.RFC3339))
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM forecasts" + whereClause
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count forecasts: %w", err)
	}

	// Query with pagination
	query := `
		SELECT id, forecast_type, entity_type, entity_id, time_horizon,
		       predicted_value, predicted_date, confidence_percentage,
		       confidence_interval_low, confidence_interval_high, model_type,
		       input_data, assumptions, created_at, expires_at
		FROM forecasts` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	paginatedArgs := append(args, limit, offset)

	rows, err := s.db.Query(query, paginatedArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list forecasts: %w", err)
	}
	defer rows.Close()

	var forecasts []*models.Forecast
	for rows.Next() {
		var forecast models.Forecast
		var predictedDate, expiresAt sql.NullString

		err := rows.Scan(
			&forecast.ID, &forecast.ForecastType, &forecast.EntityType, &forecast.EntityID,
			&forecast.TimeHorizon, &forecast.PredictedValue, &predictedDate,
			&forecast.ConfidencePercentage, &forecast.ConfidenceIntervalLow,
			&forecast.ConfidenceIntervalHigh, &forecast.ModelType, &forecast.InputData,
			&forecast.Assumptions, &forecast.CreatedAt, &expiresAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan forecast: %w", err)
		}

		if predictedDate.Valid {
			pd, _ := time.Parse(time.RFC3339, predictedDate.String)
			forecast.PredictedDate = &pd
		}
		if expiresAt.Valid {
			ea, _ := time.Parse(time.RFC3339, expiresAt.String)
			forecast.ExpiresAt = &ea
		}

		forecasts = append(forecasts, &forecast)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating forecasts: %w", err)
	}

	return forecasts, total, nil
}

// Scenario CRUD operations

func (s *ForecastingStore) CreateScenario(scenario *models.Scenario) error {
	if scenario.ID == "" {
		scenario.ID = uuid.New().String()
	}
	if scenario.CreatedAt.IsZero() {
		scenario.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO scenarios (
			id, scenario_name, scenario_type, entity_type, entity_id,
			description, parameters, baseline_forecast_id, created_by, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, scenario.ID, scenario.ScenarioName, scenario.ScenarioType, scenario.EntityType,
		scenario.EntityID, scenario.Description, scenario.Parameters,
		scenario.BaselineForecastID, scenario.CreatedBy,
		scenario.CreatedAt.UTC().Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("failed to create scenario: %w", err)
	}

	return nil
}

func (s *ForecastingStore) GetScenario(id string) (*models.Scenario, error) {
	var scenario models.Scenario

	err := s.db.QueryRow(`
		SELECT id, scenario_name, scenario_type, entity_type, entity_id,
		       description, parameters, baseline_forecast_id, created_by, created_at
		FROM scenarios WHERE id = ?
	`, id).Scan(
		&scenario.ID, &scenario.ScenarioName, &scenario.ScenarioType,
		&scenario.EntityType, &scenario.EntityID, &scenario.Description,
		&scenario.Parameters, &scenario.BaselineForecastID, &scenario.CreatedBy,
		&scenario.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get scenario: %w", err)
	}

	return &scenario, nil
}

func (s *ForecastingStore) ListScenarios(entityType, entityID string, limit, offset int) ([]*models.Scenario, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Build WHERE clause
	whereClause := " WHERE 1=1"
	args := []interface{}{}

	if entityType != "" {
		whereClause += " AND entity_type = ?"
		args = append(args, entityType)
	}
	if entityID != "" {
		whereClause += " AND entity_id = ?"
		args = append(args, entityID)
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM scenarios" + whereClause
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count scenarios: %w", err)
	}

	// Query with pagination
	query := `
		SELECT id, scenario_name, scenario_type, entity_type, entity_id,
		       description, parameters, baseline_forecast_id, created_by, created_at
		FROM scenarios` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	paginatedArgs := append(args, limit, offset)

	rows, err := s.db.Query(query, paginatedArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list scenarios: %w", err)
	}
	defer rows.Close()

	var scenarios []*models.Scenario
	for rows.Next() {
		var scenario models.Scenario
		err := rows.Scan(
			&scenario.ID, &scenario.ScenarioName, &scenario.ScenarioType,
			&scenario.EntityType, &scenario.EntityID, &scenario.Description,
			&scenario.Parameters, &scenario.BaselineForecastID, &scenario.CreatedBy,
			&scenario.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan scenario: %w", err)
		}
		scenarios = append(scenarios, &scenario)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating scenarios: %w", err)
	}

	return scenarios, total, nil
}

// ScenarioResult operations

func (s *ForecastingStore) CreateScenarioResult(result *models.ScenarioResult) error {
	if result.ID == "" {
		result.ID = uuid.New().String()
	}
	if result.ComputedAt.IsZero() {
		result.ComputedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO scenario_results (
			id, scenario_id, metric_name, predicted_value,
			difference_from_baseline, impact, computed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, result.ID, result.ScenarioID, result.MetricName, result.PredictedValue,
		result.DifferenceFromBaseline, result.Impact,
		result.ComputedAt.UTC().Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("failed to create scenario result: %w", err)
	}

	return nil
}

func (s *ForecastingStore) GetScenarioResults(scenarioID string, limit int) ([]*models.ScenarioResult, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	rows, err := s.db.Query(`
		SELECT id, scenario_id, metric_name, predicted_value,
		       difference_from_baseline, impact, computed_at
		FROM scenario_results
		WHERE scenario_id = ?
		ORDER BY computed_at DESC
		LIMIT ?
	`, scenarioID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get scenario results: %w", err)
	}
	defer rows.Close()

	var results []*models.ScenarioResult
	for rows.Next() {
		var result models.ScenarioResult
		err := rows.Scan(
			&result.ID, &result.ScenarioID, &result.MetricName, &result.PredictedValue,
			&result.DifferenceFromBaseline, &result.Impact, &result.ComputedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scenario result: %w", err)
		}
		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating scenario results: %w", err)
	}

	return results, nil
}

// RiskPrediction operations

func (s *ForecastingStore) CreateRiskPrediction(risk *models.RiskPrediction) error {
	if risk.ID == "" {
		risk.ID = uuid.New().String()
	}
	if risk.PredictedAt.IsZero() {
		risk.PredictedAt = time.Now().UTC()
	}

	var validUntil *string
	if risk.ValidUntil != nil {
		vu := risk.ValidUntil.UTC().Format(time.RFC3339)
		validUntil = &vu
	}

	_, err := s.db.Exec(`
		INSERT INTO risk_predictions (
			id, risk_type, entity_type, entity_id, risk_score,
			probability_percentage, impact_severity, contributing_factors,
			mitigation_suggestions, predicted_at, valid_until
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, risk.ID, risk.RiskType, risk.EntityType, risk.EntityID, risk.RiskScore,
		risk.ProbabilityPercentage, risk.ImpactSeverity, risk.ContributingFactors,
		risk.MitigationSuggestions, risk.PredictedAt.UTC().Format(time.RFC3339),
		validUntil)

	if err != nil {
		return fmt.Errorf("failed to create risk prediction: %w", err)
	}

	return nil
}

func (s *ForecastingStore) ListRiskPredictions(entityType, entityID, riskType string, limit, offset int) ([]*models.RiskPrediction, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Build WHERE clause
	whereClause := " WHERE 1=1"
	args := []interface{}{}

	if entityType != "" {
		whereClause += " AND entity_type = ?"
		args = append(args, entityType)
	}
	if entityID != "" {
		whereClause += " AND entity_id = ?"
		args = append(args, entityID)
	}
	if riskType != "" {
		whereClause += " AND risk_type = ?"
		args = append(args, riskType)
	}

	whereClause += " AND (valid_until IS NULL OR valid_until > ?)"
	args = append(args, time.Now().UTC().Format(time.RFC3339))

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM risk_predictions" + whereClause
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count risk predictions: %w", err)
	}

	// Query with pagination
	query := `
		SELECT id, risk_type, entity_type, entity_id, risk_score,
		       probability_percentage, impact_severity, contributing_factors,
		       mitigation_suggestions, predicted_at, valid_until
		FROM risk_predictions` + whereClause + `
		ORDER BY risk_score DESC
		LIMIT ? OFFSET ?
	`
	paginatedArgs := append(args, limit, offset)

	rows, err := s.db.Query(query, paginatedArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list risk predictions: %w", err)
	}
	defer rows.Close()

	var risks []*models.RiskPrediction
	for rows.Next() {
		var risk models.RiskPrediction
		var validUntil sql.NullString

		err := rows.Scan(
			&risk.ID, &risk.RiskType, &risk.EntityType, &risk.EntityID, &risk.RiskScore,
			&risk.ProbabilityPercentage, &risk.ImpactSeverity, &risk.ContributingFactors,
			&risk.MitigationSuggestions, &risk.PredictedAt, &validUntil,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan risk prediction: %w", err)
		}

		if validUntil.Valid {
			vu, _ := time.Parse(time.RFC3339, validUntil.String)
			risk.ValidUntil = &vu
		}

		risks = append(risks, &risk)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating risk predictions: %w", err)
	}

	return risks, total, nil
}

// ForecastAccuracy operations

func (s *ForecastingStore) CreateForecastAccuracy(accuracy *models.ForecastAccuracy) error {
	if accuracy.ID == "" {
		accuracy.ID = uuid.New().String()
	}
	if accuracy.MeasuredAt.IsZero() {
		accuracy.MeasuredAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO forecast_accuracy (
			id, forecast_id, actual_value, actual_date, error_amount,
			error_percentage, was_within_confidence_interval, measured_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, accuracy.ID, accuracy.ForecastID, accuracy.ActualValue,
		accuracy.ActualDate.UTC().Format(time.RFC3339), accuracy.ErrorAmount,
		accuracy.ErrorPercentage, accuracy.WasWithinConfidenceInterval,
		accuracy.MeasuredAt.UTC().Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("failed to create forecast accuracy: %w", err)
	}

	return nil
}

func (s *ForecastingStore) GetAccuracyByForecastID(forecastID string) (*models.ForecastAccuracy, error) {
	var accuracy models.ForecastAccuracy

	err := s.db.QueryRow(`
		SELECT id, forecast_id, actual_value, actual_date, error_amount,
		       error_percentage, was_within_confidence_interval, measured_at
		FROM forecast_accuracy
		WHERE forecast_id = ?
	`, forecastID).Scan(
		&accuracy.ID, &accuracy.ForecastID, &accuracy.ActualValue,
		&accuracy.ActualDate, &accuracy.ErrorAmount, &accuracy.ErrorPercentage,
		&accuracy.WasWithinConfidenceInterval, &accuracy.MeasuredAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get forecast accuracy: %w", err)
	}

	return &accuracy, nil
}
