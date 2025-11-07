package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/forecasting"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Forecast Handlers

func (s *Server) handleForecasts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	entityType := query.Get("entity_type")
	entityID := query.Get("entity_id")
	forecastType := query.Get("forecast_type")
	includeExpired := query.Get("include_expired") == "true"

	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	forecasts, total, err := s.forecastingStore.ListForecasts(entityType, entityID, forecastType, includeExpired, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list forecasts", err)
		return
	}

	respondPaginated(w, forecasts, total, limit, offset)
}

func (s *Server) handleForecastByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/forecasts/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Forecast ID required", http.StatusBadRequest)
		return
	}

	forecastID := parts[0]

	forecast, err := s.forecastingStore.GetForecast(forecastID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get forecast", err)
		return
	}

	if forecast == nil {
		http.Error(w, "Forecast not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, forecast)
}

func (s *Server) handleForecastSprintCompletion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/forecasts/sprint/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Sprint ID required", http.StatusBadRequest)
		return
	}

	sprintID := parts[0]

	forecast, err := s.forecastingService.ForecastSprintCompletion(sprintID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to forecast sprint completion", err)
		return
	}

	respondJSON(w, http.StatusCreated, forecast)
}

func (s *Server) handleForecastGoalCompletion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/forecasts/goal/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Goal ID required", http.StatusBadRequest)
		return
	}

	goalID := parts[0]

	forecast, err := s.forecastingService.ForecastGoalCompletion(goalID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to forecast goal completion", err)
		return
	}

	respondJSON(w, http.StatusCreated, forecast)
}

func (s *Server) handleForecastTeamVelocity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/forecasts/team/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Team ID required", http.StatusBadRequest)
		return
	}

	teamID := parts[0]

	forecast, err := s.forecastingService.ForecastTeamVelocity(teamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to forecast team velocity", err)
		return
	}

	respondJSON(w, http.StatusCreated, forecast)
}

// Risk Prediction Handlers

func (s *Server) handleRisks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	entityType := query.Get("entity_type")
	entityID := query.Get("entity_id")
	riskType := query.Get("risk_type")

	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	risks, total, err := s.forecastingStore.ListRiskPredictions(entityType, entityID, riskType, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list risk predictions", err)
		return
	}

	respondPaginated(w, risks, total, limit, offset)
}

func (s *Server) handleRiskEngineer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/risks/engineer/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Engineer ID required", http.StatusBadRequest)
		return
	}

	engineerID := parts[0]

	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	risks, total, err := s.forecastingStore.ListRiskPredictions("engineer", engineerID, "", limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get engineer risks", err)
		return
	}

	respondPaginated(w, risks, total, limit, offset)
}

func (s *Server) handleRiskTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/risks/team/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Team ID required", http.StatusBadRequest)
		return
	}

	teamID := parts[0]

	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	risks, total, err := s.forecastingStore.ListRiskPredictions("team", teamID, "", limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get team risks", err)
		return
	}

	respondPaginated(w, risks, total, limit, offset)
}

func (s *Server) handlePredictAttritionRisk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		EngineerID string `json:"engineer_id"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.EngineerID == "" {
		respondError(w, http.StatusBadRequest, "engineer_id is required", nil)
		return
	}

	risk, err := s.forecastingService.PredictAttritionRisk(req.EngineerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to predict attrition risk", err)
		return
	}

	// Store risk prediction
	if err := s.forecastingStore.CreateRiskPrediction(risk); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to store risk prediction", err)
		return
	}

	respondJSON(w, http.StatusCreated, risk)
}

func (s *Server) handlePredictTimelineMiss(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SprintID string `json:"sprint_id"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.SprintID == "" {
		respondError(w, http.StatusBadRequest, "sprint_id is required", nil)
		return
	}

	risk, err := s.forecastingService.PredictTimelineMiss(req.SprintID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to predict timeline miss", err)
		return
	}

	// Store risk prediction
	if err := s.forecastingStore.CreateRiskPrediction(risk); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to store risk prediction", err)
		return
	}

	respondJSON(w, http.StatusCreated, risk)
}

func (s *Server) handlePredictQualityDegradation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TeamID string `json:"team_id"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.TeamID == "" {
		respondError(w, http.StatusBadRequest, "team_id is required", nil)
		return
	}

	risk, err := s.forecastingService.PredictQualityDegradation(req.TeamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to predict quality degradation", err)
		return
	}

	// Store risk prediction
	if err := s.forecastingStore.CreateRiskPrediction(risk); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to store risk prediction", err)
		return
	}

	respondJSON(w, http.StatusCreated, risk)
}

// Scenario Handlers

func (s *Server) handleScenarios(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listScenarios(w, r)
	case http.MethodPost:
		s.createScenario(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listScenarios(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	entityType := query.Get("entity_type")
	entityID := query.Get("entity_id")

	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	scenarios, total, err := s.forecastingStore.ListScenarios(entityType, entityID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list scenarios", err)
		return
	}

	respondPaginated(w, scenarios, total, limit, offset)
}

func (s *Server) createScenario(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ScenarioName string                         `json:"scenario_name"`
		ScenarioType string                         `json:"scenario_type"`
		EntityType   string                         `json:"entity_type"`
		EntityID     string                         `json:"entity_id"`
		Description  string                         `json:"description"`
		CreatedBy    string                         `json:"created_by"`
		Parameters   forecasting.ScenarioParameters `json:"parameters"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.ScenarioName == "" {
		respondError(w, http.StatusBadRequest, "scenario_name is required", nil)
		return
	}

	if req.CreatedBy == "" {
		req.CreatedBy = "system"
	}

	scenario, err := s.forecastingService.CreateScenario(
		req.ScenarioName,
		req.ScenarioType,
		req.EntityType,
		req.EntityID,
		req.Description,
		req.CreatedBy,
		req.Parameters,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create scenario", err)
		return
	}

	respondJSON(w, http.StatusCreated, scenario)
}

func (s *Server) handleScenarioByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/scenarios/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Scenario ID required", http.StatusBadRequest)
		return
	}

	scenarioID := parts[0]

	scenario, err := s.forecastingStore.GetScenario(scenarioID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get scenario", err)
		return
	}

	if scenario == nil {
		http.Error(w, "Scenario not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, scenario)
}

func (s *Server) handleScenarioResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/scenarios/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Scenario ID required", http.StatusBadRequest)
		return
	}

	scenarioID := parts[0]

	results, err := s.forecastingStore.GetScenarioResults(scenarioID, 100)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get scenario results", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"scenario_id": scenarioID,
		"results":     results,
		"count":       len(results),
	})
}

func (s *Server) handleSimulateScenario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/scenarios/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Scenario ID required", http.StatusBadRequest)
		return
	}

	scenarioID := parts[0]

	results, err := s.forecastingService.SimulateScenario(scenarioID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to simulate scenario", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"scenario_id": scenarioID,
		"results":     results,
		"count":       len(results),
	})
}

func (s *Server) handleCompareScenarios(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ScenarioIDs []string `json:"scenario_ids"`
		OptimizeFor string   `json:"optimize_for"` // "cost", "speed", "velocity"
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if len(req.ScenarioIDs) == 0 {
		respondError(w, http.StatusBadRequest, "scenario_ids is required", nil)
		return
	}

	comparison, err := s.forecastingService.CompareScenarios(req.ScenarioIDs)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to compare scenarios", err)
		return
	}

	// Optionally find best scenario
	var recommendation string
	if req.OptimizeFor != "" {
		recommendation, err = s.forecastingService.RecommendBestScenario(req.ScenarioIDs, req.OptimizeFor)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to recommend best scenario", err)
			return
		}
	}

	response := map[string]interface{}{
		"comparison": comparison,
	}
	if recommendation != "" {
		response["recommended_scenario_id"] = recommendation
		response["optimized_for"] = req.OptimizeFor
	}

	respondJSON(w, http.StatusOK, response)
}

// Forecast Accuracy Handlers

func (s *Server) handleForecastAccuracy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For now, return placeholder accuracy metrics
	// In production, this would aggregate historical accuracy data
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"overall_accuracy": 85.0,
		"by_type": map[string]interface{}{
			"sprint_completion": 88.0,
			"goal_completion":   82.0,
			"team_velocity":     90.0,
			"attrition_risk":    75.0,
		},
		"total_forecasts":    150,
		"measured_forecasts": 75,
	})
}

func (s *Server) handleRecordForecastAccuracy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ForecastID  string  `json:"forecast_id"`
		ActualValue float64 `json:"actual_value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get original forecast
	forecast, err := s.forecastingStore.GetForecast(req.ForecastID)
	if err != nil || forecast == nil {
		respondError(w, http.StatusNotFound, "Forecast not found", err)
		return
	}

	// Calculate accuracy metrics
	errorAmount := req.ActualValue - forecast.PredictedValue
	errorPercentage := (errorAmount / forecast.PredictedValue) * 100

	wasWithinCI := false
	if forecast.ConfidenceIntervalLow != nil && forecast.ConfidenceIntervalHigh != nil {
		wasWithinCI = req.ActualValue >= *forecast.ConfidenceIntervalLow &&
			req.ActualValue <= *forecast.ConfidenceIntervalHigh
	}

	// Record accuracy
	accuracy, err := s.forecastingStore.GetAccuracyByForecastID(req.ForecastID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check existing accuracy", err)
		return
	}

	if accuracy == nil {
		// Create new accuracy record
		err = s.forecastingStore.CreateForecastAccuracy(&models.ForecastAccuracy{
			ForecastID:                  req.ForecastID,
			ActualValue:                 req.ActualValue,
			ActualDate:                  time.Now().UTC(),
			ErrorAmount:                 errorAmount,
			ErrorPercentage:             errorPercentage,
			WasWithinConfidenceInterval: wasWithinCI,
			MeasuredAt:                  time.Now().UTC(),
		})
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to record accuracy", err)
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"forecast_id":                req.ForecastID,
		"actual_value":               req.ActualValue,
		"predicted_value":            forecast.PredictedValue,
		"error_amount":               errorAmount,
		"error_percentage":           errorPercentage,
		"within_confidence_interval": wasWithinCI,
	})
}
