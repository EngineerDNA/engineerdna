package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/config"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// EventNormalizer is an interface for normalizing events
type EventNormalizer interface {
	NormalizeEvent(event *models.Event) error
}

type EventStore struct {
	db         *sql.DB
	normalizer EventNormalizer
}

func NewEventStore(db *sql.DB) *EventStore {
	return &EventStore{db: db}
}

// SetNormalizer sets the event normalizer for this store
func (s *EventStore) SetNormalizer(normalizer EventNormalizer) {
	s.normalizer = normalizer
}

func (s *EventStore) Create(event *models.Event) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	event.UpdatedAt = time.Now().UTC()

	// Apply normalization if normalizer is set
	if s.normalizer != nil {
		if err := s.normalizer.NormalizeEvent(event); err != nil {
			return fmt.Errorf("failed to normalize event: %w", err)
		}
	}

	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data, anonymized, normalized_type, normalized_data, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.Type, event.Source, event.SourceID, event.Timestamp, event.Actor, event.EngineerID, string(dataJSON), event.Anonymized, event.NormalizedType, event.NormalizedData, event.CreatedAt, event.UpdatedAt)

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

	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultTimeout)
	defer cancel()
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data, anonymized, normalized_type, normalized_data, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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

		// Apply normalization if normalizer is set
		if s.normalizer != nil {
			if err := s.normalizer.NormalizeEvent(event); err != nil {
				return fmt.Errorf("failed to normalize event %s: %w", event.ID, err)
			}
		}

		dataJSON, err := json.Marshal(event.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}

		_, err = stmt.Exec(event.ID, event.Type, event.Source, event.SourceID, event.Timestamp, event.Actor, event.EngineerID, string(dataJSON), event.Anonymized, event.NormalizedType, event.NormalizedData, event.CreatedAt, event.UpdatedAt)
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
		SELECT id, type, source, source_id, timestamp, actor, engineer_id, data, anonymized, normalized_type, normalized_data, created_at, updated_at
		FROM events WHERE id = ?
	`, id).Scan(&event.ID, &event.Type, &event.Source, &event.SourceID, &event.Timestamp, &event.Actor, &event.EngineerID, &dataJSON, &event.Anonymized, &event.NormalizedType, &event.NormalizedData, &event.CreatedAt, &event.UpdatedAt)

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
	query := "SELECT id, type, source, source_id, timestamp, actor, engineer_id, data, anonymized, normalized_type, normalized_data, created_at, updated_at FROM events WHERE 1=1"
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

		err := rows.Scan(&event.ID, &event.Type, &event.Source, &event.SourceID, &event.Timestamp, &event.Actor, &event.EngineerID, &dataJSON, &event.Anonymized, &event.NormalizedType, &event.NormalizedData, &event.CreatedAt, &event.UpdatedAt)
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
		UPDATE events SET type = ?, source = ?, source_id = ?, timestamp = ?, actor = ?, engineer_id = ?, data = ?, anonymized = ?, normalized_type = ?, normalized_data = ?, updated_at = ?
		WHERE id = ?
	`, event.Type, event.Source, event.SourceID, event.Timestamp, event.Actor, event.EngineerID, string(dataJSON), event.Anonymized, event.NormalizedType, event.NormalizedData, event.UpdatedAt, event.ID)

	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	return nil
}

func (s *EventStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM events WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

// GetEventsByEngineerAndTimeRange retrieves events for a specific engineer within a time range
// limit: max number of events to return (default 1000, max 10000)
func (s *EventStore) GetEventsByEngineerAndTimeRange(engineerID string, start, end time.Time, limit int) ([]*models.Event, error) {
	// Apply default and max limits for safety
	if limit <= 0 {
		limit = DefaultQueryLimit
	}
	if limit > MaxEventQueryLimit {
		limit = MaxEventQueryLimit
	}

	rows, err := s.db.Query(`
		SELECT id, type, source, source_id, timestamp, actor, engineer_id, data, anonymized, normalized_type, normalized_data, created_at, updated_at
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

		err := rows.Scan(&event.ID, &event.Type, &event.Source, &event.SourceID, &event.Timestamp, &event.Actor, &event.EngineerID, &dataJSON, &event.Anonymized, &event.NormalizedType, &event.NormalizedData, &event.CreatedAt, &event.UpdatedAt)
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
