package plugin

import (
	"fmt"

	"github.com/engineerdna/engineerdna/internal/models"
)

// ConfigurePlugin saves and configures a plugin
func (e *Executor) ConfigurePlugin(pluginName string, config map[string]string) error {
	// Get plugin info to know which fields are secret
	info, err := e.GetPluginInfo(pluginName)
	if err != nil {
		return fmt.Errorf("failed to get plugin info: %w", err)
	}

	// Identify secret fields
	var secretFields []string
	for _, field := range info.ConfigFields {
		if field.Secret {
			secretFields = append(secretFields, field.Name)
		}
	}

	// Encrypt secret fields
	encryptedConfig, err := e.keyStore.EncryptPluginConfig(config, secretFields)
	if err != nil {
		return fmt.Errorf("failed to encrypt config: %w", err)
	}

	// Save to database
	pluginConfig := &models.PluginConfig{
		Name:    pluginName,
		Type:    models.PluginType(info.Type),
		Enabled: true,
		Config:  encryptedConfig,
	}

	if err := e.pluginStore.Save(pluginConfig); err != nil {
		return fmt.Errorf("failed to save plugin config: %w", err)
	}

	// Test configuration by starting plugin and configuring it
	pluginPath, err := e.loader.GetPluginPath(pluginName)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	client, err := NewClient(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}
	defer client.Close()

	// Send decrypted config to plugin
	if err := client.Configure(config); err != nil {
		return fmt.Errorf("plugin rejected configuration: %w", err)
	}

	return nil
}
