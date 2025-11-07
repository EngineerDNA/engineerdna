package models

import "time"

type AuditLogEntry struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	PluginName    string                 `json:"plugin_name"`
	Action        string                 `json:"action"`
	EventCount    int                    `json:"event_count"`
	Anonymized    bool                   `json:"anonymized"`
	Destination   string                 `json:"destination,omitempty"`
	DataSummary   map[string]interface{} `json:"data_summary"`
	UserInitiated bool                   `json:"user_initiated"`
}
