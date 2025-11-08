package plugin

import (
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

// SyncSourcePlugin syncs data from a source plugin
func (e *Executor) SyncSourcePlugin(pluginName string, since time.Time) ([]*models.Event, error) {
	// Load plugin config
	pluginConfig, err := e.pluginStore.GetByName(pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin config: %w", err)
	}
	if pluginConfig == nil {
		return nil, fmt.Errorf("plugin not configured: %s", pluginName)
	}
	if !pluginConfig.Enabled {
		return nil, fmt.Errorf("plugin disabled: %s", pluginName)
	}

	// Get plugin info
	info, err := e.GetPluginInfo(pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin info: %w", err)
	}

	// Register event types from plugin manifest
	if err := e.RegisterEventTypesFromPlugin(pluginName, info); err != nil {
		return nil, fmt.Errorf("failed to register event types: %w", err)
	}

	// Decrypt config
	var secretFields []string
	for _, field := range info.ConfigFields {
		if field.Secret {
			secretFields = append(secretFields, field.Name)
		}
	}

	decryptedConfig, err := e.keyStore.DecryptPluginConfig(pluginConfig.Config, secretFields)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt config: %w", err)
	}

	// Start plugin
	pluginPath, err := e.loader.GetPluginPath(pluginName)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %w", err)
	}

	client, err := NewClient(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}
	defer client.Close()

	// Configure plugin
	if err := client.Configure(decryptedConfig); err != nil {
		return nil, fmt.Errorf("failed to configure plugin: %w", err)
	}

	// Sync data
	result, err := client.Sync(sdk.SyncParams{Since: since})
	if err != nil {
		return nil, fmt.Errorf("sync failed: %w", err)
	}

	// Enforce maximum event count to prevent memory exhaustion
	if len(result.Events) > MaxPluginEvents {
		return nil, fmt.Errorf("too many events returned: %d (max %d)", len(result.Events), MaxPluginEvents)
	}

	// Convert SDK events to internal events
	events := make([]*models.Event, len(result.Events))
	for i, sdkEvent := range result.Events {
		events[i] = &models.Event{
			ID:        sdkEvent.ID,
			Type:      sdkEvent.Type,
			Source:    sdkEvent.Source,
			SourceID:  sdkEvent.SourceID,
			Timestamp: sdkEvent.Timestamp,
			Actor:     sdkEvent.Actor,
			Data:      sdkEvent.Data,
		}
	}

	// Update last sync time
	now := time.Now().UTC()
	if err := e.pluginStore.UpdateLastSync(pluginName, now); err != nil {
		return nil, fmt.Errorf("failed to update last sync: %w", err)
	}

	return events, nil
}
