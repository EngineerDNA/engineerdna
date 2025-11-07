package sentiment

import (
	"fmt"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

type Analyzer struct {
	contextStore *db.ContextStore
	eventStore   *db.EventStore
}

func NewAnalyzer(contextStore *db.ContextStore, eventStore *db.EventStore) *Analyzer {
	return &Analyzer{
		contextStore: contextStore,
		eventStore:   eventStore,
	}
}

// AnalyzeCodeReviewTone analyzes sentiment from code review comments
func (a *Analyzer) AnalyzeCodeReviewTone(engineerID, timePeriod string) (*models.SentimentAnalysis, error) {
	// Get code review events for this engineer
	startDate, endDate := parsePeriod(timePeriod)

	filters := map[string]interface{}{
		"engineer_id": engineerID,
		"start_date":  startDate,
		"end_date":    endDate,
	}

	events, err := a.eventStore.List(filters, 1000, 0)
	if err != nil {
		return nil, err
	}

	// Filter for review comments
	var reviewEvents []*models.Event
	for _, event := range events {
		if event.Type == "code_review" || event.Type == "pull_request_review" {
			reviewEvents = append(reviewEvents, event)
		}
	}

	if len(reviewEvents) == 0 {
		return nil, fmt.Errorf("no code review events found for engineer %s in period %s", engineerID, timePeriod)
	}

	// Analyze tone (simplified sentiment analysis)
	positiveCount := 0
	negativeCount := 0

	for _, event := range reviewEvents {
		// Data is already a map[string]interface{}
		body, ok := event.Data["body"].(string)
		if !ok {
			continue
		}

		// Simple keyword-based sentiment
		if containsPositiveKeywords(body) {
			positiveCount++
		}
		if containsNegativeKeywords(body) {
			negativeCount++
		}
	}

	// Calculate sentiment score
	total := positiveCount + negativeCount
	if total == 0 {
		return nil, fmt.Errorf("no sentiment data in code reviews")
	}

	sentimentScore := (float64(positiveCount) - float64(negativeCount)) / float64(total)
	confidence := float64(total) / float64(len(reviewEvents))

	analysis := &models.SentimentAnalysis{
		EntityType:     "engineer",
		EntityID:       engineerID,
		TimePeriod:     timePeriod,
		Source:         "code_review_tone",
		SentimentScore: sentimentScore,
		Confidence:     confidence,
		SampleSize:     len(reviewEvents),
		Details: map[string]interface{}{
			"positive_count": positiveCount,
			"negative_count": negativeCount,
			"total_reviews":  len(reviewEvents),
		},
	}

	err = a.contextStore.CreateSentimentAnalysis(analysis)
	return analysis, err
}

// AnalyzeCommitMessages analyzes sentiment from commit messages
func (a *Analyzer) AnalyzeCommitMessages(engineerID, timePeriod string) (*models.SentimentAnalysis, error) {
	startDate, endDate := parsePeriod(timePeriod)

	filters := map[string]interface{}{
		"engineer_id": engineerID,
		"start_date":  startDate,
		"end_date":    endDate,
	}

	events, err := a.eventStore.List(filters, 1000, 0)
	if err != nil {
		return nil, err
	}

	// Filter for commit events
	var commitEvents []*models.Event
	for _, event := range events {
		if event.Type == "commit" || event.Type == "push" {
			commitEvents = append(commitEvents, event)
		}
	}

	if len(commitEvents) == 0 {
		return nil, fmt.Errorf("no commit events found for engineer %s in period %s", engineerID, timePeriod)
	}

	// Analyze commit message tone
	positiveCount := 0
	negativeCount := 0

	for _, event := range commitEvents {
		// Data is already a map[string]interface{}
		message, ok := event.Data["message"].(string)
		if !ok {
			continue
		}

		// Detect frustration or negativity in commit messages
		if containsFrustrationKeywords(message) {
			negativeCount++
		} else {
			positiveCount++
		}
	}

	total := len(commitEvents)
	sentimentScore := (float64(positiveCount) - float64(negativeCount)) / float64(total)
	confidence := 0.5 // Lower confidence for commit message analysis

	analysis := &models.SentimentAnalysis{
		EntityType:     "engineer",
		EntityID:       engineerID,
		TimePeriod:     timePeriod,
		Source:         "commit_messages",
		SentimentScore: sentimentScore,
		Confidence:     confidence,
		SampleSize:     total,
		Details: map[string]interface{}{
			"positive_count": positiveCount,
			"negative_count": negativeCount,
			"total_commits":  total,
		},
	}

	err = a.contextStore.CreateSentimentAnalysis(analysis)
	return analysis, err
}

// DetectMoodFromSurveys aggregates mood from survey responses
// Phase 2: Will integrate with survey/pulse check system
func (a *Analyzer) DetectMoodFromSurveys(entityType, entityID, timePeriod string) (*models.SentimentAnalysis, error) {
	// Survey integration requires external survey system or custom pulse check feature
	// This is a Phase 2 feature pending survey system integration

	// Return error indicating feature not available
	return nil, fmt.Errorf("survey integration not available - requires pulse check system (Phase 2)")
}

// CorrelateWithPerformance links sentiment to performance metrics
func (a *Analyzer) CorrelateWithPerformance(engineerID, timePeriod string) (map[string]interface{}, error) {
	// Get sentiment analysis
	sentimentData, err := a.contextStore.GetSentimentAnalysis("engineer", engineerID, timePeriod)
	if err != nil {
		return nil, err
	}

	if len(sentimentData) == 0 {
		return map[string]interface{}{
			"correlation": "no_data",
		}, nil
	}

	// Calculate average sentiment
	totalSentiment := 0.0
	for _, s := range sentimentData {
		totalSentiment += s.SentimentScore
	}
	avgSentiment := totalSentiment / float64(len(sentimentData))

	// Would correlate with performance scores here
	// For now, return sentiment data
	return map[string]interface{}{
		"average_sentiment": avgSentiment,
		"sample_count":      len(sentimentData),
		"correlation":       "requires_performance_data",
	}, nil
}

// PredictAttritionRisk predicts attrition risk based on sentiment
func (a *Analyzer) PredictAttritionRisk(engineerID, timePeriod string) (string, error) {
	sentimentData, err := a.contextStore.GetSentimentAnalysis("engineer", engineerID, timePeriod)
	if err != nil {
		return "unknown", err
	}

	if len(sentimentData) == 0 {
		return "unknown", nil
	}

	// Calculate average sentiment
	totalSentiment := 0.0
	for _, s := range sentimentData {
		totalSentiment += s.SentimentScore
	}
	avgSentiment := totalSentiment / float64(len(sentimentData))

	// Simple risk assessment
	if avgSentiment < -0.5 {
		return "high", nil
	} else if avgSentiment < -0.2 {
		return "medium", nil
	} else if avgSentiment < 0 {
		return "low", nil
	}

	return "none", nil
}

// Helper functions

func parsePeriod(timePeriod string) (time.Time, time.Time) {
	// Parse periods like "2024-W45" or "Q4 2024"
	// For simplicity, return last 7 days
	now := time.Now().UTC()
	return now.AddDate(0, 0, -7), now
}

func containsPositiveKeywords(text string) bool {
	lower := strings.ToLower(text)
	positiveWords := []string{
		"great", "excellent", "good", "nice", "well done",
		"thanks", "appreciate", "love", "perfect", "awesome",
	}
	for _, word := range positiveWords {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

func containsNegativeKeywords(text string) bool {
	lower := strings.ToLower(text)
	negativeWords := []string{
		"bad", "wrong", "error", "issue", "problem",
		"broken", "fail", "incorrect", "missing", "bug",
	}
	for _, word := range negativeWords {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

func containsFrustrationKeywords(text string) bool {
	lower := strings.ToLower(text)
	frustrationWords := []string{
		"wtf", "damn", "fix", "ugh", "argh",
		"finally", "again", "still broken", "why",
	}
	for _, word := range frustrationWords {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}
