package api

import (
	"net/http"
	"time"
)

// handleInsights routes insight-related requests
func (s *Server) handleInsights(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Insights retrieval not yet implemented - returns empty list
		respondJSON(w, http.StatusOK, map[string]interface{}{"insights": []interface{}{}})
	case http.MethodPost:
		s.generateInsights(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// generateInsights triggers insight generation using a processor plugin
func (s *Server) generateInsights(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Plugin       string    `json:"plugin"`
		AnalysisType string    `json:"analysis_type"`
		PeriodStart  time.Time `json:"period_start"`
		PeriodEnd    time.Time `json:"period_end"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if req.Plugin == "" || req.AnalysisType == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields: plugin, analysis_type", nil)
		return
	}

	// Get events for the period
	filters := map[string]interface{}{
		"since": req.PeriodStart,
	}
	events, err := s.eventStore.List(filters, 1000, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get events", err)
		return
	}

	// Analyze with processor
	result, err := s.executor.AnalyzeWithProcessor(req.Plugin, events, req.AnalysisType, map[string]interface{}{
		"period_start": req.PeriodStart,
		"period_end":   req.PeriodEnd,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Analysis failed", err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}
