package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	Database DatabaseConfig `json:"database"`
	Server   ServerConfig   `json:"server"`
	Sync     SyncConfig     `json:"sync"`
	Plugins  PluginsConfig  `json:"plugins"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string `json:"path"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string `json:"port"`
	Host string `json:"host"`
}

// SyncConfig holds sync configuration
type SyncConfig struct {
	Interval string `json:"interval"`
	Enabled  bool   `json:"enabled"`
}

// PluginsConfig holds plugins configuration
type PluginsConfig struct {
	Directory string `json:"directory"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		// Fallback to current directory if home directory unavailable
		homeDir = "."
	}
	engineerdnaDir := filepath.Join(homeDir, ".engineerdna")

	// Port can be configured via environment variable
	port := os.Getenv("ENGINEERDNA_PORT")
	if port == "" {
		port = "3847"
	}

	// Host is hardcoded to 127.0.0.1 for V1 security model.
	// Network access requires V2 with authentication (JWT, API keys, RBAC).
	// See CLAUDE.md rule 1: Localhost-only (V1).
	const localhostOnly = "127.0.0.1"

	return &Config{
		Database: DatabaseConfig{
			Path: filepath.Join(engineerdnaDir, "engineerdna.db"),
		},
		Server: ServerConfig{
			Port: port,
			Host: localhostOnly,
		},
		Sync: SyncConfig{
			Interval: "1h",
			Enabled:  true,
		},
		Plugins: PluginsConfig{
			Directory: filepath.Join(engineerdnaDir, "plugins"),
		},
	}
}

// EnsureDirectories creates necessary directories
func (c *Config) EnsureDirectories() error {
	dirs := []string{
		filepath.Dir(c.Database.Path),
		c.Plugins.Directory,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetPluginDirs returns all plugin directories to search
func (c *Config) GetPluginDirs() []string {
	dirs := []string{
		c.Plugins.Directory, // ~/.engineerdna/plugins
		"./plugins",         // local plugins directory
	}
	return dirs
}
