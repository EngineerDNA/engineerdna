package models

import "time"

// MetricValue represents a single time-series measurement
type MetricValue struct {
	ID          string                 `json:"id"`
	MetricName  string                 `json:"metric_name"`
	Source      string                 `json:"source"`      // Plugin name that created this
	Timestamp   time.Time              `json:"timestamp"`   // UTC timestamp
	Granularity string                 `json:"granularity"` // hourly, daily, weekly, monthly
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit,omitempty"`       // dollars, hours, count, percentage
	Dimensions  map[string]interface{} `json:"dimensions,omitempty"` // Multi-dimensional metrics
	CreatedAt   time.Time              `json:"created_at"`
}

// MetricDefinition describes a metric type that a plugin can produce
type MetricDefinition struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Unit        string   `json:"unit"`
	Granularity []string `json:"granularity"`          // Supported granularities
	Dimensions  []string `json:"dimensions,omitempty"` // Dimension names if multi-dimensional
}
