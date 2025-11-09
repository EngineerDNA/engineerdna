package models

import "time"

type InsightSeverity string

const (
	SeverityInfo    InsightSeverity = "info"
	SeverityWarning InsightSeverity = "warning"
	SeverityError   InsightSeverity = "error"
)

type Insight struct {
	ID             string                 `json:"id"`
	PluginName     string                 `json:"plugin_name"`
	GeneratedAt    time.Time              `json:"generated_at"`
	PeriodStart    time.Time              `json:"period_start"`
	PeriodEnd      time.Time              `json:"period_end"`
	Severity       InsightSeverity        `json:"severity"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Recommendation string                 `json:"recommendation,omitempty"`
	Metrics        map[string]interface{} `json:"metrics"`
	Dismissed      bool                   `json:"dismissed"`
	CreatedAt      time.Time              `json:"created_at"`
}

type Export struct {
	ID             string    `json:"id"`
	PluginName     string    `json:"plugin_name"`
	ExportedAt     time.Time `json:"exported_at"`
	Status         string    `json:"status"`
	DestinationURL string    `json:"destination_url,omitempty"`
	EventCount     int       `json:"event_count"`
	Anonymized     bool      `json:"anonymized"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}
