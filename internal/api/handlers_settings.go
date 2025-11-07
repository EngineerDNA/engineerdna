package api

import (
	"net/http"
)

// Settings Handlers

// handleSettings returns the current settings
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	settings, err := s.settingsStore.GetSettings()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve settings", err)
		return
	}

	respondJSON(w, http.StatusOK, settings)
}

func (s *Server) handleSettingsSyncSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Schedule string `json:"schedule"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate schedule value
	validSchedules := map[string]bool{
		"manual": true,
		"hourly": true,
		"daily":  true,
		"weekly": true,
	}

	if !validSchedules[req.Schedule] {
		respondError(w, http.StatusBadRequest, "Invalid schedule value. Must be one of: manual, hourly, daily, weekly", nil)
		return
	}

	// Update in settings table
	err := s.settingsStore.UpdateSyncSchedule(req.Schedule)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save sync schedule", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"schedule": req.Schedule,
	})
}

func (s *Server) handleSettingsAnonymizationStrategy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Strategy string `json:"strategy"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate strategy value
	validStrategies := map[string]bool{
		"sequential": true,
		"uuid":       true,
		"hash":       true,
	}

	if !validStrategies[req.Strategy] {
		respondError(w, http.StatusBadRequest, "Invalid strategy value. Must be one of: sequential, uuid, hash", nil)
		return
	}

	// Update in settings table
	err := s.settingsStore.UpdateAnonymizationStrategy(req.Strategy)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save anonymization strategy", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"strategy": req.Strategy,
	})
}

func (s *Server) handleSettingsPort(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Port int `json:"port"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate port range
	if req.Port < 1024 || req.Port > 65535 {
		respondError(w, http.StatusBadRequest, "Port must be between 1024 and 65535", nil)
		return
	}

	// Update in settings table
	err := s.settingsStore.UpdatePort(req.Port)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save port", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"port":    req.Port,
		"message": "Port change requires server restart to take effect",
	})
}
