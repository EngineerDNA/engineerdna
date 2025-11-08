package scoring

import (
	"database/sql"
	"testing"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	_ "modernc.org/sqlite"
)

func TestDetectBurnoutRisks(t *testing.T) {
	// Create in-memory test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()

	// Create events table
	_, err = db.Exec(`
		CREATE TABLE events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			source TEXT NOT NULL,
			source_id TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			actor TEXT NOT NULL,
			engineer_id TEXT NOT NULL,
			data TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create events table: %v", err)
	}

	service := NewScoringService(db)

	tests := []struct {
		name              string
		setupEvents       func(t *testing.T, db *sql.DB, engineerID string, weekStart time.Time)
		engineerID        string
		weekStart         time.Time
		totalScore        float64
		currentMetrics    *models.RawMetrics
		expectRisk        bool
		expectedRiskLevel string
		expectedFlagCount int
	}{
		{
			name: "no burnout - normal hours",
			setupEvents: func(t *testing.T, db *sql.DB, engineerID string, weekStart time.Time) {
				// Add commits during business hours (9am-5pm)
				for i := 0; i < 5; i++ {
					timestamp := weekStart.Add(time.Duration(i*24) * time.Hour).Add(10 * time.Hour) // 10am each day
					_, err := db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
						timestamp.Format(time.RFC3339Nano), "push", "github", "commit-"+timestamp.Format(time.RFC3339Nano),
						timestamp, "user@example.com", engineerID, "{}")
					if err != nil {
						t.Fatalf("failed to insert event: %v", err)
					}
				}
			},
			engineerID: "eng-1",
			weekStart:  time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC),
			totalScore: 125.0,
			currentMetrics: &models.RawMetrics{
				QualityBugRate: 0.5,
			},
			expectRisk:        false,
			expectedRiskLevel: "",
			expectedFlagCount: 0,
		},
		{
			name: "late night work detected",
			setupEvents: func(t *testing.T, db *sql.DB, engineerID string, weekStart time.Time) {
				// Add 4 late-night commits (11pm)
				for i := 0; i < 4; i++ {
					timestamp := weekStart.Add(time.Duration(i*24) * time.Hour).Add(23 * time.Hour) // 11pm each day
					_, err := db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
						timestamp.Format(time.RFC3339Nano), "push", "github", "commit-"+timestamp.Format(time.RFC3339Nano),
						timestamp, "user@example.com", engineerID, "{}")
					if err != nil {
						t.Fatalf("failed to insert event: %v", err)
					}
				}
			},
			engineerID: "eng-2",
			weekStart:  time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC),
			totalScore: 130.0,
			currentMetrics: &models.RawMetrics{
				QualityBugRate: 0.5,
			},
			expectRisk:        true,
			expectedRiskLevel: "low",
			expectedFlagCount: 1,
		},
		{
			name: "weekend work detected",
			setupEvents: func(t *testing.T, db *sql.DB, engineerID string, weekStart time.Time) {
				// Add 3 weekend commits (Saturday and Sunday)
				// Saturday Nov 8, 2025
				satTimestamp := weekStart.Add(5 * 24 * time.Hour).Add(10 * time.Hour) // Saturday 10am
				_, err := db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
					satTimestamp.Format(time.RFC3339Nano), "push", "github", "commit-sat-1",
					satTimestamp, "user@example.com", engineerID, "{}")
				if err != nil {
					t.Fatalf("failed to insert event: %v", err)
				}

				// Sunday Nov 9, 2025
				sunTimestamp := weekStart.Add(6 * 24 * time.Hour).Add(14 * time.Hour) // Sunday 2pm
				_, err = db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
					sunTimestamp.Format(time.RFC3339Nano), "push", "github", "commit-sun-1",
					sunTimestamp, "user@example.com", engineerID, "{}")
				if err != nil {
					t.Fatalf("failed to insert event: %v", err)
				}
			},
			engineerID: "eng-3",
			weekStart:  time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC),
			totalScore: 135.0,
			currentMetrics: &models.RawMetrics{
				QualityBugRate: 0.5,
			},
			expectRisk:        true,
			expectedRiskLevel: "low",
			expectedFlagCount: 1,
		},
		{
			name: "multiple red flags - high risk",
			setupEvents: func(t *testing.T, db *sql.DB, engineerID string, weekStart time.Time) {
				// Add late-night commits
				for i := 0; i < 4; i++ {
					timestamp := weekStart.Add(time.Duration(i*24) * time.Hour).Add(23 * time.Hour)
					_, err := db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
						timestamp.Format(time.RFC3339Nano), "push", "github", "commit-late-"+timestamp.Format(time.RFC3339Nano),
						timestamp, "user@example.com", engineerID, "{}")
					if err != nil {
						t.Fatalf("failed to insert event: %v", err)
					}
				}

				// Add weekend commits
				satTimestamp := weekStart.Add(5 * 24 * time.Hour).Add(10 * time.Hour)
				_, err := db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
					satTimestamp.Format(time.RFC3339Nano), "push", "github", "commit-sat",
					satTimestamp, "user@example.com", engineerID, "{}")
				if err != nil {
					t.Fatalf("failed to insert event: %v", err)
				}

				sunTimestamp := weekStart.Add(6 * 24 * time.Hour).Add(14 * time.Hour)
				_, err = db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
					sunTimestamp.Format(time.RFC3339Nano), "push", "github", "commit-sun",
					sunTimestamp, "user@example.com", engineerID, "{}")
				if err != nil {
					t.Fatalf("failed to insert event: %v", err)
				}

				// Add previous week data for quality decline
				previousWeekStart := weekStart.AddDate(0, 0, -7)
				// Add some PRs to previous week
				for i := 0; i < 3; i++ {
					timestamp := previousWeekStart.Add(time.Duration(i*24) * time.Hour).Add(10 * time.Hour)
					_, err := db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
						"pr-prev-"+timestamp.Format(time.RFC3339Nano), "pull_request", "github", "pr-prev-"+timestamp.Format(time.RFC3339Nano),
						timestamp, "user@example.com", engineerID, `{"action": "merged"}`)
					if err != nil {
						t.Fatalf("failed to insert previous week PR: %v", err)
					}
				}
				// Add 1 bug to previous week (bug rate = 1/3 = 0.33)
				bugTimestamp := previousWeekStart.Add(2 * 24 * time.Hour).Add(10 * time.Hour)
				_, err = db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
					"bug-prev", "issue", "github", "bug-prev",
					bugTimestamp, "user@example.com", engineerID, `{"action": "opened", "labels": ["bug"]}`)
				if err != nil {
					t.Fatalf("failed to insert previous week bug: %v", err)
				}
			},
			engineerID: "eng-4",
			weekStart:  time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC),
			totalScore: 140.0,
			currentMetrics: &models.RawMetrics{
				QualityBugRate: 0.6, // 80% increase from previous 0.33
			},
			expectRisk:        true,
			expectedRiskLevel: "high",
			expectedFlagCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test events
			tt.setupEvents(t, db, tt.engineerID, tt.weekStart)

			// Run burnout detection
			risk, err := service.DetectBurnoutRisks(tt.engineerID, tt.weekStart, tt.totalScore, tt.currentMetrics)
			if err != nil {
				t.Fatalf("DetectBurnoutRisks failed: %v", err)
			}

			// Check expectations
			if tt.expectRisk {
				if risk == nil {
					t.Errorf("expected burnout risk to be detected, got nil")
					return
				}

				if risk.RiskLevel != tt.expectedRiskLevel {
					t.Errorf("expected risk level %s, got %s", tt.expectedRiskLevel, risk.RiskLevel)
				}

				if len(risk.RedFlags) != tt.expectedFlagCount {
					t.Errorf("expected %d red flags, got %d", tt.expectedFlagCount, len(risk.RedFlags))
				}

				if risk.Score != tt.totalScore {
					t.Errorf("expected score %.2f, got %.2f", tt.totalScore, risk.Score)
				}

				if risk.Recommendation == "" {
					t.Errorf("expected recommendation to be set")
				}
			} else {
				if risk != nil {
					t.Errorf("expected no burnout risk, got %+v", risk)
				}
			}

			// Clean up for next test
			if _, err := db.Exec("DELETE FROM events WHERE engineer_id = ?", tt.engineerID); err != nil {
				t.Logf("Warning: Failed to cleanup test data: %v", err)
			}
		})
	}
}

func TestCountLateNightCommits(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			source TEXT NOT NULL,
			source_id TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			actor TEXT NOT NULL,
			engineer_id TEXT NOT NULL,
			data TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create events table: %v", err)
	}

	service := NewScoringService(db)
	engineerID := "test-eng"
	weekStart := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	// Insert test commits at various hours
	testCommits := []struct {
		hour        int
		isLateNight bool
	}{
		{hour: 9, isLateNight: false},  // 9am - normal
		{hour: 23, isLateNight: true},  // 11pm - late night
		{hour: 1, isLateNight: true},   // 1am - late night
		{hour: 14, isLateNight: false}, // 2pm - normal
		{hour: 22, isLateNight: true},  // 10pm - late night
		{hour: 5, isLateNight: false},  // 5am - normal (boundary)
	}

	for i, tc := range testCommits {
		timestamp := weekStart.Add(time.Duration(i*24) * time.Hour).Add(time.Duration(tc.hour) * time.Hour)
		_, err := db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			timestamp.Format(time.RFC3339Nano), "push", "github", "commit-"+timestamp.Format(time.RFC3339Nano),
			timestamp, "user@example.com", engineerID, "{}")
		if err != nil {
			t.Fatalf("failed to insert event: %v", err)
		}
	}

	// Count late night commits
	count, err := service.countLateNightCommits(engineerID, weekStart, weekEnd)
	if err != nil {
		t.Fatalf("countLateNightCommits failed: %v", err)
	}

	expectedCount := 3 // 23, 1, 22
	if count != expectedCount {
		t.Errorf("expected %d late night commits, got %d", expectedCount, count)
	}
}

func TestCountWeekendCommits(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			source TEXT NOT NULL,
			source_id TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			actor TEXT NOT NULL,
			engineer_id TEXT NOT NULL,
			data TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create events table: %v", err)
	}

	service := NewScoringService(db)
	engineerID := "test-eng"
	// Start on Monday Nov 3, 2025
	weekStart := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	// Insert commits across the week
	// Monday - Friday (weekdays)
	for i := 0; i < 5; i++ {
		timestamp := weekStart.Add(time.Duration(i*24) * time.Hour).Add(10 * time.Hour)
		_, err := db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			"weekday-"+timestamp.Format(time.RFC3339Nano), "push", "github", "commit-weekday-"+timestamp.Format(time.RFC3339Nano),
			timestamp, "user@example.com", engineerID, "{}")
		if err != nil {
			t.Fatalf("failed to insert weekday event: %v", err)
		}
	}

	// Saturday (day 5)
	satTimestamp := weekStart.Add(5 * 24 * time.Hour).Add(10 * time.Hour)
	_, err = db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"saturday", "push", "github", "commit-sat",
		satTimestamp, "user@example.com", engineerID, "{}")
	if err != nil {
		t.Fatalf("failed to insert Saturday event: %v", err)
	}

	// Sunday (day 6)
	sunTimestamp := weekStart.Add(6 * 24 * time.Hour).Add(14 * time.Hour)
	_, err = db.Exec(`INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"sunday", "push", "github", "commit-sun",
		sunTimestamp, "user@example.com", engineerID, "{}")
	if err != nil {
		t.Fatalf("failed to insert Sunday event: %v", err)
	}

	// Count weekend commits
	count, err := service.countWeekendCommits(engineerID, weekStart, weekEnd)
	if err != nil {
		t.Fatalf("countWeekendCommits failed: %v", err)
	}

	expectedCount := 2 // Saturday + Sunday
	if count != expectedCount {
		t.Errorf("expected %d weekend commits, got %d", expectedCount, count)
	}
}
