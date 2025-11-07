package cost

import (
	"fmt"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Analyzer handles investment analysis
type Analyzer struct {
	store      *db.CostROIStore
	eventStore *db.EventStore
	costSvc    *Service
}

// NewAnalyzer creates a new investment analyzer
func NewAnalyzer(store *db.CostROIStore, eventStore *db.EventStore, costSvc *Service) *Analyzer {
	return &Analyzer{
		store:      store,
		eventStore: eventStore,
		costSvc:    costSvc,
	}
}

// CategorizeWork classifies work into investment categories
func (a *Analyzer) CategorizeWork(event *models.Event) string {
	// Extract title and description from event data
	var title, description string
	if titleVal, ok := event.Data["title"]; ok {
		if titleStr, ok := titleVal.(string); ok {
			title = strings.ToLower(titleStr)
		}
	}
	if descVal, ok := event.Data["description"]; ok {
		if descStr, ok := descVal.(string); ok {
			description = strings.ToLower(descStr)
		}
	}

	// For PRs, analyze title and description
	if event.Type == "pull_request" {

		// Customer-facing features
		if strings.Contains(title, "feature") ||
			strings.Contains(description, "user-facing") ||
			strings.Contains(description, "customer") {
			return "customer_features"
		}

		// Bug fixes and quality
		if strings.Contains(title, "fix") ||
			strings.Contains(title, "bug") ||
			strings.Contains(title, "test") {
			return "quality"
		}

		// Refactoring / tech debt
		if strings.Contains(title, "refactor") ||
			strings.Contains(title, "tech debt") ||
			strings.Contains(title, "cleanup") {
			return "tech_debt"
		}

		// Infrastructure
		if strings.Contains(title, "ci/cd") ||
			strings.Contains(title, "docker") ||
			strings.Contains(title, "deploy") ||
			strings.Contains(title, "infra") {
			return "infrastructure"
		}

		// Internal tools
		if strings.Contains(description, "internal") ||
			strings.Contains(description, "admin") ||
			strings.Contains(title, "tool") {
			return "internal_tools"
		}
	}

	// Default
	return "operational"
}

// AnalyzeTimeAllocation analyzes where time is being spent
func (a *Analyzer) AnalyzeTimeAllocation(timePeriod string, teamID *string) (*models.InvestmentBreakdown, error) {
	// Get investment data
	investments, err := a.store.GetEngineeringInvestment(timePeriod, teamID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get investment data: %w", err)
	}

	totalHours := 0.0
	totalCost := 0.0

	for _, inv := range investments {
		totalHours += inv.Hours
		totalCost += inv.Cost
	}

	breakdown := &models.InvestmentBreakdown{
		TimePeriod: timePeriod,
		TeamID:     teamID,
		TotalHours: totalHours,
		TotalCost:  totalCost,
		Categories: investments,
		Currency:   "USD",
	}

	return breakdown, nil
}

// CompareToIndustry benchmarks against industry standards
func (a *Analyzer) CompareToIndustry(breakdown *models.InvestmentBreakdown) map[string]string {
	// Industry benchmarks (percentages)
	benchmarks := map[string]float64{
		"customer_features": 40.0, // 40% of time
		"tech_debt":         20.0, // 20% of time
		"quality":           15.0, // 15% of time
		"internal_tools":    10.0, // 10% of time
		"infrastructure":    10.0, // 10% of time
		"operational":       5.0,  // 5% of time
	}

	results := make(map[string]string)

	for _, inv := range breakdown.Categories {
		percentage := (inv.Hours / breakdown.TotalHours) * 100.0
		benchmark := benchmarks[inv.Category]

		var status string
		if percentage > benchmark*1.2 {
			status = "above_benchmark"
		} else if percentage < benchmark*0.8 {
			status = "below_benchmark"
		} else {
			status = "within_benchmark"
		}

		results[inv.Category] = status
	}

	return results
}

// IdentifyInefficiencies detects potential waste
func (a *Analyzer) IdentifyInefficiencies(breakdown *models.InvestmentBreakdown) []string {
	issues := []string{}

	for _, inv := range breakdown.Categories {
		percentage := (inv.Hours / breakdown.TotalHours) * 100.0

		// Too much operational work
		if inv.Category == "operational" && percentage > 15.0 {
			issues = append(issues, fmt.Sprintf("Operational work is %.1f%% (recommended < 15%%)", percentage))
		}

		// Too much tech debt
		if inv.Category == "tech_debt" && percentage > 30.0 {
			issues = append(issues, fmt.Sprintf("Tech debt is %.1f%% (recommended < 30%%)", percentage))
		}

		// Too little customer features
		if inv.Category == "customer_features" && percentage < 25.0 {
			issues = append(issues, fmt.Sprintf("Customer features only %.1f%% (recommended > 25%%)", percentage))
		}
	}

	return issues
}

// GenerateInvestmentReport creates executive summary
func (a *Analyzer) GenerateInvestmentReport(timePeriod string, teamID *string) (map[string]interface{}, error) {
	breakdown, err := a.AnalyzeTimeAllocation(timePeriod, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze allocation: %w", err)
	}

	comparison := a.CompareToIndustry(breakdown)
	issues := a.IdentifyInefficiencies(breakdown)

	report := map[string]interface{}{
		"time_period":         timePeriod,
		"team_id":             teamID,
		"total_hours":         breakdown.TotalHours,
		"total_cost":          breakdown.TotalCost,
		"categories":          breakdown.Categories,
		"industry_comparison": comparison,
		"inefficiencies":      issues,
		"generated_at":        time.Now().UTC(),
	}

	return report, nil
}

// ComputeInvestmentForPeriod processes all events in a period
func (a *Analyzer) ComputeInvestmentForPeriod(timePeriod string, startDate, endDate time.Time, teamID *string) error {
	// Get all events with filters
	filters := map[string]interface{}{
		"type": "pull_request",
	}

	events, err := a.eventStore.List(filters, 10000, 0)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	// Filter to date range (List doesn't support date filtering)
	var filteredEvents []*models.Event
	for _, event := range events {
		if !event.Timestamp.Before(startDate) && !event.Timestamp.After(endDate) {
			filteredEvents = append(filteredEvents, event)
		}
	}
	events = filteredEvents

	// Categorize and aggregate
	categoryData := make(map[string]*models.EngineeringInvestment)

	for _, event := range events {
		// Note: Events don't have TeamID, so we can't filter by team here
		// In a production system, we'd link events to teams via engineer memberships

		category := a.CategorizeWork(event)

		// Initialize category if needed
		if categoryData[category] == nil {
			categoryData[category] = &models.EngineeringInvestment{
				TimePeriod:   timePeriod,
				TeamID:       teamID,
				Category:     category,
				StoryPoints:  0,
				Hours:        0,
				Cost:         0,
				FeatureCount: 0,
			}
		}

		// Add hours (estimate 8 hours per PR if not available)
		hours := 8.0
		categoryData[category].Hours += hours
		categoryData[category].FeatureCount++

		// Calculate cost
		if event.Actor != "" {
			cost, err := a.costSvc.EstimateCostFromHours(hours, event.Actor)
			if err == nil {
				categoryData[category].Cost += cost
			}
		}
	}

	// Save investment data
	for _, inv := range categoryData {
		if err := a.store.CreateEngineeringInvestment(inv); err != nil {
			return fmt.Errorf("failed to save investment: %w", err)
		}
	}

	return nil
}
