package api

import (
	"net/http"
)

const (
	OnboardingKey = "onboarding_completed"
)

type OnboardingStatusResponse struct {
	Completed bool `json:"completed"`
}

func (s *Server) handleOnboardingStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	value, err := s.configStore.Get(OnboardingKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get onboarding status", err)
		return
	}

	// If key doesn't exist (empty string) or is explicitly false, check for existing data
	if value == "" || value == "false" {
		hasData := s.hasExistingData()
		if hasData {
			// User has data, mark onboarding as completed automatically
			if setErr := s.configStore.Set(OnboardingKey, "true"); setErr != nil {
				respondError(w, http.StatusInternalServerError, "Failed to set onboarding status", setErr)
				return
			}
			respondJSON(w, http.StatusOK, OnboardingStatusResponse{
				Completed: true,
			})
			return
		}
	}

	// Return current status
	completed := value == "true"

	respondJSON(w, http.StatusOK, OnboardingStatusResponse{
		Completed: completed,
	})
}

// hasExistingData checks if the user has any existing data in the system
// Returns true if events, engineers, or configured plugins exist
func (s *Server) hasExistingData() bool {
	// Check for events
	eventCount, err := s.eventStore.Count(map[string]interface{}{})
	if err == nil && eventCount > 0 {
		return true
	}

	// Check for engineers
	engineers, err := s.identityStore.ListEngineers(true, 1, 0)
	if err == nil && len(engineers) > 0 {
		return true
	}

	// Check for configured plugins
	plugins, err := s.pluginStore.List(1)
	if err == nil {
		for _, plugin := range plugins {
			if plugin.Enabled {
				return true
			}
		}
	}

	return false
}

func (s *Server) handleOnboardingComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := s.configStore.Set(OnboardingKey, "true")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save onboarding status", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
