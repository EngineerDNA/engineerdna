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
	homeDir, _ := os.UserHomeDir()
	engineerdnaDir := filepath.Join(homeDir, ".engineerdna")

	// Read host from environment (default: 127.0.0.1 for security)
	host := os.Getenv("ENGINEERDNA_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	// Read port from environment (default: 3847)
	port := os.Getenv("ENGINEERDNA_PORT")
	if port == "" {
		port = "3847"
	}

	return &Config{
		Database: DatabaseConfig{
			Path: filepath.Join(engineerdnaDir, "engineerdna.db"),
		},
		Server: ServerConfig{
			Port: port,
			Host: host,
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
