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
