package plugin

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// FieldDef defines a field in an event schema
type FieldDef struct {
	Type     string   `json:"type"`      // "string", "timestamp", "integer", "enum"
	Required bool     `json:"required"`  // Whether field is required
	Enum     []string `json:"enum"`      // Valid values for enum type
}

// EventTypeRegistration represents a mapping from source event type to normalized type
type EventTypeRegistration struct {
	PluginName       string            `json:"plugin_name"`        // Plugin that provides this event type
	SourceType       string            `json:"source_type"`        // Source-specific type (e.g., "pull_request")
	NormalizedType   string            `json:"normalized_type"`    // Normalized type (e.g., "code_review")
	Schema           map[string]FieldDef `json:"schema"`           // Field definitions
	NormalizationMap map[string]string `json:"normalization_map"`  // Normalized field -> source field mapping
}

// EventTypeRegistry manages event type registrations
type EventTypeRegistry struct {
	mu            sync.RWMutex
	registrations map[string]*EventTypeRegistration // Key: "pluginName:sourceType"
}

// NewEventTypeRegistry creates a new event type registry
func NewEventTypeRegistry() *EventTypeRegistry {
	return &EventTypeRegistry{
		registrations: make(map[string]*EventTypeRegistration),
	}
}

// RegisterEventType registers an event type mapping
func (r *EventTypeRegistry) RegisterEventType(
	pluginName string,
	sourceType string,
	normalizedType string,
	schema map[string]FieldDef,
	normalizationMap map[string]string,
) error {
	// Validate inputs
	if err := r.validatePluginName(pluginName); err != nil {
		return fmt.Errorf("invalid plugin name: %w", err)
	}
	if err := r.validateEventType(sourceType); err != nil {
		return fmt.Errorf("invalid source type: %w", err)
	}
	if err := r.validateEventType(normalizedType); err != nil {
		return fmt.Errorf("invalid normalized type: %w", err)
	}
	if err := r.validateSchema(schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}
	if err := r.validateNormalizationMap(normalizationMap, schema); err != nil {
		return fmt.Errorf("invalid normalization map: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.makeKey(pluginName, sourceType)
	r.registrations[key] = &EventTypeRegistration{
		PluginName:       pluginName,
		SourceType:       sourceType,
		NormalizedType:   normalizedType,
		Schema:           schema,
		NormalizationMap: normalizationMap,
	}

	return nil
}

// GetNormalizedType returns the normalized type for a plugin's source type
func (r *EventTypeRegistry) GetNormalizedType(pluginName, sourceType string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := r.makeKey(pluginName, sourceType)
	reg, ok := r.registrations[key]
	if !ok {
		return "", false
	}

	return reg.NormalizedType, true
}

// GetRegistration returns the full registration for a plugin's source type
func (r *EventTypeRegistry) GetRegistration(pluginName, sourceType string) (*EventTypeRegistration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := r.makeKey(pluginName, sourceType)
	reg, ok := r.registrations[key]
	if !ok {
		return nil, false
	}

	// Return a copy to prevent external modifications
	copy := *reg
	copy.Schema = make(map[string]FieldDef)
	for k, v := range reg.Schema {
		copy.Schema[k] = v
	}
	copy.NormalizationMap = make(map[string]string)
	for k, v := range reg.NormalizationMap {
		copy.NormalizationMap[k] = v
	}

	return &copy, true
}

// ListRegistrations returns all registrations for a plugin
func (r *EventTypeRegistry) ListRegistrations(pluginName string) []*EventTypeRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*EventTypeRegistration
	prefix := pluginName + ":"
	for key, reg := range r.registrations {
		if strings.HasPrefix(key, prefix) {
			copy := *reg
			copy.Schema = make(map[string]FieldDef)
			for k, v := range reg.Schema {
				copy.Schema[k] = v
			}
			copy.NormalizationMap = make(map[string]string)
			for k, v := range reg.NormalizationMap {
				copy.NormalizationMap[k] = v
			}
			results = append(results, &copy)
		}
	}

	return results
}

// UnregisterPlugin removes all registrations for a plugin
func (r *EventTypeRegistry) UnregisterPlugin(pluginName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	prefix := pluginName + ":"
	for key := range r.registrations {
		if strings.HasPrefix(key, prefix) {
			delete(r.registrations, key)
		}
	}
}

// makeKey creates a registry key from plugin name and source type
func (r *EventTypeRegistry) makeKey(pluginName, sourceType string) string {
	return pluginName + ":" + sourceType
}

// validatePluginName validates plugin name format
func (r *EventTypeRegistry) validatePluginName(name string) error {
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	// Match loader validation pattern
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid characters in plugin name (only alphanumeric, underscore, dash allowed)")
	}
	return nil
}

// validateEventType validates event type format
func (r *EventTypeRegistry) validateEventType(eventType string) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	// Event types should be lowercase with underscores
	validType := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	if !validType.MatchString(eventType) {
		return fmt.Errorf("invalid event type format (must be lowercase with underscores)")
	}
	if len(eventType) > 100 {
		return fmt.Errorf("event type too long (max 100 characters)")
	}
	return nil
}

// validateSchema validates schema definitions
func (r *EventTypeRegistry) validateSchema(schema map[string]FieldDef) error {
	if len(schema) == 0 {
		return fmt.Errorf("schema cannot be empty")
	}
	if len(schema) > 100 {
		return fmt.Errorf("schema too large (max 100 fields)")
	}

	validFieldName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	validTypes := map[string]bool{
		"string":    true,
		"timestamp": true,
		"integer":   true,
		"enum":      true,
	}

	for fieldName, fieldDef := range schema {
		// Validate field name
		if !validFieldName.MatchString(fieldName) {
			return fmt.Errorf("invalid field name: %s", fieldName)
		}
		if len(fieldName) > 50 {
			return fmt.Errorf("field name too long: %s", fieldName)
		}

		// Validate field type
		if !validTypes[fieldDef.Type] {
			return fmt.Errorf("invalid field type for %s: %s", fieldName, fieldDef.Type)
		}

		// Validate enum
		if fieldDef.Type == "enum" {
			if len(fieldDef.Enum) == 0 {
				return fmt.Errorf("enum field %s must have values", fieldName)
			}
			if len(fieldDef.Enum) > 50 {
				return fmt.Errorf("enum field %s has too many values (max 50)", fieldName)
			}
			for _, val := range fieldDef.Enum {
				if len(val) > 100 {
					return fmt.Errorf("enum value too long in field %s", fieldName)
				}
			}
		}
	}

	return nil
}

// validateNormalizationMap validates normalization mappings
func (r *EventTypeRegistry) validateNormalizationMap(
	normalizationMap map[string]string,
	schema map[string]FieldDef,
) error {
	if len(normalizationMap) == 0 {
		return fmt.Errorf("normalization map cannot be empty")
	}
	if len(normalizationMap) > 100 {
		return fmt.Errorf("normalization map too large (max 100 mappings)")
	}

	validFieldName := regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)

	for normalizedField, sourceField := range normalizationMap {
		// Validate normalized field exists in schema
		if _, ok := schema[normalizedField]; !ok {
			return fmt.Errorf("normalized field %s not in schema", normalizedField)
		}

		// Validate source field format (allows nested fields like "author.name")
		if !validFieldName.MatchString(sourceField) {
			return fmt.Errorf("invalid source field format: %s", sourceField)
		}

		// Prevent path traversal attempts
		if strings.Contains(sourceField, "..") {
			return fmt.Errorf("invalid source field (path traversal detected): %s", sourceField)
		}

		if len(sourceField) > 200 {
			return fmt.Errorf("source field path too long: %s", sourceField)
		}
	}

	return nil
}

// ValidateSchema validates a schema definition (for external use)
func (r *EventTypeRegistry) ValidateSchema(schemaJSON string) error {
	var schema map[string]FieldDef
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return r.validateSchema(schema)
}
