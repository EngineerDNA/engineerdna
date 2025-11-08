package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Goals Handlers

func (s *Server) handleGoals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listGoals(w, r)
	case http.MethodPost:
		s.createGoal(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listGoals(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	ownerType := r.URL.Query().Get("owner_type")
	ownerID := r.URL.Query().Get("owner_id")
	status := r.URL.Query().Get("status")
	timePeriod := r.URL.Query().Get("time_period")

	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = db.DefaultQueryLimit
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	goals, total, err := s.goalsStore.ListGoals(ownerType, ownerID, status, timePeriod, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list goals", err)
		return
	}

	respondPaginated(w, goals, total, limit, offset)
}

func (s *Server) createGoal(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title           string  `json:"title"`
		Description     string  `json:"description"`
		GoalType        string  `json:"goal_type"`
		OwnerType       string  `json:"owner_type"`
		OwnerID         *string `json:"owner_id"`
		TimePeriod      string  `json:"time_period"`
		StartDate       string  `json:"start_date"`
		EndDate         string  `json:"end_date"`
		TrackingMethod  string  `json:"tracking_method"`
		SuccessCriteria string  `json:"success_criteria"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "title is required", nil)
		return
	}

	if req.StartDate == "" || req.EndDate == "" {
		respondError(w, http.StatusBadRequest, "start_date and end_date are required", nil)
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid start_date format (expected YYYY-MM-DD)", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid end_date format (expected YYYY-MM-DD)", err)
		return
	}

	// Default values
	if req.TrackingMethod == "" {
		req.TrackingMethod = "manual"
	}
	if req.GoalType == "" {
		req.GoalType = "individual"
	}
	if req.OwnerType == "" {
		req.OwnerType = "engineer"
	}

	goal := &models.Goal{
		Title:           req.Title,
		Description:     req.Description,
		GoalType:        req.GoalType,
		OwnerType:       req.OwnerType,
		OwnerID:         req.OwnerID,
		TimePeriod:      req.TimePeriod,
		StartDate:       startDate.UTC(),
		EndDate:         endDate.UTC(),
		TrackingMethod:  req.TrackingMethod,
		SuccessCriteria: req.SuccessCriteria,
		Status:          "active",
	}

	if err := s.goalsStore.CreateGoal(goal); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create goal", err)
		return
	}

	respondJSON(w, http.StatusCreated, goal)
}

func (s *Server) handleGoalByID(w http.ResponseWriter, r *http.Request) {
	// Extract goal ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/goals/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Goal ID required", http.StatusBadRequest)
		return
	}

	goalID := parts[0]

	// Route based on operation
	if len(parts) > 1 {
		operation := parts[1]
		switch operation {
		case "milestones":
			s.handleGoalMilestones(w, r, goalID, parts)
		case "progress":
			s.handleGoalProgress(w, r, goalID)
		case "dependencies":
			s.handleGoalDependencies(w, r, goalID)
		case "metrics":
			s.handleGoalMetrics(w, r, goalID)
		default:
			http.Error(w, "Unknown operation", http.StatusBadRequest)
		}
	} else {
		// Single goal operations
		switch r.Method {
		case http.MethodGet:
			s.getGoal(w, r, goalID)
		case http.MethodPut:
			s.updateGoal(w, r, goalID)
		case http.MethodDelete:
			s.deleteGoal(w, r, goalID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (s *Server) getGoal(w http.ResponseWriter, r *http.Request, goalID string) {
	// Get with details
	details, err := s.goalsService.GetGoalWithDetails(goalID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get goal", err)
		return
	}

	respondJSON(w, http.StatusOK, details)
}

func (s *Server) updateGoal(w http.ResponseWriter, r *http.Request, goalID string) {
	goal, err := s.goalsStore.GetGoal(goalID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get goal", err)
		return
	}
	if goal == nil {
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}

	var req struct {
		Title              *string `json:"title"`
		Description        *string `json:"description"`
		Status             *string `json:"status"`
		ProgressPercentage *int    `json:"progress_percentage"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Update fields
	if req.Title != nil {
		goal.Title = *req.Title
	}
	if req.Description != nil {
		goal.Description = *req.Description
	}
	if req.Status != nil {
		goal.Status = *req.Status
	}
	if req.ProgressPercentage != nil {
		goal.ProgressPercentage = *req.ProgressPercentage
	}

	if err := s.goalsStore.UpdateGoal(goal); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update goal", err)
		return
	}

	respondJSON(w, http.StatusOK, goal)
}

func (s *Server) deleteGoal(w http.ResponseWriter, r *http.Request, goalID string) {
	if err := s.goalsStore.DeleteGoal(goalID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete goal", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Goal deleted successfully",
	})
}

// Milestone handlers

func (s *Server) handleGoalMilestones(w http.ResponseWriter, r *http.Request, goalID string, parts []string) {
	if len(parts) > 2 {
		// Specific milestone operation
		milestoneID := parts[2]
		if len(parts) > 3 && parts[3] == "complete" {
			s.completeMilestone(w, r, milestoneID)
			return
		}
		s.updateMilestoneByID(w, r, milestoneID)
	} else {
		// List or create milestones
		switch r.Method {
		case http.MethodGet:
			s.listMilestones(w, r, goalID)
		case http.MethodPost:
			s.createMilestone(w, r, goalID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (s *Server) listMilestones(w http.ResponseWriter, r *http.Request, goalID string) {
	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = db.DefaultQueryLimit
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	milestones, total, err := s.goalsStore.GetMilestones(goalID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get milestones", err)
		return
	}

	respondPaginated(w, milestones, total, limit, offset)
}

func (s *Server) createMilestone(w http.ResponseWriter, r *http.Request, goalID string) {
	var req struct {
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		TargetValue  *float64 `json:"target_value"`
		CurrentValue *float64 `json:"current_value"`
		Unit         string   `json:"unit"`
		DueDate      *string  `json:"due_date"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "title is required", nil)
		return
	}

	milestone := &models.GoalMilestone{
		GoalID:      goalID,
		Title:       req.Title,
		Description: req.Description,
		Unit:        req.Unit,
	}

	if req.TargetValue != nil {
		milestone.TargetValue = *req.TargetValue
	}
	if req.CurrentValue != nil {
		milestone.CurrentValue = *req.CurrentValue
	}
	if req.DueDate != nil {
		dueDate, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid due_date format", err)
			return
		}
		milestone.DueDate = &dueDate
	}

	if err := s.goalsStore.CreateMilestone(milestone); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create milestone", err)
		return
	}

	respondJSON(w, http.StatusCreated, milestone)
}

func (s *Server) updateMilestoneByID(w http.ResponseWriter, r *http.Request, milestoneID string) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Title        string     `json:"title"`
		Description  string     `json:"description"`
		TargetValue  *float64   `json:"target_value"`
		CurrentValue *float64   `json:"current_value"`
		Unit         string     `json:"unit"`
		DueDate      *time.Time `json:"due_date"`
		Completed    *bool      `json:"completed"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get existing milestone
	milestone, err := s.goalsStore.GetMilestoneByID(milestoneID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Milestone not found", err)
		return
	}

	// Update fields
	if req.Title != "" {
		milestone.Title = req.Title
	}
	if req.Description != "" {
		milestone.Description = req.Description
	}
	if req.TargetValue != nil {
		milestone.TargetValue = *req.TargetValue
	}
	if req.CurrentValue != nil {
		milestone.CurrentValue = *req.CurrentValue
	}
	if req.Unit != "" {
		milestone.Unit = req.Unit
	}
	if req.DueDate != nil {
		milestone.DueDate = req.DueDate
	}

	// Update in database
	if err := s.goalsStore.UpdateMilestone(milestone); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update milestone", err)
		return
	}

	respondJSON(w, http.StatusOK, milestone)
}

func (s *Server) completeMilestone(w http.ResponseWriter, r *http.Request, milestoneID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.goalsService.CompleteMilestone(milestoneID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to complete milestone", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Milestone completed successfully",
	})
}

// Progress handlers

func (s *Server) handleGoalProgress(w http.ResponseWriter, r *http.Request, goalID string) {
	switch r.Method {
	case http.MethodGet:
		s.getGoalProgress(w, r, goalID)
	case http.MethodPost:
		s.logGoalProgress(w, r, goalID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getGoalProgress(w http.ResponseWriter, r *http.Request, goalID string) {
	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = 20
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	logs, total, err := s.goalsStore.GetProgressLogs(goalID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get progress logs", err)
		return
	}

	respondPaginated(w, logs, total, limit, offset)
}

func (s *Server) logGoalProgress(w http.ResponseWriter, r *http.Request, goalID string) {
	var req struct {
		MilestoneID   *string  `json:"milestone_id"`
		PreviousValue *float64 `json:"previous_value"`
		NewValue      *float64 `json:"new_value"`
		ChangeType    string   `json:"change_type"`
		Evidence      string   `json:"evidence"`
		LoggedBy      *string  `json:"logged_by"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	log := &models.GoalProgressLog{
		GoalID:        goalID,
		MilestoneID:   req.MilestoneID,
		PreviousValue: req.PreviousValue,
		NewValue:      req.NewValue,
		ChangeType:    req.ChangeType,
		Evidence:      req.Evidence,
		LoggedBy:      req.LoggedBy,
	}

	if err := s.goalsStore.LogProgress(log); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to log progress", err)
		return
	}

	// Update goal progress
	if err := s.goalsService.UpdateGoalProgress(goalID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update goal progress", err)
		return
	}

	respondJSON(w, http.StatusCreated, log)
}

// Dependencies and metrics handlers

func (s *Server) handleGoalDependencies(w http.ResponseWriter, r *http.Request, goalID string) {
	switch r.Method {
	case http.MethodGet:
		limit, err := parseLimitParam(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
			return
		}
		if limit == 0 {
			limit = 100
		}

		offset, err := parseOffsetParam(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
			return
		}

		deps, total, err := s.goalsStore.GetDependencies(goalID, limit, offset)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to get dependencies", err)
			return
		}
		respondPaginated(w, deps, total, limit, offset)
	case http.MethodPost:
		var req struct {
			DependsOnGoalID string `json:"depends_on_goal_id"`
			DependencyType  string `json:"dependency_type"`
		}
		if err := decodeAndValidateJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}
		dep := &models.GoalDependency{
			GoalID:          goalID,
			DependsOnGoalID: req.DependsOnGoalID,
			DependencyType:  req.DependencyType,
			Status:          "active",
		}
		if err := s.goalsStore.CreateDependency(dep); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create dependency", err)
			return
		}
		respondJSON(w, http.StatusCreated, dep)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGoalMetrics(w http.ResponseWriter, r *http.Request, goalID string) {
	switch r.Method {
	case http.MethodGet:
		limit, err := parseLimitParam(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
			return
		}
		if limit == 0 {
			limit = 100
		}

		offset, err := parseOffsetParam(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
			return
		}

		metrics, total, err := s.goalsStore.GetGoalMetrics(goalID, limit, offset)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to get metrics", err)
			return
		}
		respondPaginated(w, metrics, total, limit, offset)
	case http.MethodPost:
		var req struct {
			MetricName  string  `json:"metric_name"`
			TargetValue float64 `json:"target_value"`
			Operator    string  `json:"operator"`
		}
		if err := decodeAndValidateJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}
		metric := &models.GoalMetric{
			GoalID:      goalID,
			MetricName:  req.MetricName,
			TargetValue: req.TargetValue,
			Operator:    req.Operator,
		}
		if err := s.goalsStore.CreateGoalMetric(metric); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create metric", err)
			return
		}
		respondJSON(w, http.StatusCreated, metric)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Summary handlers

func (s *Server) handleGoalsSummary(w http.ResponseWriter, r *http.Request) {
	ownerType := r.URL.Query().Get("owner_type")
	ownerID := r.URL.Query().Get("owner_id")

	summary, err := s.goalsService.GetGoalSummary(ownerType, ownerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get goal summary", err)
		return
	}

	respondJSON(w, http.StatusOK, summary)
}
