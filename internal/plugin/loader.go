package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/engineerdna/engineerdna/internal/models"
)

// Loader discovers and loads plugins from the filesystem
type Loader struct {
	pluginDirs []string
}

// NewLoader creates a new plugin loader
func NewLoader(pluginDirs []string) *Loader {
	return &Loader{
		pluginDirs: pluginDirs,
	}
}

// DiscoverPlugins finds all available plugins in the plugin directories
func (l *Loader) DiscoverPlugins() (map[string]*PluginEntry, error) {
	plugins := make(map[string]*PluginEntry)

	for _, dir := range l.pluginDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("failed to read plugin directory %s: %w", dir, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			pluginDir := filepath.Join(dir, entry.Name())
			pluginEntry, err := l.loadPluginEntry(pluginDir, entry.Name())
			if err != nil {
				fmt.Printf("Warning: failed to load plugin %s: %v\n", entry.Name(), err)
				continue
			}

			plugins[entry.Name()] = pluginEntry
		}
	}

	return plugins, nil
}

// GetPluginPath returns the executable path for a named plugin
func (l *Loader) GetPluginPath(name string) (string, error) {
	// Validate plugin name to prevent path traversal
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return "", fmt.Errorf("invalid plugin name: %s (must match ^[a-zA-Z0-9_-]+$)", name)
	}

	for _, dir := range l.pluginDirs {
		pluginDir := filepath.Join(dir, name)

		// Check if directory exists
		if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
			continue
		}

		// Try common executable names
		candidates := []string{
			filepath.Join(pluginDir, name),
			filepath.Join(pluginDir, "main"),
			filepath.Join(pluginDir, name+".exe"),
			filepath.Join(pluginDir, "main.exe"),
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				// Verify it's not a symlink to prevent symlink-based code execution
				// Resolve symlink and ensure it stays within plugin directory
				resolvedPath, err := filepath.EvalSymlinks(candidate)
				if err != nil {
					continue // Skip if can't resolve
				}

				// Ensure resolved path is within the plugin directory
				// Use filepath.Rel to prevent prefix matching attacks (e.g., /plugins/foo vs /plugins/foo-evil)
				relPath, err := filepath.Rel(pluginDir, resolvedPath)
				if err != nil || strings.HasPrefix(relPath, "..") {
					continue // Skip if symlink points outside plugin dir
				}

				return candidate, nil
			}
		}
	}

	return "", fmt.Errorf("plugin executable not found: %s", name)
}

func (l *Loader) loadPluginEntry(pluginDir, name string) (*PluginEntry, error) {
	// Read plugin.json metadata
	metadataPath := filepath.Join(pluginDir, "plugin.json")

	// Check file size before reading to prevent memory exhaustion
	fileInfo, err := os.Stat(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat plugin.json: %w", err)
	}
	const maxMetadataSize = 1 * 1024 * 1024 // 1MB should be more than enough for plugin metadata
	if fileInfo.Size() > maxMetadataSize {
		return nil, fmt.Errorf("plugin.json too large: %d bytes (max %d / 1MB)", fileInfo.Size(), maxMetadataSize)
	}

	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin.json: %w", err)
	}

	var metadata models.PluginMetadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse plugin.json: %w", err)
	}

	// Find executable
	execPath, err := l.GetPluginPath(name)
	if err != nil {
		return nil, err
	}

	return &PluginEntry{
		Metadata: metadata,
		ExecPath: execPath,
	}, nil
}

// PluginEntry represents a discovered plugin
type PluginEntry struct {
	Metadata models.PluginMetadata
	ExecPath string
}
