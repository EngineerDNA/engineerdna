package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/actions"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Recommendation Handlers

func (s *Server) handleRecommendations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listRecommendations(w, r)
	case http.MethodPost:
		s.createRecommendation(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listRecommendations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filters := make(map[string]string)

	if assignedTo := query.Get("assigned_to"); assignedTo != "" {
		filters["assigned_to"] = assignedTo
	}
	if status := query.Get("status"); status != "" {
		filters["status"] = status
	}
	if subjectType := query.Get("subject_type"); subjectType != "" {
		filters["subject_type"] = subjectType
	}
	if subjectID := query.Get("subject_id"); subjectID != "" {
		filters["subject_id"] = subjectID
	}
	if sourceType := query.Get("source_type"); sourceType != "" {
		filters["source_type"] = sourceType
	}

	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	recommendations, total, err := s.actionsStore.ListRecommendations(filters, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list recommendations", err)
		return
	}

	respondPaginated(w, recommendations, total, limit, offset)
}

func (s *Server) createRecommendation(w http.ResponseWriter, r *http.Request) {
	var rec models.Recommendation

	if err := decodeAndValidateJSON(r, &rec); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if rec.Title == "" {
		respondError(w, http.StatusBadRequest, "title is required", nil)
		return
	}

	if rec.RecommendationType == "" {
		respondError(w, http.StatusBadRequest, "recommendation_type is required", nil)
		return
	}

	if rec.Priority == "" {
		respondError(w, http.StatusBadRequest, "priority is required", nil)
		return
	}

	if err := s.actionsStore.CreateRecommendation(&rec); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create recommendation", err)
		return
	}

	respondJSON(w, http.StatusCreated, rec)
}

func (s *Server) handleRecommendationByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/recommendations/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Recommendation ID required", http.StatusBadRequest)
		return
	}

	recommendationID := parts[0]

	if r.Method == http.MethodGet {
		s.getRecommendation(w, r, recommendationID)
		return
	}

	if len(parts) < 2 {
		http.Error(w, "Action required", http.StatusBadRequest)
		return
	}

	action := parts[1]

	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch action {
	case "status":
		if r.Method == http.MethodPut {
			s.updateRecommendationStatus(w, r, recommendationID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "dismiss":
		s.dismissRecommendation(w, r, recommendationID)
	case "snooze":
		s.snoozeRecommendation(w, r, recommendationID)
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}

func (s *Server) getRecommendation(w http.ResponseWriter, r *http.Request, recommendationID string) {
	// Create recommendation service
	recService := actions.NewRecommendationService(s.actionsStore)

	recWithDetails, err := recService.GetRecommendationWithDetails(recommendationID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get recommendation", err)
		return
	}

	respondJSON(w, http.StatusOK, recWithDetails)
}

func (s *Server) updateRecommendationStatus(w http.ResponseWriter, r *http.Request, recommendationID string) {
	var req struct {
		Status    string `json:"status"`
		ChangedBy string `json:"changed_by"`
		Reason    string `json:"reason"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Status == "" {
		respondError(w, http.StatusBadRequest, "status is required", nil)
		return
	}

	recService := actions.NewRecommendationService(s.actionsStore)

	if err := recService.UpdateRecommendationStatus(recommendationID, req.Status, req.ChangedBy, req.Reason); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update recommendation status", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) dismissRecommendation(w http.ResponseWriter, r *http.Request, recommendationID string) {
	var req struct {
		ManagerID string `json:"manager_id"`
		Reason    string `json:"reason"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	recService := actions.NewRecommendationService(s.actionsStore)

	if err := recService.DismissRecommendation(recommendationID, req.ManagerID, req.Reason); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to dismiss recommendation", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) snoozeRecommendation(w http.ResponseWriter, r *http.Request, recommendationID string) {
	var req struct {
		ManagerID string `json:"manager_id"`
		Duration  int    `json:"duration"` // hours
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Duration <= 0 {
		req.Duration = 24 // Default 1 day
	}

	recService := actions.NewRecommendationService(s.actionsStore)

	duration := time.Duration(req.Duration) * time.Hour
	if err := recService.SnoozeRecommendation(recommendationID, req.ManagerID, duration); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to snooze recommendation", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleRecommendationsPending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	managerID := r.URL.Query().Get("manager_id")
	if managerID == "" {
		respondError(w, http.StatusBadRequest, "manager_id query parameter is required", nil)
		return
	}

	recService := actions.NewRecommendationService(s.actionsStore)

	recommendations, err := recService.GetPendingRecommendations(managerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get pending recommendations", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"recommendations": recommendations,
	})
}

// Action Handlers

func (s *Server) handleActions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listActions(w, r)
	case http.MethodPost:
		s.createAction(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listActions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filters := make(map[string]string)

	if takenBy := query.Get("taken_by"); takenBy != "" {
		filters["taken_by"] = takenBy
	}
	if recommendationID := query.Get("recommendation_id"); recommendationID != "" {
		filters["recommendation_id"] = recommendationID
	}
	if subjectType := query.Get("subject_type"); subjectType != "" {
		filters["subject_type"] = subjectType
	}
	if subjectID := query.Get("subject_id"); subjectID != "" {
		filters["subject_id"] = subjectID
	}

	limit := 100
	if limitStr := query.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	actions, err := s.actionsStore.ListActions(filters, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list actions", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"actions": actions,
	})
}

func (s *Server) createAction(w http.ResponseWriter, r *http.Request) {
	var action models.Action

	if err := decodeAndValidateJSON(r, &action); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if action.Title == "" {
		respondError(w, http.StatusBadRequest, "title is required", nil)
		return
	}

	if action.ActionType == "" {
		respondError(w, http.StatusBadRequest, "action_type is required", nil)
		return
	}

	if action.TakenBy == "" {
		respondError(w, http.StatusBadRequest, "taken_by is required", nil)
		return
	}

	actionService := actions.NewActionService(s.actionsStore)

	if err := actionService.RecordAction(&action); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create action", err)
		return
	}

	respondJSON(w, http.StatusCreated, action)
}

func (s *Server) handleActionByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/actions/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Action ID required", http.StatusBadRequest)
		return
	}

	actionID := parts[0]

	if r.Method == http.MethodGet && len(parts) == 1 {
		s.getAction(w, r, actionID)
		return
	}

	if len(parts) < 2 {
		http.Error(w, "Action required", http.StatusBadRequest)
		return
	}

	action := parts[1]

	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch action {
	case "outcome":
		if r.Method == http.MethodPost {
			s.recordActionOutcome(w, r, actionID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "effectiveness":
		if r.Method == http.MethodGet {
			s.getActionEffectiveness(w, r, actionID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "follow-up":
		if r.Method == http.MethodPost {
			s.scheduleFollowUp(w, r, actionID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}

func (s *Server) getAction(w http.ResponseWriter, r *http.Request, actionID string) {
	actionService := actions.NewActionService(s.actionsStore)

	actionWithDetails, err := actionService.GetActionWithDetails(actionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get action", err)
		return
	}

	respondJSON(w, http.StatusOK, actionWithDetails)
}

func (s *Server) recordActionOutcome(w http.ResponseWriter, r *http.Request, actionID string) {
	var outcome models.ActionOutcome

	if err := decodeAndValidateJSON(r, &outcome); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	outcome.ActionID = actionID

	if outcome.OutcomeType == "" {
		respondError(w, http.StatusBadRequest, "outcome_type is required", nil)
		return
	}

	if outcome.Effectiveness == "" {
		respondError(w, http.StatusBadRequest, "effectiveness is required", nil)
		return
	}

	actionService := actions.NewActionService(s.actionsStore)

	if err := actionService.RecordActionOutcome(&outcome); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to record outcome", err)
		return
	}

	respondJSON(w, http.StatusCreated, outcome)
}

func (s *Server) getActionEffectiveness(w http.ResponseWriter, r *http.Request, actionID string) {
	outcomes, err := s.actionsStore.GetActionOutcomes(actionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get action outcomes", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"action_id": actionID,
		"outcomes":  outcomes,
	})
}

func (s *Server) scheduleFollowUp(w http.ResponseWriter, r *http.Request, actionID string) {
	var req struct {
		FollowUpDate string `json:"follow_up_date"` // ISO 8601 format
		FollowUpType string `json:"follow_up_type"`
		Description  string `json:"description"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	followUpDate, err := time.Parse(time.RFC3339, req.FollowUpDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid follow_up_date format", err)
		return
	}

	if req.FollowUpType == "" {
		req.FollowUpType = "check_metric"
	}

	actionService := actions.NewActionService(s.actionsStore)

	if err := actionService.ScheduleFollowUp(actionID, followUpDate, req.FollowUpType, req.Description); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to schedule follow-up", err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// Follow-Up Handlers

func (s *Server) handleFollowUps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actionService := actions.NewActionService(s.actionsStore)

	followUps, err := actionService.GetDueFollowUps()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get follow-ups", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"follow_ups": followUps,
	})
}

func (s *Server) handleFollowUpByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/follow-ups/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] != "complete" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	followUpID := parts[0]

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actionService := actions.NewActionService(s.actionsStore)

	if err := actionService.CompleteFollowUp(followUpID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to complete follow-up", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Effectiveness Report Handler

func (s *Server) handleEffectivenessReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	managerID := r.URL.Query().Get("manager_id")
	if managerID == "" {
		respondError(w, http.StatusBadRequest, "manager_id query parameter is required", nil)
		return
	}

	days := 30
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil {
			days = parsedDays
		}
	}

	actionService := actions.NewActionService(s.actionsStore)

	report, err := actionService.GenerateActionReport(managerID, days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate effectiveness report", err)
		return
	}

	// Add patterns from analyzer
	analyzer := actions.NewEffectivenessAnalyzer(s.actionsStore)
	successfulPatterns, _ := analyzer.IdentifySuccessfulPatterns()
	unsuccessfulPatterns, _ := analyzer.IdentifyUnsuccessfulPatterns()

	report.SuccessfulPatterns = successfulPatterns
	report.UnsuccessfulPatterns = unsuccessfulPatterns

	respondJSON(w, http.StatusOK, report)
}
