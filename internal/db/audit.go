package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

type AuditStore struct {
	db *sql.DB
}

func NewAuditStore(db *sql.DB) *AuditStore {
	return &AuditStore{db: db}
}

// LogEntry creates an audit log entry
func (s *AuditStore) LogEntry(entry *models.AuditLogEntry) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	dataSummaryJSON, err := json.Marshal(entry.DataSummary)
	if err != nil {
		return fmt.Errorf("failed to marshal data summary: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO audit_log (id, timestamp, plugin_name, action, event_count, anonymized, destination, data_summary, user_initiated, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, entry.ID, entry.Timestamp, entry.PluginName, entry.Action, entry.EventCount, entry.Anonymized, entry.Destination, string(dataSummaryJSON), entry.UserInitiated)

	if err != nil {
		return fmt.Errorf("failed to create audit log entry: %w", err)
	}

	return nil
}

// List retrieves audit log entries with limit
func (s *AuditStore) List(limit int) ([]*models.AuditLogEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > DefaultQueryLimit {
		limit = DefaultQueryLimit
	}

	rows, err := s.db.Query(`
		SELECT id, timestamp, plugin_name, action, event_count, anonymized, destination, data_summary, user_initiated
		FROM audit_log
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit log entries: %w", err)
	}
	defer rows.Close()

	var entries []*models.AuditLogEntry
	for rows.Next() {
		var entry models.AuditLogEntry
		var dataSummaryJSON string
		var destination sql.NullString

		err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.PluginName, &entry.Action, &entry.EventCount, &entry.Anonymized, &destination, &dataSummaryJSON, &entry.UserInitiated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log entry: %w", err)
		}

		if destination.Valid {
			entry.Destination = destination.String
		}

		if err := json.Unmarshal([]byte(dataSummaryJSON), &entry.DataSummary); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data summary: %w", err)
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}
