package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

type PluginStore struct {
	db *sql.DB
}

func NewPluginStore(db *sql.DB) *PluginStore {
	return &PluginStore{db: db}
}

func (s *PluginStore) Save(config *models.PluginConfig) error {
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now().UTC()
	}
	config.UpdatedAt = time.Now().UTC()

	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO plugin_configs (name, type, enabled, config, last_sync, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			type = excluded.type,
			enabled = excluded.enabled,
			config = excluded.config,
			last_sync = excluded.last_sync,
			updated_at = excluded.updated_at
	`, config.Name, config.Type, config.Enabled, string(configJSON), config.LastSync, config.CreatedAt, config.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to save plugin config: %w", err)
	}

	return nil
}

func (s *PluginStore) GetByName(name string) (*models.PluginConfig, error) {
	var config models.PluginConfig
	var configJSON string
	var lastSync sql.NullTime

	err := s.db.QueryRow(`
		SELECT name, type, enabled, config, last_sync, created_at, updated_at
		FROM plugin_configs WHERE name = ?
	`, name).Scan(&config.Name, &config.Type, &config.Enabled, &configJSON, &lastSync, &config.CreatedAt, &config.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin config: %w", err)
	}

	if err := json.Unmarshal([]byte(configJSON), &config.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if lastSync.Valid {
		config.LastSync = &lastSync.Time
	}

	return &config, nil
}

func (s *PluginStore) List(limit int) ([]*models.PluginConfig, error) {
	// Cap at DefaultQueryLimit to prevent memory exhaustion
	if limit <= 0 || limit > DefaultQueryLimit {
		limit = DefaultQueryLimit
	}

	rows, err := s.db.Query(`
		SELECT name, type, enabled, config, last_sync, created_at, updated_at
		FROM plugin_configs ORDER BY name LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugin configs: %w", err)
	}
	defer rows.Close()

	var configs []*models.PluginConfig
	for rows.Next() {
		var config models.PluginConfig
		var configJSON string
		var lastSync sql.NullTime

		err := rows.Scan(&config.Name, &config.Type, &config.Enabled, &configJSON, &lastSync, &config.CreatedAt, &config.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plugin config: %w", err)
		}

		if err := json.Unmarshal([]byte(configJSON), &config.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		if lastSync.Valid {
			config.LastSync = &lastSync.Time
		}

		configs = append(configs, &config)
	}

	return configs, nil
}

func (s *PluginStore) UpdateLastSync(name string, lastSync time.Time) error {
	_, err := s.db.Exec(`
		UPDATE plugin_configs SET last_sync = ?, updated_at = ? WHERE name = ?
	`, lastSync, time.Now().UTC(), name)

	if err != nil {
		return fmt.Errorf("failed to update last sync: %w", err)
	}

	return nil
}

func (s *PluginStore) SetEnabled(name string, enabled bool) error {
	_, err := s.db.Exec(`
		UPDATE plugin_configs SET enabled = ?, updated_at = ? WHERE name = ?
	`, enabled, time.Now().UTC(), name)

	if err != nil {
		return fmt.Errorf("failed to set enabled: %w", err)
	}

	return nil
}
