package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// CorrelationStore handles database operations for correlations
type CorrelationStore struct {
	db *sql.DB
}

// NewCorrelationStore creates a new correlation store
func NewCorrelationStore(db *sql.DB) *CorrelationStore {
	return &CorrelationStore{db: db}
}

// CreateCorrelation inserts a new correlation definition
func (s *CorrelationStore) CreateCorrelation(corr *models.Correlation) error {
	if corr.ID == "" {
		corr.ID = uuid.New().String()
	}
	if corr.CreatedAt.IsZero() {
		corr.CreatedAt = time.Now().UTC()
	}

	definitionJSON, err := json.Marshal(corr.Definition)
	if err != nil {
		return fmt.Errorf("failed to marshal definition: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO correlations (id, name, plugin, definition, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, corr.ID, corr.Name, corr.Plugin, string(definitionJSON), corr.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create correlation: %w", err)
	}

	return nil
}

// GetCorrelationByID retrieves a correlation by ID
func (s *CorrelationStore) GetCorrelationByID(id string) (*models.Correlation, error) {
	var corr models.Correlation
	var definitionJSON string

	err := s.db.QueryRow(`
		SELECT id, name, plugin, definition, created_at
		FROM correlations WHERE id = ?
	`, id).Scan(&corr.ID, &corr.Name, &corr.Plugin, &definitionJSON, &corr.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get correlation: %w", err)
	}

	if err := json.Unmarshal([]byte(definitionJSON), &corr.Definition); err != nil {
		return nil, fmt.Errorf("failed to unmarshal definition: %w", err)
	}

	return &corr, nil
}

// GetCorrelationByName retrieves a correlation by name
func (s *CorrelationStore) GetCorrelationByName(name string) (*models.Correlation, error) {
	var corr models.Correlation
	var definitionJSON string

	err := s.db.QueryRow(`
		SELECT id, name, plugin, definition, created_at
		FROM correlations WHERE name = ?
	`, name).Scan(&corr.ID, &corr.Name, &corr.Plugin, &definitionJSON, &corr.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get correlation: %w", err)
	}

	if err := json.Unmarshal([]byte(definitionJSON), &corr.Definition); err != nil {
		return nil, fmt.Errorf("failed to unmarshal definition: %w", err)
	}

	return &corr, nil
}

// ListCorrelations retrieves all correlations
func (s *CorrelationStore) ListCorrelations() ([]*models.Correlation, error) {
	rows, err := s.db.Query(`
		SELECT id, name, plugin, definition, created_at
		FROM correlations
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list correlations: %w", err)
	}
	defer rows.Close()

	var correlations []*models.Correlation
	for rows.Next() {
		var corr models.Correlation
		var definitionJSON string

		err := rows.Scan(&corr.ID, &corr.Name, &corr.Plugin, &definitionJSON, &corr.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan correlation: %w", err)
		}

		if err := json.Unmarshal([]byte(definitionJSON), &corr.Definition); err != nil {
			return nil, fmt.Errorf("failed to unmarshal definition: %w", err)
		}

		correlations = append(correlations, &corr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating correlations: %w", err)
	}

	return correlations, nil
}

// SaveCorrelationValue inserts a correlation value
func (s *CorrelationStore) SaveCorrelationValue(value *models.CorrelationValue) error {
	if value.ID == "" {
		value.ID = uuid.New().String()
	}
	if value.CreatedAt.IsZero() {
		value.CreatedAt = time.Now().UTC()
	}

	var breakdownJSON string
	if value.Breakdown != nil {
		data, err := json.Marshal(value.Breakdown)
		if err != nil {
			return fmt.Errorf("failed to marshal breakdown: %w", err)
		}
		breakdownJSON = string(data)
	}

	_, err := s.db.Exec(`
		INSERT INTO correlation_values (id, correlation_id, timestamp, time_window, value, breakdown, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, value.ID, value.CorrelationID, value.Timestamp, value.TimeWindow, value.Value, breakdownJSON, value.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to save correlation value: %w", err)
	}

	return nil
}

// GetCorrelationValues retrieves correlation values within a time range
func (s *CorrelationStore) GetCorrelationValues(correlationID string, start, end time.Time, limit int) ([]*models.CorrelationValue, error) {
	if limit <= 0 {
		limit = MaxQueryLimit
	}
	if limit > MaxEventQueryLimit {
		limit = MaxEventQueryLimit
	}

	rows, err := s.db.Query(`
		SELECT id, correlation_id, timestamp, time_window, value, breakdown, created_at
		FROM correlation_values
		WHERE correlation_id = ? AND timestamp >= ? AND timestamp < ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, correlationID, start, end, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query correlation values: %w", err)
	}
	defer rows.Close()

	var values []*models.CorrelationValue
	for rows.Next() {
		var value models.CorrelationValue
		var breakdownJSON sql.NullString

		err := rows.Scan(&value.ID, &value.CorrelationID, &value.Timestamp, &value.TimeWindow, &value.Value, &breakdownJSON, &value.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan correlation value: %w", err)
		}

		if breakdownJSON.Valid && breakdownJSON.String != "" {
			if err := json.Unmarshal([]byte(breakdownJSON.String), &value.Breakdown); err != nil {
				return nil, fmt.Errorf("failed to unmarshal breakdown: %w", err)
			}
		}

		values = append(values, &value)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating correlation values: %w", err)
	}

	return values, nil
}
