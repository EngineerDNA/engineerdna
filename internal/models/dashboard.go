package models

import "time"

// Dashboard represents a customizable dashboard with widgets
type Dashboard struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Persona     string    `json:"persona,omitempty"` // 'ic', 'team_lead', 'director'
	IsTemplate  bool      `json:"is_template"`
	IsSystem    bool      `json:"is_system"`         // System dashboards can't be deleted
	Layout      string    `json:"layout"`            // JSON string
	Filters     string    `json:"filters,omitempty"` // JSON string
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DashboardTemplate represents a template for creating new dashboards
// Used in role-based onboarding
type DashboardTemplate struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Role        string    `json:"role,omitempty"`     // ic, manager, director, admin
	Category    string    `json:"category,omitempty"` // personal, team, org, admin
	Layout      string    `json:"layout"`             // JSON string
	IsSystem    bool      `json:"is_system"`          // System templates can't be deleted
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DashboardLayout represents the deserialized layout JSON
type DashboardLayout struct {
	Widgets []Widget `json:"widgets"`
}

// Widget represents a dashboard widget configuration
type Widget struct {
	ID                  string              `json:"id"`
	Type                string              `json:"type"` // 'number', 'timeseries', 'bar', 'table', 'status', 'feed'
	Title               string              `json:"title"`
	DataSource          string              `json:"data_source"` // API endpoint path
	QueryParams         map[string]string   `json:"query_params,omitempty"`
	VisualizationConfig VisualizationConfig `json:"visualization_config,omitempty"`
	Position            WidgetPosition      `json:"position"`
}

// WidgetPosition represents the grid position and size of a widget
type WidgetPosition struct {
	X int `json:"x"` // Grid column
	Y int `json:"y"` // Grid row
	W int `json:"w"` // Width in grid units
	H int `json:"h"` // Height in grid units
}

// VisualizationConfig represents widget-specific visualization settings
type VisualizationConfig struct {
	PrimaryColor   string                 `json:"primary_color,omitempty"`
	SecondaryColor string                 `json:"secondary_color,omitempty"`
	ShowLegend     bool                   `json:"show_legend,omitempty"`
	ShowGrid       bool                   `json:"show_grid,omitempty"`
	YAxisLabel     string                 `json:"y_axis_label,omitempty"`
	XAxisLabel     string                 `json:"x_axis_label,omitempty"`
	TimeFormat     string                 `json:"time_format,omitempty"`
	NumberFormat   string                 `json:"number_format,omitempty"`
	Columns        []TableColumn          `json:"columns,omitempty"`   // For table widgets
	Threshold      *ThresholdConfig       `json:"threshold,omitempty"` // For status widgets
	Additional     map[string]interface{} `json:"additional,omitempty"`
}

// TableColumn represents a column configuration for table widgets
type TableColumn struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Sortable bool   `json:"sortable"`
	Format   string `json:"format,omitempty"` // 'number', 'date', 'percentage', 'text'
}

// ThresholdConfig represents threshold settings for status widgets
type ThresholdConfig struct {
	GoodMin     float64 `json:"good_min"`     // >= this is good
	WarningMin  float64 `json:"warning_min"`  // >= this is warning
	CriticalMax float64 `json:"critical_max"` // <= this is critical
}

// DashboardFilters represents the deserialized filters JSON
type DashboardFilters struct {
	DateRange   *DateRangeFilter `json:"date_range,omitempty"`
	TeamIDs     []string         `json:"team_ids,omitempty"`
	EngineerIDs []string         `json:"engineer_ids,omitempty"`
}

// DateRangeFilter represents a date range filter
type DateRangeFilter struct {
	Start string `json:"start"` // ISO 8601 date
	End   string `json:"end"`   // ISO 8601 date
}

// MetricSnapshot represents a pre-computed metric value for fast queries
type MetricSnapshot struct {
	MetricName  string    `json:"metric_name"`           // 'team_score', 'pr_volume', 'cycle_time'
	EntityType  string    `json:"entity_type"`           // 'team', 'engineer', 'org'
	EntityID    string    `json:"entity_id,omitempty"`   // team_id, engineer_id, null for org
	EntityName  string    `json:"entity_name,omitempty"` // Resolved entity name (from JOIN)
	PeriodStart string    `json:"period_start"`          // ISO 8601 date
	PeriodEnd   string    `json:"period_end"`            // ISO 8601 date
	Value       float64   `json:"value"`
	Metadata    string    `json:"metadata,omitempty"` // JSON string
	ComputedAt  time.Time `json:"computed_at"`
}

// MetricSnapshotMetadata represents the deserialized metadata JSON
type MetricSnapshotMetadata struct {
	Count      int                    `json:"count,omitempty"`
	Samples    int                    `json:"samples,omitempty"`
	Min        float64                `json:"min,omitempty"`
	Max        float64                `json:"max,omitempty"`
	Avg        float64                `json:"avg,omitempty"`
	StdDev     float64                `json:"std_dev,omitempty"`
	Additional map[string]interface{} `json:"additional,omitempty"`
}
