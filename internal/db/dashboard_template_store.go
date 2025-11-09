package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// DashboardTemplateStore provides data access methods for dashboard template operations
type DashboardTemplateStore struct {
	db *sql.DB
}

func NewDashboardTemplateStore(db *sql.DB) *DashboardTemplateStore {
	return &DashboardTemplateStore{db: db}
}

// ListTemplates retrieves all dashboard templates, optionally filtered by role
func (s *DashboardTemplateStore) ListTemplates(role string) ([]*models.DashboardTemplate, error) {
	query := `
		SELECT id, name, description, role, category, layout, is_system, created_at, updated_at
		FROM dashboard_templates
		WHERE 1=1
	`
	args := []interface{}{}

	if role != "" {
		query += " AND (role = ? OR role IS NULL)"
		args = append(args, role)
	}

	query += " ORDER BY name"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.DashboardTemplate
	for rows.Next() {
		var template models.DashboardTemplate
		var description, role, category sql.NullString
		var createdAtStr, updatedAtStr string
		var isSystem int // SQLite boolean

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&description,
			&role,
			&category,
			&template.Layout,
			&isSystem,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		if description.Valid {
			template.Description = description.String
		}
		if role.Valid {
			template.Role = role.String
		}
		if category.Valid {
			template.Category = category.String
		}
		template.IsSystem = isSystem == 1

		// Parse timestamps
		template.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}
		template.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		templates = append(templates, &template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating templates: %w", err)
	}

	return templates, nil
}

// GetTemplate retrieves a single template by ID
func (s *DashboardTemplateStore) GetTemplate(id string) (*models.DashboardTemplate, error) {
	var template models.DashboardTemplate
	var description, role, category sql.NullString
	var createdAtStr, updatedAtStr string
	var isSystem int // SQLite boolean

	err := s.db.QueryRow(`
		SELECT id, name, description, role, category, layout, is_system, created_at, updated_at
		FROM dashboard_templates
		WHERE id = ?
	`, id).Scan(
		&template.ID,
		&template.Name,
		&description,
		&role,
		&category,
		&template.Layout,
		&isSystem,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if description.Valid {
		template.Description = description.String
	}
	if role.Valid {
		template.Role = role.String
	}
	if category.Valid {
		template.Category = category.String
	}
	template.IsSystem = isSystem == 1

	// Parse timestamps
	template.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	template.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &template, nil
}

// CreateTemplate creates a new dashboard template
func (s *DashboardTemplateStore) CreateTemplate(template *models.DashboardTemplate) error {
	template.ID = uuid.New().String()
	template.CreatedAt = time.Now().UTC()
	template.UpdatedAt = time.Now().UTC()

	// Convert bool to SQLite integer
	isSystemInt := 0
	if template.IsSystem {
		isSystemInt = 1
	}

	_, err := s.db.Exec(`
		INSERT INTO dashboard_templates (
			id, name, description, role, category, layout, is_system, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		template.ID,
		template.Name,
		template.Description,
		template.Role,
		template.Category,
		template.Layout,
		isSystemInt,
		template.CreatedAt.Format("2006-01-02 15:04:05"),
		template.UpdatedAt.Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

// UpdateTemplate updates an existing dashboard template
func (s *DashboardTemplateStore) UpdateTemplate(template *models.DashboardTemplate) error {
	template.UpdatedAt = time.Now().UTC()

	// Convert bool to SQLite integer
	isSystemInt := 0
	if template.IsSystem {
		isSystemInt = 1
	}

	result, err := s.db.Exec(`
		UPDATE dashboard_templates
		SET name = ?, description = ?, role = ?, category = ?, layout = ?,
		    is_system = ?, updated_at = ?
		WHERE id = ?
	`,
		template.Name,
		template.Description,
		template.Role,
		template.Category,
		template.Layout,
		isSystemInt,
		template.UpdatedAt.Format("2006-01-02 15:04:05"),
		template.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %s", template.ID)
	}

	return nil
}

// DeleteTemplate deletes a dashboard template by ID, preventing deletion of system templates
func (s *DashboardTemplateStore) DeleteTemplate(id string) error {
	// Check if template is system template
	var isSystem int
	err := s.db.QueryRow("SELECT is_system FROM dashboard_templates WHERE id = ?", id).Scan(&isSystem)
	if err == sql.ErrNoRows {
		return fmt.Errorf("template not found: %s", id)
	}
	if err != nil {
		return fmt.Errorf("failed to check template: %w", err)
	}

	if isSystem == 1 {
		return fmt.Errorf("cannot delete system template: %s", id)
	}

	result, err := s.db.Exec("DELETE FROM dashboard_templates WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %s", id)
	}

	return nil
}
