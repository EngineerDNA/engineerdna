package seed

import (
	"fmt"
)

// SeedTeams creates a multi-level team hierarchy
func SeedTeams(data *SeedData) error {
	fmt.Println("\n3. Creating multi-level team hierarchy...")

	// Level 1: Engineering Org (VP level)
	engineeringTeam, err := data.TeamStore.CreateTeam("Engineering", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create Engineering team: %w", err)
	}
	data.EngineeringTeam = engineeringTeam
	data.TeamsByName["Engineering"] = engineeringTeam
	fmt.Printf("  Created: %s (Level 1 - VP)\n", engineeringTeam.Name)

	// Level 2: Backend Engineering (Director level)
	backendEngineering, err := data.TeamStore.CreateTeam("Backend Engineering", &engineeringTeam.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create Backend Engineering team: %w", err)
	}
	data.BackendTeam = backendEngineering
	data.TeamsByName["Backend Engineering"] = backendEngineering
	fmt.Printf("  Created: %s (Level 2 - Director)\n", backendEngineering.Name)

	// Level 3: Backend sub-teams (Manager level)
	apiTeam, err := data.TeamStore.CreateTeam("API Team", &backendEngineering.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create API Team: %w", err)
	}
	data.TeamsByName["API Team"] = apiTeam
	fmt.Printf("  Created: %s (Level 3 - Manager)\n", apiTeam.Name)

	infrastructureTeam, err := data.TeamStore.CreateTeam("Infrastructure", &backendEngineering.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create Infrastructure team: %w", err)
	}
	data.TeamsByName["Infrastructure"] = infrastructureTeam
	fmt.Printf("  Created: %s (Level 3 - Manager)\n", infrastructureTeam.Name)

	// Level 2: Frontend Engineering (Director level)
	frontendEngineering, err := data.TeamStore.CreateTeam("Frontend Engineering", &engineeringTeam.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create Frontend Engineering team: %w", err)
	}
	data.FrontendTeam = frontendEngineering
	data.TeamsByName["Frontend Engineering"] = frontendEngineering
	fmt.Printf("  Created: %s (Level 2 - Director)\n", frontendEngineering.Name)

	// Level 3: Frontend sub-teams (Manager level)
	webTeam, err := data.TeamStore.CreateTeam("Web Team", &frontendEngineering.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create Web Team: %w", err)
	}
	data.TeamsByName["Web Team"] = webTeam
	fmt.Printf("  Created: %s (Level 3 - Manager)\n", webTeam.Name)

	mobileTeam, err := data.TeamStore.CreateTeam("Mobile Team", &frontendEngineering.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create Mobile Team: %w", err)
	}
	data.TeamsByName["Mobile Team"] = mobileTeam
	fmt.Printf("  Created: %s (Level 3 - Manager)\n", mobileTeam.Name)

	// Level 2: Data Engineering (Manager level)
	dataEngineering, err := data.TeamStore.CreateTeam("Data Engineering", &engineeringTeam.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create Data Engineering team: %w", err)
	}
	data.PlatformTeam = dataEngineering
	data.TeamsByName["Data Engineering"] = dataEngineering
	fmt.Printf("  Created: %s (Level 2 - Manager)\n", dataEngineering.Name)

	fmt.Printf("\nTotal teams created: %d (3-level hierarchy)\n", len(data.TeamsByName))
	return nil
}
