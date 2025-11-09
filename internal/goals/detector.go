package goals

import (
	"fmt"
	"strings"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Detector handles automatic goal progress detection from events
type Detector struct {
	store   *db.GoalsStore
	service *Service
}

// NewDetector creates a new goal detector
func NewDetector(store *db.GoalsStore, service *Service) *Detector {
	return &Detector{
		store:   store,
		service: service,
	}
}

// DetectProgressFromEvent analyzes an event and updates relevant goals
func (d *Detector) DetectProgressFromEvent(event *models.Event) error {
	// Get all active goals that use automatic tracking
	goals, _, err := d.store.ListGoals("", "", "active", "", 0, 0)
	if err != nil {
		return fmt.Errorf("failed to list goals: %w", err)
	}

	for _, goal := range goals {
		if goal.TrackingMethod == "manual" {
			continue // Skip manual goals
		}

		// Match event to goal based on owner and type
		if d.isEventRelevant(event, goal) {
			if err := d.processEventForGoal(event, goal); err != nil {
				// Log error but continue processing other goals
				fmt.Printf("Error processing event for goal %s: %v\n", goal.ID, err)
			}
		}
	}

	return nil
}

// isEventRelevant checks if an event is relevant to a goal
func (d *Detector) isEventRelevant(event *models.Event, goal *models.Goal) bool {
	// Check owner match
	if goal.OwnerType == "engineer" && goal.OwnerID != nil {
		// For engineer goals, check if event is from that engineer
		if author, ok := event.Data["author"].(string); ok {
			// Match by email or identity
			if author == *goal.OwnerID {
				return true
			}
		}
		// Also check EngineerID field
		if event.EngineerID == *goal.OwnerID {
			return true
		}
	}

	// For team goals, would need to check team membership
	// For now, simplified to engineer goals

	return false
}

// processEventForGoal processes an event for a specific goal
func (d *Detector) processEventForGoal(event *models.Event, goal *models.Goal) error {
	// Get milestones for this goal
	milestones, _, err := d.store.GetMilestones(goal.ID, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get milestones: %w", err)
	}

	// Detect different types of achievements
	switch event.Type {
	case "pull_request":
		return d.detectPullRequestProgress(event, goal, milestones)
	case "code_review":
		return d.detectReviewProgress(event, goal, milestones)
	case "issue":
		return d.detectIssueProgress(event, goal, milestones)
	default:
		return nil
	}
}

// detectPullRequestProgress detects progress from PR events
func (d *Detector) detectPullRequestProgress(event *models.Event, goal *models.Goal, milestones []*models.GoalMilestone) error {
	action, ok := event.Data["action"].(string)
	if !ok || (action != "merged" && action != "closed") {
		return nil // Only count merged/closed PRs
	}

	// Check for complex feature completion
	if d.isComplexFeature(event.Data) {
		if err := d.incrementMilestone(goal, milestones, "complex features", 1.0, event.ID); err != nil {
			return err
		}
	}

	// Check for PR count milestones
	if err := d.incrementMilestone(goal, milestones, "PRs", 1.0, event.ID); err != nil {
		return err
	}

	return nil
}

// detectReviewProgress detects progress from code review events
func (d *Detector) detectReviewProgress(event *models.Event, goal *models.Goal, milestones []*models.GoalMilestone) error {
	state, ok := event.Data["state"].(string)
	if !ok || (state != "approved" && state != "changes_requested") {
		return nil // Only count completed reviews
	}

	// Increment review count
	if err := d.incrementMilestone(goal, milestones, "reviews", 1.0, event.ID); err != nil {
		return err
	}

	return nil
}

// detectIssueProgress detects progress from issue events
func (d *Detector) detectIssueProgress(event *models.Event, goal *models.Goal, milestones []*models.GoalMilestone) error {
	action, ok := event.Data["action"].(string)
	if !ok || action != "closed" {
		return nil
	}

	// Check if it's a bug
	labels, ok := event.Data["labels"].([]interface{})
	if !ok {
		return nil
	}
	isBug := false
	for _, label := range labels {
		if labelStr, ok := label.(string); ok && strings.ToLower(labelStr) == "bug" {
			isBug = true
			break
		}
	}

	if isBug {
		if err := d.incrementMilestone(goal, milestones, "bugs", 1.0, event.ID); err != nil {
			return err
		}
	}

	return nil
}

// isComplexFeature determines if a PR represents a complex feature
func (d *Detector) isComplexFeature(eventData map[string]interface{}) bool {
	// Criteria for complex feature:
	// 1. >500 lines changed
	// 2. Touches 3+ files or services
	// 3. Has design doc or RFC link

	additions, ok := eventData["additions"].(float64)
	if !ok {
		return false
	}
	deletions, ok := eventData["deletions"].(float64)
	if !ok {
		return false
	}
	totalLines := additions + deletions

	if totalLines < 500 {
		return false
	}

	filesChanged, ok := eventData["files_changed"].(float64)
	if !ok || filesChanged < 3 {
		return false
	}

	// Check for design doc reference in PR body
	body, ok := eventData["body"].(string)
	if !ok {
		return false
	}
	hasDesignDoc := strings.Contains(strings.ToLower(body), "design doc") ||
		strings.Contains(strings.ToLower(body), "rfc") ||
		strings.Contains(strings.ToLower(body), "technical design")

	return hasDesignDoc
}

// incrementMilestone increments a milestone's current value
func (d *Detector) incrementMilestone(goal *models.Goal, milestones []*models.GoalMilestone, unit string, amount float64, eventID string) error {
	// Find milestone with matching unit
	for _, milestone := range milestones {
		if strings.Contains(strings.ToLower(milestone.Unit), strings.ToLower(unit)) ||
			strings.Contains(strings.ToLower(milestone.Title), strings.ToLower(unit)) {

			previousValue := milestone.CurrentValue
			milestone.CurrentValue += amount

			// Check if milestone is now complete
			if !milestone.Completed && milestone.TargetValue > 0 && milestone.CurrentValue >= milestone.TargetValue {
				if err := d.service.CompleteMilestone(milestone.ID); err != nil {
					return err
				}
			} else {
				if err := d.store.UpdateMilestone(milestone); err != nil {
					return err
				}
			}

			// Log the progress
			system := "system"
			evidence := fmt.Sprintf(`{"event_id": "%s", "event_type": "auto_detected"}`, eventID)
			newValue := milestone.CurrentValue
			log := &models.GoalProgressLog{
				GoalID:        goal.ID,
				MilestoneID:   &milestone.ID,
				PreviousValue: &previousValue,
				NewValue:      &newValue,
				ChangeType:    "auto_detected",
				Evidence:      evidence,
				LoggedBy:      &system,
			}
			if err := d.store.LogProgress(log); err != nil {
				return err
			}

			// Update goal progress
			if err := d.service.UpdateGoalProgress(goal.ID); err != nil {
				return err
			}

			break // Only update first matching milestone
		}
	}

	return nil
}

// DetectMentoringProgress detects mentoring activities
func (d *Detector) DetectMentoringProgress(engineerID string, pairingSessionCount int) error {
	// Get mentoring goals for this engineer
	goals, _, err := d.store.ListGoals("engineer", engineerID, "active", "", 0, 0)
	if err != nil {
		return err
	}

	for _, goal := range goals {
		if strings.Contains(strings.ToLower(goal.Title), "mentor") {
			milestones, _, err := d.store.GetMilestones(goal.ID, 100, 0)
			if err != nil {
				continue
			}

			if err := d.incrementMilestone(goal, milestones, "pairing sessions", float64(pairingSessionCount), "manual"); err != nil {
				fmt.Printf("Error incrementing mentoring milestone: %v\n", err)
			}
		}
	}

	return nil
}

// DetectSystemDesignGrowth detects system design work
func (d *Detector) DetectSystemDesignGrowth(event *models.Event) error {
	// Count services touched
	filesList, ok := event.Data["files"].([]interface{})
	if !ok {
		return nil
	}
	servicesSet := make(map[string]bool)

	for _, file := range filesList {
		if fileStr, ok := file.(string); ok {
			// Extract service name from path (e.g., "services/auth/..." -> "auth")
			parts := strings.Split(fileStr, "/")
			if len(parts) > 1 && parts[0] == "services" {
				servicesSet[parts[1]] = true
			}
		}
	}

	if len(servicesSet) >= 3 {
		// This PR touches multiple services - system design work
		author, ok := event.Data["author"].(string)
		if !ok {
			return nil
		}
		goals, _, err := d.store.ListGoals("engineer", author, "active", "", 0, 0)
		if err != nil {
			return err
		}

		for _, goal := range goals {
			if strings.Contains(strings.ToLower(goal.Title), "system design") ||
				strings.Contains(strings.ToLower(goal.Title), "architecture") {
				milestones, _, err := d.store.GetMilestones(goal.ID, 100, 0)
				if err != nil {
					continue
				}

				if err := d.incrementMilestone(goal, milestones, "services", 1.0, event.ID); err != nil {
					fmt.Printf("Error incrementing system design milestone: %v\n", err)
				}
			}
		}
	}

	return nil
}
