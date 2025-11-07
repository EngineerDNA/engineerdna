package plugin

import (
	"fmt"

	"github.com/engineerdna/engineerdna/internal/anonymization"
	"github.com/engineerdna/engineerdna/internal/config"
	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

// Executor manages plugin execution with anonymization
type Executor struct {
	loader        *Loader
	anonService   *anonymization.Service
	anonStore     *db.AnonymizationStore
	pluginStore   *db.PluginStore
	auditStore    *db.AuditStore
	keyStore      *config.KeyStore
	metadataCache *MetadataCache
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
	}
}

// GetLoader returns the plugin loader
func (e *Executor) GetLoader() *Loader {
	return e.loader
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
