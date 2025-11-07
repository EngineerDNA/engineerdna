package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// ActionService handles action management and tracking
type ActionService struct {
	actionsStore *db.ActionsStore
}

// NewActionService creates a new action service
func NewActionService(actionsStore *db.ActionsStore) *ActionService {
	return &ActionService{
		actionsStore: actionsStore,
	}
}

// RecordAction creates a new action record
func (s *ActionService) RecordAction(action *models.Action) error {
	if err := s.actionsStore.CreateAction(action); err != nil {
		return fmt.Errorf("failed to record action: %w", err)
	}

	// Auto-schedule follow-up if date specified
	if action.FollowUpDate != nil {
		followUp := &models.FollowUp{
			ActionID:     action.ID,
			FollowUpDate: *action.FollowUpDate,
			FollowUpType: "check_metric",
			Description:  "Check if action had intended outcome",
		}
		s.actionsStore.CreateFollowUp(followUp)
	}

	return nil
}

// LinkActionToRecommendation links an existing action to a recommendation
func (s *ActionService) LinkActionToRecommendation(actionID, recommendationID string) error {
	action, err := s.actionsStore.GetAction(actionID)
	if err != nil {
		return fmt.Errorf("failed to get action: %w", err)
	}
	if action == nil {
		return fmt.Errorf("action not found: %s", actionID)
	}

	// Would need UpdateAction method to modify recommendation_id
	// For now, this is handled during action creation
	return nil
}

// ScheduleFollowUp creates a follow-up reminder for an action
func (s *ActionService) ScheduleFollowUp(actionID string, followUpDate time.Time, followUpType, description string) error {
	followUp := &models.FollowUp{
		ActionID:     actionID,
		FollowUpDate: followUpDate,
		FollowUpType: followUpType,
		Description:  description,
	}

	return s.actionsStore.CreateFollowUp(followUp)
}

// GetDueFollowUps retrieves all follow-ups that are due
func (s *ActionService) GetDueFollowUps() ([]*models.FollowUp, error) {
	return s.actionsStore.GetFollowUpsDue(time.Now().UTC())
}

// CompleteFollowUp marks a follow-up as completed
func (s *ActionService) CompleteFollowUp(followUpID string) error {
	return s.actionsStore.CompleteFollowUp(followUpID)
}

// GetActionHistory retrieves all actions for a manager or subject
func (s *ActionService) GetActionHistory(filters map[string]string, limit int) ([]*models.Action, error) {
	return s.actionsStore.ListActions(filters, limit)
}

// GetActionWithDetails retrieves an action with outcomes and follow-ups
func (s *ActionService) GetActionWithDetails(actionID string) (*models.ActionWithDetails, error) {
	action, err := s.actionsStore.GetAction(actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get action: %w", err)
	}
	if action == nil {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}

	outcomes, err := s.actionsStore.GetActionOutcomes(actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get outcomes: %w", err)
	}

	// Get follow-ups for this action
	// Would need a GetFollowUpsByAction method for proper implementation
	followUps := []*models.FollowUp{}

	return &models.ActionWithDetails{
		Action:    action,
		Outcomes:  outcomes,
		FollowUps: followUps,
	}, nil
}

// RecordActionOutcome creates an outcome measurement for an action
func (s *ActionService) RecordActionOutcome(outcome *models.ActionOutcome) error {
	if err := s.actionsStore.CreateActionOutcome(outcome); err != nil {
		return fmt.Errorf("failed to record outcome: %w", err)
	}

	// If action linked to recommendation and outcome is positive, mark recommendation as completed
	action, err := s.actionsStore.GetAction(outcome.ActionID)
	if err == nil && action != nil && action.RecommendationID != nil {
		if outcome.Effectiveness == "highly_effective" || outcome.Effectiveness == "somewhat_effective" {
			s.actionsStore.UpdateRecommendationStatus(
				*action.RecommendationID,
				"completed",
				action.TakenBy,
				"Action showed positive outcome",
			)
		}
	}

	return nil
}

// MeasureActionEffectiveness computes effectiveness for an action
func (s *ActionService) MeasureActionEffectiveness(actionID string, metricName string, beforeValue, afterValue float64) (*models.ActionOutcome, error) {
	action, err := s.actionsStore.GetAction(actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get action: %w", err)
	}
	if action == nil {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}

	// Wait at least 48 hours before measuring
	timeSinceAction := time.Since(action.TakenAt)
	if timeSinceAction < 48*time.Hour {
		return nil, fmt.Errorf("too soon to measure - wait at least 48 hours after action")
	}

	// Calculate change
	changePct := ((afterValue - beforeValue) / beforeValue) * 100
	daysToImpact := int(timeSinceAction.Hours() / 24)

	// Determine effectiveness and outcome type
	effectiveness := "not_effective"
	outcomeType := "no_change"

	if changePct > 10 {
		effectiveness = "highly_effective"
		outcomeType = "metric_improvement"
	} else if changePct > 5 {
		effectiveness = "somewhat_effective"
		outcomeType = "metric_improvement"
	} else if changePct < -5 {
		outcomeType = "worsened"
	}

	outcome := &models.ActionOutcome{
		ActionID:         actionID,
		OutcomeType:      outcomeType,
		MeasuredMetric:   metricName,
		BeforeValue:      &beforeValue,
		AfterValue:       &afterValue,
		ChangePercentage: &changePct,
		TimeToImpactDays: &daysToImpact,
		Effectiveness:    effectiveness,
	}

	if err := s.actionsStore.CreateActionOutcome(outcome); err != nil {
		return nil, fmt.Errorf("failed to create outcome: %w", err)
	}

	return outcome, nil
}

// GenerateActionReport creates a summary report of action effectiveness
func (s *ActionService) GenerateActionReport(managerID string, days int) (*models.EffectivenessReport, error) {
	startDate := time.Now().UTC().AddDate(0, 0, -days)

	// Get all actions by this manager in the time period
	filters := map[string]string{
		"taken_by": managerID,
	}
	actions, err := s.actionsStore.ListActions(filters, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}

	// Filter by date
	var recentActions []*models.Action
	for _, action := range actions {
		if action.TakenAt.After(startDate) {
			recentActions = append(recentActions, action)
		}
	}

	// Count actions with outcomes and effectiveness
	var totalActions = len(recentActions)
	var actionsWithOutcomes = 0
	var highlyEffectiveCount = 0
	var somewhatEffectiveCount = 0
	var notEffectiveCount = 0

	for _, action := range recentActions {
		outcomes, err := s.actionsStore.GetActionOutcomes(action.ID)
		if err != nil || len(outcomes) == 0 {
			continue
		}

		actionsWithOutcomes++

		// Count effectiveness from latest outcome
		if len(outcomes) > 0 {
			latest := outcomes[0]
			switch latest.Effectiveness {
			case "highly_effective":
				highlyEffectiveCount++
			case "somewhat_effective":
				somewhatEffectiveCount++
			case "not_effective":
				notEffectiveCount++
			}
		}
	}

	// Calculate success rate
	successRate := 0.0
	if actionsWithOutcomes > 0 {
		successRate = float64(highlyEffectiveCount) / float64(actionsWithOutcomes) * 100
	}

	report := &models.EffectivenessReport{
		TimePeriod:             fmt.Sprintf("Last %d days", days),
		TotalActions:           totalActions,
		ActionsWithOutcomes:    actionsWithOutcomes,
		HighlyEffectiveCount:   highlyEffectiveCount,
		SomewhatEffectiveCount: somewhatEffectiveCount,
		NotEffectiveCount:      notEffectiveCount,
		SuccessRate:            successRate,
		SuccessfulPatterns:     []*models.EffectivenessPattern{},
		UnsuccessfulPatterns:   []*models.EffectivenessPattern{},
		TopActions:             []*models.Action{},
	}

	return report, nil
}

// ProcessFollowUpReminders checks for due follow-ups and processes them
func (s *ActionService) ProcessFollowUpReminders() error {
	followUps, err := s.GetDueFollowUps()
	if err != nil {
		return fmt.Errorf("failed to get due follow-ups: %w", err)
	}

	for _, followUp := range followUps {
		// In a real implementation, this would:
		// 1. Send notification to manager
		// 2. Auto-measure metrics if type is "check_metric"
		// 3. Update recommendation status if needed

		// For now, just log
		contextMap := map[string]interface{}{
			"follow_up_id":   followUp.ID,
			"action_id":      followUp.ActionID,
			"follow_up_type": followUp.FollowUpType,
			"description":    followUp.Description,
		}
		contextJSON, _ := json.Marshal(contextMap)
		fmt.Printf("Follow-up due: %s\n", string(contextJSON))
	}

	return nil
}
