package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// BriefingsStore handles weekly briefing storage operations
type BriefingsStore struct {
	db *sql.DB
}

// NewBriefingsStore creates a new briefings store
func NewBriefingsStore(db *sql.DB) *BriefingsStore {
	return &BriefingsStore{db: db}
}

// CreateBriefing stores a new weekly briefing
func (s *BriefingsStore) CreateBriefing(briefing *models.WeeklyBriefing) error {
	if briefing.ID == "" {
		briefing.ID = uuid.New().String()
	}
	if briefing.CreatedAt.IsZero() {
		briefing.CreatedAt = time.Now().UTC()
	}

	// Marshal JSON fields
	keyMetricsJSON, err := json.Marshal(briefing.KeyMetrics)
	if err != nil {
		return fmt.Errorf("failed to marshal key metrics: %w", err)
	}

	needsAttentionJSON, err := json.Marshal(briefing.NeedsAttention)
	if err != nil {
		return fmt.Errorf("failed to marshal needs attention: %w", err)
	}

	insightsJSON, err := json.Marshal(briefing.Insights)
	if err != nil {
		return fmt.Errorf("failed to marshal insights: %w", err)
	}

	trendingUpJSON, err := json.Marshal(briefing.TrendingUp)
	if err != nil {
		return fmt.Errorf("failed to marshal trending up: %w", err)
	}

	trendingDownJSON, err := json.Marshal(briefing.TrendingDown)
	if err != nil {
		return fmt.Errorf("failed to marshal trending down: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO weekly_briefings (
			id, team_id, week_start, tldr,
			key_metrics, needs_attention, insights,
			trending_up, trending_down, talking_points,
			generated_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		briefing.ID,
		briefing.TeamID,
		briefing.WeekStart.UTC(),
		briefing.TLDR,
		string(keyMetricsJSON),
		string(needsAttentionJSON),
		string(insightsJSON),
		string(trendingUpJSON),
		string(trendingDownJSON),
		briefing.TalkingPoints,
		briefing.GeneratedAt.UTC(),
		briefing.CreatedAt.UTC(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert briefing: %w", err)
	}

	return nil
}

// GetBriefing retrieves a briefing by team and week start
func (s *BriefingsStore) GetBriefing(teamID string, weekStart time.Time) (*models.WeeklyBriefing, error) {
	weekStart = weekStart.UTC().Truncate(24 * time.Hour)

	var briefing models.WeeklyBriefing
	var keyMetricsJSON, needsAttentionJSON, insightsJSON string
	var trendingUpJSON, trendingDownJSON string

	err := s.db.QueryRow(`
		SELECT
			id, team_id, week_start, tldr,
			key_metrics, needs_attention, insights,
			trending_up, trending_down, talking_points,
			generated_at, created_at
		FROM weekly_briefings
		WHERE team_id = ? AND week_start = ?
	`, teamID, weekStart).Scan(
		&briefing.ID,
		&briefing.TeamID,
		&briefing.WeekStart,
		&briefing.TLDR,
		&keyMetricsJSON,
		&needsAttentionJSON,
		&insightsJSON,
		&trendingUpJSON,
		&trendingDownJSON,
		&briefing.TalkingPoints,
		&briefing.GeneratedAt,
		&briefing.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get briefing: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(keyMetricsJSON), &briefing.KeyMetrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal key metrics: %w", err)
	}
	if err := json.Unmarshal([]byte(needsAttentionJSON), &briefing.NeedsAttention); err != nil {
		return nil, fmt.Errorf("failed to unmarshal needs attention: %w", err)
	}
	if err := json.Unmarshal([]byte(insightsJSON), &briefing.Insights); err != nil {
		return nil, fmt.Errorf("failed to unmarshal insights: %w", err)
	}
	if err := json.Unmarshal([]byte(trendingUpJSON), &briefing.TrendingUp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trending up: %w", err)
	}
	if err := json.Unmarshal([]byte(trendingDownJSON), &briefing.TrendingDown); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trending down: %w", err)
	}

	return &briefing, nil
}

// ListBriefings retrieves all briefings for a team, ordered by week start (newest first)
func (s *BriefingsStore) ListBriefings(teamID string, limit int) ([]*models.WeeklyBriefing, error) {
	// Always enforce minimum LIMIT for safety
	if limit <= 0 {
		limit = 52 // 1 year of weekly briefings
	}
	if limit > 520 {
		limit = 520 // Max 10 years
	}

	query := `
		SELECT
			id, team_id, week_start, tldr,
			key_metrics, needs_attention, insights,
			trending_up, trending_down, talking_points,
			generated_at, created_at
		FROM weekly_briefings
		WHERE team_id = ?
		ORDER BY week_start DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, teamID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query briefings: %w", err)
	}
	defer rows.Close()

	var briefings []*models.WeeklyBriefing
	for rows.Next() {
		var briefing models.WeeklyBriefing
		var keyMetricsJSON, needsAttentionJSON, insightsJSON string
		var trendingUpJSON, trendingDownJSON string

		err := rows.Scan(
			&briefing.ID,
			&briefing.TeamID,
			&briefing.WeekStart,
			&briefing.TLDR,
			&keyMetricsJSON,
			&needsAttentionJSON,
			&insightsJSON,
			&trendingUpJSON,
			&trendingDownJSON,
			&briefing.TalkingPoints,
			&briefing.GeneratedAt,
			&briefing.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan briefing: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(keyMetricsJSON), &briefing.KeyMetrics); err != nil {
			return nil, fmt.Errorf("failed to unmarshal key metrics: %w", err)
		}
		if err := json.Unmarshal([]byte(needsAttentionJSON), &briefing.NeedsAttention); err != nil {
			return nil, fmt.Errorf("failed to unmarshal needs attention: %w", err)
		}
		if err := json.Unmarshal([]byte(insightsJSON), &briefing.Insights); err != nil {
			return nil, fmt.Errorf("failed to unmarshal insights: %w", err)
		}
		if err := json.Unmarshal([]byte(trendingUpJSON), &briefing.TrendingUp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal trending up: %w", err)
		}
		if err := json.Unmarshal([]byte(trendingDownJSON), &briefing.TrendingDown); err != nil {
			return nil, fmt.Errorf("failed to unmarshal trending down: %w", err)
		}

		briefings = append(briefings, &briefing)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating briefings: %w", err)
	}

	return briefings, nil
}

// DeleteBriefing deletes a briefing
func (s *BriefingsStore) DeleteBriefing(teamID string, weekStart time.Time) error {
	weekStart = weekStart.UTC().Truncate(24 * time.Hour)

	_, err := s.db.Exec(`DELETE FROM weekly_briefings WHERE team_id = ? AND week_start = ?`,
		teamID, weekStart)
	if err != nil {
		return fmt.Errorf("failed to delete briefing: %w", err)
	}

	return nil
}

// GetLatestBriefing retrieves the most recent briefing for a team
func (s *BriefingsStore) GetLatestBriefing(teamID string) (*models.WeeklyBriefing, error) {
	briefings, err := s.ListBriefings(teamID, 1)
	if err != nil {
		return nil, err
	}

	if len(briefings) == 0 {
		return nil, nil
	}

	return briefings[0], nil
}
