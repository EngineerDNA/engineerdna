package ai

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/engineerdna/engineerdna/internal/anonymization"
	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Run migrations
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			source TEXT NOT NULL,
			source_id TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			actor TEXT NOT NULL,
			engineer_id TEXT,
			data TEXT NOT NULL,
			anonymized BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS weekly_briefings (
			id TEXT PRIMARY KEY,
			team_id TEXT NOT NULL,
			week_start DATETIME NOT NULL,
			tldr TEXT NOT NULL,
			key_metrics TEXT NOT NULL,
			needs_attention TEXT NOT NULL,
			insights TEXT NOT NULL,
			trending_up TEXT NOT NULL,
			trending_down TEXT NOT NULL,
			talking_points TEXT NOT NULL,
			generated_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(team_id, week_start)
		)`,
		`CREATE TABLE IF NOT EXISTS teams (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			parent_team_id TEXT,
			manager_id TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS team_membership (
			id TEXT PRIMARY KEY,
			team_id TEXT NOT NULL,
			member_id TEXT NOT NULL,
			role TEXT,
			joined_at DATETIME NOT NULL,
			left_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS engineers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT,
			manager TEXT,
			identifiers TEXT NOT NULL,
			active BOOLEAN NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS anonymization_mappings (
			id TEXT PRIMARY KEY,
			real_name TEXT NOT NULL,
			anonymized_id TEXT UNIQUE NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, migration := range migrations {
		if _, err := database.Exec(migration); err != nil {
			t.Fatalf("failed to run migration: %v", err)
		}
	}

	cleanup := func() {
		database.Close()
	}

	return database, cleanup
}

func TestBriefingService_CalculateTeamMetrics(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	briefingsStore := db.NewBriefingsStore(database)
	eventsStore := db.NewEventStore(database)
	teamStore := db.NewTeamStore(database)
	anonStore := db.NewAnonymizationStore(database)
	anonService := anonymization.NewService(anonStore)

	service := NewBriefingService(
		database,
		briefingsStore,
		eventsStore,
		teamStore,
		anonService,
		nil, // plugin loader not needed for this test
	)

	weekStart := time.Date(2025, 11, 4, 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	// Create test events
	events := []*models.Event{
		{
			Type:      "pull_request",
			Source:    "github",
			SourceID:  "pr-1",
			Timestamp: weekStart.Add(1 * time.Hour),
			Actor:     "user@example.com",
			Data: map[string]interface{}{
				"action":     "merged",
				"created_at": weekStart.Format(time.RFC3339),
				"merged_at":  weekStart.Add(24 * time.Hour).Format(time.RFC3339),
			},
		},
		{
			Type:      "pull_request",
			Source:    "github",
			SourceID:  "pr-2",
			Timestamp: weekStart.Add(2 * time.Hour),
			Actor:     "user2@example.com",
			Data: map[string]interface{}{
				"action": "opened",
			},
		},
		{
			Type:      "issue",
			Source:    "github",
			SourceID:  "issue-1",
			Timestamp: weekStart.Add(3 * time.Hour),
			Actor:     "user@example.com",
			Data:      map[string]interface{}{},
		},
	}

	metrics := service.calculateTeamMetrics(events, weekStart, weekEnd)

	if prCount := metrics["pr_count"].(int); prCount != 2 {
		t.Errorf("expected pr_count = 2, got %d", prCount)
	}

	if prMerged := metrics["pr_merged_count"].(int); prMerged != 1 {
		t.Errorf("expected pr_merged_count = 1, got %d", prMerged)
	}

	if issueCount := metrics["issue_count"].(int); issueCount != 1 {
		t.Errorf("expected issue_count = 1, got %d", issueCount)
	}

	if uniqueActors := metrics["unique_actors"].(int); uniqueActors != 2 {
		t.Errorf("expected unique_actors = 2, got %d", uniqueActors)
	}

	// Cycle time should be 1 day for PR-1
	if cycleTime := metrics["avg_cycle_time_days"].(float64); cycleTime != 1.0 {
		t.Errorf("expected avg_cycle_time_days = 1.0, got %f", cycleTime)
	}
}

func TestBriefingService_ParseBriefingResponse(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	briefingsStore := db.NewBriefingsStore(database)
	eventsStore := db.NewEventStore(database)
	teamStore := db.NewTeamStore(database)
	anonStore := db.NewAnonymizationStore(database)
	anonService := anonymization.NewService(anonStore)

	service := NewBriefingService(
		database,
		briefingsStore,
		eventsStore,
		teamStore,
		anonService,
		nil,
	)

	response := &models.BriefingResponse{
		TLDR: "Team is performing well",
		NeedsAttention: []models.BriefingAttentionItem{
			{
				Person:         "User_1",
				Severity:       "warning",
				Issue:          "PRs waiting on review",
				Evidence:       []string{"3 PRs in queue", "Average wait: 4 days"},
				SuggestedTopic: "Need help prioritizing?",
			},
		},
		WhyMetricsChanged: []models.MetricChange{
			{
				Metric:      "PRs merged",
				Direction:   "up",
				Explanation: "Increased velocity",
			},
		},
		Insights: []models.BriefingInsight{
			{
				Observation:    "Review times increasing",
				Context:        "Up 20% from last week",
				Recommendation: "Set review SLAs",
			},
		},
		TrendingUp:    []string{"Code quality"},
		TrendingDown:  []string{},
		TalkingPoints: "Great week team!",
	}

	teamID := "team-1"
	weekStart := time.Date(2025, 11, 4, 0, 0, 0, 0, time.UTC)

	briefing, err := service.parseBriefingResponse(teamID, weekStart, response)
	if err != nil {
		t.Fatalf("parseBriefingResponse failed: %v", err)
	}

	if briefing.TeamID != teamID {
		t.Errorf("expected team_id = %s, got %s", teamID, briefing.TeamID)
	}

	if briefing.TLDR != "Team is performing well" {
		t.Errorf("expected TLDR = 'Team is performing well', got '%s'", briefing.TLDR)
	}

	if len(briefing.NeedsAttention) != 1 {
		t.Errorf("expected 1 attention item, got %d", len(briefing.NeedsAttention))
	}

	if len(briefing.Insights) != 1 {
		t.Errorf("expected 1 insight, got %d", len(briefing.Insights))
	}

	if briefing.TalkingPoints != "Great week team!" {
		t.Errorf("expected talking_points = 'Great week team!', got '%s'", briefing.TalkingPoints)
	}
}

func TestBriefingService_GetWeeklyBriefing_Cache(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	briefingsStore := db.NewBriefingsStore(database)
	eventsStore := db.NewEventStore(database)
	teamStore := db.NewTeamStore(database)
	anonStore := db.NewAnonymizationStore(database)
	anonService := anonymization.NewService(anonStore)

	service := NewBriefingService(
		database,
		briefingsStore,
		eventsStore,
		teamStore,
		anonService,
		nil,
	)

	teamID := "team-1"
	weekStart := time.Date(2025, 11, 4, 0, 0, 0, 0, time.UTC)

	// Create a cached briefing
	cachedBriefing := &models.WeeklyBriefing{
		TeamID:        teamID,
		WeekStart:     weekStart,
		TLDR:          "Cached briefing",
		GeneratedAt:   time.Now().UTC(),
		TalkingPoints: "From cache",
	}

	if err := briefingsStore.CreateBriefing(cachedBriefing); err != nil {
		t.Fatalf("failed to create cached briefing: %v", err)
	}

	// Get briefing - should return cached version
	ctx := context.Background()
	briefing, err := service.GetWeeklyBriefing(ctx, teamID, weekStart)

	// Should return cached briefing since it's recent
	if err != nil {
		t.Fatalf("GetWeeklyBriefing failed: %v", err)
	}

	if briefing.TLDR != "Cached briefing" {
		t.Errorf("expected cached TLDR, got '%s'", briefing.TLDR)
	}
}
