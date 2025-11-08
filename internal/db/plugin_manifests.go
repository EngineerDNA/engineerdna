package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

// PluginManifestStore handles database operations for plugin manifests
type PluginManifestStore struct {
	db *sql.DB
}

// NewPluginManifestStore creates a new plugin manifest store
func NewPluginManifestStore(db *sql.DB) *PluginManifestStore {
	return &PluginManifestStore{db: db}
}

// Upsert inserts or updates a plugin manifest
func (s *PluginManifestStore) Upsert(manifest *models.PluginManifest) error {
	now := time.Now().UTC()
	if manifest.CreatedAt.IsZero() {
		manifest.CreatedAt = now
	}
	manifest.UpdatedAt = now

	// Marshal JSON fields
	capabilitiesJSON, err := json.Marshal(manifest.Capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	metricsJSON, err := json.Marshal(manifest.ProvidesMetrics)
	if err != nil {
		return fmt.Errorf("failed to marshal provides_metrics: %w", err)
	}

	eventTypesJSON, err := json.Marshal(manifest.ProvidesEventTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal provides_event_types: %w", err)
	}

	widgetsJSON, err := json.Marshal(manifest.ProvidesWidgets)
	if err != nil {
		return fmt.Errorf("failed to marshal provides_widgets: %w", err)
	}

	correlationsJSON, err := json.Marshal(manifest.ProvidesCorrelations)
	if err != nil {
		return fmt.Errorf("failed to marshal provides_correlations: %w", err)
	}

	// Upsert using INSERT OR REPLACE
	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO plugin_manifests
		(plugin_name, version, type, capabilities, provides_metrics, provides_event_types, provides_widgets, provides_correlations, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, manifest.PluginName, manifest.Version, manifest.Type, string(capabilitiesJSON), string(metricsJSON), string(eventTypesJSON), string(widgetsJSON), string(correlationsJSON), manifest.CreatedAt, manifest.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert plugin manifest: %w", err)
	}

	return nil
}

// GetByName retrieves a plugin manifest by name
func (s *PluginManifestStore) GetByName(pluginName string) (*models.PluginManifest, error) {
	var manifest models.PluginManifest
	var capabilitiesJSON, metricsJSON, eventTypesJSON, widgetsJSON, correlationsJSON string

	err := s.db.QueryRow(`
		SELECT plugin_name, version, type, capabilities, provides_metrics, provides_event_types, provides_widgets, provides_correlations, created_at, updated_at
		FROM plugin_manifests WHERE plugin_name = ?
	`, pluginName).Scan(&manifest.PluginName, &manifest.Version, &manifest.Type, &capabilitiesJSON, &metricsJSON, &eventTypesJSON, &widgetsJSON, &correlationsJSON, &manifest.CreatedAt, &manifest.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin manifest: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(capabilitiesJSON), &manifest.Capabilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	if err := json.Unmarshal([]byte(metricsJSON), &manifest.ProvidesMetrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provides_metrics: %w", err)
	}

	if err := json.Unmarshal([]byte(eventTypesJSON), &manifest.ProvidesEventTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provides_event_types: %w", err)
	}

	if err := json.Unmarshal([]byte(widgetsJSON), &manifest.ProvidesWidgets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provides_widgets: %w", err)
	}

	if err := json.Unmarshal([]byte(correlationsJSON), &manifest.ProvidesCorrelations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provides_correlations: %w", err)
	}

	return &manifest, nil
}

// List retrieves all plugin manifests
func (s *PluginManifestStore) List() ([]*models.PluginManifest, error) {
	rows, err := s.db.Query(`
		SELECT plugin_name, version, type, capabilities, provides_metrics, provides_event_types, provides_widgets, provides_correlations, created_at, updated_at
		FROM plugin_manifests
		ORDER BY plugin_name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugin manifests: %w", err)
	}
	defer rows.Close()

	var manifests []*models.PluginManifest
	for rows.Next() {
		var manifest models.PluginManifest
		var capabilitiesJSON, metricsJSON, eventTypesJSON, widgetsJSON, correlationsJSON string

		err := rows.Scan(&manifest.PluginName, &manifest.Version, &manifest.Type, &capabilitiesJSON, &metricsJSON, &eventTypesJSON, &widgetsJSON, &correlationsJSON, &manifest.CreatedAt, &manifest.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plugin manifest: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(capabilitiesJSON), &manifest.Capabilities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
		}

		if err := json.Unmarshal([]byte(metricsJSON), &manifest.ProvidesMetrics); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provides_metrics: %w", err)
		}

		if err := json.Unmarshal([]byte(eventTypesJSON), &manifest.ProvidesEventTypes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provides_event_types: %w", err)
		}

		if err := json.Unmarshal([]byte(widgetsJSON), &manifest.ProvidesWidgets); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provides_widgets: %w", err)
		}

		if err := json.Unmarshal([]byte(correlationsJSON), &manifest.ProvidesCorrelations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provides_correlations: %w", err)
		}

		manifests = append(manifests, &manifest)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plugin manifests: %w", err)
	}

	return manifests, nil
}

// Delete removes a plugin manifest
func (s *PluginManifestStore) Delete(pluginName string) error {
	_, err := s.db.Exec(`DELETE FROM plugin_manifests WHERE plugin_name = ?`, pluginName)
	if err != nil {
		return fmt.Errorf("failed to delete plugin manifest: %w", err)
	}
	return nil
}
