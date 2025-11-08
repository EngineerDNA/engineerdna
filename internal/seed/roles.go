package seed

import (
	"fmt"
)

// SeedRoles creates roles and scoring weights
func SeedRoles(data *SeedData) error {
	fmt.Println("\n1. Creating roles...")
	roles := []struct {
		name         string
		targetScore  int
		expectations map[string]float64
	}{
		{
			name:        "Junior Engineer",
			targetScore: 70,
			expectations: map[string]float64{
				"throughput":    5.0,
				"quality":       85.0,
				"speed":         3.0,
				"collaboration": 2.0,
			},
		},
		{
			name:        "Mid-Level Engineer",
			targetScore: 90,
			expectations: map[string]float64{
				"throughput":    8.0,
				"quality":       90.0,
				"speed":         2.5,
				"collaboration": 4.0,
			},
		},
		{
			name:        "Senior Engineer",
			targetScore: 100,
			expectations: map[string]float64{
				"throughput":    10.0,
				"quality":       92.0,
				"speed":         2.0,
				"collaboration": 6.0,
			},
		},
		{
			name:        "Staff Engineer",
			targetScore: 110,
			expectations: map[string]float64{
				"throughput":    12.0,
				"quality":       95.0,
				"speed":         1.5,
				"collaboration": 8.0,
			},
		},
	}

	for _, r := range roles {
		role, err := data.ScoringStore.CreateRole(r.name, r.targetScore, r.expectations)
		if err != nil {
			return fmt.Errorf("failed to create role %s: %w", r.name, err)
		}
		data.RoleMap[r.name] = role.ID
		fmt.Printf("  Created role: %s (target: %d)\n", r.name, r.targetScore)
	}

	// Create scoring weights
	fmt.Println("\n2. Creating scoring weights...")
	_, err := data.ScoringStore.UpdateWeights(30.0, 25.0, 20.0, 15.0, 10.0)
	if err != nil {
		return fmt.Errorf("failed to create scoring weights: %w", err)
	}
	fmt.Println("  Created balanced scoring weights (30/25/20/15/10)")

	return nil
}
