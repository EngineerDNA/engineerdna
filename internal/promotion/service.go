package promotion

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// Service handles promotion signal detection and management
type Service struct {
	db *sql.DB
}

// NewService creates a new promotion service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// DetectPromotionSignals scans all engineers and detects promotion signals
// based on consistent over-performance (6-8 consecutive weeks above role target)
func (s *Service) DetectPromotionSignals() ([]*models.PromotionSignal, error) {
	// Get all engineers with roles
	engineers, err := s.getEngineersWithRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to get engineers with roles: %w", err)
	}

	var signals []*models.PromotionSignal

	for _, engineer := range engineers {
		// Get last 8 weeks of scores
		scores, err := s.getRecentScores(engineer.ID, 8)
		if err != nil {
			return nil, fmt.Errorf("failed to get recent scores for engineer %s: %w", engineer.ID, err)
		}

		// Need at least 6 weeks of consecutive data
		if len(scores) < 6 {
			continue
		}

		// Check if scores are consecutive (no gaps in weeks)
		if !s.hasConsecutiveWeeks(scores) {
			continue
		}

		// Get engineer's role to calculate threshold
		role, err := s.getEngineerRole(engineer.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get role for engineer %s: %w", engineer.ID, err)
		}
		if role == nil {
			continue // Engineer has no role, skip
		}

		// Calculate threshold (120% of role target)
		threshold := float64(role.TargetScore) * 1.2

		// Check if all scores are above threshold
		allAbove := true
		totalScore := 0.0

		for _, score := range scores {
			if score.TotalScore < threshold {
				allAbove = false
				break
			}
			totalScore += score.TotalScore
		}

		// If not all scores above threshold, skip
		if !allAbove {
			continue
		}

		// Calculate average score
		avgScore := totalScore / float64(len(scores))

		// Check if signal already exists (avoid duplicates)
		existingSignal, err := s.getActiveSignal(engineer.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing signals for engineer %s: %w", engineer.ID, err)
		}
		if existingSignal != nil {
			// Signal already exists, skip
			continue
		}

		// Create new promotion signal
		signal := &models.PromotionSignal{
			ID:            uuid.New().String(),
			EngineerID:    engineer.ID,
			SignalType:    "consistent_overperformance",
			DetectedAt:    time.Now().UTC(),
			WeeksDuration: len(scores),
			AvgScore:      avgScore,
			Dismissed:     false,
		}

		// Store signal in database
		err = s.createSignal(signal)
		if err != nil {
			return nil, fmt.Errorf("failed to create signal for engineer %s: %w", engineer.ID, err)
		}

		signals = append(signals, signal)
	}

	return signals, nil
}

// GetPromotionSignals retrieves all active promotion signals
func (s *Service) GetPromotionSignals() ([]*models.PromotionSignal, error) {
	rows, err := s.db.Query(`
		SELECT id, engineer_id, signal_type, detected_at, weeks_duration, avg_score,
		       dismissed, dismissed_at, notes
		FROM promotion_signals
		WHERE dismissed = FALSE
		ORDER BY detected_at DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query promotion signals: %w", err)
	}
	defer rows.Close()

	var signals []*models.PromotionSignal
	for rows.Next() {
		signal, err := s.scanPromotionSignal(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan promotion signal: %w", err)
		}
		signals = append(signals, signal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating promotion signals: %w", err)
	}

	return signals, nil
}

// GetPromotionSignalsByEngineer retrieves all active promotion signals for a specific engineer
func (s *Service) GetPromotionSignalsByEngineer(engineerID string) ([]*models.PromotionSignal, error) {
	rows, err := s.db.Query(`
		SELECT id, engineer_id, signal_type, detected_at, weeks_duration, avg_score,
		       dismissed, dismissed_at, notes
		FROM promotion_signals
		WHERE engineer_id = ?
		  AND dismissed = FALSE
		ORDER BY detected_at DESC
		LIMIT 100
	`, engineerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query promotion signals: %w", err)
	}
	defer rows.Close()

	var signals []*models.PromotionSignal
	for rows.Next() {
		signal, err := s.scanPromotionSignal(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan promotion signal: %w", err)
		}
		signals = append(signals, signal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating promotion signals: %w", err)
	}

	return signals, nil
}

// DismissPromotionSignal dismisses a promotion signal
func (s *Service) DismissPromotionSignal(signalID string, notes string) error {
	now := time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE promotion_signals
		SET dismissed = TRUE,
		    dismissed_at = ?,
		    notes = ?
		WHERE id = ?
	`, now, notes, signalID)
	if err != nil {
		return fmt.Errorf("failed to dismiss promotion signal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("promotion signal not found: %s", signalID)
	}

	return nil
}

// getEngineersWithRoles retrieves all active engineers who have roles assigned
func (s *Service) getEngineersWithRoles() ([]*models.Engineer, error) {
	rows, err := s.db.Query(`
		SELECT id, canonical_name, email, manager, identifiers, active, created_at, updated_at
		FROM engineers
		WHERE active = TRUE
		  AND role_id IS NOT NULL
		LIMIT ?
	`, db.MaxQueryLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to query engineers: %w", err)
	}
	defer rows.Close()

	var engineers []*models.Engineer
	for rows.Next() {
		var engineer models.Engineer
		var email, manager sql.NullString
		var identifiersJSON string

		err := rows.Scan(&engineer.ID, &engineer.Name, &email, &manager, &identifiersJSON,
			&engineer.Active, &engineer.CreatedAt, &engineer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan engineer: %w", err)
		}

		if email.Valid {
			engineer.Email = email.String
		}
		if manager.Valid {
			engineer.Manager = manager.String
		}

		// Parse identifiers JSON
		if identifiersJSON != "" {
			err := json.Unmarshal([]byte(identifiersJSON), &engineer.Identifiers)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal identifiers: %w", err)
			}
		}

		engineers = append(engineers, &engineer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating engineers: %w", err)
	}

	return engineers, nil
}

// getRecentScores retrieves the most recent N weeks of performance scores for an engineer
// Queries metric_values table
func (s *Service) getRecentScores(engineerID string, weeks int) ([]*models.PerformanceScore, error) {
	// Query total scores from metric_values
	rows, err := s.db.Query(`
		SELECT id, timestamp, value, created_at
		FROM metric_values
		WHERE metric_name = 'engineer_total_score'
		  AND json_extract(dimensions, '$.engineer_id') = ?
		  AND granularity = 'weekly'
		ORDER BY timestamp DESC
		LIMIT ?
	`, engineerID, weeks)
	if err != nil {
		return nil, fmt.Errorf("failed to query performance scores: %w", err)
	}
	defer rows.Close()

	var scores []*models.PerformanceScore
	for rows.Next() {
		var score models.PerformanceScore
		score.EngineerID = engineerID
		var weekStartStr, createdAtStr string

		err := rows.Scan(&score.ID, &weekStartStr, &score.TotalScore, &createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance score: %w", err)
		}

		// Parse timestamp strings
		score.WeekStart, err = time.Parse(time.RFC3339, weekStartStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse week_start: %w", err)
		}
		score.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		// Query component scores for this week (optional, continue if not found)
		componentScores, err := s.getComponentScores(engineerID, score.WeekStart)
		if err == nil {
			// Populate component scores if available
			score.ThroughputScore = componentScores["throughput"]
			score.QualityScore = componentScores["quality"]
			score.SpeedScore = componentScores["speed"]
			score.CollaborationScore = componentScores["collaboration"]
			score.ImpactScore = componentScores["impact"]
		}
		// If component scores not found, continue with just total score

		scores = append(scores, &score)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating performance scores: %w", err)
	}

	return scores, nil
}

// getComponentScores retrieves component scores for a specific engineer and week
func (s *Service) getComponentScores(engineerID string, weekStart time.Time) (map[string]float64, error) {
	scores := make(map[string]float64)

	componentNames := []string{
		"engineer_throughput_score",
		"engineer_quality_score",
		"engineer_speed_score",
		"engineer_collaboration_score",
		"engineer_impact_score",
	}

	for _, metricName := range componentNames {
		var value sql.NullFloat64
		err := s.db.QueryRow(`
			SELECT value
			FROM metric_values
			WHERE metric_name = ?
			  AND json_extract(dimensions, '$.engineer_id') = ?
			  AND timestamp = ?
			  AND granularity = 'weekly'
			LIMIT 1
		`, metricName, engineerID, weekStart).Scan(&value)

		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to query %s: %w", metricName, err)
		}

		if value.Valid {
			// Extract component name (e.g., "throughput" from "engineer_throughput_score")
			component := metricName[9 : len(metricName)-6] // Remove "engineer_" prefix and "_score" suffix
			scores[component] = value.Float64
		}
	}

	return scores, nil
}

// hasConsecutiveWeeks checks if the scores represent consecutive weeks (no gaps)
func (s *Service) hasConsecutiveWeeks(scores []*models.PerformanceScore) bool {
	if len(scores) < 2 {
		return true // Single score or empty is considered consecutive
	}

	// Scores are ordered DESC by week_start, so we need to reverse check
	for i := 0; i < len(scores)-1; i++ {
		current := scores[i].WeekStart
		next := scores[i+1].WeekStart

		// Next week should be exactly 7 days before current
		expectedNext := current.AddDate(0, 0, -7).Truncate(24 * time.Hour)
		actualNext := next.Truncate(24 * time.Hour)

		if !expectedNext.Equal(actualNext) {
			return false // Gap detected
		}
	}

	return true
}

// getEngineerRole retrieves the role for an engineer
func (s *Service) getEngineerRole(engineerID string) (*models.Role, error) {
	// Query engineers table for role_id
	var roleID sql.NullString
	err := s.db.QueryRow(`SELECT role_id FROM engineers WHERE id = ?`, engineerID).Scan(&roleID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query engineer role: %w", err)
	}

	// If no role assigned, return nil
	if !roleID.Valid {
		return nil, nil
	}

	// Query roles table
	var role models.Role
	var expectationsJSON string
	err = s.db.QueryRow(`
		SELECT id, name, target_score, expectations, created_at, updated_at
		FROM roles WHERE id = ?
	`, roleID.String).Scan(&role.ID, &role.Name, &role.TargetScore, &expectationsJSON, &role.CreatedAt, &role.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query role: %w", err)
	}

	// Parse expectations JSON
	err = json.Unmarshal([]byte(expectationsJSON), &role.Expectations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal role expectations: %w", err)
	}

	return &role, nil
}

// getActiveSignal retrieves an active (not dismissed) promotion signal for an engineer
func (s *Service) getActiveSignal(engineerID string) (*models.PromotionSignal, error) {
	var signal models.PromotionSignal
	var dismissedAt sql.NullTime
	var notes sql.NullString

	err := s.db.QueryRow(`
		SELECT id, engineer_id, signal_type, detected_at, weeks_duration, avg_score,
		       dismissed, dismissed_at, notes
		FROM promotion_signals
		WHERE engineer_id = ?
		  AND dismissed = FALSE
		LIMIT 1
	`, engineerID).Scan(&signal.ID, &signal.EngineerID, &signal.SignalType, &signal.DetectedAt,
		&signal.WeeksDuration, &signal.AvgScore, &signal.Dismissed, &dismissedAt, &notes)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query active signal: %w", err)
	}

	if dismissedAt.Valid {
		signal.DismissedAt = &dismissedAt.Time
	}
	if notes.Valid {
		signal.Notes = notes.String
	}

	return &signal, nil
}

// createSignal creates a new promotion signal in the database
func (s *Service) createSignal(signal *models.PromotionSignal) error {
	_, err := s.db.Exec(`
		INSERT INTO promotion_signals (
			id, engineer_id, signal_type, detected_at, weeks_duration, avg_score,
			dismissed, dismissed_at, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, signal.ID, signal.EngineerID, signal.SignalType, signal.DetectedAt,
		signal.WeeksDuration, signal.AvgScore, signal.Dismissed, signal.DismissedAt, signal.Notes)

	if err != nil {
		return fmt.Errorf("failed to insert promotion signal: %w", err)
	}

	return nil
}

// scanPromotionSignal scans a promotion signal from a SQL row
func (s *Service) scanPromotionSignal(rows *sql.Rows) (*models.PromotionSignal, error) {
	var signal models.PromotionSignal
	var dismissedAt sql.NullTime
	var notes sql.NullString

	err := rows.Scan(&signal.ID, &signal.EngineerID, &signal.SignalType, &signal.DetectedAt,
		&signal.WeeksDuration, &signal.AvgScore, &signal.Dismissed, &dismissedAt, &notes)
	if err != nil {
		return nil, err
	}

	if dismissedAt.Valid {
		signal.DismissedAt = &dismissedAt.Time
	}
	if notes.Valid {
		signal.Notes = notes.String
	}

	return &signal, nil
}
