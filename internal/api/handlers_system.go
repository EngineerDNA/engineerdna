package api

import (
	"log"
	"net/http"
	"os"
)

// handleHealth returns the health status of the API
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleVersion returns the current version of the application
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"version": s.version})
}

// handleSystemInfo returns system information including database stats and plugin count
func (s *Server) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	// Get database stats
	dbPath := s.config.Database.Path
	var dbSize int64
	if fileInfo, err := os.Stat(dbPath); err == nil {
		dbSize = fileInfo.Size()
	}

	// Get event count
	eventCount, err := s.eventStore.Count(map[string]interface{}{})
	if err != nil {
		log.Printf("Warning: Failed to get event count: %v", err)
		eventCount = 0
	}

	// Get plugin count
	loader := s.executor.GetLoader()
	discoveredPlugins, err := loader.DiscoverPlugins()
	pluginCount := 0
	if err == nil {
		pluginCount = len(discoveredPlugins)
	}

	// Get master key source
	masterKeySource := "environment"
	if os.Getenv("ENGINEERDNA_MASTER_KEY") == "" {
		masterKeySource = "generated"
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"database": map[string]interface{}{
			"path":        dbPath,
			"size_bytes":  dbSize,
			"event_count": eventCount,
		},
		"encryption": map[string]interface{}{
			"master_key_source": masterKeySource,
		},
		"plugins": map[string]interface{}{
			"count": pluginCount,
		},
		"version": s.version,
	})
}
