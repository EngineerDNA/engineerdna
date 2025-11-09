package db

import (
	"database/sql"
	"fmt"
)

type HealthStore struct {
	db *sql.DB
}

func NewHealthStore(db *sql.DB) *HealthStore {
	return &HealthStore{db: db}
}

// GetSystemHealth retrieves overall system health metrics
func (s *HealthStore) GetSystemHealth() (avgScore float64, activeAlerts, goalsOffTrack int, err error) {
	// Get average score from last 7 days from metric_values table
	err = s.db.QueryRow(`
		SELECT COALESCE(AVG(value), 0) FROM metric_values
		WHERE metric_name IN ('engineer_total_score', 'team_total_score')
		  AND created_at >= datetime('now', '-7 days')
	`).Scan(&avgScore)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, 0, fmt.Errorf("failed to calculate average score: %w", err)
	}

	// Get count of critical active alerts
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM alert_instances
		WHERE resolved_at IS NULL AND severity = 'critical'
	`).Scan(&activeAlerts)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, 0, fmt.Errorf("failed to count active alerts: %w", err)
	}

	// Get count of off-track goals
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM goals
		WHERE status = 'off_track'
	`).Scan(&goalsOffTrack)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, 0, fmt.Errorf("failed to count off-track goals: %w", err)
	}

	return avgScore, activeAlerts, goalsOffTrack, nil
}
