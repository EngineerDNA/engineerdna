package api

import (
	"net/http"
	"time"
)

// handleActivityToday returns today's engineering activity events
func (s *Server) handleActivityToday(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get today's start in UTC
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Fetch today's events (limit 50 for recent activity)
	filters := map[string]interface{}{
		"since": startOfDay,
	}

	events, err := s.eventStore.List(filters, 50, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get today's activity", err)
		return
	}

	// Return response in the format expected by frontend
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
	})
}
