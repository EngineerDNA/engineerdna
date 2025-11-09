package seed

import (
	"fmt"

	"github.com/engineerdna/engineerdna/internal/models"
)

// SeedAttributes generates entity attributes (team sizes, costs, locations)
func SeedAttributes(data *SeedData) error {
	fmt.Println("\n7. Generating entity attributes...")

	// Seed team sizes
	teamSizeCount, err := seedTeamSizes(data)
	if err != nil {
		return fmt.Errorf("failed to seed team sizes: %w", err)
	}
	fmt.Printf("  Team sizes: %d attributes\n", teamSizeCount)

	// Seed engineer costs
	engineerCostCount, err := seedEngineerCosts(data)
	if err != nil {
		return fmt.Errorf("failed to seed engineer costs: %w", err)
	}
	fmt.Printf("  Engineer costs: %d attributes\n", engineerCostCount)

	// Seed locations
	locationCount, err := seedLocations(data)
	if err != nil {
		return fmt.Errorf("failed to seed locations: %w", err)
	}
	fmt.Printf("  Locations: %d attributes\n", locationCount)

	fmt.Printf("\nTotal attributes generated: %d\n", teamSizeCount+engineerCostCount+locationCount)

	return nil
}

func seedTeamSizes(data *SeedData) (int, error) {
	count := 0
	validFrom := data.Now.AddDate(-1, 0, 0) // Started 12 months ago

	for _, team := range data.TeamsByName {
		// Get current team members
		members, err := data.TeamStore.GetTeamMembers(team.ID, false)
		if err != nil {
			return count, fmt.Errorf("failed to get team members: %w", err)
		}

		if len(members) == 0 {
			continue
		}

		// Store team size as attribute
		attr := &models.EntityAttribute{
			EntityType:    "team",
			EntityID:      team.ID,
			AttributeName: "team_size",
			Value:         fmt.Sprintf("%d", len(members)),
			ValueType:     "number",
			ValidFrom:     validFrom,
			ValidUntil:    nil, // Still valid
			Source:        "seed",
		}

		if err := data.AttributeStore.Create(attr); err != nil {
			return count, fmt.Errorf("failed to create team size attribute: %w", err)
		}
		count++
	}

	return count, nil
}

func seedEngineerCosts(data *SeedData) (int, error) {
	count := 0
	validFrom := data.Now.AddDate(-1, 0, 0) // Started 12 months ago

	for _, engineerID := range data.EngineerIDs {
		eng := data.EngineerMap[engineerID]

		// Calculate monthly cost based on role (fully-loaded cost)
		monthlyCost := getMonthlyCost(eng.Role)

		// Store cost as attribute
		attr := &models.EntityAttribute{
			EntityType:    "engineer",
			EntityID:      engineerID,
			AttributeName: "monthly_cost",
			Value:         fmt.Sprintf("%.2f", monthlyCost),
			ValueType:     "currency",
			ValidFrom:     validFrom,
			ValidUntil:    nil, // Still valid
			Source:        "seed",
		}

		if err := data.AttributeStore.Create(attr); err != nil {
			return count, fmt.Errorf("failed to create cost attribute: %w", err)
		}
		count++
	}

	return count, nil
}

func seedLocations(data *SeedData) (int, error) {
	count := 0
	validFrom := data.Now.AddDate(-1, 0, 0) // Started 12 months ago

	// Distribute engineers across locations
	locations := []string{"San Francisco", "New York", "Remote"}

	for i, engineerID := range data.EngineerIDs {
		// Round-robin distribution across locations
		location := locations[i%len(locations)]

		// Store location as attribute
		attr := &models.EntityAttribute{
			EntityType:    "engineer",
			EntityID:      engineerID,
			AttributeName: "location",
			Value:         location,
			ValueType:     "string",
			ValidFrom:     validFrom,
			ValidUntil:    nil, // Still valid
			Source:        "seed",
		}

		if err := data.AttributeStore.Create(attr); err != nil {
			return count, fmt.Errorf("failed to create location attribute: %w", err)
		}
		count++
	}

	return count, nil
}

func getMonthlyCost(role string) float64 {
	// Fully-loaded cost (salary + benefits + overhead)
	// Approximate annual salaries / 12
	switch role {
	case "Junior Engineer":
		return 10000.0 // ~120k/year
	case "Mid-Level Engineer":
		return 14000.0 // ~168k/year
	case "Senior Engineer":
		return 18000.0 // ~216k/year
	case "Staff Engineer":
		return 22000.0 // ~264k/year
	default:
		return 12000.0
	}
}
