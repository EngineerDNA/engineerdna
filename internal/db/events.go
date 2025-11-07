package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

type EventStore struct {
	db *sql.DB
}

func NewEventStore(db *sql.DB) *EventStore {
	return &EventStore{db: db}
}

func (s *EventStore) Create(event *models.Event) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	event.UpdatedAt = time.Now().UTC()

	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data, anonymized, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.Type, event.Source, event.SourceID, event.Timestamp, event.Actor, event.EngineerID, string(dataJSON), event.Anonymized, event.CreatedAt, event.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// CreateBatch inserts multiple events atomically within a transaction
func (s *EventStore) CreateBatch(events []*models.Event) error {
	if len(events) == 0 {
		return nil
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data, anonymized, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		if event.ID == "" {
			event.ID = uuid.New().String()
		}
		if event.CreatedAt.IsZero() {
			event.CreatedAt = time.Now().UTC()
		}
		event.UpdatedAt = time.Now().UTC()

		dataJSON, err := json.Marshal(event.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}

		_, err = stmt.Exec(event.ID, event.Type, event.Source, event.SourceID, event.Timestamp, event.Actor, event.EngineerID, string(dataJSON), event.Anonymized, event.CreatedAt, event.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *EventStore) GetByID(id string) (*models.Event, error) {
	var event models.Event
	var dataJSON string

	err := s.db.QueryRow(`
		SELECT id, type, source, source_id, timestamp, actor, engineer_id, data, anonymized, created_at, updated_at
		FROM events WHERE id = ?
	`, id).Scan(&event.ID, &event.Type, &event.Source, &event.SourceID, &event.Timestamp, &event.Actor, &event.EngineerID, &dataJSON, &event.Anonymized, &event.CreatedAt, &event.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if err := json.Unmarshal([]byte(dataJSON), &event.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	return &event, nil
}

func (s *EventStore) List(filters map[string]interface{}, limit, offset int) ([]*models.Event, error) {
	query := "SELECT id, type, source, source_id, timestamp, actor, engineer_id, data, anonymized, created_at, updated_at FROM events WHERE 1=1"
	args := []interface{}{}

	if since, ok := filters["since"].(time.Time); ok {
		query += " AND timestamp >= ?"
		args = append(args, since)
	}
	if eventType, ok := filters["type"].(string); ok {
		query += " AND type = ?"
		args = append(args, eventType)
	}
	if source, ok := filters["source"].(string); ok {
		query += " AND source = ?"
		args = append(args, source)
	}
	if actor, ok := filters["actor"].(string); ok {
		query += " AND actor = ?"
		args = append(args, actor)
	}

	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var dataJSON string

		err := rows.Scan(&event.ID, &event.Type, &event.Source, &event.SourceID, &event.Timestamp, &event.Actor, &event.EngineerID, &dataJSON, &event.Anonymized, &event.CreatedAt, &event.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if err := json.Unmarshal([]byte(dataJSON), &event.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		events = append(events, &event)
	}

	return events, nil
}

func (s *EventStore) Count(filters map[string]interface{}) (int, error) {
	query := "SELECT COUNT(*) FROM events WHERE 1=1"
	args := []interface{}{}

	if since, ok := filters["since"].(time.Time); ok {
		query += " AND timestamp >= ?"
		args = append(args, since)
	}
	if eventType, ok := filters["type"].(string); ok {
		query += " AND type = ?"
		args = append(args, eventType)
	}
	if source, ok := filters["source"].(string); ok {
		query += " AND source = ?"
		args = append(args, source)
	}
	if actor, ok := filters["actor"].(string); ok {
		query += " AND actor = ?"
		args = append(args, actor)
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}

func (s *EventStore) Update(event *models.Event) error {
	event.UpdatedAt = time.Now().UTC()

	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	_, err = s.db.Exec(`
		UPDATE events SET type = ?, source = ?, source_id = ?, timestamp = ?, actor = ?, engineer_id = ?, data = ?, anonymized = ?, updated_at = ?
		WHERE id = ?
	`, event.Type, event.Source, event.SourceID, event.Timestamp, event.Actor, event.EngineerID, string(dataJSON), event.Anonymized, event.UpdatedAt, event.ID)

	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	return nil
}

// GetEventsByEngineerAndTimeRange retrieves events for a specific engineer within a time range
// limit: max number of events to return (default 1000, max 10000)
func (s *EventStore) GetEventsByEngineerAndTimeRange(engineerID string, start, end time.Time, limit int) ([]*models.Event, error) {
	// Apply default and max limits for safety
	if limit <= 0 {
		limit = 1000
	}
	if limit > 10000 {
		limit = 10000
	}

	rows, err := s.db.Query(`
		SELECT id, type, source, source_id, timestamp, actor, engineer_id, data, anonymized, created_at, updated_at
		FROM events
		WHERE engineer_id = ? AND timestamp >= ? AND timestamp < ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, engineerID, start, end, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var dataJSON string

		err := rows.Scan(&event.ID, &event.Type, &event.Source, &event.SourceID, &event.Timestamp, &event.Actor, &event.EngineerID, &dataJSON, &event.Anonymized, &event.CreatedAt, &event.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if err := json.Unmarshal([]byte(dataJSON), &event.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// TodayMetrics contains aggregated metrics for today's engineering activity
type TodayMetrics struct {
	TotalEvents     int
	ActiveEngineers int
	PRsCreated      int
	PRsMerged       int
	IssuesClosed    int
	CommitsToday    int
	AvgCycleTime    float64
	AvgReviewTime   float64
}

// GetTodayMetricsAggregated retrieves aggregated metrics for today using SQL aggregation
func (s *EventStore) GetTodayMetricsAggregated(startOfDay time.Time) (*TodayMetrics, error) {
	metrics := &TodayMetrics{}

	// Count total events
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM events WHERE timestamp >= ?
	`, startOfDay).Scan(&metrics.TotalEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to count total events: %w", err)
	}

	// Count unique engineers
	err = s.db.QueryRow(`
		SELECT COUNT(DISTINCT engineer_id) FROM events
		WHERE timestamp >= ? AND engineer_id != ''
	`, startOfDay).Scan(&metrics.ActiveEngineers)
	if err != nil {
		return nil, fmt.Errorf("failed to count unique engineers: %w", err)
	}

	// Count PRs created, merged, issues closed, and commits using JSON extraction
	err = s.db.QueryRow(`
		SELECT
			COUNT(CASE WHEN type = 'pull_request' AND (json_extract(data, '$.status') = 'open' OR json_extract(data, '$.status') = 'created') THEN 1 END) as prs_created,
			COUNT(CASE WHEN type = 'pull_request' AND (json_extract(data, '$.status') = 'merged' OR json_extract(data, '$.status') = 'closed') THEN 1 END) as prs_merged,
			COUNT(CASE WHEN type = 'issue' AND (json_extract(data, '$.status') = 'closed' OR json_extract(data, '$.status') = 'completed') THEN 1 END) as issues_closed,
			COUNT(CASE WHEN type = 'commit' THEN 1 END) as commits
		FROM events WHERE timestamp >= ?
	`, startOfDay).Scan(&metrics.PRsCreated, &metrics.PRsMerged, &metrics.IssuesClosed, &metrics.CommitsToday)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate event counts: %w", err)
	}

	// Calculate average cycle time (PR created to merged)
	// This is more complex and requires loading some data, but with LIMIT
	// For V1, we'll use a simplified approach
	filters := map[string]interface{}{
		"since": startOfDay,
		"type":  "pull_request",
	}
	prEvents, err := s.List(filters, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR events for cycle time: %w", err)
	}

	// Group by PR ID and calculate cycle times
	prCreatedTimes := make(map[string]time.Time)
	prMergedTimes := make(map[string]time.Time)
	for _, event := range prEvents {
		var prID string
		if id, ok := event.Data["pr_id"].(string); ok && id != "" {
			prID = id
		} else {
			prID = event.SourceID
		}

		if status, ok := event.Data["status"].(string); ok {
			if status == "open" || status == "created" {
				prCreatedTimes[prID] = event.Timestamp
			} else if status == "merged" || status == "closed" {
				prMergedTimes[prID] = event.Timestamp
			}
		}
	}

	var totalCycleTime float64
	var cycleTimeCount int
	for prID, createdTime := range prCreatedTimes {
		if mergedTime, exists := prMergedTimes[prID]; exists {
			cycleHours := mergedTime.Sub(createdTime).Hours()
			if cycleHours >= 0 {
				totalCycleTime += cycleHours
				cycleTimeCount++
			}
		}
	}

	if cycleTimeCount > 0 {
		metrics.AvgCycleTime = totalCycleTime / float64(cycleTimeCount)
		// V1 Implementation: Uses 60% of cycle time as proxy for review time
		metrics.AvgReviewTime = metrics.AvgCycleTime * 0.6
	}

	return metrics, nil
}

// ThroughputDay contains throughput metrics for a single day
type ThroughputDay struct {
	Date         string
	PullRequests int
	Issues       int
}

// GetThroughputByDay retrieves daily throughput metrics using SQL aggregation
func (s *EventStore) GetThroughputByDay(startDate, endDate time.Time) ([]ThroughputDay, error) {
	rows, err := s.db.Query(`
		SELECT
			substr(timestamp, 1, 10) as day,
			COUNT(CASE WHEN type = 'pull_request' THEN 1 END) as pr_count,
			COUNT(CASE WHEN type = 'issue' THEN 1 END) as issue_count
		FROM events
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY substr(timestamp, 1, 10)
		ORDER BY day ASC
	`, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query throughput by day: %w", err)
	}
	defer rows.Close()

	var results []ThroughputDay
	for rows.Next() {
		var day ThroughputDay
		err := rows.Scan(&day.Date, &day.PullRequests, &day.Issues)
		if err != nil {
			return nil, fmt.Errorf("failed to scan throughput day: %w", err)
		}
		results = append(results, day)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating throughput days: %w", err)
	}

	return results, nil
}

// CountOpenPRs counts the number of currently open pull requests
func (s *EventStore) CountOpenPRs() (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(DISTINCT source_id)
		FROM events
		WHERE type = 'pull_request'
		  AND (
			json_extract(data, '$.state') = 'open'
			OR json_extract(data, '$.state') = 'review_requested'
		  )
		  AND source_id NOT IN (
			SELECT source_id FROM events
			WHERE type = 'pull_request'
			  AND json_extract(data, '$.state') IN ('merged', 'closed')
		  )
		LIMIT 1000
	`).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count open PRs: %w", err)
	}

	return count, nil
}
