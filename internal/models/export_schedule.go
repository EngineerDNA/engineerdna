package models

import "time"

type ExportSchedule struct {
	ID         string     `json:"id"`
	PluginName string     `json:"plugin_name"`
	Frequency  string     `json:"frequency"`
	DayOfWeek  *int       `json:"day_of_week,omitempty"`
	TimeOfDay  string     `json:"time_of_day"`
	Enabled    bool       `json:"enabled"`
	LastRun    *time.Time `json:"last_run,omitempty"`
	NextRun    time.Time  `json:"next_run"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
