package seed

import (
	"fmt"
	"log"

	"github.com/engineerdna/engineerdna/internal/models"
)

// SeedSprints creates sprints and stories
func SeedSprints(data *SeedData) error {
	fmt.Println("\n8. Creating sprints...")

	sprintConfigs := []struct {
		name            string
		weeksAgo        int
		status          string
		committedPoints int
		completedPoints int
		capacity        int
	}{
		{"Sprint 45", 4, "completed", 85, 82, 30},
		{"Sprint 46", 2, "active", 90, 45, 30},
		{"Sprint 47", 0, "planning", 0, 0, 30},
		{"Sprint 48", -2, "planning", 0, 0, 30},
	}

	sprintTeams := []struct {
		id   string
		name string
	}{
		{data.BackendTeam.ID, "Backend"},
		{data.FrontendTeam.ID, "Frontend"},
		{data.PlatformTeam.ID, "Platform"},
	}

	for _, team := range sprintTeams {
		for _, s := range sprintConfigs {
			startDate := data.Now.AddDate(0, 0, -7*s.weeksAgo)
			endDate := startDate.AddDate(0, 0, 14) // 2-week sprints

			sprint := &models.Sprint{
				Name:            fmt.Sprintf("%s - %s", team.name, s.name),
				StartDate:       startDate,
				EndDate:         endDate,
				CommittedPoints: s.committedPoints,
				CompletedPoints: s.completedPoints,
				TeamCapacity:    s.capacity,
				Status:          s.status,
				TeamID:          team.id,
			}

			if err := data.PlanningStore.CreateSprint(sprint); err != nil {
				log.Printf("Warning: Failed to create sprint: %v", err)
			} else {
				data.SprintIDs = append(data.SprintIDs, sprint.ID)
				fmt.Printf("  Created sprint: %s (%s, %d/%d points)\n", sprint.Name, s.status, s.completedPoints, s.committedPoints)
			}
		}
	}

	// 9. Create Stories
	fmt.Println("\n9. Creating stories...")
	stories := []struct {
		title       string
		points      int
		status      string
		sprintIndex int
		assignee    int // engineer index
	}{
		{"Implement user authentication", 8, "completed", 0, 1},
		{"Add OAuth2 support", 5, "completed", 0, 2},
		{"Create API rate limiting", 5, "completed", 0, 1},
		{"Build dashboard UI", 8, "completed", 0, 4},
		{"Add chart components", 5, "completed", 0, 5},
		{"Implement data export", 5, "completed", 0, 7},

		{"Refactor backend services", 13, "in_progress", 1, 1},
		{"Optimize database queries", 8, "in_progress", 1, 2},
		{"Add frontend tests", 5, "in_progress", 1, 5},
		{"Update documentation", 3, "in_progress", 1, 8},
		{"Fix critical bugs", 8, "completed", 1, 3},

		{"Build notification system", 13, "todo", 2, -1},
		{"Add email templates", 5, "todo", 2, -1},
		{"Implement webhooks", 8, "todo", 2, -1},
		{"Create admin panel", 13, "todo", 2, -1},
	}

	for i, s := range stories {
		var assigneeID *string
		if s.assignee >= 0 && s.assignee < len(data.EngineerIDs) {
			assigneeID = &data.EngineerIDs[s.assignee]
		}

		var sprintID *string
		if s.sprintIndex < len(data.SprintIDs) && data.SprintIDs[s.sprintIndex] != "" {
			sprintID = &data.SprintIDs[s.sprintIndex]
		}

		story := &models.Story{
			Title:       s.title,
			StoryPoints: s.points,
			Status:      s.status,
			SprintID:    sprintID,
			AssigneeID:  assigneeID,
		}

		if err := data.PlanningStore.CreateStory(story); err != nil {
			log.Printf("Warning: Failed to create story %d: %v", i, err)
		}
	}
	fmt.Printf("  Created %d stories across sprints\n", len(stories))

	return nil
}
