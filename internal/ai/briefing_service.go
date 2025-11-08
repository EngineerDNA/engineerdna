package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/anonymization"
	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/internal/plugin"
)

// BriefingService handles AI-generated weekly briefings
type BriefingService struct {
	database        *sql.DB
	briefingsStore  *db.BriefingsStore
	eventsStore     *db.EventStore
	teamStore       *db.TeamStore
	anonService     *anonymization.Service
	pluginLoader    *plugin.Loader
	cacheTimeoutMin int // Cache timeout in minutes
}

// NewBriefingService creates a new briefing service
func NewBriefingService(
	database *sql.DB,
	briefingsStore *db.BriefingsStore,
	eventsStore *db.EventStore,
	teamStore *db.TeamStore,
	anonService *anonymization.Service,
	pluginLoader *plugin.Loader,
) *BriefingService {
	return &BriefingService{
		database:        database,
		briefingsStore:  briefingsStore,
		eventsStore:     eventsStore,
		teamStore:       teamStore,
		anonService:     anonService,
		pluginLoader:    pluginLoader,
		cacheTimeoutMin: 60, // 1 hour cache by default
	}
}

// GenerateWeeklyBriefing generates a new AI briefing for a team for the given week
func (s *BriefingService) GenerateWeeklyBriefing(ctx context.Context, teamID string, weekStart time.Time) (*models.WeeklyBriefing, error) {
	// Normalize week start to beginning of day UTC
	weekStart = weekStart.UTC().Truncate(24 * time.Hour)
	weekEnd := weekStart.AddDate(0, 0, 7)

	// Get team with members
	teamWithMembers, err := s.teamStore.GetTeamWithMembers(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	if teamWithMembers == nil {
		return nil, fmt.Errorf("team not found: %s", teamID)
	}

	team := &teamWithMembers.Team
	members := teamWithMembers.Members

	// Collect member IDs for event query
	memberIDs := make([]string, len(members))
	for i, member := range members {
		memberIDs[i] = member.ID
	}

	// Fetch events for the week
	events, err := s.fetchTeamEvents(memberIDs, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	// Calculate team metrics
	currentMetrics := s.calculateTeamMetrics(events, weekStart, weekEnd)

	// Get previous week metrics for comparison
	prevWeekStart := weekStart.AddDate(0, 0, -7)
	prevWeekEnd := prevWeekStart.AddDate(0, 0, 7)
	prevEvents, err := s.fetchTeamEvents(memberIDs, prevWeekStart, prevWeekEnd)
	if err != nil {
		// Log but continue if previous week events fail
		prevEvents = []*models.Event{}
	}
	prevMetrics := s.calculateTeamMetrics(prevEvents, prevWeekStart, prevWeekEnd)

	// Anonymize events for AI processing
	anonEvents, err := s.anonService.CloneAndAnonymizeEvents(
		events,
		[]string{"actor", "author", "assignee", "email", "name"},
		models.StrategySequential,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to anonymize events: %w", err)
	}

	// Build AI prompt
	prompt := s.buildPrompt(team, members, anonEvents, currentMetrics, prevMetrics, weekStart, weekEnd)

	// Call AI plugin
	response, err := s.callAIPlugin(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI plugin: %w", err)
	}

	// Parse response and build briefing
	briefing, err := s.parseBriefingResponse(teamID, weekStart, response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Deanonymize engineer names in attention items
	for i := range briefing.NeedsAttention {
		// Try to deanonymize the engineer name
		realName, err := s.anonService.Deanonymize(briefing.NeedsAttention[i].EngineerName)
		if err == nil && realName != "" {
			briefing.NeedsAttention[i].EngineerName = realName
		}
	}

	// Store briefing
	if err := s.briefingsStore.CreateBriefing(briefing); err != nil {
		return nil, fmt.Errorf("failed to store briefing: %w", err)
	}

	return briefing, nil
}

// GetWeeklyBriefing retrieves a briefing from the database
// Briefings should be pre-generated via GenerateWeeklyBriefing or seed data
func (s *BriefingService) GetWeeklyBriefing(ctx context.Context, teamID string, weekStart time.Time) (*models.WeeklyBriefing, error) {
	weekStart = weekStart.UTC().Truncate(24 * time.Hour)

	// Get briefing from database
	briefing, err := s.briefingsStore.GetBriefing(teamID, weekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get briefing: %w", err)
	}

	// Return nil, nil if no briefing exists (not an error condition)
	// The caller can decide how to handle missing briefings
	return briefing, nil
}

// fetchTeamEvents fetches events for team members in the given time range
func (s *BriefingService) fetchTeamEvents(memberIDs []string, start, end time.Time) ([]*models.Event, error) {
	var allEvents []*models.Event

	for _, memberID := range memberIDs {
		events, err := s.eventsStore.GetEventsByEngineerAndTimeRange(memberID, start, end, db.MaxQueryLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to get events for engineer %s: %w", memberID, err)
		}
		allEvents = append(allEvents, events...)
	}

	return allEvents, nil
}

// calculateTeamMetrics calculates key metrics from events
func (s *BriefingService) calculateTeamMetrics(events []*models.Event, start, end time.Time) map[string]interface{} {
	metrics := make(map[string]interface{})

	prCount := 0
	prMergedCount := 0
	issueCount := 0
	commitCount := 0
	reviewCount := 0
	uniqueActors := make(map[string]bool)

	var totalCycleTime float64
	var cycleTimeCount int

	for _, event := range events {
		uniqueActors[event.Actor] = true

		switch event.Type {
		case "pull_request":
			prCount++
			if action, ok := event.Data["action"].(string); ok && action == "merged" {
				prMergedCount++

				// Calculate cycle time if we have created_at and merged_at
				if createdAtStr, ok := event.Data["created_at"].(string); ok {
					if mergedAtStr, ok := event.Data["merged_at"].(string); ok {
						createdAt, err1 := time.Parse(time.RFC3339, createdAtStr)
						mergedAt, err2 := time.Parse(time.RFC3339, mergedAtStr)
						if err1 == nil && err2 == nil {
							cycleTimeDays := mergedAt.Sub(createdAt).Hours() / 24.0
							totalCycleTime += cycleTimeDays
							cycleTimeCount++
						}
					}
				}
			}
		case "issue":
			issueCount++
		case "commit":
			commitCount++
		case "pull_request_review":
			reviewCount++
		}
	}

	metrics["pr_count"] = prCount
	metrics["pr_merged_count"] = prMergedCount
	metrics["issue_count"] = issueCount
	metrics["commit_count"] = commitCount
	metrics["review_count"] = reviewCount
	metrics["unique_actors"] = len(uniqueActors)

	if cycleTimeCount > 0 {
		metrics["avg_cycle_time_days"] = totalCycleTime / float64(cycleTimeCount)
	} else {
		metrics["avg_cycle_time_days"] = 0.0
	}

	return metrics
}

// buildPrompt constructs the AI prompt from team data and events
func (s *BriefingService) buildPrompt(
	team *models.Team,
	members []models.Engineer,
	events []*models.Event,
	currentMetrics map[string]interface{},
	prevMetrics map[string]interface{},
	weekStart time.Time,
	weekEnd time.Time,
) string {
	// Build context section
	context := fmt.Sprintf(`Context:
- Team: %s
- Team size: %d engineers
- Time period: %s to %s (7 days)
- Previous week for comparison: %s to %s

Current Week Metrics:
- PRs: %v (merged: %v)
- Issues: %v
- Commits: %v
- Reviews: %v
- Active engineers: %v
- Avg cycle time: %.1f days

Previous Week Metrics:
- PRs: %v (merged: %v)
- Issues: %v
- Avg cycle time: %.1f days
`,
		team.Name,
		len(members),
		weekStart.Format("2006-01-02"),
		weekEnd.Format("2006-01-02"),
		weekStart.AddDate(0, 0, -7).Format("2006-01-02"),
		weekStart.Format("2006-01-02"),
		currentMetrics["pr_count"],
		currentMetrics["pr_merged_count"],
		currentMetrics["issue_count"],
		currentMetrics["commit_count"],
		currentMetrics["review_count"],
		currentMetrics["unique_actors"],
		currentMetrics["avg_cycle_time_days"],
		prevMetrics["pr_count"],
		prevMetrics["pr_merged_count"],
		prevMetrics["issue_count"],
		prevMetrics["avg_cycle_time_days"],
	)

	// Build events summary (limit to 100 most recent events for token efficiency)
	eventsData := events
	if len(eventsData) > 100 {
		eventsData = eventsData[:100]
	}

	eventsJSON, _ := json.Marshal(eventsData)

	// Build full prompt
	prompt := fmt.Sprintf(`You are an AI engineering coach analyzing team metrics.

%s

Events (anonymized, JSON array):
%s

Tasks:
1. Summarize: What's the TL;DR? (2 sentences max)
2. Identify: Who needs attention and why? (Top 2-3 people)
3. Explain: Why did metrics change? (Root causes)
4. Suggest: What should the manager do? (Concrete 1:1 topics)
5. Trends: What's improving/declining? (2-3 patterns)
6. Talking points: What to say in standup? (3 sentences)

Rules:
- Be specific (cite actual PRs, numbers, dates)
- Be actionable (suggest concrete discussion topics)
- Be empathetic (frame issues as "needs support" not "poor performance")
- Link every claim to data (provide evidence)
- Don't mention score numbers in briefing (save for scorecard)

Respond in JSON:
{
  "tldr": "...",
  "needs_attention": [{"person": "User_1", "severity": "warning|info|critical", "issue": "...", "evidence": ["...", "..."], "suggested_topic": "..."}],
  "why_metrics_changed": [{"metric": "PRs merged", "direction": "up", "explanation": "..."}],
  "insights": [{"observation": "...", "context": "...", "recommendation": "..."}],
  "trending_up": ["..."],
  "trending_down": ["..."],
  "talking_points": "..."
}`,
		context,
		string(eventsJSON),
	)

	return prompt
}

// callAIPlugin invokes the ai-insights plugin to generate the briefing
func (s *BriefingService) callAIPlugin(ctx context.Context, prompt string) (*models.BriefingResponse, error) {
	// V1 Implementation Note: This returns sample data for demonstration purposes.
	// Real AI plugin integration will be added in a future release when plugin SDK
	// supports custom prompts or when we implement direct Claude API integration.
	//
	// Future implementation will:
	// 1. Get plugin path from loader
	// 2. Call plugin with custom method for briefing generation
	// 3. Parse JSON response
	// 4. Handle errors gracefully

	mockResponse := &models.BriefingResponse{
		TLDR: "Team is performing well with steady metrics. Review times need attention.",
		NeedsAttention: []models.BriefingAttentionItem{
			{
				Person:         "Maya Patel",
				Severity:       "warning",
				Issue:          "PRs waiting on review for >3 days",
				Evidence:       []string{"3 PRs in review queue", "Average review wait time: 4.2 days"},
				SuggestedTopic: "Do you need help prioritizing reviews?",
			},
		},
		WhyMetricsChanged: []models.MetricChange{
			{
				Metric:      "PRs merged",
				Direction:   "stable",
				Explanation: "Consistent output with 15 PRs merged this week",
			},
		},
		Insights: []models.BriefingInsight{
			{
				Observation:    "Review times are increasing",
				Context:        "Average review time up 20% from last week",
				Recommendation: "Consider setting review SLAs",
			},
		},
		TrendingUp:   []string{"Code quality remains high", "Team collaboration strong"},
		TrendingDown: []string{},
		TalkingPoints: "Team shipped 15 PRs this week maintaining steady velocity. " +
			"Main focus: reducing review queue to keep momentum.",
	}

	return mockResponse, nil
}

// parseBriefingResponse converts AI response to WeeklyBriefing model
func (s *BriefingService) parseBriefingResponse(
	teamID string,
	weekStart time.Time,
	response *models.BriefingResponse,
) (*models.WeeklyBriefing, error) {
	// Build key metrics from changes
	keyMetrics := []models.MetricSummary{}
	for _, change := range response.WhyMetricsChanged {
		keyMetrics = append(keyMetrics, models.MetricSummary{
			Name:      change.Metric,
			Value:     change.Explanation,
			Direction: change.Direction,
		})
	}

	// Convert attention items
	needsAttention := []models.AttentionItem{}
	for _, item := range response.NeedsAttention {
		needsAttention = append(needsAttention, models.AttentionItem{
			EngineerName:   item.Person,
			Severity:       item.Severity,
			Issue:          item.Issue,
			Evidence:       item.Evidence,
			SuggestedTopic: item.SuggestedTopic,
		})
	}

	// Convert insights
	insights := []models.AIInsight{}
	for _, item := range response.Insights {
		insights = append(insights, models.AIInsight{
			Observation:    item.Observation,
			Context:        item.Context,
			Recommendation: item.Recommendation,
		})
	}

	briefing := &models.WeeklyBriefing{
		TeamID:         teamID,
		WeekStart:      weekStart,
		TLDR:           response.TLDR,
		KeyMetrics:     keyMetrics,
		NeedsAttention: needsAttention,
		Insights:       insights,
		TrendingUp:     response.TrendingUp,
		TrendingDown:   response.TrendingDown,
		TalkingPoints:  response.TalkingPoints,
		GeneratedAt:    time.Now().UTC(),
	}

	return briefing, nil
}
