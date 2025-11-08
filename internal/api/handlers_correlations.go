package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/services"
)

// handleCorrelations lists all available correlations
func (s *Server) handleCorrelations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	correlations, err := s.correlationStore.ListCorrelations()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list correlations", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"correlations": correlations,
		"count":        len(correlations),
	})
}

// handleCorrelationByName handles requests for a specific correlation
// Path format: /api/correlations/{name}/calculate or /api/correlations/{name}/values
func (s *Server) handleCorrelationByName(w http.ResponseWriter, r *http.Request) {
	// Extract correlation name from path
	path := strings.TrimPrefix(r.URL.Path, "/api/correlations/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		respondError(w, http.StatusBadRequest, "Correlation name is required", nil)
		return
	}

	correlationName := parts[0]
	if correlationName == "" {
		respondError(w, http.StatusBadRequest, "Correlation name is required", nil)
		return
	}

	// Determine action
	var action string
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "calculate":
		s.handleCorrelationCalculate(w, r, correlationName)
	case "values":
		s.handleCorrelationValues(w, r, correlationName)
	default:
		respondError(w, http.StatusBadRequest, "Invalid action (must be: calculate, values)", nil)
	}
}

// handleCorrelationCalculate calculates a correlation on-demand
// GET /api/correlations/{name}/calculate?start=...&end=...
func (s *Server) handleCorrelationCalculate(w http.ResponseWriter, r *http.Request, correlationName string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get correlation definition from database
	correlation, err := s.correlationStore.GetCorrelationByName(correlationName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get correlation", err)
		return
	}
	if correlation == nil {
		respondError(w, http.StatusNotFound, "Correlation not found", nil)
		return
	}

	// Parse correlation definition
	definitionJSON, err := json.Marshal(correlation.Definition)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to marshal correlation definition", err)
		return
	}

	definition, err := services.ParseCorrelationDefinition(definitionJSON)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid correlation definition", err)
		return
	}

	// Parse time range parameters
	params := make(map[string]interface{})
	if startStr := r.URL.Query().Get("start"); startStr != "" {
		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid start timestamp", err)
			return
		}
		params["start"] = start
	}
	if endStr := r.URL.Query().Get("end"); endStr != "" {
		end, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid end timestamp", err)
			return
		}
		params["end"] = end
	}

	// Check if correlation engine is available
	if s.correlationEngine == nil {
		respondError(w, http.StatusInternalServerError, "Correlation engine not available", nil)
		return
	}

	// Calculate correlation
	result, err := s.correlationEngine.Calculate(definition, params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to calculate correlation", err)
		return
	}

	// Save result to database
	if err := s.correlationEngine.SaveCorrelationValue(correlation.ID, result); err != nil {
		s.logger.Printf("Failed to save correlation value: %v", err)
		// Don't fail the request if save fails
	}

	respondJSON(w, http.StatusOK, result)
}

// handleCorrelationValues retrieves historical correlation values
// GET /api/correlations/{name}/values?start=...&end=...&limit=...
func (s *Server) handleCorrelationValues(w http.ResponseWriter, r *http.Request, correlationName string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get correlation definition
	correlation, err := s.correlationStore.GetCorrelationByName(correlationName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get correlation", err)
		return
	}
	if correlation == nil {
		respondError(w, http.StatusNotFound, "Correlation not found", nil)
		return
	}

	// Parse time range (default to last 30 days)
	now := time.Now().UTC()
	start := now.AddDate(0, 0, -30)
	end := now

	if startStr := r.URL.Query().Get("start"); startStr != "" {
		parsed, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid start timestamp", err)
			return
		}
		start = parsed
	}

	if endStr := r.URL.Query().Get("end"); endStr != "" {
		parsed, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid end timestamp", err)
			return
		}
		end = parsed
	}

	// Parse limit
	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = db.DefaultQueryLimit
	}
	if limit > db.MaxQueryLimit {
		limit = db.MaxQueryLimit
	}

	// Get correlation values
	values, err := s.correlationStore.GetCorrelationValues(correlation.ID, start, end, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get correlation values", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"correlation_name": correlationName,
		"values":           values,
		"count":            len(values),
		"period": map[string]string{
			"start": start.Format(time.RFC3339),
			"end":   end.Format(time.RFC3339),
		},
	})
}
