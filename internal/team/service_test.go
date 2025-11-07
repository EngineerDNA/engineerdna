package team

import (
	"database/sql"
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

	// Create minimal schema for testing
	schema := `
		CREATE TABLE teams (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			parent_team_id TEXT,
			manager_id TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE team_membership (
			id TEXT PRIMARY KEY,
			team_id TEXT NOT NULL,
			member_id TEXT NOT NULL,
			role TEXT,
			joined_at DATETIME NOT NULL,
			left_at DATETIME
		);

		CREATE TABLE engineers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT,
			manager TEXT,
			identifiers TEXT,
			active BOOLEAN DEFAULT TRUE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE performance_scores (
			id TEXT PRIMARY KEY,
			engineer_id TEXT NOT NULL,
			week_start DATE NOT NULL,
			total_score REAL,
			throughput_score REAL,
			quality_score REAL,
			speed_score REAL,
			collaboration_score REAL,
			impact_score REAL,
			raw_metrics TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(engineer_id, week_start)
		);

		CREATE TABLE team_performance_scores (
			id TEXT PRIMARY KEY,
			team_id TEXT NOT NULL,
			week_start DATE NOT NULL,
			total_score REAL,
			member_count INTEGER,
			throughput_score REAL,
			quality_score REAL,
			speed_score REAL,
			collaboration_score REAL,
			impact_score REAL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(team_id, week_start)
		);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create test schema: %v", err)
	}

	return db
}

func TestCalculateTeamScore_EmptyTeam(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a team with no members
	_, err := db.Exec(`INSERT INTO teams (id, name) VALUES ('team1', 'Empty Team')`)
	if err != nil {
		t.Fatalf("failed to create test team: %v", err)
	}

	weekStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	score, err := service.CalculateTeamScore("team1", weekStart)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if score != nil {
		t.Errorf("expected nil score for empty team, got %+v", score)
	}
}

func TestCalculateTeamScore_WithMembers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create team
	_, err := db.Exec(`INSERT INTO teams (id, name) VALUES ('team1', 'Backend Team')`)
	if err != nil {
		t.Fatalf("failed to create test team: %v", err)
	}

	// Create engineers
	_, err = db.Exec(`
		INSERT INTO engineers (id, name, identifiers) VALUES
		('eng1', 'Alice Smith', '{}'),
		('eng2', 'Bob Jones', '{}')
	`)
	if err != nil {
		t.Fatalf("failed to create test engineers: %v", err)
	}

	// Add team members
	now := time.Now().UTC()
	_, err = db.Exec(`
		INSERT INTO team_membership (id, team_id, member_id, joined_at) VALUES
		('mem1', 'team1', 'eng1', ?),
		('mem2', 'team1', 'eng2', ?)
	`, now, now)
	if err != nil {
		t.Fatalf("failed to create team membership: %v", err)
	}

	// Create performance scores
	weekStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err = db.Exec(`
		INSERT INTO performance_scores (id, engineer_id, week_start, total_score, throughput_score, quality_score, speed_score, collaboration_score, impact_score, raw_metrics) VALUES
		('score1', 'eng1', ?, 95.0, 100.0, 90.0, 95.0, 85.0, 105.0, '{}'),
		('score2', 'eng2', ?, 85.0, 90.0, 80.0, 85.0, 90.0, 95.0, '{}')
	`, weekStart, weekStart)
	if err != nil {
		t.Fatalf("failed to create performance scores: %v", err)
	}

	// Calculate team score
	score, err := service.CalculateTeamScore("team1", weekStart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if score == nil {
		t.Fatal("expected non-nil score, got nil")
	}

	// Verify averages: (95 + 85) / 2 = 90
	expectedTotal := 90.0
	if score.TotalScore != expectedTotal {
		t.Errorf("expected total score %f, got %f", expectedTotal, score.TotalScore)
	}

	// Verify member count
	if score.MemberCount != 2 {
		t.Errorf("expected member count 2, got %d", score.MemberCount)
	}

	// Verify component scores
	expectedThroughput := 95.0 // (100 + 90) / 2
	if score.ThroughputScore != expectedThroughput {
		t.Errorf("expected throughput score %f, got %f", expectedThroughput, score.ThroughputScore)
	}
}

func TestGetTeamHierarchy(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create team hierarchy: Engineering > Backend > API
	_, err := db.Exec(`
		INSERT INTO teams (id, name, parent_team_id) VALUES
		('eng', 'Engineering', NULL),
		('backend', 'Backend', 'eng'),
		('api', 'API', 'backend')
	`)
	if err != nil {
		t.Fatalf("failed to create teams: %v", err)
	}

	// Get hierarchy starting from root
	hierarchy, err := service.GetTeamHierarchy("eng")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hierarchy == nil {
		t.Fatal("expected non-nil hierarchy, got nil")
	}

	// Verify root team
	if hierarchy.Team.ID != "eng" {
		t.Errorf("expected root team ID 'eng', got '%s'", hierarchy.Team.ID)
	}

	// Verify children
	if len(hierarchy.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(hierarchy.Children))
	}

	if len(hierarchy.Children) > 0 {
		backend := hierarchy.Children[0]
		if backend.Team.ID != "backend" {
			t.Errorf("expected child team ID 'backend', got '%s'", backend.Team.ID)
		}

		// Verify grandchildren
		if len(backend.Children) != 1 {
			t.Errorf("expected 1 grandchild, got %d", len(backend.Children))
		}

		if len(backend.Children) > 0 {
			api := backend.Children[0]
			if api.Team.ID != "api" {
				t.Errorf("expected grandchild team ID 'api', got '%s'", api.Team.ID)
			}
		}
	}
}

func TestGetTeamHierarchy_CircularReference(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create circular reference: team1 -> team2 -> team1 (via direct updates to bypass FK)
	_, err := db.Exec(`
		INSERT INTO teams (id, name, parent_team_id) VALUES
		('team1', 'Team 1', 'team2'),
		('team2', 'Team 2', 'team1')
	`)
	if err != nil {
		t.Fatalf("failed to create teams: %v", err)
	}

	// Should detect circular reference
	_, err = service.GetTeamHierarchy("team1")
	if err == nil {
		t.Error("expected error for circular reference, got nil")
	}
}

func TestGetOrgWideScores(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create top-level teams
	_, err := db.Exec(`
		INSERT INTO teams (id, name, parent_team_id) VALUES
		('backend', 'Backend', NULL),
		('frontend', 'Frontend', NULL)
	`)
	if err != nil {
		t.Fatalf("failed to create teams: %v", err)
	}

	// Create team scores for current and previous week
	weekStart := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	prevWeekStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err = db.Exec(`
		INSERT INTO team_performance_scores (id, team_id, week_start, total_score, member_count) VALUES
		('score1', 'backend', ?, 90.0, 5),
		('score2', 'backend', ?, 95.0, 5),
		('score3', 'frontend', ?, 85.0, 3),
		('score4', 'frontend', ?, 88.0, 3)
	`, prevWeekStart, weekStart, prevWeekStart, weekStart)
	if err != nil {
		t.Fatalf("failed to create team scores: %v", err)
	}

	// Get org-wide scores for current week
	scorecards, err := service.GetOrgWideScores(weekStart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scorecards) != 2 {
		t.Fatalf("expected 2 scorecards, got %d", len(scorecards))
	}

	// Verify backend team
	var backendCard *models.TeamScorecard
	var frontendCard *models.TeamScorecard
	for _, card := range scorecards {
		if card.Team.ID == "backend" {
			backendCard = card
		} else if card.Team.ID == "frontend" {
			frontendCard = card
		}
	}

	if backendCard == nil {
		t.Fatal("backend scorecard not found")
	}
	if frontendCard == nil {
		t.Fatal("frontend scorecard not found")
	}

	// Verify backend scores and change
	if backendCard.Score != 95.0 {
		t.Errorf("expected backend score 95.0, got %f", backendCard.Score)
	}
	if backendCard.PrevWeekScore != 90.0 {
		t.Errorf("expected backend prev score 90.0, got %f", backendCard.PrevWeekScore)
	}
	if backendCard.ScoreChange != 5.0 {
		t.Errorf("expected backend score change 5.0, got %f", backendCard.ScoreChange)
	}

	// Verify frontend scores and change
	if frontendCard.Score != 88.0 {
		t.Errorf("expected frontend score 88.0, got %f", frontendCard.Score)
	}
	if frontendCard.PrevWeekScore != 85.0 {
		t.Errorf("expected frontend prev score 85.0, got %f", frontendCard.PrevWeekScore)
	}
	if frontendCard.ScoreChange != 3.0 {
		t.Errorf("expected frontend score change 3.0, got %f", frontendCard.ScoreChange)
	}
}
