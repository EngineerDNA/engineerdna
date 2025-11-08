package seed

import (
	"fmt"
	"log"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/internal/utils"
)

// SeedGoals creates goal tracking data
func SeedGoals(data *SeedData) (goalIDs []string, milestoneCount, progressLogCount int) {
	fmt.Println("\n11. Creating goal tracking data...")

	goals := []struct {
		title       string
		goalType    string
		ownerType   string
		ownerID     *string
		timePeriod  string
		tracking    string
		status      string
		progress    int
		weeksOffset int
	}{
		{"Improve team velocity by 20%", "team", "team", &data.BackendTeam.ID, "Q1 2025", "automatic", "active", 45, -4},
		{"Reduce average cycle time to under 36 hours", "team", "team", &data.FrontendTeam.ID, "Q1 2025", "automatic", "active", 60, -4},
		{"Achieve 90% test coverage", "team", "team", &data.PlatformTeam.ID, "Q1 2025", "automatic", "at_risk", 30, -4},
		{"Lead 2 major feature initiatives", "individual", "engineer", &data.EngineerIDs[1], "Q1 2025", "manual", "active", 50, -4},
		{"Mentor 2 junior engineers", "individual", "engineer", &data.EngineerIDs[2], "Q1 2025", "hybrid", "active", 75, -4},
		{"Learn Kubernetes and deploy 3 services", "individual", "engineer", &data.EngineerIDs[8], "Q1 2025", "manual", "active", 33, -4},
		{"Improve code review response time", "team", "team", &data.EngineeringTeam.ID, "Q1 2025", "automatic", "active", 55, -4},
		{"Complete Go certification", "individual", "engineer", &data.EngineerIDs[3], "Q1 2025", "manual", "completed", 100, -8},
	}

	goalIDs = make([]string, 0)
	for _, g := range goals {
		startDate := data.Now.AddDate(0, 0, g.weeksOffset*7)
		endDate := startDate.AddDate(0, 3, 0)

		goal := &models.Goal{
			Title:              g.title,
			Description:        fmt.Sprintf("Goal for %s to %s", g.ownerType, g.title),
			GoalType:           g.goalType,
			OwnerType:          g.ownerType,
			OwnerID:            g.ownerID,
			TimePeriod:         g.timePeriod,
			StartDate:          startDate,
			EndDate:            endDate,
			TrackingMethod:     g.tracking,
			Status:             g.status,
			ProgressPercentage: g.progress,
		}
		if err := data.GoalsStore.CreateGoal(goal); err != nil {
			log.Printf("Warning: Failed to create goal: %v", err)
		} else {
			goalIDs = append(goalIDs, goal.ID)
		}
	}
	fmt.Printf("  Created %d goals\n", len(goalIDs))

	// Goal Milestones
	for i, goalID := range goalIDs {
		numMilestones := 2
		if i < 3 {
			numMilestones = 3
		}
		for j := 0; j < numMilestones; j++ {
			completed := j == 0 && goals[i].progress > 30
			milestone := &models.GoalMilestone{
				GoalID:       goalID,
				Title:        fmt.Sprintf("Milestone %d", j+1),
				Description:  fmt.Sprintf("Key milestone %d for goal", j+1),
				TargetValue:  float64(100 / numMilestones),
				CurrentValue: float64(goals[i].progress / numMilestones),
				Unit:         "percent",
				Completed:    completed,
			}
			if completed {
				compTime := data.Now.AddDate(0, 0, -7*(numMilestones-j))
				milestone.CompletedAt = &compTime
			}
			if err := data.GoalsStore.CreateMilestone(milestone); err != nil {
				log.Printf("Warning: Failed to create milestone: %v", err)
			} else {
				milestoneCount++
			}
		}
	}
	fmt.Printf("  Created %d goal milestones\n", milestoneCount)

	// Goal Progress Logs
	for i, goalID := range goalIDs {
		numLogs := 3
		for j := 0; j < numLogs; j++ {
			prevValue := float64((j * 100) / numLogs)
			newValue := float64(((j + 1) * 100) / numLogs)
			if newValue > float64(goals[i].progress) {
				newValue = float64(goals[i].progress)
			}

			progressLog := &models.GoalProgressLog{
				GoalID:        goalID,
				PreviousValue: &prevValue,
				NewValue:      &newValue,
				ChangeType:    "auto_detected",
				Evidence:      fmt.Sprintf(`{"event_count": %d, "source": "automatic"}`, j+1),
				LoggedAt:      data.Now.AddDate(0, 0, -7*(numLogs-j)),
				LoggedBy:      utils.PtrString("system"),
			}
			if err := data.GoalsStore.LogProgress(progressLog); err != nil {
				log.Printf("Warning: Failed to create progress log: %v", err)
			} else {
				progressLogCount++
			}
		}
	}
	fmt.Printf("  Created %d goal progress logs\n", progressLogCount)

	// Goal Dependencies
	if len(goalIDs) >= 2 {
		dep := &models.GoalDependency{
			GoalID:          goalIDs[1],
			DependsOnGoalID: goalIDs[0],
			DependencyType:  "relates_to",
			Status:          "active",
		}
		if err := data.GoalsStore.CreateDependency(dep); err != nil {
			log.Printf("Warning: Failed to create goal dependency: %v", err)
		}
	}
	fmt.Printf("  Created goal dependencies\n")

	return goalIDs, milestoneCount, progressLogCount
}
