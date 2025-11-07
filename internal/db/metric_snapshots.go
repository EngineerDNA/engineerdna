package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

// Metric Snapshots

// CreateMetricSnapshot creates or replaces a metric snapshot in the database.
func (s *DashboardStore) CreateMetricSnapshot(snapshot *models.MetricSnapshot) error {
	snapshot.ComputedAt = time.Now().UTC()

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO metric_snapshots (
			metric_name, entity_type, entity_id, period_start, period_end,
			value, metadata, computed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, snapshot.MetricName, snapshot.EntityType, snapshot.EntityID, snapshot.PeriodStart,
		snapshot.PeriodEnd, snapshot.Value, snapshot.Metadata, snapshot.ComputedAt.Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("failed to create metric snapshot: %w", err)
	}

	return nil
}

// CreateMetricSnapshotsBatch efficiently inserts multiple metric snapshots in a single transaction.
func (s *DashboardStore) CreateMetricSnapshotsBatch(snapshots []*models.MetricSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO metric_snapshots (
			metric_name, entity_type, entity_id, period_start, period_end,
			value, metadata, computed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	computedAt := time.Now().UTC()
	for _, snapshot := range snapshots {
		snapshot.ComputedAt = computedAt
		_, err = stmt.Exec(snapshot.MetricName, snapshot.EntityType, snapshot.EntityID,
			snapshot.PeriodStart, snapshot.PeriodEnd, snapshot.Value, snapshot.Metadata,
			snapshot.ComputedAt.Format(time.RFC3339))
		if err != nil {
			return fmt.Errorf("failed to insert snapshot: %w", err)
		}
	}

	return tx.Commit()
}

// GetMetricSnapshot retrieves a single metric snapshot by its primary key components.
func (s *DashboardStore) GetMetricSnapshot(metricName, entityType, entityID, periodStart string) (*models.MetricSnapshot, error) {
	var snapshot models.MetricSnapshot
	var entityIDNull, metadata sql.NullString
	var computedAtStr string

	err := s.db.QueryRow(`
		SELECT metric_name, entity_type, entity_id, period_start, period_end,
		       value, metadata, computed_at
		FROM metric_snapshots
		WHERE metric_name = ? AND entity_type = ? AND entity_id = ? AND period_start = ?
	`, metricName, entityType, entityID, periodStart).Scan(&snapshot.MetricName, &snapshot.EntityType,
		&entityIDNull, &snapshot.PeriodStart, &snapshot.PeriodEnd, &snapshot.Value, &metadata,
		&computedAtStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get metric snapshot: %w", err)
	}

	if entityIDNull.Valid {
		snapshot.EntityID = entityIDNull.String
	}
	if metadata.Valid {
		snapshot.Metadata = metadata.String
	}

	// Parse timestamp
	snapshot.ComputedAt, err = time.Parse(time.RFC3339, computedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse computed_at: %w", err)
	}

	return &snapshot, nil
}

// GetMetricSnapshots retrieves metric snapshots within a date range, ordered by value descending.
func (s *DashboardStore) GetMetricSnapshots(metricName, entityType string, startDate, endDate string, limit int) ([]*models.MetricSnapshot, error) {
	if limit <= 0 {
		limit = DefaultQueryLimit
	}
	if limit > MaxQueryLimit {
		limit = MaxQueryLimit
	}

	// Use JOIN to get entity names in single query - eliminates N+1
	query := `
		SELECT
			ms.metric_name, ms.entity_type, ms.entity_id, ms.period_start, ms.period_end,
			ms.value, ms.metadata, ms.computed_at,
			CASE
				WHEN ms.entity_type = 'team' THEN t.name
				WHEN ms.entity_type = 'engineer' THEN e.canonical_name
				ELSE ms.entity_id
			END as entity_name
		FROM metric_snapshots ms
		LEFT JOIN teams t ON ms.entity_type = 'team' AND ms.entity_id = t.id
		LEFT JOIN engineers e ON ms.entity_type = 'engineer' AND ms.entity_id = e.id
		WHERE ms.metric_name = ? AND ms.entity_type = ?
		  AND ms.period_start >= ? AND ms.period_start <= ?
		ORDER BY ms.value DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, metricName, entityType, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query metric snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []*models.MetricSnapshot
	for rows.Next() {
		var snapshot models.MetricSnapshot
		var entityIDNull, metadata, entityName sql.NullString
		var computedAtStr string

		err := rows.Scan(&snapshot.MetricName, &snapshot.EntityType, &entityIDNull,
			&snapshot.PeriodStart, &snapshot.PeriodEnd, &snapshot.Value, &metadata,
			&computedAtStr, &entityName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric snapshot: %w", err)
		}

		if entityIDNull.Valid {
			snapshot.EntityID = entityIDNull.String
		}
		if metadata.Valid {
			snapshot.Metadata = metadata.String
		}
		if entityName.Valid {
			snapshot.EntityName = entityName.String
		}

		// Parse timestamp
		snapshot.ComputedAt, err = time.Parse(time.RFC3339, computedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse computed_at: %w", err)
		}

		snapshots = append(snapshots, &snapshot)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metric snapshots: %w", err)
	}

	return snapshots, nil
}

// GetMetricSnapshotsInRange retrieves metric snapshots within a date range for a specific entity.
// This method is optimized for timeseries queries, fetching all periods in a single query
// instead of N individual queries (eliminates N+1 problem).
// Results are ordered by period_start for consistent timeseries display.
func (s *DashboardStore) GetMetricSnapshotsInRange(metricName, entityType, entityID, startDate, endDate string, limit int) ([]*models.MetricSnapshot, error) {
	if limit <= 0 {
		limit = DefaultQueryLimit
	}
	if limit > MaxQueryLimit {
		limit = MaxQueryLimit
	}

	query := `
		SELECT metric_name, entity_type, entity_id, period_start, period_end,
		       value, metadata, computed_at
		FROM metric_snapshots
		WHERE metric_name = ? AND entity_type = ? AND entity_id = ?
		  AND period_start >= ? AND period_start <= ?
		ORDER BY period_start ASC
		LIMIT ?
	`

	rows, err := s.db.Query(query, metricName, entityType, entityID, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query metric snapshots in range: %w", err)
	}
	defer rows.Close()

	var snapshots []*models.MetricSnapshot
	for rows.Next() {
		var snapshot models.MetricSnapshot
		var entityIDNull, metadata sql.NullString
		var computedAtStr string

		err := rows.Scan(&snapshot.MetricName, &snapshot.EntityType, &entityIDNull,
			&snapshot.PeriodStart, &snapshot.PeriodEnd, &snapshot.Value, &metadata,
			&computedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric snapshot: %w", err)
		}

		if entityIDNull.Valid {
			snapshot.EntityID = entityIDNull.String
		}
		if metadata.Valid {
			snapshot.Metadata = metadata.String
		}

		// Parse timestamp
		snapshot.ComputedAt, err = time.Parse(time.RFC3339, computedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse computed_at: %w", err)
		}

		snapshots = append(snapshots, &snapshot)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metric snapshots: %w", err)
	}

	return snapshots, nil
}
