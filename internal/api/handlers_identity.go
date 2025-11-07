package api

import (
	"net/http"
	"strconv"

	"github.com/engineerdna/engineerdna/internal/models"
)

// Identity Resolution Handlers

func (s *Server) handleIdentityUnresolved(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	unresolved, err := s.identityService.GetUnresolved()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get unresolved identities", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"unresolved": unresolved,
	})
}

func (s *Server) handleIdentityResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UnresolvedID string  `json:"unresolved_id"`
		EngineerID   *string `json:"engineer_id"`
		Name         string  `json:"name"` // Optional: name for new engineer
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.UnresolvedID == "" {
		respondError(w, http.StatusBadRequest, "unresolved_id is required", nil)
		return
	}

	// Look up unresolved identity first
	unresolved, err := s.identityService.GetUnresolvedByID(req.UnresolvedID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get unresolved identity", err)
		return
	}
	if unresolved == nil {
		respondError(w, http.StatusNotFound, "Unresolved identity not found", nil)
		return
	}

	var engineerID string

	// If engineer_id is provided (not null and not empty), assign to existing engineer
	// If not provided or null, create a new engineer
	if req.EngineerID != nil && *req.EngineerID != "" {
		// Assign to existing engineer
		engineerID = *req.EngineerID
		err = s.identityService.AssignIdentity(engineerID, unresolved.Source, unresolved.Identifier)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to assign identity", err)
			return
		}
	} else {
		// Create new engineer with this identity
		// Use user-provided name if available, otherwise fall back to identifier
		name := req.Name
		if name == "" {
			name = unresolved.Identifier
		}

		// Use transaction-wrapped method for atomicity
		engineerID, err = s.identityService.CreateEngineerFromUnresolved(
			name,                  // name (user-provided or identifier)
			"",                    // email (empty)
			"",                    // manager (empty)
			unresolved.Source,     // source
			unresolved.Identifier, // identifier
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create engineer", err)
			return
		}
	}

	// Get the engineer to return
	engineer, err := s.identityService.GetEngineer(engineerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated engineer", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"engineer": engineer,
	})
}

func (s *Server) handleIdentitySuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	unresolvedID := r.URL.Query().Get("unresolved_id")
	engineerID := r.URL.Query().Get("engineer_id")
	minConfidenceStr := r.URL.Query().Get("min_confidence")

	suggestions, err := s.identityService.SuggestMatches()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate suggestions", err)
		return
	}

	// Filter suggestions based on query parameters
	filtered := suggestions
	if unresolvedID != "" {
		filtered = filterSuggestionsByUnresolved(filtered, unresolvedID)
	}
	if engineerID != "" {
		filtered = filterSuggestionsByEngineer(filtered, engineerID)
	}
	if minConfidenceStr != "" {
		if minConfidence, err := strconv.ParseFloat(minConfidenceStr, 64); err == nil {
			filtered = filterSuggestionsByConfidence(filtered, minConfidence)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"suggestions": filtered,
	})
}

func filterSuggestionsByUnresolved(suggestions []*models.MatchSuggestion, unresolvedID string) []*models.MatchSuggestion {
	result := []*models.MatchSuggestion{}
	for _, s := range suggestions {
		if s.UnresolvedID == unresolvedID {
			result = append(result, s)
		}
	}
	return result
}

func filterSuggestionsByEngineer(suggestions []*models.MatchSuggestion, engineerID string) []*models.MatchSuggestion {
	result := []*models.MatchSuggestion{}
	for _, s := range suggestions {
		if s.EngineerID == engineerID {
			result = append(result, s)
		}
	}
	return result
}

func filterSuggestionsByConfidence(suggestions []*models.MatchSuggestion, minConfidence float64) []*models.MatchSuggestion {
	result := []*models.MatchSuggestion{}
	for _, s := range suggestions {
		if s.Confidence >= minConfidence {
			result = append(result, s)
		}
	}
	return result
}

func (s *Server) handleIdentityMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		KeepID  string `json:"keep_id"`
		MergeID string `json:"merge_id"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.KeepID == "" || req.MergeID == "" {
		respondError(w, http.StatusBadRequest, "keep_id and merge_id are required", nil)
		return
	}

	// Prevent self-merge
	if req.KeepID == req.MergeID {
		respondError(w, http.StatusBadRequest, "Cannot merge an engineer with itself", nil)
		return
	}

	err := s.identityService.MergeEngineers(req.KeepID, req.MergeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to merge engineers", err)
		return
	}

	// Get updated engineer to return in response
	keepEngineer, err := s.identityService.GetEngineer(req.KeepID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated engineer", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"engineer": keepEngineer,
	})
}

func (s *Server) handleIdentityIgnore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UnresolvedID string `json:"unresolved_id"`
		Ignored      bool   `json:"ignored"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.UnresolvedID == "" {
		respondError(w, http.StatusBadRequest, "unresolved_id is required", nil)
		return
	}

	err := s.identityService.SetUnresolvedIgnored(req.UnresolvedID, req.Ignored)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to set ignored status", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (s *Server) handleIdentityIgnoredList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ignored, err := s.identityService.GetIgnored()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get ignored identities", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"ignored": ignored,
	})
}
