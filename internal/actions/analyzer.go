package actions

import (
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// EffectivenessAnalyzer analyzes action effectiveness and identifies patterns
type EffectivenessAnalyzer struct {
	actionsStore *db.ActionsStore
}

// NewEffectivenessAnalyzer creates a new effectiveness analyzer
func NewEffectivenessAnalyzer(actionsStore *db.ActionsStore) *EffectivenessAnalyzer {
	return &EffectivenessAnalyzer{
		actionsStore: actionsStore,
	}
}

// AnalyzeActionImpact analyzes the impact of an action by comparing before/after metrics
func (a *EffectivenessAnalyzer) AnalyzeActionImpact(actionID string) (*models.ActionOutcome, error) {
	action, err := a.actionsStore.GetAction(actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get action: %w", err)
	}
	if action == nil {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}

	// Check if enough time has passed
	timeSinceAction := time.Since(action.TakenAt)
	if timeSinceAction < 48*time.Hour {
		return nil, fmt.Errorf("insufficient time to measure impact - wait at least 48 hours")
	}

	// Get existing outcomes
	outcomes, err := a.actionsStore.GetActionOutcomes(actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get outcomes: %w", err)
	}

	if len(outcomes) > 0 {
		return outcomes[0], nil // Return most recent outcome
	}

	return nil, fmt.Errorf("no outcomes recorded for action: %s", actionID)
}

// CalculateTimeToImpact calculates how long it took for an action to show results
func (a *EffectivenessAnalyzer) CalculateTimeToImpact(actionID string) (*int, error) {
	action, err := a.actionsStore.GetAction(actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get action: %w", err)
	}
	if action == nil {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}

	outcomes, err := a.actionsStore.GetActionOutcomes(actionID)
	if err != nil || len(outcomes) == 0 {
		return nil, fmt.Errorf("no outcomes found for action: %s", actionID)
	}

	// Use the first positive outcome
	for _, outcome := range outcomes {
		if outcome.Effectiveness == "highly_effective" || outcome.Effectiveness == "somewhat_effective" {
			days := int(outcome.MeasuredAt.Sub(action.TakenAt).Hours() / 24)
			return &days, nil
		}
	}

	return nil, fmt.Errorf("no effective outcomes found for action: %s", actionID)
}

// DetermineEffectiveness categorizes the effectiveness of an action
func (a *EffectivenessAnalyzer) DetermineEffectiveness(beforeValue, afterValue float64) string {
	if beforeValue == 0 {
		return "not_effective"
	}

	changePct := ((afterValue - beforeValue) / beforeValue) * 100

	if changePct > 15 {
		return "highly_effective"
	} else if changePct > 5 {
		return "somewhat_effective"
	} else {
		return "not_effective"
	}
}

// IdentifySuccessfulPatterns finds patterns in highly effective actions
func (a *EffectivenessAnalyzer) IdentifySuccessfulPatterns() ([]*models.EffectivenessPattern, error) {
	// Get all actions
	actions, err := a.actionsStore.ListActions(map[string]string{}, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}

	// Group by action type and recommendation type
	patternMap := make(map[string]*patternData)

	for _, action := range actions {
		outcomes, err := a.actionsStore.GetActionOutcomes(action.ID)
		if err != nil || len(outcomes) == 0 {
			continue
		}

		// Get recommendation type if linked
		recType := "ad_hoc"
		if action.RecommendationID != nil {
			rec, err := a.actionsStore.GetRecommendation(*action.RecommendationID)
			if err == nil && rec != nil {
				recType = rec.RecommendationType
			}
		}

		key := fmt.Sprintf("%s:%s", action.ActionType, recType)

		if _, exists := patternMap[key]; !exists {
			patternMap[key] = &patternData{
				actionType:         action.ActionType,
				recommendationType: recType,
				outcomes:           []string{},
				timeToImpacts:      []int{},
			}
		}

		// Add outcome data
		latest := outcomes[0]
		patternMap[key].outcomes = append(patternMap[key].outcomes, latest.Effectiveness)
		if latest.TimeToImpactDays != nil {
			patternMap[key].timeToImpacts = append(patternMap[key].timeToImpacts, *latest.TimeToImpactDays)
		}
	}

	// Convert to patterns (only those with >= 3 samples)
	var patterns []*models.EffectivenessPattern

	for _, data := range patternMap {
		if len(data.outcomes) < 3 {
			continue
		}

		successCount := 0
		for _, outcome := range data.outcomes {
			if outcome == "highly_effective" {
				successCount++
			}
		}

		successRate := float64(successCount) / float64(len(data.outcomes))

		// Only include patterns with >50% success rate
		if successRate > 0.5 {
			avgTimeToImpact := 0.0
			if len(data.timeToImpacts) > 0 {
				sum := 0
				for _, t := range data.timeToImpacts {
					sum += t
				}
				avgTimeToImpact = float64(sum) / float64(len(data.timeToImpacts))
			}

			pattern := &models.EffectivenessPattern{
				ActionType:         data.actionType,
				RecommendationType: data.recommendationType,
				SuccessRate:        successRate * 100,
				SampleSize:         len(data.outcomes),
				AvgTimeToImpact:    avgTimeToImpact,
				Recommendation: fmt.Sprintf(
					"When addressing %s issues, %s actions are %.0f%% effective (based on %d cases)",
					data.recommendationType,
					data.actionType,
					successRate*100,
					len(data.outcomes),
				),
			}

			patterns = append(patterns, pattern)
		}
	}

	return patterns, nil
}

// IdentifyUnsuccessfulPatterns finds patterns in ineffective actions
func (a *EffectivenessAnalyzer) IdentifyUnsuccessfulPatterns() ([]*models.EffectivenessPattern, error) {
	// Get all actions
	actions, err := a.actionsStore.ListActions(map[string]string{}, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}

	// Group by action type and recommendation type
	patternMap := make(map[string]*patternData)

	for _, action := range actions {
		outcomes, err := a.actionsStore.GetActionOutcomes(action.ID)
		if err != nil || len(outcomes) == 0 {
			continue
		}

		// Get recommendation type if linked
		recType := "ad_hoc"
		if action.RecommendationID != nil {
			rec, err := a.actionsStore.GetRecommendation(*action.RecommendationID)
			if err == nil && rec != nil {
				recType = rec.RecommendationType
			}
		}

		key := fmt.Sprintf("%s:%s", action.ActionType, recType)

		if _, exists := patternMap[key]; !exists {
			patternMap[key] = &patternData{
				actionType:         action.ActionType,
				recommendationType: recType,
				outcomes:           []string{},
			}
		}

		// Add outcome data
		latest := outcomes[0]
		patternMap[key].outcomes = append(patternMap[key].outcomes, latest.Effectiveness)
	}

	// Convert to patterns (only those with >= 3 samples and <50% success)
	var patterns []*models.EffectivenessPattern

	for _, data := range patternMap {
		if len(data.outcomes) < 3 {
			continue
		}

		successCount := 0
		for _, outcome := range data.outcomes {
			if outcome == "highly_effective" {
				successCount++
			}
		}

		successRate := float64(successCount) / float64(len(data.outcomes))

		// Only include patterns with <50% success rate (unsuccessful)
		if successRate < 0.5 {
			pattern := &models.EffectivenessPattern{
				ActionType:         data.actionType,
				RecommendationType: data.recommendationType,
				SuccessRate:        successRate * 100,
				SampleSize:         len(data.outcomes),
				Recommendation: fmt.Sprintf(
					"%s actions for %s issues are only %.0f%% effective - consider different approach",
					data.actionType,
					data.recommendationType,
					successRate*100,
				),
			}

			patterns = append(patterns, pattern)
		}
	}

	return patterns, nil
}

// GenerateRecommendationsFromPatterns creates new recommendations based on learned patterns
func (a *EffectivenessAnalyzer) GenerateRecommendationsFromPatterns() ([]*models.Recommendation, error) {
	successfulPatterns, err := a.IdentifySuccessfulPatterns()
	if err != nil {
		return nil, fmt.Errorf("failed to identify patterns: %w", err)
	}

	var recommendations []*models.Recommendation

	for _, pattern := range successfulPatterns {
		// Only create recommendations for very successful patterns (>80%)
		if pattern.SuccessRate > 80 {
			rec := &models.Recommendation{
				SourceType:         "ai_insight",
				RecommendationType: "skill_development",
				Priority:           "low",
				SubjectType:        "process",
				Title:              "Successful pattern identified",
				Description:        pattern.Recommendation,
				SuggestedActions: []string{
					"Review pattern data",
					"Train team on successful approach",
					"Standardize this approach",
				},
				Status: "pending",
			}

			recommendations = append(recommendations, rec)
		}
	}

	return recommendations, nil
}

// Helper struct for pattern analysis
type patternData struct {
	actionType         string
	recommendationType string
	outcomes           []string
	timeToImpacts      []int
}
