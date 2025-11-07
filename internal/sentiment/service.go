package sentiment

import (
	"fmt"
	"math"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

type Service struct {
	contextStore *db.ContextStore
	teamStore    *db.TeamStore
}

func NewService(contextStore *db.ContextStore, teamStore *db.TeamStore) *Service {
	return &Service{
		contextStore: contextStore,
		teamStore:    teamStore,
	}
}

// CreatePulseSurvey creates a weekly pulse survey
func (s *Service) CreatePulseSurvey(createdBy string) (*models.SentimentSurvey, error) {
	survey := &models.SentimentSurvey{
		SurveyType:     "weekly_pulse",
		Title:          fmt.Sprintf("Weekly Pulse - Week %s", time.Now().UTC().Format("2006-W02")),
		Description:    "Quick check-in on how you're feeling this week",
		Questions:      GetDefaultPulseQuestions(),
		TargetAudience: "all",
		Anonymous:      true,
		Active:         true,
		CreatedBy:      createdBy,
		ExpiresAt:      expiresIn(7 * 24 * time.Hour),
	}

	err := s.contextStore.CreateSentimentSurvey(survey)
	if err != nil {
		return nil, err
	}

	return survey, nil
}

// SubmitResponse submits a survey response (anonymous or identified)
func (s *Service) SubmitResponse(surveyID string, respondentID *string, responses map[string]interface{}, moodRating *int, textFeedback string) error {
	response := &models.SurveyResponse{
		SurveyID:     surveyID,
		RespondentID: respondentID,
		Responses:    responses,
		MoodRating:   moodRating,
		TextFeedback: textFeedback,
	}

	return s.contextStore.CreateSurveyResponse(response)
}

// AnalyzeSurveyResults aggregates survey responses
func (s *Service) AnalyzeSurveyResults(surveyID string) (map[string]interface{}, error) {
	responses, err := s.contextStore.GetSurveyResponses(surveyID)
	if err != nil {
		return nil, err
	}

	if len(responses) == 0 {
		return map[string]interface{}{
			"total_responses": 0,
			"analysis":        "No responses yet",
		}, nil
	}

	// Calculate average mood rating
	totalMood := 0.0
	moodCount := 0
	for _, resp := range responses {
		if resp.MoodRating != nil {
			totalMood += float64(*resp.MoodRating)
			moodCount++
		}
	}

	avgMood := 0.0
	if moodCount > 0 {
		avgMood = totalMood / float64(moodCount)
	}

	// Collect text feedback
	var feedback []string
	for _, resp := range responses {
		if resp.TextFeedback != "" {
			feedback = append(feedback, resp.TextFeedback)
		}
	}

	return map[string]interface{}{
		"total_responses": len(responses),
		"average_mood":    avgMood,
		"mood_count":      moodCount,
		"text_feedback":   feedback,
	}, nil
}

// ComputeSentimentScore calculates team sentiment for a time period
func (s *Service) ComputeSentimentScore(teamID, timePeriod string) (*models.SentimentAnalysis, error) {
	// Get all survey responses for this team in this period
	// For simplicity, we'll compute based on active surveys
	surveys, _, err := s.contextStore.ListActiveSurveys(100, 0)
	if err != nil {
		return nil, err
	}

	totalMood := 0.0
	responseCount := 0

	for _, survey := range surveys {
		// Check if survey targets this team
		if survey.TargetAudience != "all" && survey.TargetAudience != fmt.Sprintf("team:%s", teamID) {
			continue
		}

		responses, err := s.contextStore.GetSurveyResponses(survey.ID)
		if err != nil {
			continue
		}

		for _, resp := range responses {
			if resp.MoodRating != nil {
				totalMood += float64(*resp.MoodRating)
				responseCount++
			}
		}
	}

	if responseCount == 0 {
		return nil, fmt.Errorf("no survey responses found for team %s in period %s", teamID, timePeriod)
	}

	avgMood := totalMood / float64(responseCount)

	// Convert 1-5 scale to -1 to 1
	sentimentScore := (avgMood - 3.0) / 2.0

	// Calculate confidence based on response rate
	// For now, use a simple heuristic
	confidence := math.Min(float64(responseCount)/10.0, 1.0)

	analysis := &models.SentimentAnalysis{
		EntityType:     "team",
		EntityID:       teamID,
		TimePeriod:     timePeriod,
		Source:         "survey",
		SentimentScore: sentimentScore,
		Confidence:     confidence,
		SampleSize:     responseCount,
		Details: map[string]interface{}{
			"average_mood":   avgMood,
			"response_count": responseCount,
		},
	}

	err = s.contextStore.CreateSentimentAnalysis(analysis)
	if err != nil {
		return nil, err
	}

	return analysis, nil
}

// DetectSentimentTrends analyzes changes over time
func (s *Service) DetectSentimentTrends(entityType, entityID string, periods []string) (string, error) {
	if len(periods) < 2 {
		return "stable", nil
	}

	var scores []float64
	for _, period := range periods {
		analyses, err := s.contextStore.GetSentimentAnalysis(entityType, entityID, period)
		if err != nil {
			continue
		}

		for _, analysis := range analyses {
			if analysis.Source == "survey" {
				scores = append(scores, analysis.SentimentScore)
				break
			}
		}
	}

	if len(scores) < 2 {
		return "stable", nil
	}

	// Simple trend detection: compare first half to second half
	mid := len(scores) / 2
	firstHalf := average(scores[:mid])
	secondHalf := average(scores[mid:])

	diff := secondHalf - firstHalf

	if diff > 0.1 {
		return "improving", nil
	} else if diff < -0.1 {
		return "declining", nil
	}

	return "stable", nil
}

// GenerateSentimentReport creates a manager dashboard report
func (s *Service) GenerateSentimentReport(entityType, entityID, timePeriod string) (*models.SentimentSummary, error) {
	analyses, err := s.contextStore.GetSentimentAnalysis(entityType, entityID, timePeriod)
	if err != nil {
		return nil, err
	}

	if len(analyses) == 0 {
		return nil, fmt.Errorf("no sentiment data found for %s %s in period %s", entityType, entityID, timePeriod)
	}

	// Aggregate scores by source
	sourceBreakdown := make(map[string]float64)
	totalScore := 0.0
	totalConfidence := 0.0

	for _, analysis := range analyses {
		sourceBreakdown[analysis.Source] = analysis.SentimentScore
		totalScore += analysis.SentimentScore * analysis.Confidence
		totalConfidence += analysis.Confidence
	}

	overallScore := 0.0
	if totalConfidence > 0 {
		overallScore = totalScore / totalConfidence
	}

	// Detect trend (simplified - would need historical periods)
	trend := "stable"

	summary := &models.SentimentSummary{
		EntityType:      entityType,
		EntityID:        entityID,
		TimePeriod:      timePeriod,
		OverallScore:    overallScore,
		Confidence:      totalConfidence / float64(len(analyses)),
		SourceBreakdown: sourceBreakdown,
		Trend:           trend,
		ComputedAt:      time.Now().UTC(),
	}

	return summary, nil
}

// Helper functions

func expiresIn(duration time.Duration) *time.Time {
	t := time.Now().UTC().Add(duration)
	return &t
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
