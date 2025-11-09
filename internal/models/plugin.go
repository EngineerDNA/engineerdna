package models

import "time"

type PluginType string

const (
	PluginTypeSource          PluginType = "source"
	PluginTypeDestination     PluginType = "destination"
	PluginTypeProcessor       PluginType = "processor"
	PluginTypeMetricSource    PluginType = "metric_source"
	PluginTypeAttributeSource PluginType = "attribute_source"
)

type PluginConfig struct {
	Name      string            `json:"name"`
	Type      PluginType        `json:"type"`
	Enabled   bool              `json:"enabled"`
	Config    map[string]string `json:"config"`
	LastSync  *time.Time        `json:"last_sync"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type PluginMetadata struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Type                 PluginType        `json:"type"`
	Description          string            `json:"description"`
	Author               string            `json:"author"`
	ConfigFields         []ConfigField     `json:"config_fields"`
	Anonymization        AnonymizationSpec `json:"anonymization"`
	OAuth                *OAuthSpec        `json:"oauth,omitempty"`
	Schedule             *ScheduleSpec     `json:"schedule,omitempty"`
	ProvidesMetrics      []MetricSpec      `json:"provides_metrics,omitempty"`
	ProvidesEventTypes   []EventTypeSpec   `json:"provides_event_types,omitempty"`
	ProvidesWidgets      []WidgetSpec      `json:"provides_widgets,omitempty"`
	ProvidesCorrelations []CorrelationSpec `json:"provides_correlations,omitempty"`
}

type ConfigField struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
	Default     string   `json:"default,omitempty"`
	Options     []string `json:"options,omitempty"`
	Secret      bool     `json:"secret"`
}

type AnonymizationSpec struct {
	Required bool     `json:"required"`
	Reason   string   `json:"reason"`
	Strategy string   `json:"strategy"`
	Fields   []string `json:"fields"`
}

type OAuthSpec struct {
	Provider string   `json:"provider"`
	Scopes   []string `json:"scopes"`
}

type ScheduleSpec struct {
	Supported bool   `json:"supported"`
	Default   string `json:"default"`
}

// MetricSpec describes a metric that a plugin can produce
type MetricSpec struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Unit        string   `json:"unit"`
	Granularity []string `json:"granularity"`
	Dimensions  []string `json:"dimensions,omitempty"`
}

// EventTypeSpec describes an event type that a plugin can produce
type EventTypeSpec struct {
	Type             string            `json:"type"`                        // Source event type (e.g., "pull_request")
	Description      string            `json:"description"`                 // Human-readable description
	Schema           string            `json:"schema"`                      // JSON schema for field definitions
	NormalizedType   string            `json:"normalized_type,omitempty"`   // Normalized type (e.g., "code_review")
	NormalizationMap map[string]string `json:"normalization_map,omitempty"` // Mapping from normalized fields to source fields
}

// WidgetSpec describes a widget type that a plugin provides
type WidgetSpec struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CorrelationSpec describes a correlation that a plugin can compute
type CorrelationSpec struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Inputs      []string `json:"inputs"` // Metric names or event types required
}

// PluginManifest represents plugin capabilities stored in database
type PluginManifest struct {
	PluginName           string            `json:"plugin_name"`
	Version              string            `json:"version"`
	Type                 PluginType        `json:"type"`
	Capabilities         []string          `json:"capabilities,omitempty"`
	ProvidesMetrics      []MetricSpec      `json:"provides_metrics,omitempty"`
	ProvidesEventTypes   []EventTypeSpec   `json:"provides_event_types,omitempty"`
	ProvidesWidgets      []WidgetSpec      `json:"provides_widgets,omitempty"`
	ProvidesCorrelations []CorrelationSpec `json:"provides_correlations,omitempty"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}
