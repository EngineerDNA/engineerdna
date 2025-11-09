package forecasting

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// ForecastingService handles predictive analytics and forecasting
type ForecastingService struct {
	db               *sql.DB
	forecastingStore *db.ForecastingStore
	planningStore    *db.PlanningStore
	scoringStore     *db.ScoringStore
	metricStore      *db.MetricStore
	goalsStore       *db.GoalsStore
}

// NewForecastingService creates a new forecasting service
func NewForecastingService(database *sql.DB, forecastingStore *db.ForecastingStore, planningStore *db.PlanningStore, scoringStore *db.ScoringStore, metricStore *db.MetricStore, goalsStore *db.GoalsStore) *ForecastingService {
	return &ForecastingService{
		db:               database,
		forecastingStore: forecastingStore,
		planningStore:    planningStore,
		scoringStore:     scoringStore,
		metricStore:      metricStore,
		goalsStore:       goalsStore,
	}
}

// ForecastSprintCompletion predicts when a sprint will complete
func (s *ForecastingService) ForecastSprintCompletion(sprintID string) (*models.Forecast, error) {
	sprint, err := s.planningStore.GetSprint(sprintID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sprint: %w", err)
	}
	if sprint == nil {
		return nil, fmt.Errorf("sprint not found: %s", sprintID)
	}

	// Calculate current progress
	completedPoints := s.calculateCompletedPoints(sprintID)
	totalPoints := float64(sprint.CommittedPoints)

	if totalPoints == 0 {
		return nil, fmt.Errorf("sprint has no committed points")
	}

	// Calculate current pace
	daysElapsed := time.Since(sprint.StartDate).Hours() / 24
	if daysElapsed <= 0 {
		daysElapsed = MinimumDaysElapsed
	}

	currentPace := completedPoints / daysElapsed
	if currentPace == 0 {
		currentPace = MinimumPace
	}

	// Predict remaining time
	pointsRemaining := totalPoints - completedPoints
	daysNeeded := pointsRemaining / currentPace

	predictedDate := time.Now().UTC().Add(time.Duration(daysNeeded*24) * time.Hour)

	// Calculate confidence based on historical variance
	historicalVariance := s.calculateSprintVariance(sprint.TeamID)
	confidencePct := MaxProgressPercent - (historicalVariance * 10) // Higher variance = lower confidence
	if confidencePct < ConfidenceMinimum {
		confidencePct = ConfidenceMinimum
	}

	// Monte Carlo for confidence interval
	low, high := s.monteCarloSprintSimulation(sprint, completedPoints, currentPace, DefaultMonteCarloSimulations)

	// Prepare input data
	inputData := map[string]interface{}{
		"sprint_id":        sprintID,
		"completed_points": completedPoints,
		"total_points":     totalPoints,
		"days_elapsed":     daysElapsed,
		"current_pace":     currentPace,
		"points_remaining": pointsRemaining,
	}
	inputDataJSON, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input data: %w", err)
	}

	forecast := &models.Forecast{
		ID:                     uuid.New().String(),
		ForecastType:           "sprint_completion",
		EntityType:             "sprint",
		EntityID:               &sprintID,
		TimeHorizon:            "end_of_sprint",
		PredictedValue:         completedPoints + (pointsRemaining * (currentPace / currentPace)),
		PredictedDate:          &predictedDate,
		ConfidencePercentage:   confidencePct,
		ConfidenceIntervalLow:  &low,
		ConfidenceIntervalHigh: &high,
		ModelType:              "monte_carlo",
		InputData:              string(inputDataJSON),
		CreatedAt:              time.Now().UTC(),
		ExpiresAt:              &sprint.EndDate,
	}

	if err := s.forecastingStore.CreateForecast(forecast); err != nil {
		return nil, fmt.Errorf("failed to store forecast: %w", err)
	}

	return forecast, nil
}

// ForecastGoalCompletion predicts when a goal will be achieved
func (s *ForecastingService) ForecastGoalCompletion(goalID string) (*models.Forecast, error) {
	goal, err := s.goalsStore.GetGoal(goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}
	if goal == nil {
		return nil, fmt.Errorf("goal not found: %s", goalID)
	}

	currentProgress := float64(goal.ProgressPercentage)
	targetProgress := 100.0

	// Calculate rate of progress
	daysElapsed := time.Since(goal.StartDate).Hours() / 24
	if daysElapsed <= 0 {
		daysElapsed = MinimumDaysElapsed
	}

	progressRate := currentProgress / daysElapsed
	if progressRate == 0 {
		progressRate = MinimumProgressRate
	}

	// Predict completion date
	progressRemaining := targetProgress - currentProgress
	daysNeeded := progressRemaining / progressRate

	predictedDate := time.Now().UTC().Add(time.Duration(daysNeeded*24) * time.Hour)

	// Calculate confidence
	confidencePct := ConfidenceMedium
	if goal.Status == "at_risk" {
		confidencePct = ConfidenceLow
	} else if goal.Status == "off_track" {
		confidencePct = ConfidenceDefault
	}

	// Confidence interval (±20%)
	low := currentProgress + (progressRemaining * ConfidenceIntervalLowMultiplier)
	high := currentProgress + (progressRemaining * ConfidenceIntervalHighMultiplier)

	inputData := map[string]interface{}{
		"goal_id":            goalID,
		"current_progress":   currentProgress,
		"days_elapsed":       daysElapsed,
		"progress_rate":      progressRate,
		"progress_remaining": progressRemaining,
	}
	inputDataJSON, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input data: %w", err)
	}

	forecast := &models.Forecast{
		ID:                     uuid.New().String(),
		ForecastType:           "goal_completion",
		EntityType:             "goal",
		EntityID:               &goalID,
		TimeHorizon:            "goal_end_date",
		PredictedValue:         targetProgress,
		PredictedDate:          &predictedDate,
		ConfidencePercentage:   confidencePct,
		ConfidenceIntervalLow:  &low,
		ConfidenceIntervalHigh: &high,
		ModelType:              "linear_regression",
		InputData:              string(inputDataJSON),
		CreatedAt:              time.Now().UTC(),
		ExpiresAt:              &goal.EndDate,
	}

	if err := s.forecastingStore.CreateForecast(forecast); err != nil {
		return nil, fmt.Errorf("failed to store forecast: %w", err)
	}

	return forecast, nil
}

// ForecastTeamVelocity predicts team velocity for next sprint
func (s *ForecastingService) ForecastTeamVelocity(teamID string) (*models.Forecast, error) {
	// Get recent sprint velocities
	velocities := s.getRecentVelocities(teamID, DefaultHistoricalSprintCount)
	if len(velocities) == 0 {
		return nil, fmt.Errorf("no historical velocity data for team %s", teamID)
	}

	// Calculate moving average
	avgVelocity := calculateMovingAverage(velocities)

	// Calculate trend
	trend := calculateTrend(velocities)

	// Predict next sprint velocity
	predictedVelocity := avgVelocity + trend

	// Calculate confidence based on variance
	variance := calculateVariance(velocities)
	stdDev := calculateStdDev(variance)
	confidencePct := ConfidenceHigh - (stdDev * ConfidenceStdDevFactor)
	if confidencePct < ConfidenceMinimum {
		confidencePct = ConfidenceMinimum
	}

	// Confidence interval (±1 std dev)
	low := predictedVelocity - stdDev
	high := predictedVelocity + stdDev

	inputData := map[string]interface{}{
		"team_id":               teamID,
		"historical_velocities": velocities,
		"average_velocity":      avgVelocity,
		"trend":                 trend,
		"variance":              variance,
		"std_dev":               stdDev,
	}
	inputDataJSON, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input data: %w", err)
	}

	nextSprint := time.Now().UTC().Add(14 * 24 * time.Hour)

	forecast := &models.Forecast{
		ID:                     uuid.New().String(),
		ForecastType:           "team_velocity",
		EntityType:             "team",
		EntityID:               &teamID,
		TimeHorizon:            "next_sprint",
		PredictedValue:         predictedVelocity,
		PredictedDate:          &nextSprint,
		ConfidencePercentage:   confidencePct,
		ConfidenceIntervalLow:  &low,
		ConfidenceIntervalHigh: &high,
		ModelType:              "moving_average",
		InputData:              string(inputDataJSON),
		CreatedAt:              time.Now().UTC(),
		ExpiresAt:              &nextSprint,
	}

	if err := s.forecastingStore.CreateForecast(forecast); err != nil {
		return nil, fmt.Errorf("failed to store forecast: %w", err)
	}

	return forecast, nil
}

// Helper functions

func (s *ForecastingService) calculateCompletedPoints(sprintID string) float64 {
	var completed float64
	s.db.QueryRow(`
		SELECT COALESCE(SUM(points), 0)
		FROM sprint_tasks
		WHERE sprint_id = ? AND status = 'completed'
	`, sprintID).Scan(&completed)
	return completed
}

func (s *ForecastingService) calculateSprintVariance(teamID string) float64 {
	velocities := s.getRecentVelocities(teamID, MaxHistoricalSprintCount)
	if len(velocities) < 2 {
		return DefaultVariance
	}
	return calculateVariance(velocities)
}

func (s *ForecastingService) getRecentVelocities(teamID string, limit int) []float64 {
	rows, err := s.db.Query(`
		SELECT actual_velocity
		FROM sprints
		WHERE team_id = ? AND actual_velocity IS NOT NULL
		ORDER BY end_date DESC
		LIMIT ?
	`, teamID, limit)
	if err != nil {
		return []float64{}
	}
	defer rows.Close()

	var velocities []float64
	for rows.Next() {
		var velocity float64
		if rows.Scan(&velocity) == nil {
			velocities = append(velocities, velocity)
		}
	}
	return velocities
}

func (s *ForecastingService) monteCarloSprintSimulation(sprint *models.Sprint, completedPoints, currentPace float64, simulations int) (low, high float64) {
	results := make([]float64, simulations)

	for i := 0; i < simulations; i++ {
		// Simulate with variance
		pace := currentPace * (0.8 + (0.4 * float64(i%simulations) / float64(simulations)))
		pointsRemaining := float64(sprint.CommittedPoints) - completedPoints
		daysNeeded := pointsRemaining / pace
		results[i] = daysNeeded
	}

	// Calculate percentiles
	low = calculatePercentile(results, 10)
	high = calculatePercentile(results, 90)

	return low, high
}
