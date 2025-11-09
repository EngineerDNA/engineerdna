package goals

import (
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Service handles goal business logic
type Service struct {
	store        *db.GoalsStore
	scoringStore *db.ScoringStore
	metricStore  *db.MetricStore
}

// NewService creates a new goal service
func NewService(store *db.GoalsStore, scoringStore *db.ScoringStore, metricStore *db.MetricStore) *Service {
	return &Service{
		store:        store,
		scoringStore: scoringStore,
		metricStore:  metricStore,
	}
}

// CalculateGoalProgress computes progress percentage from milestones
func (s *Service) CalculateGoalProgress(goalID string) (int, error) {
	milestones, _, err := s.store.GetMilestones(goalID, 100, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get milestones: %w", err)
	}

	if len(milestones) == 0 {
		return 0, nil
	}

	totalProgress := 0.0
	for _, milestone := range milestones {
		if milestone.Completed {
			totalProgress += 100.0
		} else if milestone.TargetValue > 0 {
			progress := (milestone.CurrentValue / milestone.TargetValue) * 100.0
			if progress > 100.0 {
				progress = 100.0
			}
			totalProgress += progress
		}
	}

	averageProgress := totalProgress / float64(len(milestones))
	return int(averageProgress), nil
}

// EvaluateGoalStatus determines if goal is on track, at risk, or off track
func (s *Service) EvaluateGoalStatus(goal *models.Goal) string {
	now := time.Now().UTC()

	// If completed, return completed
	if goal.Status == "completed" || goal.ProgressPercentage >= 100 {
		return "completed"
	}

	// If archived, return archived
	if goal.Status == "archived" {
		return "archived"
	}

	// Calculate time metrics
	totalDuration := goal.EndDate.Sub(goal.StartDate).Hours() / 24
	elapsedDuration := now.Sub(goal.StartDate).Hours() / 24
	remainingDuration := goal.EndDate.Sub(now).Hours() / 24

	// If not yet started
	if elapsedDuration < 0 {
		return "active"
	}

	// If past due date
	if remainingDuration < 0 {
		if goal.ProgressPercentage < 100 {
			return "off_track"
		}
		return "completed"
	}

	// Calculate expected progress
	elapsedPercent := elapsedDuration / totalDuration
	expectedProgress := elapsedPercent * 100

	// Status logic
	actualProgress := float64(goal.ProgressPercentage)

	if actualProgress >= expectedProgress*0.9 {
		return "active" // On track
	} else if actualProgress >= expectedProgress*0.7 {
		return "at_risk"
	} else {
		return "off_track"
	}
}

// UpdateGoalProgress updates goal progress and status
func (s *Service) UpdateGoalProgress(goalID string) error {
	goal, err := s.store.GetGoal(goalID)
	if err != nil {
		return fmt.Errorf("failed to get goal: %w", err)
	}
	if goal == nil {
		return fmt.Errorf("goal not found: %s", goalID)
	}

	// Calculate new progress
	progress, err := s.CalculateGoalProgress(goalID)
	if err != nil {
		return fmt.Errorf("failed to calculate progress: %w", err)
	}

	// Update goal
	goal.ProgressPercentage = progress
	goal.Status = s.EvaluateGoalStatus(goal)

	if err := s.store.UpdateGoal(goal); err != nil {
		return fmt.Errorf("failed to update goal: %w", err)
	}

	return nil
}

// UpdateGoalMetrics updates current values for goal metrics
func (s *Service) UpdateGoalMetrics(goalID string) error {
	goal, err := s.store.GetGoal(goalID)
	if err != nil {
		return fmt.Errorf("failed to get goal: %w", err)
	}
	if goal == nil {
		return fmt.Errorf("goal not found: %s", goalID)
	}

	metrics, _, err := s.store.GetGoalMetrics(goalID, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	// Get scoring data for the owner
	if goal.OwnerID == nil {
		return nil // Org-level goals don't have owner-specific metrics
	}

	for _, metric := range metrics {
		currentValue, err := s.fetchMetricValue(metric.MetricName, *goal.OwnerID, goal.OwnerType)
		if err != nil {
			continue // Skip metrics we can't fetch
		}

		metric.CurrentValue = &currentValue
		if err := s.store.UpdateGoalMetric(metric); err != nil {
			return fmt.Errorf("failed to update metric: %w", err)
		}
	}

	return nil
}

// fetchMetricValue retrieves current value for a metric
func (s *Service) fetchMetricValue(metricName, ownerID, ownerType string) (float64, error) {
	if ownerType != "engineer" {
		return 0, fmt.Errorf("metric fetch only supported for engineers")
	}

	// Get recent scoring data (last 4 weeks)
	now := time.Now().UTC()
	endDate := now
	startDate := now.AddDate(0, 0, -28) // 4 weeks ago

	metrics, err := s.metricStore.GetEngineerScores(ownerID, startDate, endDate, 100)
	if err != nil || len(metrics) == 0 {
		return 0, fmt.Errorf("no scoring data available")
	}

	// Extract total score from most recent metric
	totalScore := float64(0)
	for _, m := range metrics {
		if m.MetricName == "engineer_total_score" {
			totalScore = m.Value
			break
		}
	}

	// For now, return total score as percentage (simplified)
	// TODO: Implement proper metric parsing and calculation from dimensions
	// For most metric names, we'll return the total score
	return totalScore, nil
}

// CompleteMilestone marks a milestone as complete
func (s *Service) CompleteMilestone(milestoneID string) error {
	// Get milestone to get goal_id
	milestones, _, err := s.store.GetMilestones("", db.MaxQueryLimit, 0)
	if err != nil {
		return fmt.Errorf("failed to get milestones: %w", err)
	}

	var milestone *models.GoalMilestone
	for _, m := range milestones {
		if m.ID == milestoneID {
			milestone = m
			break
		}
	}

	if milestone == nil {
		return fmt.Errorf("milestone not found: %s", milestoneID)
	}

	// Mark complete
	now := time.Now().UTC()
	milestone.Completed = true
	milestone.CompletedAt = &now

	if err := s.store.UpdateMilestone(milestone); err != nil {
		return fmt.Errorf("failed to update milestone: %w", err)
	}

	// Log progress
	system := "system"
	log := &models.GoalProgressLog{
		GoalID:      milestone.GoalID,
		MilestoneID: &milestone.ID,
		ChangeType:  "milestone_complete",
		LoggedBy:    &system,
	}
	if err := s.store.LogProgress(log); err != nil {
		return fmt.Errorf("failed to log progress: %w", err)
	}

	// Update goal progress
	if err := s.UpdateGoalProgress(milestone.GoalID); err != nil {
		return fmt.Errorf("failed to update goal progress: %w", err)
	}

	return nil
}

// GetGoalSummary generates summary statistics for goals
func (s *Service) GetGoalSummary(ownerType, ownerID string) (*models.GoalSummary, error) {
	goals, _, err := s.store.ListGoals(ownerType, ownerID, "", "", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list goals: %w", err)
	}

	summary := &models.GoalSummary{
		OwnerType: ownerType,
	}
	if ownerID != "" {
		summary.OwnerID = &ownerID
	}

	totalProgress := 0
	for _, goal := range goals {
		summary.TotalGoals++
		totalProgress += goal.ProgressPercentage

		switch goal.Status {
		case "active":
			summary.ActiveGoals++
		case "completed":
			summary.CompletedGoals++
		case "at_risk":
			summary.AtRiskGoals++
		case "off_track":
			summary.OffTrackGoals++
		}
	}

	if summary.TotalGoals > 0 {
		summary.AverageProgress = float64(totalProgress) / float64(summary.TotalGoals)
	}

	return summary, nil
}

// GetGoalWithDetails retrieves a goal with all related data
func (s *Service) GetGoalWithDetails(goalID string) (*models.GoalWithDetails, error) {
	goal, err := s.store.GetGoal(goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}
	if goal == nil {
		return nil, fmt.Errorf("goal not found: %s", goalID)
	}

	milestones, _, err := s.store.GetMilestones(goalID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestones: %w", err)
	}

	metrics, _, err := s.store.GetGoalMetrics(goalID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	dependencies, _, err := s.store.GetDependencies(goalID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}

	progressLogs, _, err := s.store.GetProgressLogs(goalID, 20, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get progress logs: %w", err)
	}

	return &models.GoalWithDetails{
		Goal:         goal,
		Milestones:   milestones,
		Metrics:      metrics,
		Dependencies: dependencies,
		ProgressLogs: progressLogs,
	}, nil
}

// ForecastGoalCompletion predicts completion date based on current pace
func (s *Service) ForecastGoalCompletion(goalID string) (*time.Time, error) {
	goal, err := s.store.GetGoal(goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}
	if goal == nil {
		return nil, fmt.Errorf("goal not found: %s", goalID)
	}

	// If already completed
	if goal.Status == "completed" || goal.ProgressPercentage >= 100 {
		return &goal.EndDate, nil
	}

	// Get progress logs to calculate velocity
	logs, _, err := s.store.GetProgressLogs(goalID, 10, 0)
	if err != nil || len(logs) < 2 {
		// Not enough data, return planned end date
		return &goal.EndDate, nil
	}

	// Calculate average progress per day
	oldestLog := logs[len(logs)-1]
	newestLog := logs[0]

	if newestLog.NewValue == nil || oldestLog.NewValue == nil {
		return &goal.EndDate, nil
	}

	progressMade := *newestLog.NewValue - *oldestLog.NewValue
	daysElapsed := newestLog.LoggedAt.Sub(oldestLog.LoggedAt).Hours() / 24

	if daysElapsed <= 0 || progressMade <= 0 {
		return &goal.EndDate, nil
	}

	velocityPerDay := progressMade / daysElapsed
	remainingProgress := 100.0 - float64(goal.ProgressPercentage)
	daysRemaining := remainingProgress / velocityPerDay

	forecast := time.Now().UTC().AddDate(0, 0, int(daysRemaining))
	return &forecast, nil
}
