package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

type ScoringStore struct {
	db *sql.DB
}

func NewScoringStore(db *sql.DB) *ScoringStore {
	return &ScoringStore{db: db}
}

// Role CRUD operations

func (s *ScoringStore) CreateRole(name string, targetScore int, expectations map[string]float64) (*models.Role, error) {
	role := &models.Role{
		ID:           uuid.New().String(),
		Name:         name,
		TargetScore:  targetScore,
		Expectations: expectations,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	expectationsJSON, err := json.Marshal(expectations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal expectations: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO roles (id, name, target_score, expectations, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, role.ID, role.Name, role.TargetScore, string(expectationsJSON), role.CreatedAt, role.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

func (s *ScoringStore) GetRole(id string) (*models.Role, error) {
	var role models.Role
	var expectationsJSON string

	err := s.db.QueryRow(`
		SELECT id, name, target_score, expectations, created_at, updated_at
		FROM roles
		WHERE id = ?
	`, id).Scan(&role.ID, &role.Name, &role.TargetScore, &expectationsJSON, &role.CreatedAt, &role.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	if err := json.Unmarshal([]byte(expectationsJSON), &role.Expectations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal expectations: %w", err)
	}

	return &role, nil
}

func (s *ScoringStore) ListRoles(limit int) ([]*models.Role, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = MaxQueryLimit
	}

	rows, err := s.db.Query(`
		SELECT id, name, target_score, expectations, created_at, updated_at
		FROM roles
		ORDER BY name
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		var role models.Role
		var expectationsJSON string

		err := rows.Scan(&role.ID, &role.Name, &role.TargetScore, &expectationsJSON, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		if err := json.Unmarshal([]byte(expectationsJSON), &role.Expectations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal expectations: %w", err)
		}

		roles = append(roles, &role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

func (s *ScoringStore) UpdateRole(id, name string, targetScore int, expectations map[string]float64) error {
	expectationsJSON, err := json.Marshal(expectations)
	if err != nil {
		return fmt.Errorf("failed to marshal expectations: %w", err)
	}

	result, err := s.db.Exec(`
		UPDATE roles
		SET name = ?, target_score = ?, expectations = ?, updated_at = ?
		WHERE id = ?
	`, name, targetScore, string(expectationsJSON), time.Now().UTC(), id)

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found: %s", id)
	}

	return nil
}

func (s *ScoringStore) DeleteRole(id string) error {
	result, err := s.db.Exec(`DELETE FROM roles WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found: %s", id)
	}

	return nil
}

// Scoring Weights operations

func (s *ScoringStore) GetWeights() (*models.ScoringWeights, error) {
	var weights models.ScoringWeights

	err := s.db.QueryRow(`
		SELECT id, throughput_weight, quality_weight, speed_weight, collaboration_weight, impact_weight, updated_at
		FROM scoring_weights
		ORDER BY updated_at DESC
		LIMIT 1
	`).Scan(&weights.ID, &weights.ThroughputWeight, &weights.QualityWeight, &weights.SpeedWeight, &weights.CollaborationWeight, &weights.ImpactWeight, &weights.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get weights: %w", err)
	}

	return &weights, nil
}

func (s *ScoringStore) UpdateWeights(throughput, quality, speed, collaboration, impact float64) (*models.ScoringWeights, error) {
	// Delete existing weights and insert new ones (simpler than update)
	_, err := s.db.Exec(`DELETE FROM scoring_weights`)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old weights: %w", err)
	}

	weights := &models.ScoringWeights{
		ID:                  uuid.New().String(),
		ThroughputWeight:    throughput,
		QualityWeight:       quality,
		SpeedWeight:         speed,
		CollaborationWeight: collaboration,
		ImpactWeight:        impact,
		UpdatedAt:           time.Now().UTC(),
	}

	_, err = s.db.Exec(`
		INSERT INTO scoring_weights (id, throughput_weight, quality_weight, speed_weight, collaboration_weight, impact_weight, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, weights.ID, weights.ThroughputWeight, weights.QualityWeight, weights.SpeedWeight, weights.CollaborationWeight, weights.ImpactWeight, weights.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert weights: %w", err)
	}

	return weights, nil
}

// Performance Scores operations

func (s *ScoringStore) GetPerformanceScores(engineerID string, startDate, endDate time.Time, limit int) ([]*models.PerformanceScore, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = MaxQueryLimit
	}

	query := `
		SELECT id, engineer_id, week_start, total_score,
		       throughput_score, quality_score, speed_score, collaboration_score, impact_score,
		       raw_metrics, burnout_risk, created_at
		FROM performance_scores
		WHERE engineer_id = ?
		  AND week_start >= ?
		  AND week_start <= ?
		ORDER BY week_start DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, engineerID, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query performance scores: %w", err)
	}
	defer rows.Close()

	var scores []*models.PerformanceScore
	for rows.Next() {
		var score models.PerformanceScore
		var burnoutRisk sql.NullString
		err := rows.Scan(
			&score.ID, &score.EngineerID, &score.WeekStart, &score.TotalScore,
			&score.ThroughputScore, &score.QualityScore, &score.SpeedScore, &score.CollaborationScore, &score.ImpactScore,
			&score.RawMetrics, &burnoutRisk, &score.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance score: %w", err)
		}
		if burnoutRisk.Valid {
			score.BurnoutRisk = burnoutRisk.String
		}
		scores = append(scores, &score)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating performance scores: %w", err)
	}

	return scores, nil
}
