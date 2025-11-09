package models

import "time"

// Settings represents user-configurable application settings
// Singleton pattern: only one row with id='singleton'
type Settings struct {
	ID                    string    `json:"id"`
	SyncSchedule          string    `json:"sync_schedule"`
	AnonymizationStrategy string    `json:"anonymization_strategy"`
	Port                  int       `json:"port"`
	Role                  string    `json:"role"`                           // ic, manager, director, admin
	PrimaryDashboardID    *string   `json:"primary_dashboard_id,omitempty"` // Nullable
	FavoriteDashboards    []string  `json:"favorite_dashboards,omitempty"`  // JSON array
	OnboardingCompleted   bool      `json:"onboarding_completed"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
