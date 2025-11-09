package goals

import (
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// CreateDefaultGoals creates sample goals for demonstration
func CreateDefaultGoals(store *db.GoalsStore, teamID, engineerID string) error {
	// Create team goal: Reduce cycle time
	teamGoal := &models.Goal{
		Title:              "Reduce average cycle time to 2.0 days",
		Description:        "Improve team velocity by reducing PR cycle time from current 3.5 days to target 2.0 days by end of quarter.",
		GoalType:           "team",
		OwnerType:          "team",
		OwnerID:            &teamID,
		TimePeriod:         "Q1 2025",
		StartDate:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:            time.Date(2025, 3, 31, 23, 59, 59, 0, time.UTC),
		TrackingMethod:     "automatic",
		Status:             "active",
		ProgressPercentage: 0,
	}

	if err := store.CreateGoal(teamGoal); err != nil {
		return fmt.Errorf("failed to create team goal: %w", err)
	}

	// Add metric for cycle time
	cycleTimeMetric := &models.GoalMetric{
		GoalID:      teamGoal.ID,
		MetricName:  "cycle_time",
		TargetValue: 2.0,
		Operator:    "decrease_to",
	}
	if err := store.CreateGoalMetric(cycleTimeMetric); err != nil {
		return fmt.Errorf("failed to create metric: %w", err)
	}

	// Add milestones for cycle time goal
	milestones := []*models.GoalMilestone{
		{
			GoalID:       teamGoal.ID,
			Title:        "Reduce cycle time to 3.0 days",
			Description:  "First milestone: 15% improvement",
			TargetValue:  3.0,
			CurrentValue: 3.5,
			Unit:         "days",
			DueDate:      timePtr(time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)),
		},
		{
			GoalID:       teamGoal.ID,
			Title:        "Reduce cycle time to 2.5 days",
			Description:  "Second milestone: 30% improvement",
			TargetValue:  2.5,
			CurrentValue: 3.5,
			Unit:         "days",
			DueDate:      timePtr(time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)),
		},
		{
			GoalID:       teamGoal.ID,
			Title:        "Reach target of 2.0 days",
			Description:  "Final milestone: 43% improvement",
			TargetValue:  2.0,
			CurrentValue: 3.5,
			Unit:         "days",
			DueDate:      timePtr(time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC)),
		},
	}

	for _, milestone := range milestones {
		if err := store.CreateMilestone(milestone); err != nil {
			return fmt.Errorf("failed to create milestone: %w", err)
		}
	}

	// Create individual goal: Lead complex features
	individualGoal := &models.Goal{
		Title:              "Lead 2 complex features from design to deployment",
		Description:        "Demonstrate technical leadership by owning end-to-end delivery of 2 complex features requiring system design and cross-team coordination.",
		GoalType:           "individual",
		OwnerType:          "engineer",
		OwnerID:            &engineerID,
		TimePeriod:         "Q1 2025",
		StartDate:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:            time.Date(2025, 3, 31, 23, 59, 59, 0, time.UTC),
		TrackingMethod:     "automatic",
		SuccessCriteria:    `["Feature has design doc", "Touches 3+ services", ">500 lines of code", "Successfully deployed to production"]`,
		Status:             "active",
		ProgressPercentage: 0,
	}

	if err := store.CreateGoal(individualGoal); err != nil {
		return fmt.Errorf("failed to create individual goal: %w", err)
	}

	// Add milestone for complex features
	featureMilestone := &models.GoalMilestone{
		GoalID:       individualGoal.ID,
		Title:        "Complete complex features",
		Description:  "Lead and deliver 2 complex features",
		TargetValue:  2.0,
		CurrentValue: 0.0,
		Unit:         "complex features",
		DueDate:      timePtr(time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC)),
	}
	if err := store.CreateMilestone(featureMilestone); err != nil {
		return fmt.Errorf("failed to create feature milestone: %w", err)
	}

	// Create mentoring goal
	mentoringGoal := &models.Goal{
		Title:              "Mentor junior engineer",
		Description:        "Support the growth of a junior engineer through regular pairing sessions and code reviews.",
		GoalType:           "individual",
		OwnerType:          "engineer",
		OwnerID:            &engineerID,
		TimePeriod:         "Q1 2025",
		StartDate:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:            time.Date(2025, 3, 31, 23, 59, 59, 0, time.UTC),
		TrackingMethod:     "hybrid",
		Status:             "active",
		ProgressPercentage: 0,
	}

	if err := store.CreateGoal(mentoringGoal); err != nil {
		return fmt.Errorf("failed to create mentoring goal: %w", err)
	}

	// Add milestones for mentoring
	mentoringMilestones := []*models.GoalMilestone{
		{
			GoalID:       mentoringGoal.ID,
			Title:        "Conduct pairing sessions",
			Description:  "Hold at least 12 pairing sessions (1 per week)",
			TargetValue:  12.0,
			CurrentValue: 0.0,
			Unit:         "pairing sessions",
			DueDate:      timePtr(time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC)),
		},
		{
			GoalID:       mentoringGoal.ID,
			Title:        "Provide thorough code reviews",
			Description:  "Review mentee's PRs with detailed feedback (20+ reviews)",
			TargetValue:  20.0,
			CurrentValue: 0.0,
			Unit:         "reviews",
			DueDate:      timePtr(time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC)),
		},
	}

	for _, milestone := range mentoringMilestones {
		if err := store.CreateMilestone(milestone); err != nil {
			return fmt.Errorf("failed to create mentoring milestone: %w", err)
		}
	}

	// Create org-level goal
	orgGoal := &models.Goal{
		Title:              "Improve engineering satisfaction to 4.5/5",
		Description:        "Increase team satisfaction scores through better processes, tools, and culture initiatives.",
		GoalType:           "org",
		OwnerType:          "org",
		OwnerID:            nil,
		TimePeriod:         "2025",
		StartDate:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:            time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
		TrackingMethod:     "manual",
		Status:             "active",
		ProgressPercentage: 0,
	}

	if err := store.CreateGoal(orgGoal); err != nil {
		return fmt.Errorf("failed to create org goal: %w", err)
	}

	return nil
}

// timePtr returns a pointer to a time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}
