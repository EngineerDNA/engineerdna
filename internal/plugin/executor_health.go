package plugin

import (
	"fmt"
	"os"

	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

// HealthCheck checks if a plugin is healthy
func (e *Executor) HealthCheck(pluginName string) (*sdk.HealthResult, error) {
	pluginPath, err := e.loader.GetPluginPath(pluginName)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %w", err)
	}

	client, err := NewClient(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}
	defer client.Close()

	// Get config if exists
	pluginConfig, err := e.pluginStore.GetByName(pluginName)
	if err == nil && pluginConfig != nil {
		// Get plugin info to know secret fields
		info, err := e.GetPluginInfo(pluginName)
		if err == nil {
			var secretFields []string
			for _, field := range info.ConfigFields {
				if field.Secret {
					secretFields = append(secretFields, field.Name)
				}
			}

			// Decrypt and configure
			decryptedConfig, err := e.keyStore.DecryptPluginConfig(pluginConfig.Config, secretFields)
			if err == nil {
				if err := client.Configure(decryptedConfig); err != nil {
					// Log configuration error but continue with health check
					// (health check should still run even if configuration fails)
					fmt.Fprintf(os.Stderr, "Warning: failed to configure plugin %s for health check: %v\n", pluginName, err)
				}
			}
		}
	}

	return client.Health()
}
