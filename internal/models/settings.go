package models

import "time"

// Settings represents user-configurable application settings
// Singleton pattern: only one row with id='singleton'
type Settings struct {
	ID                    string    `json:"id"`
	SyncSchedule          string    `json:"sync_schedule"`
	AnonymizationStrategy string    `json:"anonymization_strategy"`
	Port                  int       `json:"port"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
