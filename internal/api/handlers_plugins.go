package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
	sdk "github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

// handlePlugins routes plugin list requests
func (s *Server) handlePlugins(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listPlugins(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listPlugins returns all discovered plugins with their configuration status
func (s *Server) listPlugins(w http.ResponseWriter, r *http.Request) {
	// Get all discovered plugins from the loader
	loader := s.executor.GetLoader()
	discoveredPlugins, err := loader.DiscoverPlugins()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to discover plugins", err)
		return
	}

	// Get configured plugins from database
	configuredPlugins, err := s.pluginStore.List(db.MaxQueryLimit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list configured plugins", err)
		return
	}

	// Build map of configured plugins for quick lookup
	configMap := make(map[string]*models.PluginConfig)
	for _, cfg := range configuredPlugins {
		configMap[cfg.Name] = cfg
	}

	// Response structure
	type PluginStatus struct {
		Name         string            `json:"name"`
		Type         string            `json:"type"`
		Version      string            `json:"version"`
		Description  string            `json:"description"`
		Enabled      bool              `json:"enabled"`
		Configured   bool              `json:"configured"`
		LastSync     *time.Time        `json:"last_sync,omitempty"`
		Health       string            `json:"health"`
		ConfigFields []sdk.ConfigField `json:"config_fields"`
	}

	plugins := make([]PluginStatus, 0, len(discoveredPlugins))
	for name := range discoveredPlugins {
		// Get plugin info
		info, err := s.executor.GetPluginInfo(name)
		if err != nil {
			log.Printf("Warning: Failed to get info for plugin %s: %v", name, err)
			continue
		}

		// Check if configured
		cfg, isConfigured := configMap[name]

		// Determine health status
		health := "not_configured"
		enabled := false
		var lastSync *time.Time

		if isConfigured {
			enabled = cfg.Enabled
			lastSync = cfg.LastSync

			if cfg.Enabled {
				healthResult, err := s.executor.HealthCheck(name)
				if err == nil && healthResult.Healthy {
					health = "healthy"
				} else {
					health = "unhealthy"
				}
			} else {
				health = "disabled"
			}
		}

		plugins = append(plugins, PluginStatus{
			Name:         name,
			Type:         info.Type,
			Version:      info.Version,
			Description:  info.Description,
			Enabled:      enabled,
			Configured:   isConfigured,
			LastSync:     lastSync,
			Health:       health,
			ConfigFields: info.ConfigFields,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"plugins": plugins})
}

// handlePluginOperations routes plugin-specific operations
func (s *Server) handlePluginOperations(w http.ResponseWriter, r *http.Request) {
	// Extract plugin name from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/plugins/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Plugin name required", http.StatusBadRequest)
		return
	}

	pluginName := parts[0]

	// Validate plugin name to prevent path traversal
	if err := validatePluginName(pluginName); err != nil {
		http.Error(w, "Invalid plugin name", http.StatusBadRequest)
		return
	}

	// Route based on operation
	if len(parts) > 1 {
		operation := parts[1]
		switch operation {
		case "configure":
			s.configurePlugin(w, r, pluginName)
		case "sync":
			s.syncPlugin(w, r, pluginName)
		case "test":
			s.testPlugin(w, r, pluginName)
		case "enable":
			s.enablePlugin(w, r, pluginName, true)
		case "disable":
			s.enablePlugin(w, r, pluginName, false)
		default:
			http.Error(w, "Unknown operation", http.StatusBadRequest)
		}
	} else {
		// Get plugin details
		s.getPlugin(w, r, pluginName)
	}
}

// getPlugin retrieves detailed information about a specific plugin
func (s *Server) getPlugin(w http.ResponseWriter, r *http.Request, pluginName string) {
	info, err := s.executor.GetPluginInfo(pluginName)
	if err != nil {
		respondError(w, http.StatusNotFound, "Plugin not found", err)
		return
	}

	config, _ := s.pluginStore.GetByName(pluginName)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"info":   info,
		"config": config,
	})
}

// configurePlugin saves plugin configuration
func (s *Server) configurePlugin(w http.ResponseWriter, r *http.Request, pluginName string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config map[string]string
	if err := decodeAndValidateJSON(r, &config); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.executor.ConfigurePlugin(pluginName, config); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to configure plugin", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// syncPlugin triggers a data sync from a source plugin
func (s *Server) syncPlugin(w http.ResponseWriter, r *http.Request, pluginName string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get last sync time
	config, _ := s.pluginStore.GetByName(pluginName)
	since := time.Now().Add(-24 * time.Hour) // Default to 24 hours ago
	if config != nil && config.LastSync != nil {
		since = *config.LastSync
	}

	events, err := s.executor.SyncSourcePlugin(pluginName, since)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Sync failed", err)
		return
	}

	// Resolve identities for all events before saving
	for _, event := range events {
		engineerID, _ := s.identityService.ResolveIdentity(event.Source, event.Actor)
		event.EngineerID = engineerID
	}

	// Save events atomically in a transaction
	if err := s.eventStore.CreateBatch(events); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save events", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "ok",
		"events_synced": len(events),
	})
}

// testPlugin runs a health check on a plugin
func (s *Server) testPlugin(w http.ResponseWriter, r *http.Request, pluginName string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health, err := s.executor.HealthCheck(pluginName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Health check failed", err)
		return
	}

	respondJSON(w, http.StatusOK, health)
}

// enablePlugin enables or disables a plugin
func (s *Server) enablePlugin(w http.ResponseWriter, r *http.Request, pluginName string, enabled bool) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.pluginStore.SetEnabled(pluginName, enabled); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update plugin", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
