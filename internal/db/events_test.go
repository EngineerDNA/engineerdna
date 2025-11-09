package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	_ "modernc.org/sqlite"
)

func setupEventTestDB(t *testing.T) (*sql.DB, *EventStore) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create events table
	_, err = db.Exec(`
		CREATE TABLE events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			source TEXT NOT NULL,
			source_id TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			actor TEXT NOT NULL,
			engineer_id TEXT,
			data TEXT NOT NULL,
			anonymized BOOLEAN DEFAULT FALSE,
			normalized_type TEXT,
			normalized_data TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db, NewEventStore(db)
}

func TestAvgFieldValue_SQLInjectionPrevention(t *testing.T) {
	db, store := setupEventTestDB(t)
	defer db.Close()

	// Insert test events with numeric field
	now := time.Now().UTC()
	testEvents := []*models.Event{
		{
			ID:        "event1",
			Type:      "test",
			Source:    "test-source",
			SourceID:  "1",
			Timestamp: now,
			Actor:     "test-actor",
			Data: map[string]interface{}{
				"duration": 100.0,
				"count":    5.0,
			},
		},
		{
			ID:        "event2",
			Type:      "test",
			Source:    "test-source",
			SourceID:  "2",
			Timestamp: now.Add(1 * time.Hour),
			Actor:     "test-actor",
			Data: map[string]interface{}{
				"duration": 200.0,
				"count":    10.0,
			},
		},
	}

	for _, event := range testEvents {
		if err := store.Create(event); err != nil {
			t.Fatalf("failed to create test event: %v", err)
		}
	}

	tests := []struct {
		name      string
		field     string
		wantErr   bool
		wantAvg   float64
		errSubstr string
	}{
		{
			name:    "valid field - duration",
			field:   "duration",
			wantErr: false,
			wantAvg: 150.0,
		},
		{
			name:    "valid field - count",
			field:   "count",
			wantErr: false,
			wantAvg: 7.5,
		},
		{
			name:    "valid field with underscore",
			field:   "my_field_123",
			wantErr: false,
			wantAvg: 0, // Field doesn't exist, but validation passes
		},
		{
			name:      "SQL injection attempt - single quote",
			field:     "x') OR 1=1 OR ('1'='1",
			wantErr:   true,
			errSubstr: "invalid field name",
		},
		{
			name:      "SQL injection attempt - semicolon",
			field:     "duration'; DROP TABLE events; --",
			wantErr:   true,
			errSubstr: "invalid field name",
		},
		{
			name:      "SQL injection attempt - comment",
			field:     "duration--",
			wantErr:   true,
			errSubstr: "invalid field name",
		},
		{
			name:      "SQL injection attempt - union",
			field:     "duration UNION SELECT password FROM users",
			wantErr:   true,
			errSubstr: "invalid field name",
		},
		{
			name:      "invalid field - special characters",
			field:     "duration@field",
			wantErr:   true,
			errSubstr: "invalid field name",
		},
		{
			name:      "invalid field - parentheses",
			field:     "duration()",
			wantErr:   true,
			errSubstr: "invalid field name",
		},
		{
			name:      "empty field name",
			field:     "",
			wantErr:   true,
			errSubstr: "field name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avg, err := store.AvgFieldValue(tt.field, map[string]interface{}{})

			if tt.wantErr {
				if err == nil {
					t.Errorf("AvgFieldValue() expected error containing %q, got nil", tt.errSubstr)
					return
				}
				if tt.errSubstr != "" && !contains(err.Error(), tt.errSubstr) {
					t.Errorf("AvgFieldValue() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("AvgFieldValue() unexpected error = %v", err)
					return
				}
				if avg != tt.wantAvg {
					t.Errorf("AvgFieldValue() = %v, want %v", avg, tt.wantAvg)
				}
			}
		})
	}
}

func TestAvgFieldValue_WithFilters(t *testing.T) {
	db, store := setupEventTestDB(t)
	defer db.Close()

	// Insert test events
	now := time.Now().UTC()
	testEvents := []*models.Event{
		{
			ID:        "event1",
			Type:      "type_a",
			Source:    "source_a",
			SourceID:  "1",
			Timestamp: now,
			Actor:     "actor1",
			Data:      map[string]interface{}{"score": 80.0},
		},
		{
			ID:        "event2",
			Type:      "type_a",
			Source:    "source_a",
			SourceID:  "2",
			Timestamp: now.Add(1 * time.Hour),
			Actor:     "actor1",
			Data:      map[string]interface{}{"score": 90.0},
		},
		{
			ID:        "event3",
			Type:      "type_b",
			Source:    "source_b",
			SourceID:  "3",
			Timestamp: now.Add(2 * time.Hour),
			Actor:     "actor2",
			Data:      map[string]interface{}{"score": 100.0},
		},
	}

	for _, event := range testEvents {
		if err := store.Create(event); err != nil {
			t.Fatalf("failed to create test event: %v", err)
		}
	}

	tests := []struct {
		name    string
		filters map[string]interface{}
		wantAvg float64
	}{
		{
			name:    "no filters",
			filters: map[string]interface{}{},
			wantAvg: 90.0, // (80 + 90 + 100) / 3
		},
		{
			name: "filter by type",
			filters: map[string]interface{}{
				"type": "type_a",
			},
			wantAvg: 85.0, // (80 + 90) / 2
		},
		{
			name: "filter by source",
			filters: map[string]interface{}{
				"source": "source_a",
			},
			wantAvg: 85.0, // (80 + 90) / 2
		},
		{
			name: "filter by time range",
			filters: map[string]interface{}{
				"start_time": now,
				"end_time":   now.Add(90 * time.Minute),
			},
			wantAvg: 85.0, // (80 + 90) / 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avg, err := store.AvgFieldValue("score", tt.filters)
			if err != nil {
				t.Errorf("AvgFieldValue() unexpected error = %v", err)
				return
			}
			if avg != tt.wantAvg {
				t.Errorf("AvgFieldValue() = %v, want %v", avg, tt.wantAvg)
			}
		})
	}
}

func TestIsValidFieldName(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		want      bool
	}{
		{"valid lowercase", "fieldname", true},
		{"valid uppercase", "FIELDNAME", true},
		{"valid mixed case", "FieldName", true},
		{"valid with underscore", "field_name", true},
		{"valid with numbers", "field123", true},
		{"valid alphanumeric underscore", "my_field_123", true},
		{"empty string", "", false},
		{"with space", "field name", false},
		{"with dash", "field-name", false},
		{"with dot", "field.name", false},
		{"with single quote", "field'name", false},
		{"with double quote", "field\"name", false},
		{"with parentheses", "field()", false},
		{"with semicolon", "field;", false},
		{"with asterisk", "field*", false},
		{"with dollar", "$field", false},
		{"too long", "this_is_a_very_long_field_name_that_exceeds_fifty_characters_limit", false},
		{"SQL injection 1", "x') OR 1=1--", false},
		{"SQL injection 2", "'; DROP TABLE events; --", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidFieldName(tt.fieldName); got != tt.want {
				t.Errorf("isValidFieldName(%q) = %v, want %v", tt.fieldName, got, tt.want)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
