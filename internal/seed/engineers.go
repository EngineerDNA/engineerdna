package seed

import (
	"fmt"
	"log"

	"github.com/engineerdna/engineerdna/internal/models"
)

// SeedEngineers creates engineers and assigns them to teams
func SeedEngineers(data *SeedData) error {
	fmt.Println("\n4. Creating engineers...")

	engineers := []EngineerDef{
		{
			Name:    "Sarah Chen",
			Email:   "sarah.chen@example.com",
			Manager: "",
			Role:    "Staff Engineer",
			TeamID:  data.EngineeringTeam.ID,
			Identifiers: map[string]string{
				"github": "schen",
				"jira":   "sarah.chen@example.com",
			},
		},
		{
			Name:    "Marcus Johnson",
			Email:   "marcus.johnson@example.com",
			Manager: "Sarah Chen",
			Role:    "Senior Engineer",
			TeamID:  data.BackendTeam.ID,
			Identifiers: map[string]string{
				"github": "mjohnson",
				"jira":   "marcus.johnson@example.com",
			},
		},
		{
			Name:    "Elena Rodriguez",
			Email:   "elena.rodriguez@example.com",
			Manager: "Sarah Chen",
			Role:    "Senior Engineer",
			TeamID:  data.BackendTeam.ID,
			Identifiers: map[string]string{
				"github": "erodriguez",
				"jira":   "elena.rodriguez@example.com",
			},
		},
		{
			Name:    "David Kim",
			Email:   "david.kim@example.com",
			Manager: "Sarah Chen",
			Role:    "Mid-Level Engineer",
			TeamID:  data.BackendTeam.ID,
			Identifiers: map[string]string{
				"github": "dkim",
				"jira":   "david.kim@example.com",
			},
		},
		{
			Name:    "Priya Sharma",
			Email:   "priya.sharma@example.com",
			Manager: "Sarah Chen",
			Role:    "Senior Engineer",
			TeamID:  data.FrontendTeam.ID,
			Identifiers: map[string]string{
				"github": "psharma",
				"jira":   "priya.sharma@example.com",
			},
		},
		{
			Name:    "Alex Thompson",
			Email:   "alex.thompson@example.com",
			Manager: "Sarah Chen",
			Role:    "Mid-Level Engineer",
			TeamID:  data.FrontendTeam.ID,
			Identifiers: map[string]string{
				"github": "athompson",
				"jira":   "alex.thompson@example.com",
			},
		},
		{
			Name:    "Maya Patel",
			Email:   "maya.patel@example.com",
			Manager: "Sarah Chen",
			Role:    "Junior Engineer",
			TeamID:  data.FrontendTeam.ID,
			Identifiers: map[string]string{
				"github": "mpatel",
				"jira":   "maya.patel@example.com",
			},
		},
		{
			Name:    "James Wilson",
			Email:   "james.wilson@example.com",
			Manager: "Sarah Chen",
			Role:    "Senior Engineer",
			TeamID:  data.PlatformTeam.ID,
			Identifiers: map[string]string{
				"github": "jwilson",
				"jira":   "james.wilson@example.com",
			},
		},
		{
			Name:    "Li Wei",
			Email:   "li.wei@example.com",
			Manager: "Sarah Chen",
			Role:    "Mid-Level Engineer",
			TeamID:  data.PlatformTeam.ID,
			Identifiers: map[string]string{
				"github": "lwei",
				"jira":   "li.wei@example.com",
			},
		},
		{
			Name:    "Nina Okafor",
			Email:   "nina.okafor@example.com",
			Manager: "Sarah Chen",
			Role:    "Junior Engineer",
			TeamID:  data.PlatformTeam.ID,
			Identifiers: map[string]string{
				"github": "nokafor",
				"jira":   "nina.okafor@example.com",
			},
		},
	}

	data.EngineerIDs = make([]string, len(engineers))
	for i, eng := range engineers {
		engineerID, err := data.IdentityService.CreateEngineer(eng.Name, eng.Email, eng.Manager, eng.Identifiers)
		if err != nil {
			return fmt.Errorf("failed to create engineer %s: %w", eng.Name, err)
		}
		data.EngineerIDs[i] = engineerID
		data.EngineerMap[engineerID] = eng

		// Assign role
		roleID := data.RoleMap[eng.Role]
		_, err = data.Database.DB.Exec("UPDATE engineers SET role_id = ? WHERE id = ?", roleID, engineerID)
		if err != nil {
			return fmt.Errorf("failed to assign role to %s: %w", eng.Name, err)
		}

		// Add to team
		_, err = data.TeamStore.AddTeamMember(eng.TeamID, engineerID, eng.Role)
		if err != nil {
			return fmt.Errorf("failed to add %s to team: %w", eng.Name, err)
		}

		fmt.Printf("  Created: %s (%s) -> %s team\n", eng.Name, eng.Role, getTeamName(eng.TeamID, data.EngineeringTeam, data.BackendTeam, data.FrontendTeam, data.PlatformTeam))
	}

	// Update team manager IDs
	sarahID := data.EngineerIDs[0] // Sarah Chen is the manager
	err := data.TeamStore.UpdateTeam(data.EngineeringTeam.ID, data.EngineeringTeam.Name, nil, &sarahID)
	if err != nil {
		log.Printf("Warning: Failed to set team manager: %v", err)
	}

	return nil
}

func getTeamName(teamID string, engineering, backend, frontend, platform *models.Team) string {
	switch teamID {
	case engineering.ID:
		return "Engineering"
	case backend.ID:
		return "Backend"
	case frontend.ID:
		return "Frontend"
	case platform.ID:
		return "Platform"
	default:
		return "Unknown"
	}
}
