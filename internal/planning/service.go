package planning

import (
	"fmt"
	"math"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Service handles planning calculations and forecasting
type Service struct {
	store *db.PlanningStore
}

// NewService creates a new planning service
func NewService(store *db.PlanningStore) *Service {
	return &Service{store: store}
}

// CalculateSprintHealth assesses the health and risk level of a sprint
func (s *Service) CalculateSprintHealth(sprintID string) (*models.SprintHealth, error) {
	// Get sprint details
	sprint, err := s.store.GetSprint(sprintID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sprint: %w", err)
	}
	if sprint == nil {
		return nil, fmt.Errorf("sprint not found: %s", sprintID)
	}

	// Calculate time metrics
	now := time.Now().UTC()
	totalDuration := sprint.EndDate.Sub(sprint.StartDate)
	elapsed := now.Sub(sprint.StartDate)
	remaining := sprint.EndDate.Sub(now)

	totalDays := int(totalDuration.Hours() / 24)
	daysElapsed := int(elapsed.Hours() / 24)
	daysRemaining := int(remaining.Hours() / 24)

	// Prevent division by zero
	if totalDays == 0 {
		totalDays = 1
	}

	elapsedPct := float64(daysElapsed) / float64(totalDays)
	if elapsedPct > 1.0 {
		elapsedPct = 1.0
	}
	if elapsedPct < 0.0 {
		elapsedPct = 0.0
	}

	// Calculate percent complete
	var percentComplete float64
	if sprint.CommittedPoints > 0 {
		percentComplete = float64(sprint.CompletedPoints) / float64(sprint.CommittedPoints)
	}

	// Get historical average for the team
	historicalAvg := 0.0
	if sprint.TeamID != "" {
		velocity, err := s.GetHistoricalVelocity(sprint.TeamID, 12)
		if err == nil && velocity != nil {
			historicalAvg = velocity.AveragePoints
		}
	}

	// If no historical data, use team capacity as baseline
	if historicalAvg == 0.0 && sprint.TeamCapacity > 0 {
		historicalAvg = float64(sprint.TeamCapacity)
	}

	// Assess risk level
	riskLevel := "on_track"
	expectedComplete := elapsedPct * float64(sprint.CommittedPoints)
	behindPace := float64(sprint.CompletedPoints) < expectedComplete

	overcommitment := 0.0
	if historicalAvg > 0 {
		overcommitment = (float64(sprint.CommittedPoints) - historicalAvg) / historicalAvg * 100
	}

	// Risk assessment logic
	if historicalAvg > 0 {
		if float64(sprint.CommittedPoints) > historicalAvg*1.4 && behindPace {
			riskLevel = "high_risk"
		} else if float64(sprint.CommittedPoints) > historicalAvg*1.2 || behindPace {
			riskLevel = "at_risk"
		}
	} else {
		// No historical data, assess based on pace only
		if behindPace && elapsedPct > 0.5 {
			riskLevel = "at_risk"
		}
	}

	// Generate recommendation
	recommendation := s.generateRecommendation(sprint, riskLevel, overcommitment, historicalAvg)

	return &models.SprintHealth{
		Sprint:          sprint,
		RiskLevel:       riskLevel,
		PercentComplete: percentComplete,
		DaysRemaining:   daysRemaining,
		DaysElapsed:     daysElapsed,
		TotalDays:       totalDays,
		HistoricalAvg:   historicalAvg,
		Overcommitment:  overcommitment,
		Recommendation:  recommendation,
	}, nil
}

// GetHistoricalVelocity calculates team velocity trend
func (s *Service) GetHistoricalVelocity(teamID string, numSprints int) (*models.VelocityTrend, error) {
	// Get last N sprints for the team
	sprints, err := s.store.GetSprintsByTeam(teamID, numSprints)
	if err != nil {
		return nil, fmt.Errorf("failed to get sprints: %w", err)
	}

	if len(sprints) == 0 {
		return nil, fmt.Errorf("no sprints found for team: %s", teamID)
	}

	// Collect completed points
	var completedPoints []int
	for _, sprint := range sprints {
		completedPoints = append(completedPoints, sprint.CompletedPoints)
	}

	// Calculate average
	avg := average(completedPoints)

	// Calculate standard deviation
	stdDev := standardDeviation(completedPoints, avg)

	// Determine trend (compare first half vs second half)
	trend := determineTrend(completedPoints)

	// Determine confidence level based on standard deviation
	confidenceLevel := "medium"
	if stdDev < avg*0.15 {
		confidenceLevel = "high" // Low variance = high confidence
	} else if stdDev > avg*0.30 {
		confidenceLevel = "low" // High variance = low confidence
	}

	return &models.VelocityTrend{
		TeamID:          teamID,
		AveragePoints:   avg,
		Trend:           trend,
		Last12Sprints:   completedPoints,
		StdDeviation:    stdDev,
		ConfidenceLevel: confidenceLevel,
	}, nil
}

// EstimateTimeline forecasts delivery timeline for a feature
func (s *Service) EstimateTimeline(featureName string, estimatedPoints int, teamID string) (*models.TimelineEstimate, error) {
	// Get team's historical velocity
	velocity, err := s.GetHistoricalVelocity(teamID, 12)
	if err != nil || velocity == nil {
		// No historical data, provide a cautious estimate
		return &models.TimelineEstimate{
			FeatureName:     featureName,
			EstimatedPoints: estimatedPoints,
			BestCase:        0,
			Likely:          0,
			WorstCase:       0,
			ConfidenceLevel: "low",
			Recommendation:  "No historical data available. Complete a few sprints first to establish baseline velocity.",
			TeamID:          teamID,
			TeamVelocity:    0,
		}, nil
	}

	// Calculate timeline estimates
	// Assume 2-week sprints (standard)
	weeksPerSprint := 2.0

	// Best case: velocity + 1 std dev
	bestCaseVelocity := velocity.AveragePoints + velocity.StdDeviation
	if bestCaseVelocity <= 0 {
		bestCaseVelocity = velocity.AveragePoints
	}
	bestCaseSprints := float64(estimatedPoints) / bestCaseVelocity
	bestCaseWeeks := int(math.Ceil(bestCaseSprints * weeksPerSprint))

	// Likely case: average velocity
	likelyVelocity := velocity.AveragePoints
	if likelyVelocity <= 0 {
		likelyVelocity = 1 // Minimum to avoid division by zero
	}
	likelySprints := float64(estimatedPoints) / likelyVelocity
	likelyWeeks := int(math.Ceil(likelySprints * weeksPerSprint))

	// Worst case: velocity - 1 std dev
	worstCaseVelocity := velocity.AveragePoints - velocity.StdDeviation
	if worstCaseVelocity <= 0 {
		worstCaseVelocity = velocity.AveragePoints * 0.5 // At least half of average
	}
	worstCaseSprints := float64(estimatedPoints) / worstCaseVelocity
	worstCaseWeeks := int(math.Ceil(worstCaseSprints * weeksPerSprint))

	// Generate recommendation
	recommendation := s.generateTimelineRecommendation(estimatedPoints, velocity, likelyWeeks)

	return &models.TimelineEstimate{
		FeatureName:     featureName,
		EstimatedPoints: estimatedPoints,
		BestCase:        bestCaseWeeks,
		Likely:          likelyWeeks,
		WorstCase:       worstCaseWeeks,
		ConfidenceLevel: velocity.ConfidenceLevel,
		Recommendation:  recommendation,
		TeamID:          teamID,
		TeamVelocity:    velocity.AveragePoints,
	}, nil
}

// Helper functions

func (s *Service) generateRecommendation(sprint *models.Sprint, riskLevel string, overcommitment, historicalAvg float64) string {
	if riskLevel == "on_track" {
		return "Sprint is on track to meet commitment. Continue current pace."
	}

	recommendation := ""

	if riskLevel == "high_risk" {
		recommendation = "High risk of missing commitment. "
	} else {
		recommendation = "Sprint is at risk. "
	}

	if overcommitment > 20 {
		pointsOver := sprint.CommittedPoints - int(historicalAvg)
		recommendation += fmt.Sprintf("Consider moving %d points to next sprint. ", pointsOver)
		if historicalAvg > 0 {
			recommendation += fmt.Sprintf("Team historically completes %.0f points per sprint. ", historicalAvg)
		}
	}

	if sprint.CommittedPoints > 0 {
		remaining := sprint.CommittedPoints - sprint.CompletedPoints
		recommendation += fmt.Sprintf("Focus on completing %d remaining points. ", remaining)
	}

	recommendation += "Review scope in daily standup and adjust as needed."

	return recommendation
}

func (s *Service) generateTimelineRecommendation(estimatedPoints int, velocity *models.VelocityTrend, likelyWeeks int) string {
	recommendation := fmt.Sprintf("Based on team velocity of %.0f points per sprint, ", velocity.AveragePoints)

	if velocity.Trend == "decreasing" {
		recommendation += "note that velocity is declining. Consider addressing capacity constraints. "
	} else if velocity.Trend == "increasing" {
		recommendation += "velocity is improving. Timeline may be optimistic. "
	}

	if likelyWeeks > 12 {
		recommendation += "This is a large feature. Consider breaking into smaller deliverables or adding engineers."
	} else if likelyWeeks > 8 {
		recommendation += "Consider phased delivery to show progress incrementally."
	} else {
		recommendation += "Timeline is reasonable for current team capacity."
	}

	return recommendation
}

// average calculates the mean of a slice of integers
func average(values []int) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0
	for _, v := range values {
		sum += v
	}

	return float64(sum) / float64(len(values))
}

// standardDeviation calculates the standard deviation
func standardDeviation(values []int, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := float64(v) - mean
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(values))
	return math.Sqrt(variance)
}

// determineTrend analyzes whether velocity is increasing, decreasing, or stable
func determineTrend(values []int) string {
	if len(values) < 4 {
		return "stable" // Not enough data to determine trend
	}

	// Split into first half and second half
	// values are in DESC order (newest first), so:
	// - values[:midpoint] are the NEWER sprints
	// - values[midpoint:] are the OLDER sprints
	midpoint := len(values) / 2
	newerSprints := values[:midpoint] // First half of array = newer
	olderSprints := values[midpoint:] // Second half of array = older

	newerAvg := average(newerSprints)
	olderAvg := average(olderSprints)

	// Calculate percent change from older to newer
	if olderAvg == 0 {
		return "stable"
	}

	percentChange := (newerAvg - olderAvg) / olderAvg * 100

	// Trend threshold: ±10%
	if percentChange > 10 {
		return "increasing"
	} else if percentChange < -10 {
		return "decreasing"
	}

	return "stable"
}
