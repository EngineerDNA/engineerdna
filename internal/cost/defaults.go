package cost

import (
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// CreateDefaultCostConfiguration sets up industry standard costs
func CreateDefaultCostConfiguration(store *db.CostROIStore) error {
	defaultCosts := map[string]float64{
		"junior":    15000.0, // $15k/month
		"mid":       20000.0, // $20k/month
		"senior":    30000.0, // $30k/month
		"staff":     40000.0, // $40k/month
		"principal": 50000.0, // $50k/month
	}

	now := time.Now().UTC()

	// Check if defaults already exist
	existing, _, err := store.ListCostConfigurations("org", 0, 0)
	if err != nil {
		return fmt.Errorf("failed to check existing configurations: %w", err)
	}

	if len(existing) > 0 {
		// Defaults already exist
		return nil
	}

	// Create default configs for each role
	for role, cost := range defaultCosts {
		config := &models.CostConfiguration{
			EntityType:    "org",
			EntityID:      nil, // org-wide default
			Role:          &role,
			MonthlyCost:   cost,
			Currency:      "USD",
			EffectiveFrom: now,
			EffectiveTo:   nil, // current
			Notes:         "Industry standard fully-loaded cost (salary + benefits + overhead)",
		}

		if err := store.CreateCostConfiguration(config); err != nil {
			return fmt.Errorf("failed to create default cost for role %s: %w", role, err)
		}
	}

	return nil
}

// GetRoleForEngineer determines role based on engineer data
// This is a simplified version - in production, this would be more sophisticated
func GetRoleForEngineer(engineerID string) string {
	// Default to mid-level
	// In production, this would query the engineers table or identity system
	return "mid"
}

// EstimateFeatureValue provides default value estimates based on feature type
func EstimateFeatureValue(featureType string) float64 {
	estimates := map[string]float64{
		"customer_acquisition":  50000.0,  // $50k ARR impact
		"retention_improvement": 30000.0,  // $30k ARR impact
		"efficiency_gain":       20000.0,  // $20k cost savings
		"arr_impact":            100000.0, // $100k ARR impact
	}

	if value, ok := estimates[featureType]; ok {
		return value
	}

	// Default estimate
	return 25000.0
}

// StandardCategories returns standard investment categories
func StandardCategories() []string {
	return []string{
		"customer_features",
		"internal_tools",
		"quality",
		"tech_debt",
		"infrastructure",
		"operational",
	}
}
