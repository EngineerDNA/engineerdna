package config

import "time"

const (
	// DefaultTimeout is the default timeout for database and API operations
	DefaultTimeout = 30 * time.Second

	// PluginTimeout is the timeout for plugin execution
	PluginTimeout = 30 * time.Second
)
