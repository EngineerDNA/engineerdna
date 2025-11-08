package models

import "time"

// Correlation defines a relationship between different data types
// (e.g., cost per feature, velocity vs team size)
type Correlation struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Plugin     string                 `json:"plugin"`     // Plugin that defines this
	Definition map[string]interface{} `json:"definition"` // How to compute correlation
	CreatedAt  time.Time              `json:"created_at"`
}

// CorrelationValue represents a computed correlation result
type CorrelationValue struct {
	ID            string                 `json:"id"`
	CorrelationID string                 `json:"correlation_id"`
	Timestamp     time.Time              `json:"timestamp"`
	TimeWindow    string                 `json:"time_window"` // day, week, month, quarter
	Value         float64                `json:"value"`
	Breakdown     map[string]interface{} `json:"breakdown,omitempty"` // Breakdown by dimension
	CreatedAt     time.Time              `json:"created_at"`
}
