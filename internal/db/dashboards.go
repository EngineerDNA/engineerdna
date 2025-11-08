package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// DashboardStore provides data access methods for dashboard and metric snapshot operations.
type DashboardStore struct {
	db *sql.DB
}

func NewDashboardStore(db *sql.DB) *DashboardStore {
	return &DashboardStore{db: db}
}

// Dashboard CRUD

// CreateDashboard creates a new dashboard in the database with generated ID and timestamps.
func (s *DashboardStore) CreateDashboard(dashboard *models.Dashboard) error {
	dashboard.ID = uuid.New().String()
	dashboard.CreatedAt = time.Now().UTC()
	dashboard.UpdatedAt = time.Now().UTC()

	_, err := s.db.Exec(`
		INSERT INTO dashboards (
			id, name, description, persona, is_template, is_system,
			layout, filters, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, dashboard.ID, dashboard.Name, dashboard.Description, dashboard.Persona,
		dashboard.IsTemplate, dashboard.IsSystem, dashboard.Layout, dashboard.Filters,
		dashboard.CreatedAt.Format(time.RFC3339), dashboard.UpdatedAt.Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("failed to create dashboard: %w", err)
	}

	return nil
}

// GetDashboard retrieves a dashboard by ID, returning nil if not found.
func (s *DashboardStore) GetDashboard(id string) (*models.Dashboard, error) {
	var dashboard models.Dashboard
	var description, persona, filters sql.NullString
	var createdAtStr, updatedAtStr string

	err := s.db.QueryRow(`
		SELECT id, name, description, persona, is_template, is_system,
		       layout, filters, created_at, updated_at
		FROM dashboards
		WHERE id = ?
	`, id).Scan(&dashboard.ID, &dashboard.Name, &description, &persona,
		&dashboard.IsTemplate, &dashboard.IsSystem, &dashboard.Layout, &filters,
		&createdAtStr, &updatedAtStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	if description.Valid {
		dashboard.Description = description.String
	}
	if persona.Valid {
		dashboard.Persona = persona.String
	}
	if filters.Valid {
		dashboard.Filters = filters.String
	}

	// Parse timestamps
	dashboard.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	dashboard.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &dashboard, nil
}

// UpdateDashboard updates an existing dashboard, setting the updated_at timestamp.
func (s *DashboardStore) UpdateDashboard(dashboard *models.Dashboard) error {
	dashboard.UpdatedAt = time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE dashboards
		SET name = ?, description = ?, persona = ?, is_template = ?, is_system = ?,
		    layout = ?, filters = ?, updated_at = ?
		WHERE id = ?
	`, dashboard.Name, dashboard.Description, dashboard.Persona, dashboard.IsTemplate,
		dashboard.IsSystem, dashboard.Layout, dashboard.Filters, dashboard.UpdatedAt.Format(time.RFC3339),
		dashboard.ID)

	if err != nil {
		return fmt.Errorf("failed to update dashboard: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dashboard not found: %s", dashboard.ID)
	}

	return nil
}

// DeleteDashboard deletes a dashboard by ID, preventing deletion of system dashboards.
func (s *DashboardStore) DeleteDashboard(id string) error {
	// Check if dashboard is system dashboard
	var isSystem bool
	err := s.db.QueryRow("SELECT is_system FROM dashboards WHERE id = ?", id).Scan(&isSystem)
	if err == sql.ErrNoRows {
		return fmt.Errorf("dashboard not found: %s", id)
	}
	if err != nil {
		return fmt.Errorf("failed to check dashboard: %w", err)
	}

	if isSystem {
		return fmt.Errorf("cannot delete system dashboard: %s", id)
	}

	result, err := s.db.Exec("DELETE FROM dashboards WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dashboard not found: %s", id)
	}

	return nil
}

// ListDashboards retrieves a paginated list of dashboards, optionally filtered by persona and template status.
func (s *DashboardStore) ListDashboards(persona string, templatesOnly bool, limit, offset int) ([]*models.Dashboard, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = DefaultQueryLimit
	}
	if limit > MaxQueryLimit {
		limit = MaxQueryLimit
	}

	// Build query with filters
	countQuery := "SELECT COUNT(*) FROM dashboards WHERE 1=1"
	query := `
		SELECT id, name, description, persona, is_template, is_system,
		       layout, filters, created_at, updated_at
		FROM dashboards WHERE 1=1
	`
	args := []interface{}{}

	if persona != "" {
		countQuery += " AND persona = ?"
		query += " AND persona = ?"
		args = append(args, persona)
	}

	if templatesOnly {
		countQuery += " AND is_template = true"
		query += " AND is_template = true"
	}

	// Count total matching records
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count dashboards: %w", err)
	}

	// Query with pagination
	query += " ORDER BY name LIMIT ? OFFSET ?"
	queryArgs := append(args, limit, offset)

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list dashboards: %w", err)
	}
	defer rows.Close()

	var dashboards []*models.Dashboard
	for rows.Next() {
		var dashboard models.Dashboard
		var description, persona, filters sql.NullString
		var createdAtStr, updatedAtStr string

		err := rows.Scan(&dashboard.ID, &dashboard.Name, &description, &persona,
			&dashboard.IsTemplate, &dashboard.IsSystem, &dashboard.Layout, &filters,
			&createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan dashboard: %w", err)
		}

		if description.Valid {
			dashboard.Description = description.String
		}
		if persona.Valid {
			dashboard.Persona = persona.String
		}
		if filters.Valid {
			dashboard.Filters = filters.String
		}

		// Parse timestamps
		dashboard.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse created_at: %w", err)
		}
		dashboard.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		dashboards = append(dashboards, &dashboard)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating dashboards: %w", err)
	}

	return dashboards, total, nil
}

// CloneDashboard creates a copy of an existing dashboard with a new name.
func (s *DashboardStore) CloneDashboard(sourceID string, newName string) (*models.Dashboard, error) {
	// Get source dashboard
	source, err := s.GetDashboard(sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source dashboard: %w", err)
	}
	if source == nil {
		return nil, fmt.Errorf("source dashboard not found: %s", sourceID)
	}

	// Create new dashboard from source
	clone := &models.Dashboard{
		Name:        newName,
		Description: source.Description,
		Persona:     source.Persona,
		IsTemplate:  false, // Clones are never templates
		IsSystem:    false, // Clones are never system dashboards
		Layout:      source.Layout,
		Filters:     source.Filters,
	}

	err = s.CreateDashboard(clone)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloned dashboard: %w", err)
	}

	return clone, nil
}

// Metric Computation Methods

// TeamMetrics holds aggregated metrics for a team
type TeamMetrics struct {
	TeamID     string
	TeamScore  float64
	PRCount    float64
	AlertCount float64
}

// EngineerMetrics holds aggregated metrics for an engineer
type EngineerMetrics struct {
	EngineerID    string
	EngineerScore float64
	PRCount       float64
}

// CountPRsByDate counts pull requests for a specific date
func (s *DashboardStore) CountPRsByDate(date string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE type = 'pull_request'
		  AND DATE(timestamp) = ?
	`, date).Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to count PRs by date: %w", err)
	}

	return count, nil
}

// CountActiveAlerts counts unresolved alerts in a date range
func (s *DashboardStore) CountActiveAlerts(start, end string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM alert_instances
		WHERE fired_at >= ? AND fired_at < ?
		  AND resolved_at IS NULL
	`, start, end).Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to count active alerts: %w", err)
	}

	return count, nil
}

// CountActiveGoals counts goals with active status
func (s *DashboardStore) CountActiveGoals() (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM goals
		WHERE status = 'active'
	`).Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to count active goals: %w", err)
	}

	return count, nil
}

// GetTeamMetricsForDate retrieves aggregated metrics for all teams on a specific date
func (s *DashboardStore) GetTeamMetricsForDate(periodStart, periodEnd string) ([]TeamMetrics, error) {
	query := `
		SELECT
			t.id,
			COALESCE(tps.total_score, 0) as team_score,
			COUNT(DISTINCT CASE WHEN e.type = 'pull_request' AND DATE(e.timestamp) = ? THEN e.id END) as pr_count,
			COUNT(DISTINCT CASE WHEN ai.fired_at >= ? AND ai.fired_at < ? AND ai.resolved_at IS NULL THEN ai.id END) as alert_count
		FROM teams t
		LEFT JOIN (
			SELECT team_id, total_score, created_at,
				ROW_NUMBER() OVER (PARTITION BY team_id ORDER BY created_at DESC) as rn
			FROM team_performance_scores
		) tps ON t.id = tps.team_id AND tps.rn = 1
		LEFT JOIN team_membership tm ON t.id = tm.team_id AND tm.left_at IS NULL
		LEFT JOIN engineers eng ON tm.member_id = eng.id
		LEFT JOIN events e ON eng.email = e.actor
		LEFT JOIN alert_instances ai ON ai.entity_type = 'team' AND ai.entity_id = t.id
		GROUP BY t.id, tps.total_score
	`

	rows, err := s.db.Query(query, periodStart, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get team metrics: %w", err)
	}
	defer rows.Close()

	var metrics []TeamMetrics
	for rows.Next() {
		var m TeamMetrics
		if err := rows.Scan(&m.TeamID, &m.TeamScore, &m.PRCount, &m.AlertCount); err != nil {
			return nil, fmt.Errorf("failed to scan team metrics: %w", err)
		}
		metrics = append(metrics, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team metrics: %w", err)
	}

	return metrics, nil
}

// GetEngineerMetricsForDate retrieves aggregated metrics for engineers on a specific date
func (s *DashboardStore) GetEngineerMetricsForDate(periodStart string, limit int) ([]EngineerMetrics, error) {
	if limit <= 0 {
		limit = MaxQueryLimit
	}

	query := `
		SELECT
			e.id,
			COALESCE(ps.total_score, 0) as engineer_score,
			COUNT(DISTINCT CASE WHEN ev.type = 'pull_request' AND DATE(ev.timestamp) = ? THEN ev.id END) as pr_count
		FROM engineers e
		LEFT JOIN (
			SELECT engineer_id, total_score, created_at,
				ROW_NUMBER() OVER (PARTITION BY engineer_id ORDER BY created_at DESC) as rn
			FROM performance_scores
		) ps ON e.id = ps.engineer_id AND ps.rn = 1
		LEFT JOIN events ev ON e.email = ev.actor
		GROUP BY e.id, ps.total_score
		LIMIT ?
	`

	rows, err := s.db.Query(query, periodStart, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get engineer metrics: %w", err)
	}
	defer rows.Close()

	var metrics []EngineerMetrics
	for rows.Next() {
		var m EngineerMetrics
		if err := rows.Scan(&m.EngineerID, &m.EngineerScore, &m.PRCount); err != nil {
			return nil, fmt.Errorf("failed to scan engineer metrics: %w", err)
		}
		metrics = append(metrics, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating engineer metrics: %w", err)
	}

	return metrics, nil
}

// GetLatestTeamScore retrieves the most recent team performance score
func (s *DashboardStore) GetLatestTeamScore(teamID string) (float64, error) {
	var score sql.NullFloat64
	err := s.db.QueryRow(`
		SELECT total_score FROM team_performance_scores
		WHERE team_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, teamID).Scan(&score)

	if err == sql.ErrNoRows || !score.Valid {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get latest team score: %w", err)
	}

	return score.Float64, nil
}

// CountOrgPRsInRange counts organization-wide PRs in a date range
func (s *DashboardStore) CountOrgPRsInRange(start, end string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE type = 'pull_request'
		  AND DATE(timestamp) >= ? AND DATE(timestamp) <= ?
	`, start, end).Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to count org PRs in range: %w", err)
	}

	return count, nil
}

// CountTeamPRsInRange counts team PRs in a date range
func (s *DashboardStore) CountTeamPRsInRange(teamID, start, end string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(DISTINCT e.id) FROM events e
		INNER JOIN engineers eng ON e.actor = eng.email
		INNER JOIN team_membership tm ON eng.id = tm.member_id AND tm.left_at IS NULL
		WHERE tm.team_id = ?
		  AND e.type = 'pull_request'
		  AND DATE(e.timestamp) >= ? AND DATE(e.timestamp) <= ?
	`, teamID, start, end).Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to count team PRs in range: %w", err)
	}

	return count, nil
}

// GetEngineerEmail retrieves an engineer's email by their ID
func (s *DashboardStore) GetEngineerEmail(engineerID string) (string, error) {
	var email string
	err := s.db.QueryRow(`SELECT email FROM engineers WHERE id = ?`, engineerID).Scan(&email)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("engineer not found: %s", engineerID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get engineer email: %w", err)
	}

	return email, nil
}

// CountEngineerPRsInRange counts engineer PRs in a date range
func (s *DashboardStore) CountEngineerPRsInRange(email, start, end string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE actor = ?
		  AND type = 'pull_request'
		  AND DATE(timestamp) >= ? AND DATE(timestamp) <= ?
	`, email, start, end).Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to count engineer PRs in range: %w", err)
	}

	return count, nil
}

// CalculateAvgCycleTime calculates average PR cycle time (open to merge) in hours
func (s *DashboardStore) CalculateAvgCycleTime(start, end string) (float64, error) {
	query := `
		SELECT
			AVG(CAST((julianday(merged_at) - julianday(created_at)) * 24 AS REAL)) as avg_hours
		FROM (
			SELECT
				json_extract(data, '$.created_at') as created_at,
				json_extract(data, '$.merged_at') as merged_at
			FROM events
			WHERE type = 'pull_request'
			  AND json_extract(data, '$.merged') = 1
			  AND DATE(timestamp) >= ? AND DATE(timestamp) <= ?
		)`

	var avgCycleTime sql.NullFloat64
	err := s.db.QueryRow(query, start, end).Scan(&avgCycleTime)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to calculate avg cycle time: %w", err)
	}

	if !avgCycleTime.Valid {
		return 0, nil
	}

	return avgCycleTime.Float64, nil
}
