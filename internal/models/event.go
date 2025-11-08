package models

import "time"

type Event struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	Source         string                 `json:"source"`
	SourceID       string                 `json:"source_id"`
	Timestamp      time.Time              `json:"timestamp"`
	Actor          string                 `json:"actor"`
	EngineerID     string                 `json:"engineer_id,omitempty"`
	Data           map[string]interface{} `json:"data"`
	Anonymized     bool                   `json:"anonymized"`
	NormalizedType string                 `json:"normalized_type,omitempty"` // Normalized event type for multi-source
	NormalizedData string                 `json:"normalized_data,omitempty"` // Normalized data as JSON string
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}
