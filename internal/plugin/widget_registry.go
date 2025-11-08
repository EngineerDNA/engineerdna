package plugin

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// WidgetRegistry manages widget registrations from plugins
type WidgetRegistry struct {
	manifestStore *db.PluginManifestStore
	mu            sync.RWMutex
	widgets       map[string]*WidgetRegistration // key: widget_id (plugin:widget_name)
}

// WidgetRegistration represents a registered widget
type WidgetRegistration struct {
	ID               string               `json:"id"` // Unique ID: plugin_name:widget_name
	PluginName       string               `json:"plugin_name"`
	Name             string               `json:"name"`
	Type             string               `json:"type"` // "number", "chart", "table", "custom"
	Description      string               `json:"description"`
	DataRequirements DataRequirements     `json:"data_requirements"`
	Configuration    []models.ConfigField `json:"configuration,omitempty"`
}

// DataRequirements specifies what data a widget needs
type DataRequirements struct {
	Metrics      []string `json:"metrics,omitempty"`
	Events       []string `json:"events,omitempty"`
	Correlations []string `json:"correlations,omitempty"`
	Attributes   []string `json:"attributes,omitempty"`
}

// NewWidgetRegistry creates a new widget registry
func NewWidgetRegistry(manifestStore *db.PluginManifestStore) *WidgetRegistry {
	return &WidgetRegistry{
		manifestStore: manifestStore,
		widgets:       make(map[string]*WidgetRegistration),
	}
}

// RegisterWidget registers a widget from a plugin
func (r *WidgetRegistry) RegisterWidget(pluginName string, widget *WidgetRegistration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate widget
	if widget.Name == "" {
		return fmt.Errorf("widget name is required")
	}
	if widget.Type == "" {
		return fmt.Errorf("widget type is required")
	}

	// Validate widget type
	validTypes := map[string]bool{
		"number": true,
		"chart":  true,
		"table":  true,
		"custom": true,
	}
	if !validTypes[widget.Type] {
		return fmt.Errorf("invalid widget type: %s (must be: number, chart, table, custom)", widget.Type)
	}

	// Set plugin name and ID
	widget.PluginName = pluginName
	widget.ID = fmt.Sprintf("%s:%s", pluginName, widget.Name)

	// Store in memory
	r.widgets[widget.ID] = widget

	return nil
}

// RegisterWidgetsFromManifest registers all widgets from a plugin manifest
func (r *WidgetRegistry) RegisterWidgetsFromManifest(manifest *models.PluginManifest) error {
	if manifest == nil || len(manifest.ProvidesWidgets) == 0 {
		return nil
	}

	for _, widgetSpec := range manifest.ProvidesWidgets {
		widget := &WidgetRegistration{
			Name:        widgetSpec.Name,
			Type:        widgetSpec.Type,
			Description: widgetSpec.Description,
			// DataRequirements populated from manifest if available
		}

		if err := r.RegisterWidget(manifest.PluginName, widget); err != nil {
			return fmt.Errorf("failed to register widget %s: %w", widgetSpec.Name, err)
		}
	}

	return nil
}

// GetWidget retrieves a widget by ID
func (r *WidgetRegistry) GetWidget(widgetID string) (*WidgetRegistration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	widget, ok := r.widgets[widgetID]
	return widget, ok
}

// ListWidgets returns all registered widgets
func (r *WidgetRegistry) ListWidgets() []*WidgetRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	widgets := make([]*WidgetRegistration, 0, len(r.widgets))
	for _, widget := range r.widgets {
		widgets = append(widgets, widget)
	}
	return widgets
}

// ListWidgetsByPlugin returns widgets for a specific plugin
func (r *WidgetRegistry) ListWidgetsByPlugin(pluginName string) []*WidgetRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	widgets := make([]*WidgetRegistration, 0)
	for _, widget := range r.widgets {
		if widget.PluginName == pluginName {
			widgets = append(widgets, widget)
		}
	}
	return widgets
}

// UnregisterPlugin removes all widgets for a plugin
func (r *WidgetRegistry) UnregisterPlugin(pluginName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove all widgets for this plugin
	for widgetID, widget := range r.widgets {
		if widget.PluginName == pluginName {
			delete(r.widgets, widgetID)
		}
	}
}

// LoadFromDatabase loads widget registrations from plugin manifests in database
func (r *WidgetRegistry) LoadFromDatabase() error {
	manifests, err := r.manifestStore.List()
	if err != nil {
		return fmt.Errorf("failed to list plugin manifests: %w", err)
	}

	for _, manifest := range manifests {
		if err := r.RegisterWidgetsFromManifest(manifest); err != nil {
			// Log error but continue loading other widgets
			continue
		}
	}

	return nil
}

// ValidateWidgetConfig validates widget configuration
func (r *WidgetRegistry) ValidateWidgetConfig(widgetID string, config map[string]interface{}) error {
	widget, ok := r.GetWidget(widgetID)
	if !ok {
		return fmt.Errorf("widget not found: %s", widgetID)
	}

	// Validate required configuration fields
	for _, field := range widget.Configuration {
		if field.Required {
			if _, ok := config[field.Name]; !ok {
				return fmt.Errorf("required configuration field missing: %s", field.Name)
			}
		}
	}

	return nil
}

// ParseWidgetSpec parses a widget specification from JSON
func ParseWidgetSpec(data []byte) (*WidgetRegistration, error) {
	var widget WidgetRegistration
	if err := json.Unmarshal(data, &widget); err != nil {
		return nil, fmt.Errorf("failed to parse widget spec: %w", err)
	}
	return &widget, nil
}

// GetWidgetsByType returns widgets of a specific type
func (r *WidgetRegistry) GetWidgetsByType(widgetType string) []*WidgetRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	widgets := make([]*WidgetRegistration, 0)
	for _, widget := range r.widgets {
		if widget.Type == widgetType {
			widgets = append(widgets, widget)
		}
	}
	return widgets
}

// CheckDataAvailability checks if required data is available for a widget
func (r *WidgetRegistry) CheckDataAvailability(widgetID string, availableMetrics, availableEvents []string) bool {
	widget, ok := r.GetWidget(widgetID)
	if !ok {
		return false
	}

	// Check if all required metrics are available
	for _, requiredMetric := range widget.DataRequirements.Metrics {
		found := false
		for _, available := range availableMetrics {
			if available == requiredMetric {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check if all required events are available
	for _, requiredEvent := range widget.DataRequirements.Events {
		found := false
		for _, available := range availableEvents {
			if available == requiredEvent {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
