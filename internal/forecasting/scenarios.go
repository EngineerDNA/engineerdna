package forecasting

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/internal/utils"
	"github.com/google/uuid"
)

// ScenarioParameters represents the parameters for a what-if scenario
type ScenarioParameters struct {
	// Capacity changes
	NewEngineers     int    `json:"new_engineers,omitempty"`
	RemovedEngineers int    `json:"removed_engineers,omitempty"`
	EngineerRole     string `json:"engineer_role,omitempty"` // "junior", "mid", "senior"
	RampWeeks        int    `json:"ramp_weeks,omitempty"`

	// Timeline changes
	ExtendDays  int    `json:"extend_days,omitempty"`
	ShortenDays int    `json:"shortened_days,omitempty"`
	NewDeadline string `json:"new_deadline,omitempty"`

	// Scope changes
	ScopeIncrease float64 `json:"scope_increase,omitempty"` // percentage
	ScopeDecrease float64 `json:"scope_decrease,omitempty"` // percentage

	// Cost changes
	BudgetIncrease float64 `json:"budget_increase,omitempty"` // dollars
	BudgetCut      float64 `json:"budget_cut,omitempty"`      // dollars
}

// CreateScenario creates a new what-if scenario
func (s *ForecastingService) CreateScenario(scenarioName, scenarioType, entityType, entityID, description, createdBy string, parameters ScenarioParameters) (*models.Scenario, error) {
	parametersJSON, err := json.Marshal(parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	scenario := &models.Scenario{
		ID:           uuid.New().String(),
		ScenarioName: scenarioName,
		ScenarioType: scenarioType,
		EntityType:   entityType,
		EntityID:     &entityID,
		Description:  description,
		Parameters:   string(parametersJSON),
		CreatedBy:    createdBy,
		CreatedAt:    time.Now().UTC(),
	}

	// Get baseline forecast if applicable
	if entityType == "sprint" || entityType == "goal" || entityType == "team" {
		forecasts, _, err := s.forecastingStore.ListForecasts(entityType, entityID, "", false, 1, 0)
		if err == nil && len(forecasts) > 0 {
			scenario.BaselineForecastID = &forecasts[0].ID
		}
	}

	if err := s.forecastingStore.CreateScenario(scenario); err != nil {
		return nil, fmt.Errorf("failed to create scenario: %w", err)
	}

	return scenario, nil
}

// SimulateScenario computes the outcomes of a scenario
func (s *ForecastingService) SimulateScenario(scenarioID string) ([]*models.ScenarioResult, error) {
	scenario, err := s.forecastingStore.GetScenario(scenarioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scenario: %w", err)
	}
	if scenario == nil {
		return nil, fmt.Errorf("scenario not found: %s", scenarioID)
	}

	var params ScenarioParameters
	if err := json.Unmarshal([]byte(scenario.Parameters), &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	// Get baseline forecast
	var baseline *models.Forecast
	if scenario.BaselineForecastID != nil {
		baseline, err = s.forecastingStore.GetForecast(*scenario.BaselineForecastID)
		if err != nil {
			return nil, fmt.Errorf("failed to get baseline forecast: %w", err)
		}
	}

	var results []*models.ScenarioResult

	switch scenario.ScenarioType {
	case "capacity_change":
		results, err = s.simulateCapacityChange(scenario, params, baseline)
	case "timeline_adjustment":
		results, err = s.simulateTimelineAdjustment(scenario, params, baseline)
	case "cost_optimization":
		results, err = s.simulateCostOptimization(scenario, params, baseline)
	default:
		return nil, fmt.Errorf("unknown scenario type: %s", scenario.ScenarioType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to simulate scenario: %w", err)
	}

	// Store results
	for _, result := range results {
		if err := s.forecastingStore.CreateScenarioResult(result); err != nil {
			return nil, fmt.Errorf("failed to store scenario result: %w", err)
		}
	}

	return results, nil
}

// simulateCapacityChange simulates adding or removing engineers
func (s *ForecastingService) simulateCapacityChange(scenario *models.Scenario, params ScenarioParameters, baseline *models.Forecast) ([]*models.ScenarioResult, error) {
	results := []*models.ScenarioResult{}

	// Get current team velocity
	teamID := ""
	if scenario.EntityID != nil {
		teamID = *scenario.EntityID
	}

	currentVelocity := s.calculateCurrentTeamVelocity(teamID)
	if currentVelocity == 0 {
		currentVelocity = DefaultVelocity
	}

	// Calculate velocity impact
	newVelocity := currentVelocity

	if params.NewEngineers > 0 {
		// New engineers add capacity but with ramp time
		rampWeeks := params.RampWeeks
		if rampWeeks == 0 {
			rampWeeks = DefaultRampWeeks
		}

		// Role-based productivity
		productivity := ProductivityMid
		switch params.EngineerRole {
		case "junior":
			productivity = ProductivityJunior
		case "mid":
			productivity = ProductivityMid
		case "senior":
			productivity = ProductivitySenior
		}

		// During ramp: 50% productivity
		rampVelocity := currentVelocity + (float64(params.NewEngineers) * productivity * ProductivityDuringRamp)
		// Post-ramp: full productivity
		fullVelocity := currentVelocity + (float64(params.NewEngineers) * productivity)

		newVelocity = fullVelocity

		results = append(results, &models.ScenarioResult{
			ID:                     uuid.New().String(),
			ScenarioID:             scenario.ID,
			MetricName:             "velocity_during_ramp",
			PredictedValue:         rampVelocity,
			DifferenceFromBaseline: utils.PtrFloat64(rampVelocity - currentVelocity),
			Impact:                 determineImpact(rampVelocity, currentVelocity),
			ComputedAt:             time.Now().UTC(),
		})
	}

	if params.RemovedEngineers > 0 {
		// Removing engineers reduces capacity
		reductionFactor := float64(params.RemovedEngineers) / DefaultTeamSize
		newVelocity = currentVelocity * (1.0 - reductionFactor)
	}

	// Velocity result
	results = append(results, &models.ScenarioResult{
		ID:                     uuid.New().String(),
		ScenarioID:             scenario.ID,
		MetricName:             "new_velocity",
		PredictedValue:         newVelocity,
		DifferenceFromBaseline: utils.PtrFloat64(newVelocity - currentVelocity),
		Impact:                 determineImpact(newVelocity, currentVelocity),
		ComputedAt:             time.Now().UTC(),
	})

	// Cost result
	costPerEngineer := DefaultCostPerEngineer
	newCost := 0.0
	currentCost := 0.0

	if params.NewEngineers > 0 {
		newCost = float64(params.NewEngineers) * costPerEngineer
	}
	if params.RemovedEngineers > 0 {
		newCost = -float64(params.RemovedEngineers) * costPerEngineer
	}

	results = append(results, &models.ScenarioResult{
		ID:                     uuid.New().String(),
		ScenarioID:             scenario.ID,
		MetricName:             "monthly_cost_delta",
		PredictedValue:         newCost,
		DifferenceFromBaseline: utils.PtrFloat64(newCost - currentCost),
		Impact:                 determineImpact(currentCost, newCost), // Lower cost is better
		ComputedAt:             time.Now().UTC(),
	})

	// Completion time result
	if baseline != nil && baseline.PredictedDate != nil {
		// Recalculate completion date with new velocity
		daysNeeded := (baseline.PredictedValue / currentVelocity) * SprintDurationDays
		newDaysNeeded := (baseline.PredictedValue / newVelocity) * SprintDurationDays

		results = append(results, &models.ScenarioResult{
			ID:                     uuid.New().String(),
			ScenarioID:             scenario.ID,
			MetricName:             "days_to_completion",
			PredictedValue:         newDaysNeeded,
			DifferenceFromBaseline: utils.PtrFloat64(newDaysNeeded - daysNeeded),
			Impact:                 determineImpact(daysNeeded, newDaysNeeded), // Fewer days is better
			ComputedAt:             time.Now().UTC(),
		})
	}

	return results, nil
}

// simulateTimelineAdjustment simulates changing deadlines
func (s *ForecastingService) simulateTimelineAdjustment(scenario *models.Scenario, params ScenarioParameters, baseline *models.Forecast) ([]*models.ScenarioResult, error) {
	results := []*models.ScenarioResult{}

	if baseline == nil {
		return results, fmt.Errorf("timeline adjustment requires baseline forecast")
	}

	currentDays := 0.0
	if baseline.PredictedDate != nil {
		currentDays = time.Until(*baseline.PredictedDate).Hours() / 24
	}

	newDays := currentDays
	if params.ExtendDays > 0 {
		newDays += float64(params.ExtendDays)
	}
	if params.ShortenDays > 0 {
		newDays -= float64(params.ShortenDays)
	}

	results = append(results, &models.ScenarioResult{
		ID:                     uuid.New().String(),
		ScenarioID:             scenario.ID,
		MetricName:             "new_timeline_days",
		PredictedValue:         newDays,
		DifferenceFromBaseline: utils.PtrFloat64(newDays - currentDays),
		Impact:                 determineImpact(currentDays, newDays),
		ComputedAt:             time.Now().UTC(),
	})

	// Calculate required velocity change
	requiredVelocity := baseline.PredictedValue / newDays * SprintDurationDays
	currentVelocity := baseline.PredictedValue / currentDays * SprintDurationDays

	results = append(results, &models.ScenarioResult{
		ID:                     uuid.New().String(),
		ScenarioID:             scenario.ID,
		MetricName:             "required_velocity",
		PredictedValue:         requiredVelocity,
		DifferenceFromBaseline: utils.PtrFloat64(requiredVelocity - currentVelocity),
		Impact:                 determineImpact(currentVelocity, requiredVelocity),
		ComputedAt:             time.Now().UTC(),
	})

	// Calculate feasibility
	teamID := ""
	if scenario.EntityID != nil {
		teamID = *scenario.EntityID
	}
	maxVelocity := s.calculateCurrentTeamVelocity(teamID) * MaxVelocityStretchFactor

	feasibility := DefaultFeasibility
	if requiredVelocity > maxVelocity {
		feasibility = (maxVelocity / requiredVelocity) * MaxProgressPercent
	}

	results = append(results, &models.ScenarioResult{
		ID:                     uuid.New().String(),
		ScenarioID:             scenario.ID,
		MetricName:             "feasibility_percentage",
		PredictedValue:         feasibility,
		DifferenceFromBaseline: utils.PtrFloat64(feasibility - DefaultFeasibility),
		Impact:                 determineImpact(DefaultFeasibility, feasibility),
		ComputedAt:             time.Now().UTC(),
	})

	return results, nil
}

// simulateCostOptimization simulates cost changes
func (s *ForecastingService) simulateCostOptimization(scenario *models.Scenario, params ScenarioParameters, baseline *models.Forecast) ([]*models.ScenarioResult, error) {
	results := []*models.ScenarioResult{}

	currentBudget := DefaultBudget
	newBudget := currentBudget

	if params.BudgetIncrease > 0 {
		newBudget += params.BudgetIncrease
	}
	if params.BudgetCut > 0 {
		newBudget -= params.BudgetCut
	}

	results = append(results, &models.ScenarioResult{
		ID:                     uuid.New().String(),
		ScenarioID:             scenario.ID,
		MetricName:             "new_budget",
		PredictedValue:         newBudget,
		DifferenceFromBaseline: utils.PtrFloat64(newBudget - currentBudget),
		Impact:                 determineImpact(currentBudget, newBudget),
		ComputedAt:             time.Now().UTC(),
	})

	// Calculate impact on delivery
	budgetRatio := newBudget / currentBudget
	capacityImpact := (budgetRatio - 1) * EfficiencyFactor

	results = append(results, &models.ScenarioResult{
		ID:                     uuid.New().String(),
		ScenarioID:             scenario.ID,
		MetricName:             "capacity_change_percentage",
		PredictedValue:         capacityImpact * MaxProgressPercent,
		DifferenceFromBaseline: utils.PtrFloat64(capacityImpact * MaxProgressPercent),
		Impact:                 determineImpact(0, capacityImpact*MaxProgressPercent),
		ComputedAt:             time.Now().UTC(),
	})

	return results, nil
}

// CompareScenarios compares multiple scenarios side by side
func (s *ForecastingService) CompareScenarios(scenarioIDs []string) (map[string][]*models.ScenarioResult, error) {
	comparison := make(map[string][]*models.ScenarioResult)

	for _, id := range scenarioIDs {
		results, err := s.forecastingStore.GetScenarioResults(id, 100)
		if err != nil {
			return nil, fmt.Errorf("failed to get results for scenario %s: %w", id, err)
		}
		comparison[id] = results
	}

	return comparison, nil
}

// RecommendBestScenario identifies the optimal scenario based on criteria
func (s *ForecastingService) RecommendBestScenario(scenarioIDs []string, optimizeFor string) (string, error) {
	comparison, err := s.CompareScenarios(scenarioIDs)
	if err != nil {
		return "", err
	}

	bestScenario := ""
	bestScore := -999999.0

	for scenarioID, results := range comparison {
		score := 0.0

		for _, result := range results {
			switch optimizeFor {
			case "cost":
				if result.MetricName == "monthly_cost_delta" {
					score -= result.PredictedValue // Lower cost is better
				}
			case "speed":
				if result.MetricName == "days_to_completion" {
					score -= result.PredictedValue // Fewer days is better
				}
			case "velocity":
				if result.MetricName == "new_velocity" {
					score += result.PredictedValue // Higher velocity is better
				}
			}
		}

		if score > bestScore {
			bestScore = score
			bestScenario = scenarioID
		}
	}

	return bestScenario, nil
}

// Helper functions

func (s *ForecastingService) calculateCurrentTeamVelocity(teamID string) float64 {
	velocities := s.getRecentVelocities(teamID, 3)
	if len(velocities) == 0 {
		return DefaultVelocity
	}
	return calculateMovingAverage(velocities)
}

func determineImpact(baseline, predicted float64) string {
	if predicted > baseline {
		return "positive"
	} else if predicted < baseline {
		return "negative"
	}
	return "neutral"
}
