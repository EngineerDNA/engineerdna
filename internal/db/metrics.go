package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// MetricStore handles database operations for metrics
type MetricStore struct {
	db *sql.DB
}

// NewMetricStore creates a new metric store
func NewMetricStore(db *sql.DB) *MetricStore {
	return &MetricStore{db: db}
}

// Create inserts a new metric value
func (s *MetricStore) Create(metric *models.MetricValue) error {
	if metric.ID == "" {
		metric.ID = uuid.New().String()
	}
	if metric.CreatedAt.IsZero() {
		metric.CreatedAt = time.Now().UTC()
	}

	var dimensionsJSON string
	if metric.Dimensions != nil {
		data, err := json.Marshal(metric.Dimensions)
		if err != nil {
			return fmt.Errorf("failed to marshal dimensions: %w", err)
		}
		dimensionsJSON = string(data)
	}

	_, err := s.db.Exec(`
		INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, metric.ID, metric.MetricName, metric.Source, metric.Timestamp, metric.Granularity, metric.Value, metric.Unit, dimensionsJSON, metric.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create metric: %w", err)
	}

	return nil
}

// GetByID retrieves a metric value by ID
func (s *MetricStore) GetByID(id string) (*models.MetricValue, error) {
	var metric models.MetricValue
	var dimensionsJSON sql.NullString

	err := s.db.QueryRow(`
		SELECT id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at
		FROM metric_values WHERE id = ?
	`, id).Scan(&metric.ID, &metric.MetricName, &metric.Source, &metric.Timestamp, &metric.Granularity, &metric.Value, &metric.Unit, &dimensionsJSON, &metric.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	if dimensionsJSON.Valid && dimensionsJSON.String != "" {
		if err := json.Unmarshal([]byte(dimensionsJSON.String), &metric.Dimensions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal dimensions: %w", err)
		}
	}

	return &metric, nil
}

// List retrieves metrics with filters
func (s *MetricStore) List(filters map[string]interface{}, limit, offset int) ([]*models.MetricValue, error) {
	query := "SELECT id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at FROM metric_values WHERE 1=1"
	args := []interface{}{}

	if metricName, ok := filters["metric_name"].(string); ok {
		query += " AND metric_name = ?"
		args = append(args, metricName)
	}
	if source, ok := filters["source"].(string); ok {
		query += " AND source = ?"
		args = append(args, source)
	}
	if since, ok := filters["since"].(time.Time); ok {
		query += " AND timestamp >= ?"
		args = append(args, since)
	}
	if until, ok := filters["until"].(time.Time); ok {
		query += " AND timestamp < ?"
		args = append(args, until)
	}
	if granularity, ok := filters["granularity"].(string); ok {
		query += " AND granularity = ?"
		args = append(args, granularity)
	}

	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*models.MetricValue
	for rows.Next() {
		var metric models.MetricValue
		var dimensionsJSON sql.NullString

		err := rows.Scan(&metric.ID, &metric.MetricName, &metric.Source, &metric.Timestamp, &metric.Granularity, &metric.Value, &metric.Unit, &dimensionsJSON, &metric.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}

		if dimensionsJSON.Valid && dimensionsJSON.String != "" {
			if err := json.Unmarshal([]byte(dimensionsJSON.String), &metric.Dimensions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal dimensions: %w", err)
			}
		}

		metrics = append(metrics, &metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metrics: %w", err)
	}

	return metrics, nil
}

// GetByMetricName retrieves metrics for a specific metric name within a time range
func (s *MetricStore) GetByMetricName(name string, start, end time.Time, limit int) ([]*models.MetricValue, error) {
	if limit <= 0 {
		limit = 1000
	}
	if limit > 10000 {
		limit = 10000
	}

	rows, err := s.db.Query(`
		SELECT id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at
		FROM metric_values
		WHERE metric_name = ? AND timestamp >= ? AND timestamp < ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, name, start, end, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*models.MetricValue
	for rows.Next() {
		var metric models.MetricValue
		var dimensionsJSON sql.NullString

		err := rows.Scan(&metric.ID, &metric.MetricName, &metric.Source, &metric.Timestamp, &metric.Granularity, &metric.Value, &metric.Unit, &dimensionsJSON, &metric.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}

		if dimensionsJSON.Valid && dimensionsJSON.String != "" {
			if err := json.Unmarshal([]byte(dimensionsJSON.String), &metric.Dimensions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal dimensions: %w", err)
			}
		}

		metrics = append(metrics, &metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metrics: %w", err)
	}

	return metrics, nil
}
