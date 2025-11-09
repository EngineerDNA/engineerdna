package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

// ExportsStore handles export record storage operations
type ExportsStore struct {
	db *sql.DB
}

// NewExportsStore creates a new exports store
func NewExportsStore(db *sql.DB) *ExportsStore {
	return &ExportsStore{db: db}
}

// CreateExport records an export operation
func (s *ExportsStore) CreateExport(export *models.Export) error {
	_, err := s.db.Exec(`
		INSERT INTO exports (
			id, plugin_name, exported_at, status, destination_url,
			event_count, anonymized, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		export.ID,
		export.PluginName,
		export.ExportedAt.UTC(),
		export.Status,
		export.DestinationURL,
		export.EventCount,
		export.Anonymized,
		export.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to create export record: %w", err)
	}

	return nil
}

// ListExports retrieves all export records with optional filters
func (s *ExportsStore) ListExports(limit, offset int) ([]*models.Export, error) {
	query := `
		SELECT id, plugin_name, exported_at, status, destination_url,
		       event_count, anonymized, error_message, created_at
		FROM exports
		ORDER BY exported_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query exports: %w", err)
	}
	defer rows.Close()

	var exports []*models.Export
	for rows.Next() {
		var export models.Export
		var createdAt time.Time

		err := rows.Scan(
			&export.ID,
			&export.PluginName,
			&export.ExportedAt,
			&export.Status,
			&export.DestinationURL,
			&export.EventCount,
			&export.Anonymized,
			&export.ErrorMessage,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan export: %w", err)
		}

		exports = append(exports, &export)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return exports, nil
}

// CountExports returns the total count of all exports
func (s *ExportsStore) CountExports() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM exports").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count exports: %w", err)
	}
	return count, nil
}

// GetExport retrieves a single export by ID
func (s *ExportsStore) GetExport(id string) (*models.Export, error) {
	var export models.Export
	var createdAt time.Time

	err := s.db.QueryRow(`
		SELECT id, plugin_name, exported_at, status, destination_url,
		       event_count, anonymized, error_message, created_at
		FROM exports
		WHERE id = ?
	`, id).Scan(
		&export.ID,
		&export.PluginName,
		&export.ExportedAt,
		&export.Status,
		&export.DestinationURL,
		&export.EventCount,
		&export.Anonymized,
		&export.ErrorMessage,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("export not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get export: %w", err)
	}

	return &export, nil
}

// GetExportsByPlugin retrieves exports for a specific plugin
func (s *ExportsStore) GetExportsByPlugin(pluginName string, limit int) ([]*models.Export, error) {
	query := `
		SELECT id, plugin_name, exported_at, status, destination_url,
		       event_count, anonymized, error_message, created_at
		FROM exports
		WHERE plugin_name = ?
		ORDER BY exported_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, pluginName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query exports by plugin: %w", err)
	}
	defer rows.Close()

	var exports []*models.Export
	for rows.Next() {
		var export models.Export
		var createdAt time.Time

		err := rows.Scan(
			&export.ID,
			&export.PluginName,
			&export.ExportedAt,
			&export.Status,
			&export.DestinationURL,
			&export.EventCount,
			&export.Anonymized,
			&export.ErrorMessage,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan export: %w", err)
		}

		exports = append(exports, &export)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return exports, nil
}

// DeleteOldExports removes export records older than the specified days
func (s *ExportsStore) DeleteOldExports(daysOld int) (int64, error) {
	cutoffDate := time.Now().UTC().AddDate(0, 0, -daysOld)

	result, err := s.db.Exec(`
		DELETE FROM exports
		WHERE exported_at < ?
	`, cutoffDate)

	if err != nil {
		return 0, fmt.Errorf("failed to delete old exports: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
