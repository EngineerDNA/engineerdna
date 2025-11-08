package forecasting

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/constants"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// PredictAttritionRisk predicts engineer flight risk
func (s *ForecastingService) PredictAttritionRisk(engineerID string) (*models.RiskPrediction, error) {
	factors := []models.RiskFactor{}
	riskScore := 0.0

	// Factor 1: Performance disengagement (declining scores)
	scoreHistory := s.getScoreHistory(engineerID, 90) // last 90 days
	if len(scoreHistory) > 0 && isDecliningSteadily(scoreHistory) {
		decline := (scoreHistory[0] - scoreHistory[len(scoreHistory)-1]) / scoreHistory[0] * 100
		factors = append(factors, models.RiskFactor{
			Name:     "declining_performance",
			Weight:   0.3,
			Evidence: fmt.Sprintf("Score declined %.1f%% over 3 months", decline),
		})
		riskScore += 0.3
	}

	// Factor 2: Sentiment (from surveys/feedback)
	sentiment := s.getRecentSentiment(engineerID, 30)
	if sentiment < -0.3 {
		factors = append(factors, models.RiskFactor{
			Name:     "negative_sentiment",
			Weight:   0.25,
			Evidence: fmt.Sprintf("Survey sentiment at %.2f (negative)", sentiment),
		})
		riskScore += 0.25
	}

	// Factor 3: Behavioral changes (code review participation)
	reviewParticipation := s.getReviewParticipation(engineerID, 30)
	if reviewParticipation < 0.5 {
		factors = append(factors, models.RiskFactor{
			Name:     "disengagement",
			Weight:   0.2,
			Evidence: fmt.Sprintf("Code review participation at %.0f%%", reviewParticipation*100),
		})
		riskScore += 0.2
	}

	// Factor 4: Burnout indicators
	burnoutRisk := s.getBurnoutRisk(engineerID)
	if burnoutRisk > 0.5 {
		factors = append(factors, models.RiskFactor{
			Name:     "burnout_indicators",
			Weight:   0.15,
			Evidence: fmt.Sprintf("Burnout risk at %.0f%%", burnoutRisk*100),
		})
		riskScore += 0.15
	}

	// Factor 5: Low growth opportunities
	skillGrowth := s.getSkillGrowth(engineerID, 180)
	if skillGrowth < 0.1 {
		factors = append(factors, models.RiskFactor{
			Name:     "stagnant_growth",
			Weight:   0.1,
			Evidence: "Limited skill development in past 6 months",
		})
		riskScore += 0.1
	}

	// Convert to probability (0-100%)
	probability := riskScore * 100

	// Determine severity
	severity := "low"
	if riskScore > 0.7 {
		severity = "critical"
	} else if riskScore > 0.5 {
		severity = "high"
	} else if riskScore > 0.3 {
		severity = "medium"
	}

	// Generate mitigations
	mitigations := s.generateAttritionMitigations(factors)

	factorsJSON, err := json.Marshal(factors)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal risk factors: %w", err)
	}

	mitigationsJSON, err := json.Marshal(mitigations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mitigations: %w", err)
	}

	validUntil := time.Now().UTC().Add(constants.RiskValidityPeriod)

	return &models.RiskPrediction{
		ID:                    uuid.New().String(),
		RiskType:              "attrition",
		EntityType:            "engineer",
		EntityID:              &engineerID,
		RiskScore:             riskScore,
		ProbabilityPercentage: probability,
		ImpactSeverity:        severity,
		ContributingFactors:   string(factorsJSON),
		MitigationSuggestions: string(mitigationsJSON),
		PredictedAt:           time.Now().UTC(),
		ValidUntil:            &validUntil,
	}, nil
}

// PredictTimelineMiss predicts risk of project delay
func (s *ForecastingService) PredictTimelineMiss(sprintID string) (*models.RiskPrediction, error) {
	sprint, err := s.planningStore.GetSprint(sprintID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sprint: %w", err)
	}
	if sprint == nil {
		return nil, fmt.Errorf("sprint not found: %s", sprintID)
	}

	factors := []models.RiskFactor{}
	riskScore := 0.0

	// Factor 1: Current pace vs required pace
	daysElapsed := time.Since(sprint.StartDate).Hours() / 24
	daysTotal := sprint.EndDate.Sub(sprint.StartDate).Hours() / 24
	percentTimeElapsed := daysElapsed / daysTotal

	completedPoints := float64(sprint.CompletedPoints)
	totalPoints := float64(sprint.CommittedPoints)
	percentComplete := completedPoints / totalPoints

	if percentTimeElapsed > percentComplete+0.2 {
		factors = append(factors, models.RiskFactor{
			Name:     "behind_schedule",
			Weight:   0.4,
			Evidence: fmt.Sprintf("%.0f%% time elapsed but only %.0f%% complete", percentTimeElapsed*100, percentComplete*100),
		})
		riskScore += 0.4
	}

	// Factor 2: Historical sprint completion rate
	teamCompletionRate := s.getTeamCompletionRate(sprint.TeamID)
	if teamCompletionRate < 0.8 {
		factors = append(factors, models.RiskFactor{
			Name:     "low_completion_rate",
			Weight:   0.3,
			Evidence: fmt.Sprintf("Team completes %.0f%% of committed work on average", teamCompletionRate*100),
		})
		riskScore += 0.3
	}

	// Factor 3: Scope creep
	originalScope := s.getOriginalSprintScope(sprintID)
	currentScope := totalPoints
	if currentScope > originalScope*1.2 {
		factors = append(factors, models.RiskFactor{
			Name:     "scope_creep",
			Weight:   0.2,
			Evidence: fmt.Sprintf("Scope increased %.0f%%", ((currentScope-originalScope)/originalScope)*100),
		})
		riskScore += 0.2
	}

	// Factor 4: Team capacity issues
	teamCapacity := s.getTeamCapacity(sprint.TeamID)
	if teamCapacity < 0.7 {
		factors = append(factors, models.RiskFactor{
			Name:     "reduced_capacity",
			Weight:   0.1,
			Evidence: fmt.Sprintf("Team at %.0f%% capacity", teamCapacity*100),
		})
		riskScore += 0.1
	}

	probability := riskScore * 100

	severity := "low"
	if riskScore > 0.7 {
		severity = "critical"
	} else if riskScore > 0.5 {
		severity = "high"
	} else if riskScore > 0.3 {
		severity = "medium"
	}

	mitigations := s.generateTimelineMitigations(factors, sprint)

	factorsJSON, err := json.Marshal(factors)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal risk factors: %w", err)
	}

	mitigationsJSON, err := json.Marshal(mitigations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mitigations: %w", err)
	}

	return &models.RiskPrediction{
		ID:                    uuid.New().String(),
		RiskType:              "timeline_miss",
		EntityType:            "sprint",
		EntityID:              &sprintID,
		RiskScore:             riskScore,
		ProbabilityPercentage: probability,
		ImpactSeverity:        severity,
		ContributingFactors:   string(factorsJSON),
		MitigationSuggestions: string(mitigationsJSON),
		PredictedAt:           time.Now().UTC(),
		ValidUntil:            &sprint.EndDate,
	}, nil
}

// PredictQualityDegradation predicts risk of quality decline
func (s *ForecastingService) PredictQualityDegradation(teamID string) (*models.RiskPrediction, error) {
	factors := []models.RiskFactor{}
	riskScore := 0.0

	// Factor 1: Bug rate increasing
	bugRateTrend := s.getBugRateTrend(teamID, 60)
	if bugRateTrend > 0.2 {
		factors = append(factors, models.RiskFactor{
			Name:     "increasing_bugs",
			Weight:   0.35,
			Evidence: fmt.Sprintf("Bug rate increased %.0f%% over 2 months", bugRateTrend*100),
		})
		riskScore += 0.35
	}

	// Factor 2: Review quality declining
	reviewDepth := s.getAverageReviewDepth(teamID, 30)
	if reviewDepth < 2.0 {
		factors = append(factors, models.RiskFactor{
			Name:     "shallow_reviews",
			Weight:   0.25,
			Evidence: fmt.Sprintf("Average %.1f comments per review (low)", reviewDepth),
		})
		riskScore += 0.25
	}

	// Factor 3: Velocity pressure
	velocityPressure := s.getVelocityPressure(teamID)
	if velocityPressure > 1.2 {
		factors = append(factors, models.RiskFactor{
			Name:     "velocity_pressure",
			Weight:   0.2,
			Evidence: fmt.Sprintf("Team pushed %.0f%% above sustainable pace", (velocityPressure-1)*100),
		})
		riskScore += 0.2
	}

	// Factor 4: Test coverage declining
	testCoverageTrend := s.getTestCoverageTrend(teamID, 60)
	if testCoverageTrend < -0.1 {
		factors = append(factors, models.RiskFactor{
			Name:     "declining_coverage",
			Weight:   0.2,
			Evidence: "Test coverage decreased over time",
		})
		riskScore += 0.2
	}

	probability := riskScore * 100

	severity := "low"
	if riskScore > 0.7 {
		severity = "critical"
	} else if riskScore > 0.5 {
		severity = "high"
	} else if riskScore > 0.3 {
		severity = "medium"
	}

	mitigations := s.generateQualityMitigations(factors)

	factorsJSON, err := json.Marshal(factors)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal risk factors: %w", err)
	}

	mitigationsJSON, err := json.Marshal(mitigations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mitigations: %w", err)
	}

	validUntil := time.Now().UTC().Add(constants.RiskValidityPeriod)

	return &models.RiskPrediction{
		ID:                    uuid.New().String(),
		RiskType:              "quality_degradation",
		EntityType:            "team",
		EntityID:              &teamID,
		RiskScore:             riskScore,
		ProbabilityPercentage: probability,
		ImpactSeverity:        severity,
		ContributingFactors:   string(factorsJSON),
		MitigationSuggestions: string(mitigationsJSON),
		PredictedAt:           time.Now().UTC(),
		ValidUntil:            &validUntil,
	}, nil
}

// Helper functions for risk prediction

func (s *ForecastingService) getScoreHistory(engineerID string, days int) []float64 {
	startDate := time.Now().UTC().AddDate(0, 0, -days)
	scores, err := s.scoringStore.GetPerformanceScores(engineerID, startDate, time.Now().UTC(), 100)
	if err != nil {
		return []float64{}
	}

	history := make([]float64, len(scores))
	for i, score := range scores {
		history[i] = score.TotalScore
	}
	return history
}

func (s *ForecastingService) getRecentSentiment(engineerID string, days int) float64 {
	// Placeholder - would integrate with survey/feedback system
	return 0.0
}

func (s *ForecastingService) getReviewParticipation(engineerID string, days int) float64 {
	// Placeholder - would query code review participation from events
	return 0.8
}

func (s *ForecastingService) getBurnoutRisk(engineerID string) float64 {
	// Placeholder - would query latest burnout risk from scoring
	return 0.0
}

func (s *ForecastingService) getSkillGrowth(engineerID string, days int) float64 {
	// Placeholder - would query skill tracking system
	return 0.5
}

func (s *ForecastingService) getTeamCompletionRate(teamID string) float64 {
	var avgRate float64
	s.db.QueryRow(`
		SELECT COALESCE(AVG(CAST(completed_points AS REAL) / CAST(committed_points AS REAL)), 0.8)
		FROM sprints
		WHERE team_id = ? AND committed_points > 0
		LIMIT 10
	`, teamID).Scan(&avgRate)
	return avgRate
}

func (s *ForecastingService) getOriginalSprintScope(sprintID string) float64 {
	// Placeholder - would track original scope
	return 50.0
}

func (s *ForecastingService) getTeamCapacity(teamID string) float64 {
	// Placeholder - would query team capacity
	return 0.9
}

func (s *ForecastingService) getBugRateTrend(teamID string, days int) float64 {
	// Placeholder - would analyze bug rate trend
	return 0.1
}

func (s *ForecastingService) getAverageReviewDepth(teamID string, days int) float64 {
	// Placeholder - would analyze review depth
	return 3.5
}

func (s *ForecastingService) getVelocityPressure(teamID string) float64 {
	// Placeholder - would compare recent vs sustainable velocity
	return 1.0
}

func (s *ForecastingService) getTestCoverageTrend(teamID string, days int) float64 {
	// Placeholder - would analyze test coverage trend
	return 0.0
}

func isDecliningSteadily(values []float64) bool {
	if len(values) < 3 {
		return false
	}

	trend := calculateTrend(values)
	return trend < -0.5 // Negative trend
}

// Mitigation generators

func (s *ForecastingService) generateAttritionMitigations(factors []models.RiskFactor) []models.Mitigation {
	mitigations := []models.Mitigation{}

	for _, factor := range factors {
		switch factor.Name {
		case "declining_performance":
			mitigations = append(mitigations, models.Mitigation{
				Action:      "Schedule 1:1 to discuss career goals and blockers",
				Priority:    "high",
				Description: "Understand reasons for performance decline and provide support",
			})
		case "negative_sentiment":
			mitigations = append(mitigations, models.Mitigation{
				Action:      "Address concerns raised in surveys",
				Priority:    "high",
				Description: "Follow up on specific feedback and take action",
			})
		case "disengagement":
			mitigations = append(mitigations, models.Mitigation{
				Action:      "Assign more engaging projects or leadership opportunities",
				Priority:    "medium",
				Description: "Increase engagement through challenging work",
			})
		case "stagnant_growth":
			mitigations = append(mitigations, models.Mitigation{
				Action:      "Create development plan with new skill acquisition goals",
				Priority:    "medium",
				Description: "Provide growth opportunities and training",
			})
		}
	}

	return mitigations
}

func (s *ForecastingService) generateTimelineMitigations(factors []models.RiskFactor, sprint *models.Sprint) []models.Mitigation {
	mitigations := []models.Mitigation{}

	for _, factor := range factors {
		switch factor.Name {
		case "behind_schedule":
			mitigations = append(mitigations, models.Mitigation{
				Action:      "Reduce sprint scope to match current velocity",
				Priority:    "high",
				Description: "Move lower-priority stories to backlog",
			})
		case "scope_creep":
			mitigations = append(mitigations, models.Mitigation{
				Action:      "Freeze scope and defer new requests to next sprint",
				Priority:    "high",
				Description: "Protect team from mid-sprint additions",
			})
		case "reduced_capacity":
			mitigations = append(mitigations, models.Mitigation{
				Action:      "Adjust expectations or extend timeline",
				Priority:    "medium",
				Description: "Align commitments with actual capacity",
			})
		}
	}

	return mitigations
}

func (s *ForecastingService) generateQualityMitigations(factors []models.RiskFactor) []models.Mitigation {
	mitigations := []models.Mitigation{}

	for _, factor := range factors {
		switch factor.Name {
		case "increasing_bugs":
			mitigations = append(mitigations, models.Mitigation{
				Action:      "Schedule bug bash and root cause analysis",
				Priority:    "high",
				Description: "Address bug backlog and identify systemic issues",
			})
		case "shallow_reviews":
			mitigations = append(mitigations, models.Mitigation{
				Action:      "Implement code review guidelines and training",
				Priority:    "high",
				Description: "Improve review thoroughness",
			})
		case "velocity_pressure":
			mitigations = append(mitigations, models.Mitigation{
				Action:      "Reduce sprint commitments to sustainable levels",
				Priority:    "high",
				Description: "Prevent quality shortcuts from velocity pressure",
			})
		case "declining_coverage":
			mitigations = append(mitigations, models.Mitigation{
				Action:      "Enforce test coverage requirements in CI",
				Priority:    "medium",
				Description: "Make testing part of definition of done",
			})
		}
	}

	return mitigations
}
