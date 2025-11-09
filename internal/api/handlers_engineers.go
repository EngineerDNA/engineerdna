package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/engineerdna/engineerdna/internal/db"
)

// Engineer Management Handlers

func (s *Server) handleEngineers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listEngineers(w, r)
	case http.MethodPost:
		s.createEngineer(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listEngineers(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") != "false"

	// Parse pagination parameters
	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = db.DefaultQueryLimit
	}
	// Cap at MaxQueryLimit to prevent resource exhaustion
	if limit > db.MaxQueryLimit {
		limit = db.MaxQueryLimit
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	engineers, err := s.identityService.ListEngineers(activeOnly, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list engineers", err)
		return
	}

	total, err := s.identityService.CountEngineers(activeOnly)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to count engineers", err)
		return
	}

	respondPaginated(w, engineers, total, limit, offset)
}

func (s *Server) createEngineer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string            `json:"name"`
		Email       string            `json:"email"`
		Manager     string            `json:"manager"`
		Identifiers map[string]string `json:"identifiers"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required", nil)
		return
	}

	engineerID, err := s.identityService.CreateEngineer(req.Name, req.Email, req.Manager, req.Identifiers)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create engineer", err)
		return
	}

	engineer, err := s.identityService.GetEngineer(engineerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get created engineer", err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"engineer": engineer,
	})
}

func (s *Server) handleEngineerOperations(w http.ResponseWriter, r *http.Request) {
	// Extract engineer ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/engineers/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Engineer ID required", http.StatusBadRequest)
		return
	}

	engineerID := parts[0]

	// Route based on operation
	if len(parts) > 1 {
		operation := parts[1]
		switch operation {
		case "activity":
			s.getEngineerActivity(w, r, engineerID)
		case "skills":
			s.handleEngineerSkillOperations(w, r, engineerID, parts)
		default:
			http.Error(w, "Unknown operation", http.StatusBadRequest)
		}
	} else {
		// Single engineer operations
		switch r.Method {
		case http.MethodGet:
			s.getEngineer(w, r, engineerID)
		case http.MethodPut:
			s.updateEngineer(w, r, engineerID)
		case http.MethodDelete:
			s.deleteEngineer(w, r, engineerID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (s *Server) getEngineer(w http.ResponseWriter, r *http.Request, engineerID string) {
	engineer, err := s.identityService.GetEngineer(engineerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get engineer", err)
		return
	}

	if engineer == nil {
		http.Error(w, "Engineer not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"engineer": engineer,
	})
}

func (s *Server) updateEngineer(w http.ResponseWriter, r *http.Request, engineerID string) {
	var req struct {
		Name        string            `json:"name"`
		Email       string            `json:"email"`
		Manager     string            `json:"manager"`
		Identifiers map[string]string `json:"identifiers"`
		Active      *bool             `json:"active"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required", nil)
		return
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	err := s.identityService.UpdateEngineer(engineerID, req.Name, req.Email, req.Manager, req.Identifiers, active)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update engineer", err)
		return
	}

	engineer, err := s.identityService.GetEngineer(engineerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated engineer", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"engineer": engineer,
	})
}

func (s *Server) deleteEngineer(w http.ResponseWriter, r *http.Request, engineerID string) {
	err := s.identityService.DeleteEngineer(engineerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete engineer", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) getEngineerActivity(w http.ResponseWriter, r *http.Request, engineerID string) {
	// Parse days parameter with bounds
	daysStr := r.URL.Query().Get("days")
	days := DefaultActivityDays
	if daysStr != "" {
		if parsed, err := strconv.Atoi(daysStr); err == nil && parsed > 0 {
			days = parsed
			// Cap at MaxActivityDays to prevent excessive database queries
			if days > MaxActivityDays {
				days = MaxActivityDays
			}
		}
	}

	metrics, err := s.identityService.GetEngineerActivity(engineerID, days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get engineer activity", err)
		return
	}

	// Transform to frontend-expected format
	response := map[string]interface{}{
		"engineer": map[string]interface{}{
			"id":   metrics.EngineerID,
			"name": metrics.EngineerName,
		},
		"period": map[string]interface{}{
			"start": metrics.PeriodStart.Format("2006-01-02T15:04:05Z07:00"),
			"end":   metrics.PeriodEnd.Format("2006-01-02T15:04:05Z07:00"),
		},
		"metrics": map[string]interface{}{
			"pull_requests": metrics.PullRequests,
			"reviews":       metrics.CodeReviews,
			"issues":        metrics.Issues,
			"commits":       metrics.Commits,
		},
		"recent_events": []interface{}{},
	}

	respondJSON(w, http.StatusOK, response)
}
