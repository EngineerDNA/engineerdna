package seed

import (
	"fmt"
)

// SeedTeams creates the team hierarchy
func SeedTeams(data *SeedData) error {
	fmt.Println("\n3. Creating team hierarchy...")

	// Engineering parent team
	engineeringTeam, err := data.TeamStore.CreateTeam("Engineering", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create Engineering team: %w", err)
	}
	data.EngineeringTeam = engineeringTeam
	fmt.Printf("  Created team: %s\n", engineeringTeam.Name)

	// Child teams
	backendTeam, err := data.TeamStore.CreateTeam("Backend", &engineeringTeam.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create Backend team: %w", err)
	}
	data.BackendTeam = backendTeam
	fmt.Printf("  Created team: %s (parent: Engineering)\n", backendTeam.Name)

	frontendTeam, err := data.TeamStore.CreateTeam("Frontend", &engineeringTeam.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create Frontend team: %w", err)
	}
	data.FrontendTeam = frontendTeam
	fmt.Printf("  Created team: %s (parent: Engineering)\n", frontendTeam.Name)

	platformTeam, err := data.TeamStore.CreateTeam("Platform", &engineeringTeam.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create Platform team: %w", err)
	}
	data.PlatformTeam = platformTeam
	fmt.Printf("  Created team: %s (parent: Engineering)\n", platformTeam.Name)

	return nil
}
