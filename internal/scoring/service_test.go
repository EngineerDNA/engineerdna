package scoring

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create events table
	_, err = db.Exec(`
		CREATE TABLE events (
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
		)
	`)
	if err != nil {
		t.Fatalf("failed to create events table: %v", err)
	}

	return db
}

func insertEvent(t *testing.T, db *sql.DB, event *models.Event) {
	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		t.Fatalf("failed to marshal event data: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data, anonymized, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.Type, event.Source, event.SourceID, event.Timestamp, event.Actor, event.EngineerID, string(dataJSON), event.Anonymized, event.CreatedAt, event.UpdatedAt)
	if err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}
}

func TestCalculateCycleTime(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewScoringService(db)
	engineerID := "eng-123"
	weekStart := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	// Insert PR events with different cycle times
	prs := []struct {
		createdAt    string
		mergedAt     string
		expectedDays float64
	}{
		{"2025-11-01T10:00:00Z", "2025-11-03T10:00:00Z", 2.0}, // 2 days
		{"2025-11-02T10:00:00Z", "2025-11-06T10:00:00Z", 4.0}, // 4 days
	}

	for i, pr := range prs {
		mergedAt, _ := time.Parse(time.RFC3339, pr.mergedAt)
		insertEvent(t, db, &models.Event{
			ID:         string(rune('a' + i)),
			Type:       "pull_request",
			Source:     "github",
			SourceID:   "PR-" + string(rune('1'+i)),
			Timestamp:  mergedAt,
			Actor:      "user@example.com",
			EngineerID: engineerID,
			Data: map[string]interface{}{
				"action":     "merged",
				"created_at": pr.createdAt,
				"merged_at":  pr.mergedAt,
				"repo":       "org/repo",
			},
			CreatedAt: mergedAt,
			UpdatedAt: mergedAt,
		})
	}

	// Calculate cycle time
	cycleTime, err := svc.calculateCycleTime(engineerID, weekStart, weekEnd)
	if err != nil {
		t.Fatalf("calculateCycleTime failed: %v", err)
	}

	// Expected average: (2 + 4) / 2 = 3 days
	expected := 3.0
	if cycleTime != expected {
		t.Errorf("expected cycle time %v, got %v", expected, cycleTime)
	}
}

func TestCalculateBugRate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewScoringService(db)
	engineerID := "eng-123"
	weekStart := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	// Insert 3 merged PRs
	for i := 0; i < 3; i++ {
		timestamp := weekStart.Add(time.Duration(i) * 24 * time.Hour)
		insertEvent(t, db, &models.Event{
			ID:         "pr-" + string(rune('a'+i)),
			Type:       "pull_request",
			Source:     "github",
			SourceID:   "PR-" + string(rune('1'+i)),
			Timestamp:  timestamp,
			Actor:      "user@example.com",
			EngineerID: engineerID,
			Data: map[string]interface{}{
				"action": "merged",
				"repo":   "org/repo",
			},
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		})
	}

	// Insert 2 bug issues
	for i := 0; i < 2; i++ {
		timestamp := weekStart.Add(time.Duration(i) * 24 * time.Hour)
		insertEvent(t, db, &models.Event{
			ID:         "issue-" + string(rune('a'+i)),
			Type:       "issue",
			Source:     "github",
			SourceID:   "ISSUE-" + string(rune('1'+i)),
			Timestamp:  timestamp,
			Actor:      "user@example.com",
			EngineerID: engineerID,
			Data: map[string]interface{}{
				"action": "opened",
				"labels": []string{"bug", "priority-high"},
				"repo":   "org/repo",
			},
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		})
	}

	// Calculate bug rate
	bugRate, err := svc.calculateBugRate(engineerID, weekStart, weekEnd)
	if err != nil {
		t.Fatalf("calculateBugRate failed: %v", err)
	}

	// Expected: 2 bugs / 3 PRs = 0.666...
	expected := 2.0 / 3.0
	if bugRate != expected {
		t.Errorf("expected bug rate %v, got %v", expected, bugRate)
	}
}

func TestCalculateServicesTouched(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewScoringService(db)
	engineerID := "eng-123"
	weekStart := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)

	// Insert events across different repos in the last 30 days
	repos := []string{"org/repo1", "org/repo2", "org/repo3", "org/repo1"} // repo1 appears twice
	for i, repo := range repos {
		timestamp := weekStart.AddDate(0, 0, -i*7) // Spread across last 4 weeks
		insertEvent(t, db, &models.Event{
			ID:         "event-" + string(rune('a'+i)),
			Type:       "pull_request",
			Source:     "github",
			SourceID:   "PR-" + string(rune('1'+i)),
			Timestamp:  timestamp,
			Actor:      "user@example.com",
			EngineerID: engineerID,
			Data: map[string]interface{}{
				"action": "merged",
				"repo":   repo,
			},
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		})
	}

	// Calculate services touched
	services, err := svc.calculateServicesTouched(engineerID, weekStart)
	if err != nil {
		t.Fatalf("calculateServicesTouched failed: %v", err)
	}

	// Expected: 3 distinct repos
	expected := 3.0
	if services != expected {
		t.Errorf("expected %v services, got %v", expected, services)
	}
}

func TestCalculateRawMetrics(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewScoringService(db)
	engineerID := "eng-123"
	weekStart := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	// Insert a merged PR with cycle time
	createdAt := "2025-11-01T10:00:00Z"
	mergedAt := "2025-11-03T10:00:00Z"
	mergedAtTime, _ := time.Parse(time.RFC3339, mergedAt)

	insertEvent(t, db, &models.Event{
		ID:         "pr-1",
		Type:       "pull_request",
		Source:     "github",
		SourceID:   "PR-1",
		Timestamp:  mergedAtTime,
		Actor:      "user@example.com",
		EngineerID: engineerID,
		Data: map[string]interface{}{
			"action":     "merged",
			"created_at": createdAt,
			"merged_at":  mergedAt,
			"repo":       "org/repo1",
		},
		CreatedAt: mergedAtTime,
		UpdatedAt: mergedAtTime,
	})

	// Insert a bug issue
	bugTime := weekStart.Add(12 * time.Hour)
	insertEvent(t, db, &models.Event{
		ID:         "issue-1",
		Type:       "issue",
		Source:     "github",
		SourceID:   "ISSUE-1",
		Timestamp:  bugTime,
		Actor:      "user@example.com",
		EngineerID: engineerID,
		Data: map[string]interface{}{
			"action": "opened",
			"labels": []string{"bug"},
			"repo":   "org/repo1",
		},
		CreatedAt: bugTime,
		UpdatedAt: bugTime,
	})

	// Calculate raw metrics
	metrics, err := svc.calculateRawMetrics(engineerID, weekStart, weekEnd)
	if err != nil {
		t.Fatalf("calculateRawMetrics failed: %v", err)
	}

	// Verify metrics
	if metrics.ThroughputPRsPerWeek != 1.0 {
		t.Errorf("expected 1 PR, got %v", metrics.ThroughputPRsPerWeek)
	}
	if metrics.SpeedCycleTimeDays != 2.0 {
		t.Errorf("expected 2 day cycle time, got %v", metrics.SpeedCycleTimeDays)
	}
	if metrics.QualityBugRate != 1.0 {
		t.Errorf("expected bug rate 1.0, got %v", metrics.QualityBugRate)
	}
	if metrics.ImpactServicesTouched != 1.0 {
		t.Errorf("expected 1 service, got %v", metrics.ImpactServicesTouched)
	}
}
