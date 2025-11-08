package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

// EventQueryFilters contains filters for event queries
type EventQueryFilters struct {
	Since          *time.Time             // Events after this time
	Until          *time.Time             // Events before this time
	EngineerID     string                 // Filter by engineer
	Actor          string                 // Filter by actor
	NormalizedData map[string]interface{} // Filter by normalized data fields
}

// Pagination contains pagination parameters
type Pagination struct {
	Limit  int // Max results (default 100, max 10000)
	Offset int // Skip this many results
}

// GetEventsByNormalizedType retrieves events by normalized type
func (s *EventStore) GetEventsByNormalizedType(
	normalizedType string,
	filters *EventQueryFilters,
	pagination *Pagination,
) ([]*models.Event, error) {
	// Validate normalized type
	if normalizedType == "" {
		return nil, fmt.Errorf("normalized type cannot be empty")
	}

	// Apply pagination defaults and limits
	limit := 100
	offset := 0
	if pagination != nil {
		if pagination.Limit > 0 {
			limit = pagination.Limit
		}
		if pagination.Limit > 10000 {
			limit = 10000
		}
		offset = pagination.Offset
	}

	// Build query
	query := `
		SELECT id, type, source, source_id, timestamp, actor, engineer_id, data,
		       anonymized, normalized_type, normalized_data, created_at, updated_at
		FROM events
		WHERE normalized_type = ?
	`
	args := []interface{}{normalizedType}

	// Add filters
	if filters != nil {
		if filters.Since != nil {
			query += " AND timestamp >= ?"
			args = append(args, *filters.Since)
		}
		if filters.Until != nil {
			query += " AND timestamp < ?"
			args = append(args, *filters.Until)
		}
		if filters.EngineerID != "" {
			query += " AND engineer_id = ?"
			args = append(args, filters.EngineerID)
		}
		if filters.Actor != "" {
			query += " AND actor = ?"
			args = append(args, filters.Actor)
		}

		// Normalized data filters using json_extract
		if len(filters.NormalizedData) > 0 {
			for field, value := range filters.NormalizedData {
				// Validate field name (prevent injection)
				if !isValidFieldName(field) {
					return nil, fmt.Errorf("invalid field name: %s", field)
				}
				query += fmt.Sprintf(" AND json_extract(normalized_data, '$.%s') = ?", field)
				args = append(args, value)
			}
		}
	}

	// Add ordering and pagination
	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// Execute query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	// Scan results
	var events []*models.Event
	for rows.Next() {
		event, err := s.scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// CountEventsByNormalizedType counts events by normalized type
func (s *EventStore) CountEventsByNormalizedType(
	normalizedType string,
	filters *EventQueryFilters,
) (int, error) {
	// Validate normalized type
	if normalizedType == "" {
		return 0, fmt.Errorf("normalized type cannot be empty")
	}

	// Build query
	query := "SELECT COUNT(*) FROM events WHERE normalized_type = ?"
	args := []interface{}{normalizedType}

	// Add filters
	if filters != nil {
		if filters.Since != nil {
			query += " AND timestamp >= ?"
			args = append(args, *filters.Since)
		}
		if filters.Until != nil {
			query += " AND timestamp < ?"
			args = append(args, *filters.Until)
		}
		if filters.EngineerID != "" {
			query += " AND engineer_id = ?"
			args = append(args, filters.EngineerID)
		}
		if filters.Actor != "" {
			query += " AND actor = ?"
			args = append(args, filters.Actor)
		}

		// Normalized data filters
		if len(filters.NormalizedData) > 0 {
			for field, value := range filters.NormalizedData {
				if !isValidFieldName(field) {
					return 0, fmt.Errorf("invalid field name: %s", field)
				}
				query += fmt.Sprintf(" AND json_extract(normalized_data, '$.%s') = ?", field)
				args = append(args, value)
			}
		}
	}

	// Execute query
	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}

// GetNormalizedTypeSummary returns summary statistics for a normalized type
func (s *EventStore) GetNormalizedTypeSummary(normalizedType string) (map[string]interface{}, error) {
	if normalizedType == "" {
		return nil, fmt.Errorf("normalized type cannot be empty")
	}

	summary := make(map[string]interface{})

	// Total count
	var totalCount int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM events WHERE normalized_type = ?
	`, normalizedType).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count events: %w", err)
	}
	summary["total_count"] = totalCount

	// Count by source
	rows, err := s.db.Query(`
		SELECT source, COUNT(*) as count
		FROM events
		WHERE normalized_type = ?
		GROUP BY source
		ORDER BY count DESC
	`, normalizedType)
	if err != nil {
		return nil, fmt.Errorf("failed to query by source: %w", err)
	}
	defer rows.Close()

	sourceCount := make(map[string]int)
	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			return nil, fmt.Errorf("failed to scan source count: %w", err)
		}
		sourceCount[source] = count
	}
	summary["by_source"] = sourceCount

	// Date range
	var firstTimestamp, lastTimestamp sql.NullTime
	err = s.db.QueryRow(`
		SELECT MIN(timestamp), MAX(timestamp)
		FROM events
		WHERE normalized_type = ?
	`, normalizedType).Scan(&firstTimestamp, &lastTimestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to query date range: %w", err)
	}
	if firstTimestamp.Valid {
		summary["first_event"] = firstTimestamp.Time
	}
	if lastTimestamp.Valid {
		summary["last_event"] = lastTimestamp.Time
	}

	return summary, nil
}

// scanEvent scans a single event from a row
func (s *EventStore) scanEvent(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Event, error) {
	var event models.Event
	var dataJSON string

	err := scanner.Scan(
		&event.ID,
		&event.Type,
		&event.Source,
		&event.SourceID,
		&event.Timestamp,
		&event.Actor,
		&event.EngineerID,
		&dataJSON,
		&event.Anonymized,
		&event.NormalizedType,
		&event.NormalizedData,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal data JSON
	if err := json.Unmarshal([]byte(dataJSON), &event.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	return &event, nil
}

// isValidFieldName validates a field name to prevent SQL injection
func isValidFieldName(name string) bool {
	// Allow only alphanumeric and underscore
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}
	return len(name) > 0 && len(name) <= 50
}
