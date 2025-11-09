package cost

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Service handles cost calculation business logic
type Service struct {
	store          *db.CostROIStore
	teamStore      *db.TeamStore
	attributeStore *db.AttributeStore
}

// NewService creates a new cost service
func NewService(store *db.CostROIStore, teamStore *db.TeamStore, attributeStore *db.AttributeStore) *Service {
	return &Service{
		store:          store,
		teamStore:      teamStore,
		attributeStore: attributeStore,
	}
}

// CalculateEngineerCost gets current cost for an engineer
func (s *Service) CalculateEngineerCost(engineerID string) (float64, error) {
	// Get cost from entity_attributes
	attr, err := s.attributeStore.GetCurrentAttributeValue("engineer", engineerID, "monthly_cost")
	if err != nil {
		return 0, fmt.Errorf("failed to get engineer cost: %w", err)
	}

	if attr != nil {
		// Parse cost value
		var cost float64
		if costStr, ok := attr.(string); ok {
			_, err := fmt.Sscanf(costStr, "%f", &cost)
			if err == nil {
				return cost, nil
			}
		}
	}

	// If no specific config, try role-based default
	role := GetRoleForEngineer(engineerID)
	attr, err = s.attributeStore.GetCurrentAttributeValue("role", role, "monthly_cost")
	if err != nil {
		return 0, fmt.Errorf("failed to get role cost: %w", err)
	}

	if attr != nil {
		var cost float64
		if costStr, ok := attr.(string); ok {
			_, err := fmt.Sscanf(costStr, "%f", &cost)
			if err == nil {
				return cost, nil
			}
		}
	}

	return 0, fmt.Errorf("no cost configuration found for engineer: %s", engineerID)
}

// CalculateTeamCost aggregates team costs
func (s *Service) CalculateTeamCost(teamID string) (float64, error) {
	// Get team members (active only)
	members, err := s.teamStore.GetTeamMembers(teamID, false)
	if err != nil {
		return 0, fmt.Errorf("failed to get team members: %w", err)
	}

	totalCost := 0.0
	for _, member := range members {
		cost, err := s.CalculateEngineerCost(member.MemberID)
		if err != nil {
			// Skip if no cost found, log warning
			continue
		}
		totalCost += cost
	}

	return totalCost, nil
}

// GetCostPerStoryPoint derives cost per point for a team
func (s *Service) GetCostPerStoryPoint(teamID string, timePeriod string) (float64, error) {
	// Get team cost
	teamCost, err := s.CalculateTeamCost(teamID)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate team cost: %w", err)
	}

	// Get investment data for period
	investments, err := s.store.GetEngineeringInvestment(timePeriod, &teamID, 100)
	if err != nil {
		return 0, fmt.Errorf("failed to get investment data: %w", err)
	}

	totalPoints := 0
	for _, inv := range investments {
		totalPoints += inv.StoryPoints
	}

	if totalPoints == 0 {
		return 0, fmt.Errorf("no story points tracked for period: %s", timePeriod)
	}

	return teamCost / float64(totalPoints), nil
}

// GetCostPerPR calculates cost per PR for a team
func (s *Service) GetCostPerPR(teamID string, prCount int) (float64, error) {
	if prCount == 0 {
		return 0, fmt.Errorf("no PRs in period")
	}

	teamCost, err := s.CalculateTeamCost(teamID)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate team cost: %w", err)
	}

	return teamCost / float64(prCount), nil
}

// CalculateFeatureCost computes total cost for a feature
func (s *Service) CalculateFeatureCost(feature *models.FeatureValue) (*models.FeatureCost, error) {
	// Get work items for feature
	workItems, err := s.store.GetFeatureWorkItems(feature.ID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get work items: %w", err)
	}

	if len(workItems) == 0 {
		return nil, fmt.Errorf("no work items linked to feature: %s", feature.ID)
	}

	totalPoints := 0
	totalHours := 0.0
	costBreakdown := make(map[string]float64)

	for _, item := range workItems {
		// Add story points
		if item.StoryPoints != nil {
			totalPoints += *item.StoryPoints
		}

		// Add hours
		if item.ActualHours != nil {
			totalHours += *item.ActualHours
		}

		// Calculate cost if engineer is known
		if item.EngineerID != nil && item.ActualHours != nil {
			engineerCost, err := s.CalculateEngineerCost(*item.EngineerID)
			if err != nil {
				// Skip if no cost found
				continue
			}

			// Convert monthly to hourly: 160 hours/month
			hourlyRate := engineerCost / 160.0

			// Calculate cost for this work item
			itemCost := hourlyRate * *item.ActualHours

			// Add to breakdown by team
			teamKey := "unknown"
			if item.TeamID != nil {
				teamKey = *item.TeamID
			}
			costBreakdown[teamKey] += itemCost
		}
	}

	// Calculate total cost
	totalCost := 0.0
	for _, cost := range costBreakdown {
		totalCost += cost
	}

	// Marshal breakdown to JSON
	breakdownJSON, err := json.Marshal(costBreakdown)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cost breakdown: %w", err)
	}

	cost := &models.FeatureCost{
		FeatureID:         feature.ID,
		TotalStoryPoints:  totalPoints,
		TotalHours:        totalHours,
		TotalCost:         totalCost,
		CostBreakdown:     string(breakdownJSON),
		ComputationMethod: "hours",
		ComputedAt:        time.Now().UTC(),
	}

	// Save to database
	if err := s.store.CreateFeatureCost(cost); err != nil {
		return nil, fmt.Errorf("failed to save feature cost: %w", err)
	}

	return cost, nil
}

// GetDefaultCosts returns industry default costs by role
func (s *Service) GetDefaultCosts() map[string]float64 {
	return map[string]float64{
		"junior":    15000.0, // $15k/month
		"mid":       20000.0, // $20k/month
		"senior":    30000.0, // $30k/month
		"staff":     40000.0, // $40k/month
		"principal": 50000.0, // $50k/month
	}
}

// EstimateCostFromPoints estimates cost from story points
func (s *Service) EstimateCostFromPoints(points int, teamID string, timePeriod string) (float64, error) {
	costPerPoint, err := s.GetCostPerStoryPoint(teamID, timePeriod)
	if err != nil {
		// Use default if no historical data
		// Assume $200 per story point as default
		costPerPoint = 200.0
	}

	return costPerPoint * float64(points), nil
}

// EstimateCostFromHours estimates cost from hours
func (s *Service) EstimateCostFromHours(hours float64, engineerID string) (float64, error) {
	monthlyCost, err := s.CalculateEngineerCost(engineerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get engineer cost: %w", err)
	}

	// Convert monthly to hourly: 160 hours/month
	hourlyRate := monthlyCost / 160.0

	return hourlyRate * hours, nil
}
