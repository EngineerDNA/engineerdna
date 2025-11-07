package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

type SettingsStore struct {
	db *sql.DB
}

func NewSettingsStore(db *sql.DB) *SettingsStore {
	return &SettingsStore{db: db}
}

// GetSettings retrieves the singleton settings row
func (s *SettingsStore) GetSettings() (*models.Settings, error) {
	var settings models.Settings
	var createdAtStr, updatedAtStr string

	err := s.db.QueryRow(`
		SELECT id, sync_schedule, anonymization_strategy, port, created_at, updated_at
		FROM settings
		WHERE id = 'singleton'
	`).Scan(
		&settings.ID,
		&settings.SyncSchedule,
		&settings.AnonymizationStrategy,
		&settings.Port,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("settings not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// Parse timestamps
	settings.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	settings.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &settings, nil
}

// UpdateSyncSchedule updates the sync schedule setting
func (s *SettingsStore) UpdateSyncSchedule(schedule string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	result, err := s.db.Exec(`
		UPDATE settings
		SET sync_schedule = ?, updated_at = ?
		WHERE id = 'singleton'
	`, schedule, now)

	if err != nil {
		return fmt.Errorf("failed to update sync schedule: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("settings not found")
	}

	return nil
}

// UpdateAnonymizationStrategy updates the anonymization strategy setting
func (s *SettingsStore) UpdateAnonymizationStrategy(strategy string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	result, err := s.db.Exec(`
		UPDATE settings
		SET anonymization_strategy = ?, updated_at = ?
		WHERE id = 'singleton'
	`, strategy, now)

	if err != nil {
		return fmt.Errorf("failed to update anonymization strategy: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("settings not found")
	}

	return nil
}

// UpdatePort updates the port setting
func (s *SettingsStore) UpdatePort(port int) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	result, err := s.db.Exec(`
		UPDATE settings
		SET port = ?, updated_at = ?
		WHERE id = 'singleton'
	`, port, now)

	if err != nil {
		return fmt.Errorf("failed to update port: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("settings not found")
	}

	return nil
}
