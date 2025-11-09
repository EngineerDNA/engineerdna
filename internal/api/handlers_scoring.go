package api

import (
	"net/http"
	"strings"
	"time"
)

// Roles Handlers

func (s *Server) handleRoles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listRoles(w, r)
	case http.MethodPost:
		s.createRole(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listRoles(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	roles, err := s.scoringStore.ListRoles(limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list roles", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"roles": roles,
	})
}

func (s *Server) createRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string             `json:"name"`
		TargetScore  int                `json:"target_score"`
		Expectations map[string]float64 `json:"expectations"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required", nil)
		return
	}

	if req.Expectations == nil {
		req.Expectations = make(map[string]float64)
	}

	role, err := s.scoringStore.CreateRole(req.Name, req.TargetScore, req.Expectations)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create role", err)
		return
	}

	respondJSON(w, http.StatusCreated, role)
}

func (s *Server) handleRoleByID(w http.ResponseWriter, r *http.Request) {
	// Extract role ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/roles/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Role ID required", http.StatusBadRequest)
		return
	}

	roleID := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.getRole(w, r, roleID)
	case http.MethodPut:
		s.updateRole(w, r, roleID)
	case http.MethodDelete:
		s.deleteRole(w, r, roleID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getRole(w http.ResponseWriter, r *http.Request, roleID string) {
	role, err := s.scoringStore.GetRole(roleID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get role", err)
		return
	}

	if role == nil {
		http.Error(w, "Role not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, role)
}

func (s *Server) updateRole(w http.ResponseWriter, r *http.Request, roleID string) {
	var req struct {
		Name         string             `json:"name"`
		TargetScore  int                `json:"target_score"`
		Expectations map[string]float64 `json:"expectations"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required", nil)
		return
	}

	if req.Expectations == nil {
		req.Expectations = make(map[string]float64)
	}

	err := s.scoringStore.UpdateRole(roleID, req.Name, req.TargetScore, req.Expectations)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update role", err)
		return
	}

	// Fetch updated role
	role, err := s.scoringStore.GetRole(roleID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated role", err)
		return
	}

	respondJSON(w, http.StatusOK, role)
}

func (s *Server) deleteRole(w http.ResponseWriter, r *http.Request, roleID string) {
	err := s.scoringStore.DeleteRole(roleID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete role", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Scoring Weights Handlers

func (s *Server) handleScoringWeights(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getScoringWeights(w, r)
	case http.MethodPut:
		s.updateScoringWeights(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getScoringWeights(w http.ResponseWriter, r *http.Request) {
	weights, err := s.scoringStore.GetWeights()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get scoring weights", err)
		return
	}

	if weights == nil {
		// Return default weights if none configured
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"throughput_weight":    30.0,
			"quality_weight":       25.0,
			"speed_weight":         20.0,
			"collaboration_weight": 15.0,
			"impact_weight":        10.0,
		})
		return
	}

	respondJSON(w, http.StatusOK, weights)
}

func (s *Server) updateScoringWeights(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ThroughputWeight    float64 `json:"throughput_weight"`
		QualityWeight       float64 `json:"quality_weight"`
		SpeedWeight         float64 `json:"speed_weight"`
		CollaborationWeight float64 `json:"collaboration_weight"`
		ImpactWeight        float64 `json:"impact_weight"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate weights are non-negative
	if req.ThroughputWeight < 0 || req.QualityWeight < 0 || req.SpeedWeight < 0 || req.CollaborationWeight < 0 || req.ImpactWeight < 0 {
		respondError(w, http.StatusBadRequest, "Weights must be non-negative", nil)
		return
	}

	weights, err := s.scoringStore.UpdateWeights(
		req.ThroughputWeight,
		req.QualityWeight,
		req.SpeedWeight,
		req.CollaborationWeight,
		req.ImpactWeight,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update weights", err)
		return
	}

	respondJSON(w, http.StatusOK, weights)
}

// Score Calculation Handler

func (s *Server) handleCalculateScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		EngineerID string `json:"engineer_id"`
		WeekStart  string `json:"week_start"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.EngineerID == "" {
		respondError(w, http.StatusBadRequest, "engineer_id is required", nil)
		return
	}

	if req.WeekStart == "" {
		respondError(w, http.StatusBadRequest, "week_start is required", nil)
		return
	}

	// Parse week_start date
	weekStart, err := time.Parse("2006-01-02", req.WeekStart)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid week_start format (expected YYYY-MM-DD)", err)
		return
	}

	// Calculate score
	score, err := s.scoringService.CalculateIndividualScore(req.EngineerID, weekStart)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to calculate score", err)
		return
	}

	respondJSON(w, http.StatusOK, score)
}

// Performance Scores Handlers

func (s *Server) handlePerformanceIndividual(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract engineer ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/performance/individual/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Engineer ID required", http.StatusBadRequest)
		return
	}

	engineerID := parts[0]

	// Parse date range from query parameters
	query := r.URL.Query()
	endDate := time.Now().UTC()
	startDate := endDate.AddDate(0, 0, -90)

	if startDateStr := query.Get("start_date"); startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid start_date format (expected YYYY-MM-DD)", err)
			return
		}
		startDate = parsed
	}

	if endDateStr := query.Get("end_date"); endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid end_date format (expected YYYY-MM-DD)", err)
			return
		}
		endDate = parsed
	}

	// Get scores from metric_values
	metrics, err := s.metricStore.GetEngineerScores(engineerID, startDate, endDate, 100)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get performance scores", err)
		return
	}

	// Return metric_values directly (no legacy conversion)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"engineer_id": engineerID,
		"start_date":  startDate.Format("2006-01-02"),
		"end_date":    endDate.Format("2006-01-02"),
		"metrics":     metrics,
	})
}

func (s *Server) handlePerformanceTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract team ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/performance/team/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Team ID required", http.StatusBadRequest)
		return
	}

	teamID := parts[0]

	// Placeholder for Phase 2 - team scoring not yet implemented
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"team_id": teamID,
		"message": "Team performance scoring is not yet implemented (Phase 2)",
		"scores":  []interface{}{},
	})
}
