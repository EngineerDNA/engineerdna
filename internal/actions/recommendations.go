package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// RecommendationService handles recommendation management
type RecommendationService struct {
	actionsStore *db.ActionsStore
}

// NewRecommendationService creates a new recommendation service
func NewRecommendationService(actionsStore *db.ActionsStore) *RecommendationService {
	return &RecommendationService{
		actionsStore: actionsStore,
	}
}

// GenerateRecommendationFromAlert creates a recommendation from an alert
func (s *RecommendationService) GenerateRecommendationFromAlert(alert *models.AlertInstance, managerID string) (*models.Recommendation, error) {
	priority := mapSeverityToPriority(alert.Severity)

	// Determine recommendation type based on alert type
	recType := "check_in"
	suggestedActions := []string{
		"Have 1:1 conversation to understand root cause",
		"Document context in manager notes",
		"Create action plan with timeline",
	}

	contextMap := map[string]interface{}{
		"alert_id":    alert.ID,
		"severity":    alert.Severity,
		"fired_at":    alert.FiredAt,
		"entity_type": alert.EntityType,
		"entity_id":   alert.EntityID,
	}
	contextJSON, _ := json.Marshal(contextMap)

	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour) // 1 week expiry

	rec := &models.Recommendation{
		SourceType:         "alert",
		SourceID:           &alert.ID,
		RecommendationType: recType,
		Priority:           priority,
		SubjectType:        alert.EntityType,
		SubjectID:          &alert.EntityID,
		Title:              fmt.Sprintf("Address: %s", alert.Title),
		Description:        alert.Message,
		SuggestedActions:   suggestedActions,
		Context:            string(contextJSON),
		AssignedTo:         &managerID,
		ExpiresAt:          &expiresAt,
		Status:             "pending",
	}

	if err := s.actionsStore.CreateRecommendation(rec); err != nil {
		return nil, fmt.Errorf("failed to create recommendation: %w", err)
	}

	return rec, nil
}

// GenerateRecommendationFromBriefing creates recommendations from briefing attention items
func (s *RecommendationService) GenerateRecommendationFromBriefing(
	briefing *models.WeeklyBriefing,
	item *models.AttentionItem,
	managerID string,
) (*models.Recommendation, error) {
	recType := "check_in"
	priority := mapBriefingSeverityToPriority(item.Severity)

	contextMap := map[string]interface{}{
		"briefing_week": briefing.WeekStart,
		"team_id":       briefing.TeamID,
		"severity":      item.Severity,
	}
	contextJSON, _ := json.Marshal(contextMap)

	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)

	rec := &models.Recommendation{
		SourceType:         "briefing",
		SourceID:           &briefing.TeamID,
		RecommendationType: recType,
		Priority:           priority,
		SubjectType:        "engineer",
		SubjectID:          nil, // Would need engineer ID mapping
		Title:              fmt.Sprintf("Check in: %s", item.EngineerName),
		Description:        item.Issue,
		SuggestedActions:   []string{item.SuggestedTopic},
		Context:            string(contextJSON),
		AssignedTo:         &managerID,
		ExpiresAt:          &expiresAt,
		Status:             "pending",
	}

	if err := s.actionsStore.CreateRecommendation(rec); err != nil {
		return nil, fmt.Errorf("failed to create recommendation: %w", err)
	}

	return rec, nil
}

// AssignRecommendation assigns a recommendation to a manager
func (s *RecommendationService) AssignRecommendation(recommendationID, managerID string) error {
	rec, err := s.actionsStore.GetRecommendation(recommendationID)
	if err != nil {
		return fmt.Errorf("failed to get recommendation: %w", err)
	}
	if rec == nil {
		return fmt.Errorf("recommendation not found: %s", recommendationID)
	}

	// Update in database would require an UpdateRecommendation method
	// For now, we track this in history
	history := &models.RecommendationHistory{
		RecommendationID: recommendationID,
		StatusChange:     "assigned",
		ChangedBy:        &managerID,
		Reason:           "Assigned to manager",
	}

	return s.actionsStore.CreateRecommendationHistory(history)
}

// UpdateRecommendationStatus updates the status of a recommendation
func (s *RecommendationService) UpdateRecommendationStatus(recommendationID, status, managerID, reason string) error {
	return s.actionsStore.UpdateRecommendationStatus(recommendationID, status, managerID, reason)
}

// DismissRecommendation dismisses a recommendation with a reason
func (s *RecommendationService) DismissRecommendation(recommendationID, managerID, reason string) error {
	return s.actionsStore.UpdateRecommendationStatus(recommendationID, "dismissed", managerID, reason)
}

// SnoozeRecommendation temporarily hides a recommendation
func (s *RecommendationService) SnoozeRecommendation(recommendationID, managerID string, duration time.Duration) error {
	return s.actionsStore.UpdateRecommendationStatus(
		recommendationID,
		"snoozed",
		managerID,
		fmt.Sprintf("Snoozed for %s", duration),
	)
}

// GetPendingRecommendations retrieves all pending recommendations for a manager
func (s *RecommendationService) GetPendingRecommendations(managerID string) ([]*models.Recommendation, error) {
	filters := map[string]string{
		"assigned_to": managerID,
		"status":      "pending",
	}

	recommendations, _, err := s.actionsStore.ListRecommendations(filters, 100, 0)
	return recommendations, err
}

// GetRecommendationWithDetails retrieves a recommendation with actions and history
func (s *RecommendationService) GetRecommendationWithDetails(recommendationID string) (*models.RecommendationWithDetails, error) {
	rec, err := s.actionsStore.GetRecommendation(recommendationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendation: %w", err)
	}
	if rec == nil {
		return nil, fmt.Errorf("recommendation not found: %s", recommendationID)
	}

	// Get related actions
	actions, err := s.actionsStore.ListActions(map[string]string{
		"recommendation_id": recommendationID,
	}, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}

	// Get history
	history, err := s.actionsStore.GetRecommendationHistory(recommendationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	return &models.RecommendationWithDetails{
		Recommendation: rec,
		Actions:        actions,
		History:        history,
	}, nil
}

// Helper functions

func mapSeverityToPriority(severity string) string {
	switch severity {
	case "critical":
		return "urgent"
	case "warning":
		return "high"
	case "info":
		return "medium"
	default:
		return "low"
	}
}

func mapBriefingSeverityToPriority(severity string) string {
	switch severity {
	case "critical":
		return "urgent"
	case "warning":
		return "high"
	case "info":
		return "medium"
	default:
		return "low"
	}
}
