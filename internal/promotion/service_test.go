package promotion

import (
	"database/sql"
	"testing"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Create in-memory database
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migrations
	if err := db.RunMigrations(database); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

func TestHasConsecutiveWeeks(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		scores   []*models.PerformanceScore
		expected bool
	}{
		{
			name:     "empty scores",
			scores:   []*models.PerformanceScore{},
			expected: true,
		},
		{
			name: "single score",
			scores: []*models.PerformanceScore{
				{WeekStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
			expected: true,
		},
		{
			name: "consecutive weeks",
			scores: []*models.PerformanceScore{
				{WeekStart: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
				{WeekStart: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)},
				{WeekStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
			expected: true,
		},
		{
			name: "gap in weeks",
			scores: []*models.PerformanceScore{
				{WeekStart: time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)},
				{WeekStart: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
				{WeekStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}, // Gap between 1/1 and 1/15
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.hasConsecutiveWeeks(tt.scores)
			if result != tt.expected {
				t.Errorf("hasConsecutiveWeeks() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestDetectPromotionSignals(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	service := NewService(database)

	// Create a role
	roleID := uuid.New().String()
	_, err := database.Exec(`
		INSERT INTO roles (id, name, target_score, expectations, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, roleID, "Junior Engineer", 70, `{"throughput_prs_per_week": 3.0}`, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	// Create an engineer with role
	engineerID := uuid.New().String()
	_, err = database.Exec(`
		INSERT INTO engineers (id, canonical_name, email, role_id, identifiers, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, engineerID, "Test Engineer", "test@example.com", roleID, `{}`, true, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to create engineer: %v", err)
	}

	// Create 8 weeks of performance scores in metric_values table
	now := time.Now().UTC()
	for i := 0; i < 8; i++ {
		weekStart := now.AddDate(0, 0, -7*i).Truncate(24 * time.Hour)
		scoreID := uuid.New().String()

		_, err := database.Exec(`
			INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, scoreID, "engineer_total_score", "scoring_system", weekStart.Format(time.RFC3339), "weekly", 88.0, "score",
			`{"engineer_id":"`+engineerID+`"}`, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			t.Fatalf("Failed to create performance score: %v", err)
		}
	}

	// Detect promotion signals
	signals, err := service.DetectPromotionSignals()
	if err != nil {
		t.Fatalf("DetectPromotionSignals failed: %v", err)
	}

	// Should detect one signal
	if len(signals) != 1 {
		t.Errorf("Expected 1 signal, got %d", len(signals))
	}

	if len(signals) > 0 {
		signal := signals[0]
		if signal.EngineerID != engineerID {
			t.Errorf("Expected engineer_id %s, got %s", engineerID, signal.EngineerID)
		}
		if signal.WeeksDuration != 8 {
			t.Errorf("Expected weeks_duration 8, got %d", signal.WeeksDuration)
		}
		if signal.AvgScore != 88.0 {
			t.Errorf("Expected avg_score 88.0, got %f", signal.AvgScore)
		}
		if signal.SignalType != "consistent_overperformance" {
			t.Errorf("Expected signal_type 'consistent_overperformance', got %s", signal.SignalType)
		}
	}
}

func TestDetectPromotionSignalsNoSignal(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	service := NewService(database)

	// Create a role
	roleID := uuid.New().String()
	_, err := database.Exec(`
		INSERT INTO roles (id, name, target_score, expectations, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, roleID, "Junior Engineer", 70, `{"throughput_prs_per_week": 3.0}`, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	// Create an engineer with role
	engineerID := uuid.New().String()
	_, err = database.Exec(`
		INSERT INTO engineers (id, canonical_name, email, role_id, identifiers, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, engineerID, "Test Engineer", "test@example.com", roleID, `{}`, true, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to create engineer: %v", err)
	}

	// Create 8 weeks of performance scores in metric_values table, but some below threshold
	now := time.Now().UTC()
	for i := 0; i < 8; i++ {
		weekStart := now.AddDate(0, 0, -7*i).Truncate(24 * time.Hour)
		scoreID := uuid.New().String()

		// Scores alternate between 88 and 80 (80 < 84 threshold)
		score := 88.0
		if i%2 == 0 {
			score = 80.0
		}

		_, err := database.Exec(`
			INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, scoreID, "engineer_total_score", "scoring_system", weekStart.Format(time.RFC3339), "weekly", score, "score",
			`{"engineer_id":"`+engineerID+`"}`, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			t.Fatalf("Failed to create performance score: %v", err)
		}
	}

	// Detect promotion signals
	signals, err := service.DetectPromotionSignals()
	if err != nil {
		t.Fatalf("DetectPromotionSignals failed: %v", err)
	}

	// Should not detect any signals
	if len(signals) != 0 {
		t.Errorf("Expected 0 signals, got %d", len(signals))
	}
}

func TestDismissPromotionSignal(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	service := NewService(database)

	// Create a role
	roleID := uuid.New().String()
	_, err := database.Exec(`
		INSERT INTO roles (id, name, target_score, expectations, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, roleID, "Junior Engineer", 70, `{"throughput_prs_per_week": 3.0}`, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	// Create an engineer
	engineerID := uuid.New().String()
	_, err = database.Exec(`
		INSERT INTO engineers (id, canonical_name, email, role_id, identifiers, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, engineerID, "Test Engineer", "test@example.com", roleID, `{}`, true, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to create engineer: %v", err)
	}

	// Create a promotion signal
	signalID := uuid.New().String()
	_, err = database.Exec(`
		INSERT INTO promotion_signals (id, engineer_id, signal_type, detected_at, weeks_duration, avg_score, dismissed)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, signalID, engineerID, "consistent_overperformance", time.Now().UTC(), 8, 88.0, false)
	if err != nil {
		t.Fatalf("Failed to create promotion signal: %v", err)
	}

	// Dismiss the signal
	err = service.DismissPromotionSignal(signalID, "Not ready yet")
	if err != nil {
		t.Fatalf("DismissPromotionSignal failed: %v", err)
	}

	// Verify signal is dismissed
	var dismissed bool
	var notes string
	err = database.QueryRow(`SELECT dismissed, notes FROM promotion_signals WHERE id = ?`, signalID).Scan(&dismissed, &notes)
	if err != nil {
		t.Fatalf("Failed to query promotion signal: %v", err)
	}

	if !dismissed {
		t.Error("Expected signal to be dismissed")
	}
	if notes != "Not ready yet" {
		t.Errorf("Expected notes 'Not ready yet', got %s", notes)
	}
}
