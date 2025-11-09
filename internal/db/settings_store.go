package db

import (
	"database/sql"
	"encoding/json"
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
	var primaryDashboardID, favoriteDashboardsJSON sql.NullString
	var onboardingCompleted int // SQLite boolean (0/1)

	err := s.db.QueryRow(`
		SELECT id, sync_schedule, anonymization_strategy, port, role,
		       primary_dashboard_id, favorite_dashboards, onboarding_completed,
		       created_at, updated_at
		FROM settings
		WHERE id = 'singleton'
	`).Scan(
		&settings.ID,
		&settings.SyncSchedule,
		&settings.AnonymizationStrategy,
		&settings.Port,
		&settings.Role,
		&primaryDashboardID,
		&favoriteDashboardsJSON,
		&onboardingCompleted,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("settings not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// Handle nullable primary_dashboard_id
	if primaryDashboardID.Valid {
		settings.PrimaryDashboardID = &primaryDashboardID.String
	}

	// Parse favorite_dashboards JSON array
	if favoriteDashboardsJSON.Valid && favoriteDashboardsJSON.String != "" {
		var favorites []string
		if err := json.Unmarshal([]byte(favoriteDashboardsJSON.String), &favorites); err != nil {
			return nil, fmt.Errorf("failed to parse favorite_dashboards: %w", err)
		}
		settings.FavoriteDashboards = favorites
	}

	// Convert SQLite boolean
	settings.OnboardingCompleted = onboardingCompleted == 1

	// Parse timestamps from SQLite DATETIME format
	settings.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	settings.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", updatedAtStr)
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

// UpdateRole updates the role setting
func (s *SettingsStore) UpdateRole(role string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	result, err := s.db.Exec(`
		UPDATE settings
		SET role = ?, updated_at = ?
		WHERE id = 'singleton'
	`, role, now)

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
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

// UpdatePrimaryDashboard updates the primary dashboard ID
func (s *SettingsStore) UpdatePrimaryDashboard(dashboardID *string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	result, err := s.db.Exec(`
		UPDATE settings
		SET primary_dashboard_id = ?, updated_at = ?
		WHERE id = 'singleton'
	`, dashboardID, now)

	if err != nil {
		return fmt.Errorf("failed to update primary dashboard: %w", err)
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

// UpdateFavoriteDashboards updates the favorite dashboards list
func (s *SettingsStore) UpdateFavoriteDashboards(dashboardIDs []string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	// Marshal to JSON
	favoritesJSON, err := json.Marshal(dashboardIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal favorite dashboards: %w", err)
	}

	result, err := s.db.Exec(`
		UPDATE settings
		SET favorite_dashboards = ?, updated_at = ?
		WHERE id = 'singleton'
	`, string(favoritesJSON), now)

	if err != nil {
		return fmt.Errorf("failed to update favorite dashboards: %w", err)
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

// UpdateOnboardingCompleted marks onboarding as completed
func (s *SettingsStore) UpdateOnboardingCompleted(completed bool) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	// Convert bool to SQLite integer
	onboardingInt := 0
	if completed {
		onboardingInt = 1
	}

	result, err := s.db.Exec(`
		UPDATE settings
		SET onboarding_completed = ?, updated_at = ?
		WHERE id = 'singleton'
	`, onboardingInt, now)

	if err != nil {
		return fmt.Errorf("failed to update onboarding completed: %w", err)
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
