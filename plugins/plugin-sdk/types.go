package sdk

import "time"

// Request represents a JSON-RPC request from the host to the plugin
type Request struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	ID     string      `json:"id"`
}

// Response represents a JSON-RPC response from the plugin to the host
type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  *RPCError   `json:"error,omitempty"`
	ID     string      `json:"id"`
}

// RPCError represents an error in JSON-RPC format
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error codes
const (
	ErrCodeConfig         = 1000
	ErrCodeAuth           = 1001
	ErrCodeRateLimit      = 1002
	ErrCodeNetwork        = 1003
	ErrCodePluginInternal = 2000
)

// Event represents a single engineering event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	SourceID  string                 `json:"source_id"`
	Timestamp time.Time              `json:"timestamp"`
	Actor     string                 `json:"actor"`
	Data      map[string]interface{} `json:"data"`
}

// Metric represents a time-series measurement
type Metric struct {
	MetricName  string                 `json:"metric_name"`
	Timestamp   time.Time              `json:"timestamp"`
	Granularity string                 `json:"granularity"` // hourly, daily, weekly, monthly
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit,omitempty"`       // dollars, hours, count, percentage
	Dimensions  map[string]interface{} `json:"dimensions,omitempty"` // Multi-dimensional metrics
}

// Attribute represents a fact about an entity
type Attribute struct {
	EntityType    string     `json:"entity_type"` // team, engineer, org
	EntityID      string     `json:"entity_id"`
	AttributeName string     `json:"attribute_name"`
	Value         string     `json:"value"`
	ValueType     string     `json:"value_type"`            // string, number, boolean, json
	ValidFrom     time.Time  `json:"valid_from"`            // When this value became valid
	ValidUntil    *time.Time `json:"valid_until,omitempty"` // When it stopped being valid (NULL = current)
}

// PluginInfo contains metadata about the plugin
type PluginInfo struct {
	Name               string            `json:"name"`
	Version            string            `json:"version"`
	Type               string            `json:"type"` // source, destination, processor
	Description        string            `json:"description"`
	Author             string            `json:"author"`
	ConfigFields       []ConfigField     `json:"config_fields"`
	Anonymization      AnonymizationSpec `json:"anonymization"`
	OAuth              *OAuthSpec        `json:"oauth,omitempty"`
	Schedule           *ScheduleSpec     `json:"schedule,omitempty"`
	Capabilities       []string          `json:"capabilities,omitempty"`
	ProvidesEventTypes []EventTypeSpec   `json:"provides_event_types,omitempty"`
}

// EventTypeSpec describes an event type that a plugin can produce
type EventTypeSpec struct {
	Type             string            `json:"type"`                        // Source event type (e.g., "pull_request")
	Description      string            `json:"description"`                 // Human-readable description
	Schema           string            `json:"schema"`                      // JSON schema for field definitions
	NormalizedType   string            `json:"normalized_type,omitempty"`   // Normalized type (e.g., "code_review")
	NormalizationMap map[string]string `json:"normalization_map,omitempty"` // Mapping from normalized fields to source fields
}

// ConfigField describes a configuration field required by the plugin
type ConfigField struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // string, password, boolean, select, json
	Required    bool     `json:"required"`
	Description string   `json:"description"`
	Default     string   `json:"default,omitempty"`
	Options     []string `json:"options,omitempty"`
	Secret      bool     `json:"secret"`
}

// AnonymizationSpec describes anonymization requirements
type AnonymizationSpec struct {
	Required bool     `json:"required"`
	Reason   string   `json:"reason"`
	Strategy string   `json:"strategy"`
	Fields   []string `json:"fields"`
}

// OAuthSpec describes OAuth requirements
type OAuthSpec struct {
	Provider string   `json:"provider"`
	Scopes   []string `json:"scopes"`
}

// ScheduleSpec describes scheduling capabilities
type ScheduleSpec struct {
	Supported bool   `json:"supported"`
	Default   string `json:"default"`
}

// HealthResult represents the result of a health check
type HealthResult struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message"`
}

// SyncParams contains parameters for syncing source data
type SyncParams struct {
	Since time.Time `json:"since"`
}

// SyncResult contains the result of a sync operation
// Plugins can return events, metrics, attributes, or any combination
type SyncResult struct {
	Events     []Event     `json:"events,omitempty"`
	Metrics    []Metric    `json:"metrics,omitempty"`
	Attributes []Attribute `json:"attributes,omitempty"`
	Warnings   []string    `json:"warnings,omitempty"` // Track skipped/failed rows
}

// TestResult represents the result of a destination test
type TestResult struct {
	Connected   bool   `json:"connected"`
	Destination string `json:"destination"`
}

// ExportParams contains parameters for exporting data
type ExportParams struct {
	DataType string                 `json:"data_type"`
	Data     map[string]interface{} `json:"data"`
	Options  map[string]interface{} `json:"options"`
}

// ExportResult contains the result of an export operation
type ExportResult struct {
	Status      string `json:"status"`
	URL         string `json:"url,omitempty"`
	RowsWritten int    `json:"rows_written,omitempty"`
}

// AnalyzeRequest contains parameters for processor analysis
type AnalyzeRequest struct {
	AnalysisType string                 `json:"analysis_type"`
	Events       []Event                `json:"events"`
	Context      map[string]interface{} `json:"context"`
}

// AnalyzeResult contains the result of a processor analysis
type AnalyzeResult struct {
	Insights []Insight `json:"insights"`
	Summary  string    `json:"summary,omitempty"`
}

// Insight represents an AI-generated insight
type Insight struct {
	Severity       string                 `json:"severity"` // info, warning, error
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Recommendation string                 `json:"recommendation,omitempty"`
	Metrics        map[string]interface{} `json:"metrics,omitempty"`
}

// PluginError allows plugins to return errors with specific error codes
type PluginError struct {
	Code    int
	Message string
}

func (e *PluginError) Error() string {
	return e.Message
}

// NewAuthError creates an authentication error
func NewAuthError(message string) *PluginError {
	return &PluginError{Code: ErrCodeAuth, Message: message}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string) *PluginError {
	return &PluginError{Code: ErrCodeRateLimit, Message: message}
}

// NewNetworkError creates a network error
func NewNetworkError(message string) *PluginError {
	return &PluginError{Code: ErrCodeNetwork, Message: message}
}

// NewConfigError creates a configuration error
func NewConfigError(message string) *PluginError {
	return &PluginError{Code: ErrCodeConfig, Message: message}
}
