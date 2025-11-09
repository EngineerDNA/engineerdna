package seed

import (
	"fmt"
	"log"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/internal/utils"
)

// SeedForecasting creates predictive analytics data
func SeedForecasting(data *SeedData, goalIDs []string) (forecastCount, riskCount int) {
	fmt.Println("\n15. Creating predictive analytics data...")

	// Forecasts
	forecasts := []struct {
		forecastType string
		entityType   string
		entityID     *string
		timeHorizon  string
		value        float64
		confidence   float64
		modelType    string
	}{
		{"sprint_completion", "sprint", &data.SprintIDs[0], "next_sprint", 85.0, 75.0, "linear_regression"},
		{"team_velocity", "team", &data.BackendTeam.ID, "next_quarter", 120.0, 70.0, "moving_average"},
		{"goal_completion", "goal", &goalIDs[0], "Q1_end", 92.0, 65.0, "monte_carlo"},
		{"attrition_risk", "engineer", &data.EngineerIDs[6], "next_quarter", 0.35, 60.0, "linear_regression"},
		{"cost_projection", "team", &data.FrontendTeam.ID, "next_quarter", 375000.0, 80.0, "linear_regression"},
	}

	for _, f := range forecasts {
		forecast := &models.Forecast{
			ForecastType:           f.forecastType,
			EntityType:             f.entityType,
			EntityID:               f.entityID,
			TimeHorizon:            f.timeHorizon,
			PredictedValue:         f.value,
			PredictedDate:          utils.PtrTime(data.Now.AddDate(0, 3, 0)),
			ConfidencePercentage:   f.confidence,
			ConfidenceIntervalLow:  utils.PtrFloat64(f.value * 0.9),
			ConfidenceIntervalHigh: utils.PtrFloat64(f.value * 1.1),
			ModelType:              f.modelType,
			InputData:              `{"historical_data_points": 12}`,
			Assumptions:            `{"assumes_stable_team": true}`,
			ExpiresAt:              utils.PtrTime(data.Now.AddDate(0, 3, 0)),
		}
		if err := data.ForecastingStore.CreateForecast(forecast); err != nil {
			log.Printf("Warning: Failed to create forecast: %v", err)
		} else {
			forecastCount++
		}
	}
	fmt.Printf("  Created %d forecasts\n", forecastCount)

	// Risk Predictions
	risks := []struct {
		riskType    string
		entityType  string
		entityID    *string
		riskScore   float64
		probability float64
		severity    string
	}{
		{"attrition", "engineer", &data.EngineerIDs[6], 0.35, 35.0, "medium"},
		{"timeline_miss", "team", &data.BackendTeam.ID, 0.28, 28.0, "low"},
		{"burnout", "engineer", &data.EngineerIDs[1], 0.42, 42.0, "medium"},
		{"quality_degradation", "team", &data.FrontendTeam.ID, 0.18, 18.0, "low"},
	}

	for _, r := range risks {
		risk := &models.RiskPrediction{
			RiskType:              r.riskType,
			EntityType:            r.entityType,
			EntityID:              r.entityID,
			RiskScore:             r.riskScore,
			ProbabilityPercentage: r.probability,
			ImpactSeverity:        r.severity,
			ContributingFactors:   `["high_workload", "long_hours"]`,
			MitigationSuggestions: `["redistribute_work", "add_resources"]`,
			PredictedAt:           data.Now,
			ValidUntil:            utils.PtrTime(data.Now.AddDate(0, 1, 0)),
		}
		if err := data.ForecastingStore.CreateRiskPrediction(risk); err != nil {
			log.Printf("Warning: Failed to create risk prediction: %v", err)
		} else {
			riskCount++
		}
	}
	fmt.Printf("  Created %d risk predictions\n", riskCount)

	return forecastCount, riskCount
}

// SeedActions creates action tracking data
func SeedActions(data *SeedData) (recommendationCount, actionCount int) {
	fmt.Println("\n16. Creating action tracking data...")

	// Recommendations
	recommendations := []struct {
		sourceType  string
		recType     string
		priority    string
		subjectType string
		subjectID   *string
		title       string
		description string
		status      string
	}{
		{"ai_insight", "check_in", "high", "engineer", &data.EngineerIDs[6], "Check in with Maya Patel", "High workload detected. Schedule 1-on-1 to discuss priorities and workload distribution.", "in_progress"},
		{"alert", "adjust_workload", "high", "team", &data.FrontendTeam.ID, "Address PR review backlog", "12 PRs waiting for review. Consider pair review sessions or dedicated review time.", "completed"},
		{"briefing", "recognize_achievement", "medium", "engineer", &data.EngineerIDs[2], "Recognize Elena's mentoring", "Elena has been providing excellent mentorship to David. Consider recognition.", "pending"},
		{"ai_insight", "skill_development", "medium", "engineer", &data.EngineerIDs[3], "System design training for David", "David shows interest in architecture. Recommend system design course.", "pending"},
		{"manual", "address_blocker", "high", "team", &data.BackendTeam.ID, "Address tech debt", "Tech debt is impacting velocity. Schedule tech debt sprint.", "in_progress"},
	}

	recommendationIDs := make([]string, 0)
	for _, r := range recommendations {
		rec := &models.Recommendation{
			SourceType:         r.sourceType,
			RecommendationType: r.recType,
			Priority:           r.priority,
			SubjectType:        r.subjectType,
			SubjectID:          r.subjectID,
			Title:              r.title,
			Description:        r.description,
			SuggestedActions:   []string{"Schedule meeting", "Discuss options", "Document decision"},
			Context:            `{"confidence": "high", "data_points": 15}`,
			AssignedTo:         &data.EngineerIDs[0],
			Status:             r.status,
		}
		if err := data.ActionsStore.CreateRecommendation(rec); err != nil {
			log.Printf("Warning: Failed to create recommendation: %v", err)
		} else {
			recommendationIDs = append(recommendationIDs, rec.ID)
		}
	}
	fmt.Printf("  Created %d recommendations\n", len(recommendationIDs))

	// Actions
	actions := []struct {
		recIndex    int
		actionType  string
		title       string
		description string
	}{
		{0, "conversation", "1-on-1 with Maya about workload", "Discussed current workload, identified blockers, reprioritized tasks"},
		{1, "process_change", "Implemented daily review time", "Team now dedicates 30 minutes daily to PR reviews"},
		{4, "workload_adjustment", "Scheduled tech debt sprint", "Allocated Sprint 8 for tech debt paydown"},
	}

	for _, a := range actions {
		if a.recIndex >= len(recommendationIDs) {
			continue
		}
		action := &models.Action{
			RecommendationID: &recommendationIDs[a.recIndex],
			ActionType:       a.actionType,
			SubjectType:      recommendations[a.recIndex].subjectType,
			SubjectID:        recommendations[a.recIndex].subjectID,
			Title:            a.title,
			Description:      a.description,
			TakenBy:          data.EngineerIDs[0],
			TakenAt:          data.Now.AddDate(0, 0, -3),
			Evidence:         `{"meeting_notes": "Positive discussion, action items identified"}`,
			ExpectedOutcome:  "Improved workload balance and reduced stress",
			FollowUpDate:     utils.PtrTime(data.Now.AddDate(0, 0, 14)),
		}
		if err := data.ActionsStore.CreateAction(action); err != nil {
			log.Printf("Warning: Failed to create action: %v", err)
		} else {
			actionCount++
		}
	}
	fmt.Printf("  Created %d actions\n", actionCount)

	return len(recommendationIDs), actionCount
}
