package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Planning Handlers

func (s *Server) handleSprints(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listSprints(w, r)
	case http.MethodPost:
		s.createSprint(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listSprints(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	teamID := r.URL.Query().Get("team_id")
	status := r.URL.Query().Get("status")
	limit := 100 // Default limit

	var sprints []*models.Sprint
	var err error

	// Filter by team if provided
	if teamID != "" {
		sprints, err = s.planningStore.GetSprintsByTeam(teamID, limit)
	} else {
		sprints, err = s.planningStore.ListSprints(limit)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list sprints", err)
		return
	}

	// Filter by status if provided
	if status != "" {
		var filtered []*models.Sprint
		for _, sprint := range sprints {
			if sprint.Status == status {
				filtered = append(filtered, sprint)
			}
		}
		sprints = filtered
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"sprints": sprints,
	})
}

func (s *Server) createSprint(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string `json:"name"`
		StartDate       string `json:"start_date"`
		EndDate         string `json:"end_date"`
		CommittedPoints int    `json:"committed_points"`
		CompletedPoints int    `json:"completed_points"`
		TeamCapacity    int    `json:"team_capacity"`
		Status          string `json:"status"`
		TeamID          string `json:"team_id"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required", nil)
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

	// Default status
	if req.Status == "" {
		req.Status = "planned"
	}

	sprint := &models.Sprint{
		Name:            req.Name,
		StartDate:       startDate.UTC(),
		EndDate:         endDate.UTC(),
		CommittedPoints: req.CommittedPoints,
		CompletedPoints: req.CompletedPoints,
		TeamCapacity:    req.TeamCapacity,
		Status:          req.Status,
		TeamID:          req.TeamID,
	}

	if err := s.planningStore.CreateSprint(sprint); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create sprint", err)
		return
	}

	respondJSON(w, http.StatusCreated, sprint)
}

func (s *Server) handleSprintByID(w http.ResponseWriter, r *http.Request) {
	// Extract sprint ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/planning/sprints/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Sprint ID required", http.StatusBadRequest)
		return
	}

	sprintID := parts[0]

	// Route based on operation
	if len(parts) > 1 {
		operation := parts[1]
		switch operation {
		case "health":
			s.getSprintHealth(w, r, sprintID)
		default:
			http.Error(w, "Unknown operation", http.StatusBadRequest)
		}
	} else {
		// Single sprint operations
		switch r.Method {
		case http.MethodGet:
			s.getSprint(w, r, sprintID)
		case http.MethodPut:
			s.updateSprint(w, r, sprintID)
		case http.MethodDelete:
			s.deleteSprint(w, r, sprintID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (s *Server) getSprint(w http.ResponseWriter, r *http.Request, sprintID string) {
	sprint, err := s.planningStore.GetSprint(sprintID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get sprint", err)
		return
	}

	if sprint == nil {
		http.Error(w, "Sprint not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, sprint)
}

func (s *Server) updateSprint(w http.ResponseWriter, r *http.Request, sprintID string) {
	var req struct {
		Name            string `json:"name"`
		StartDate       string `json:"start_date"`
		EndDate         string `json:"end_date"`
		CommittedPoints int    `json:"committed_points"`
		CompletedPoints int    `json:"completed_points"`
		TeamCapacity    int    `json:"team_capacity"`
		Status          string `json:"status"`
		TeamID          string `json:"team_id"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get existing sprint
	sprint, err := s.planningStore.GetSprint(sprintID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get sprint", err)
		return
	}

	if sprint == nil {
		http.Error(w, "Sprint not found", http.StatusNotFound)
		return
	}

	// Update fields - only update if provided
	if req.Name != "" {
		sprint.Name = req.Name
	}
	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid start_date format (expected YYYY-MM-DD)", err)
			return
		}
		sprint.StartDate = startDate.UTC()
	}
	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid end_date format (expected YYYY-MM-DD)", err)
			return
		}
		sprint.EndDate = endDate.UTC()
	}
	// Only update numeric fields if they're non-zero (allows setting to 0 explicitly if needed)
	// For proper partial updates, consider using pointers in the request struct
	if req.CommittedPoints > 0 {
		sprint.CommittedPoints = req.CommittedPoints
	}
	if req.CompletedPoints >= 0 {
		sprint.CompletedPoints = req.CompletedPoints
	}
	if req.TeamCapacity > 0 {
		sprint.TeamCapacity = req.TeamCapacity
	}
	if req.Status != "" {
		sprint.Status = req.Status
	}
	if req.TeamID != "" {
		sprint.TeamID = req.TeamID
	}

	if err := s.planningStore.UpdateSprint(sprint); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update sprint", err)
		return
	}

	respondJSON(w, http.StatusOK, sprint)
}

func (s *Server) deleteSprint(w http.ResponseWriter, r *http.Request, sprintID string) {
	// Note: DeleteSprint needs to be implemented in PlanningStore
	// For now, return not implemented
	respondError(w, http.StatusNotImplemented, "Delete sprint not yet implemented", nil)
}

func (s *Server) getSprintHealth(w http.ResponseWriter, r *http.Request, sprintID string) {
	health, err := s.planningService.CalculateSprintHealth(sprintID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to calculate sprint health", err)
		return
	}

	respondJSON(w, http.StatusOK, health)
}

// Velocity Handlers

func (s *Server) handleVelocity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract team ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/planning/velocity/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Team ID required", http.StatusBadRequest)
		return
	}

	teamID := parts[0]

	// Parse num_sprints parameter
	numSprints := 12 // Default
	if numSprintsStr := r.URL.Query().Get("num_sprints"); numSprintsStr != "" {
		parsed, err := strconv.Atoi(numSprintsStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid num_sprints parameter", err)
			return
		}
		numSprints = parsed
	}

	velocity, err := s.planningService.GetHistoricalVelocity(teamID, numSprints)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get velocity", err)
		return
	}

	// Get sprint details for recent_sprints field
	sprints, err := s.planningStore.GetSprintsByTeam(teamID, numSprints)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get sprint details", err)
		return
	}

	// Build recent_sprints array with full sprint objects
	recentSprints := make([]map[string]interface{}, 0, len(sprints))
	for _, sprint := range sprints {
		recentSprints = append(recentSprints, map[string]interface{}{
			"sprint_id":   sprint.ID,
			"sprint_name": sprint.Name,
			"points":      sprint.CompletedPoints,
			"start_date":  sprint.StartDate.Format(time.RFC3339),
			"end_date":    sprint.EndDate.Format(time.RFC3339),
		})
	}

	// Convert confidence_level string to numeric confidence
	var confidence float64
	switch velocity.ConfidenceLevel {
	case "high":
		confidence = 1.0
	case "medium":
		confidence = 0.67
	case "low":
		confidence = 0.33
	default:
		confidence = 0.0
	}

	// Transform response to match frontend interface
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"team_id":        velocity.TeamID,
		"average_points": velocity.AveragePoints,
		"std_dev":        velocity.StdDeviation, // Renamed from std_deviation
		"trend":          velocity.Trend,
		"confidence":     confidence,    // Changed from confidence_level, now numeric
		"recent_sprints": recentSprints, // Changed from last_12_sprints, full objects
	})
}

// Timeline Estimation Handlers

func (s *Server) handleEstimate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TeamID          string `json:"team_id"`
		FeatureName     string `json:"feature_name"`
		EstimatedPoints int    `json:"estimated_points"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.TeamID == "" {
		respondError(w, http.StatusBadRequest, "team_id is required", nil)
		return
	}

	if req.FeatureName == "" {
		respondError(w, http.StatusBadRequest, "feature_name is required", nil)
		return
	}

	if req.EstimatedPoints <= 0 {
		respondError(w, http.StatusBadRequest, "estimated_points must be greater than 0", nil)
		return
	}

	estimate, err := s.planningService.EstimateTimeline(req.FeatureName, req.EstimatedPoints, req.TeamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to estimate timeline", err)
		return
	}

	respondJSON(w, http.StatusOK, estimate)
}

// Capacity History Handlers

func (s *Server) handleCapacity(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getCapacityHistory(w, r)
	case http.MethodPost:
		s.createCapacityHistory(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getCapacityHistory(w http.ResponseWriter, r *http.Request) {
	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		respondError(w, http.StatusBadRequest, "team_id parameter is required", nil)
		return
	}

	// Parse weeks parameter
	weeks := 12 // Default
	if weeksStr := r.URL.Query().Get("weeks"); weeksStr != "" {
		parsed, err := strconv.Atoi(weeksStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid weeks parameter", err)
			return
		}
		weeks = parsed
	}

	history, err := s.planningStore.GetCapacityHistory(teamID, weeks)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get capacity history", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"history": history,
	})
}

func (s *Server) createCapacityHistory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WeekStart          string  `json:"week_start"`
		TeamID             string  `json:"team_id"`
		TeamSize           int     `json:"team_size"`
		AvailableEngineers float64 `json:"available_engineers"`
		CompletedPoints    int     `json:"completed_points"`
		MeetingHours       float64 `json:"meeting_hours"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.WeekStart == "" {
		respondError(w, http.StatusBadRequest, "week_start is required", nil)
		return
	}

	if req.TeamID == "" {
		respondError(w, http.StatusBadRequest, "team_id is required", nil)
		return
	}

	// Parse week_start date
	weekStart, err := time.Parse("2006-01-02", req.WeekStart)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid week_start format (expected YYYY-MM-DD)", err)
		return
	}

	history := &models.CapacityHistory{
		WeekStart:          weekStart.UTC(),
		TeamID:             req.TeamID,
		TeamSize:           req.TeamSize,
		AvailableEngineers: req.AvailableEngineers,
		CompletedPoints:    req.CompletedPoints,
		MeetingHours:       req.MeetingHours,
	}

	if err := s.planningStore.CreateCapacityHistory(history); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create capacity history", err)
		return
	}

	respondJSON(w, http.StatusCreated, history)
}

// handleSprintBurndown returns current sprint burndown chart data
func (s *Server) handleSprintBurndown(w http.ResponseWriter, r *http.Request) {
	// Find the current active sprint (where NOW() is between start_date and end_date)
	now := time.Now().UTC()

	// Get all active sprints
	sprints, err := s.planningStore.ListSprints(db.MaxQueryLimit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get sprints", err)
		return
	}

	// Find the first active sprint
	var activeSprint *models.Sprint
	for _, sprint := range sprints {
		if now.After(sprint.StartDate) && now.Before(sprint.EndDate) {
			activeSprint = sprint
			break
		}
	}

	// If no active sprint, return empty response (not 404)
	if activeSprint == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"sprint_id":        "",
			"sprint_name":      "",
			"start_date":       "",
			"end_date":         "",
			"total_points":     0,
			"completed_points": 0,
			"risk_level":       "on_track",
			"burndown_points":  []interface{}{},
		})
		return
	}

	// Get all stories for this sprint
	stories, err := s.planningStore.GetStoriesBySprint(activeSprint.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stories", err)
		return
	}

	// Calculate total points and completed points
	var totalPoints int
	var completedPoints int
	for _, story := range stories {
		totalPoints += story.StoryPoints
		if story.CompletedAt != nil {
			completedPoints += story.StoryPoints
		}
	}

	// Generate daily burndown data points from sprint start to today
	type dataPoint struct {
		Date   string  `json:"date"`
		Ideal  float64 `json:"ideal"`
		Actual int     `json:"actual"`
	}

	var dataPoints []dataPoint

	// Calculate sprint duration in days
	sprintDuration := activeSprint.EndDate.Sub(activeSprint.StartDate).Hours() / 24
	dailyIdealBurn := float64(totalPoints) / sprintDuration

	// Iterate from sprint start to today (or end date if sprint ended)
	endDate := now
	if now.After(activeSprint.EndDate) {
		endDate = activeSprint.EndDate
	}

	currentDate := activeSprint.StartDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate.Truncate(24*time.Hour)) {
		// Calculate completed points up to this date
		completedUpToDate := 0
		for _, story := range stories {
			if story.CompletedAt != nil && !story.CompletedAt.Truncate(24*time.Hour).After(currentDate.Truncate(24*time.Hour)) {
				completedUpToDate += story.StoryPoints
			}
		}

		remainingPoints := totalPoints - completedUpToDate

		// Calculate ideal remaining
		daysElapsed := currentDate.Sub(activeSprint.StartDate).Hours() / 24
		idealRemaining := float64(totalPoints) - (dailyIdealBurn * daysElapsed)
		if idealRemaining < 0 {
			idealRemaining = 0
		}

		dataPoints = append(dataPoints, dataPoint{
			Date:   currentDate.Format(time.RFC3339),
			Ideal:  idealRemaining,
			Actual: remainingPoints,
		})

		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// Calculate risk level based on progress vs ideal
	riskLevel := "on_track"
	if len(dataPoints) > 0 {
		latestPoint := dataPoints[len(dataPoints)-1]
		if totalPoints > 0 {
			// Calculate percentage behind/ahead of ideal
			actual := float64(latestPoint.Actual)
			ideal := latestPoint.Ideal
			percentageBehind := ((actual - ideal) / float64(totalPoints)) * 100

			if percentageBehind > 25 {
				riskLevel = "high_risk"
			} else if percentageBehind > 10 {
				riskLevel = "at_risk"
			}
		}
	}

	// Return response with snake_case field names
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"sprint_id":        activeSprint.ID,
		"sprint_name":      activeSprint.Name,
		"start_date":       activeSprint.StartDate.Format(time.RFC3339),
		"end_date":         activeSprint.EndDate.Format(time.RFC3339),
		"total_points":     totalPoints,
		"completed_points": completedPoints,
		"risk_level":       riskLevel,
		"burndown_points":  dataPoints,
	})
}
