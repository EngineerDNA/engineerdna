package api

import (
	"net/http"
	"strconv"

	"github.com/engineerdna/engineerdna/internal/db"
)

// handleAnonymizationPolicies retrieves anonymization policies
func (s *Server) handleAnonymizationPolicies(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		policies, err := s.anonStore.ListPolicies(db.MaxQueryLimit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to list policies", err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"policies": policies})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAnonymizationMappings retrieves anonymization mappings
func (s *Server) handleAnonymizationMappings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mappings, err := s.anonStore.ListMappings(db.MaxQueryLimit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to list mappings", err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"mappings": mappings})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAnonymizationAudit retrieves the anonymization audit log
func (s *Server) handleAnonymizationAudit(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	entries, err := s.auditStore.List(limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve audit log", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"entries": entries})
}
