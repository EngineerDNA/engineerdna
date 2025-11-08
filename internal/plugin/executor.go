package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/engineerdna/engineerdna/internal/anonymization"
	"github.com/engineerdna/engineerdna/internal/config"
	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

// Executor manages plugin execution with anonymization and event normalization
type Executor struct {
	loader        *Loader
	anonService   *anonymization.Service
	anonStore     *db.AnonymizationStore
	pluginStore   *db.PluginStore
	auditStore    *db.AuditStore
	keyStore      *config.KeyStore
	metadataCache *MetadataCache
	eventRegistry *EventTypeRegistry
}

// NewExecutor creates a new plugin executor
func NewExecutor(
	loader *Loader,
	anonService *anonymization.Service,
	anonStore *db.AnonymizationStore,
	pluginStore *db.PluginStore,
	auditStore *db.AuditStore,
	keyStore *config.KeyStore,
) *Executor {
	return &Executor{
		loader:        loader,
		anonService:   anonService,
		anonStore:     anonStore,
		pluginStore:   pluginStore,
		auditStore:    auditStore,
		keyStore:      keyStore,
		metadataCache: NewMetadataCache(),
		eventRegistry: NewEventTypeRegistry(),
	}
}

// GetLoader returns the plugin loader
func (e *Executor) GetLoader() *Loader {
	return e.loader
}

// GetEventRegistry returns the event type registry
func (e *Executor) GetEventRegistry() *EventTypeRegistry {
	return e.eventRegistry
}

// RegisterEventTypesFromPlugin registers event types from a plugin's metadata
func (e *Executor) RegisterEventTypesFromPlugin(pluginName string, info *sdk.PluginInfo) error {
	if info.ProvidesEventTypes == nil || len(info.ProvidesEventTypes) == 0 {
		return nil
	}

	for _, eventTypeSpec := range info.ProvidesEventTypes {
		// Skip if no normalization configured
		if eventTypeSpec.NormalizedType == "" {
			continue
		}

		// Parse schema JSON
		var schema map[string]FieldDef
		if err := json.Unmarshal([]byte(eventTypeSpec.Schema), &schema); err != nil {
			return fmt.Errorf("failed to parse schema for event type %s: %w", eventTypeSpec.Type, err)
		}

		// Register the event type
		if err := e.eventRegistry.RegisterEventType(
			pluginName,
			eventTypeSpec.Type,
			eventTypeSpec.NormalizedType,
			schema,
			eventTypeSpec.NormalizationMap,
		); err != nil {
			return fmt.Errorf("failed to register event type %s: %w", eventTypeSpec.Type, err)
		}
	}

	return nil
}

// GetPluginInfo retrieves plugin metadata with caching
func (e *Executor) GetPluginInfo(pluginName string) (*sdk.PluginInfo, error) {
	// Check cache first
	if cached, ok := e.metadataCache.Get(pluginName); ok {
		return cached, nil
	}

	// Cache miss - fetch from plugin
	pluginPath, err := e.loader.GetPluginPath(pluginName)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %w", err)
	}

	client, err := NewClient(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}
	defer client.Close()

	info, err := client.GetInfo()
	if err != nil {
		return nil, err
	}

	// Store in cache
	e.metadataCache.Set(pluginName, *info)

	return info, nil
}
