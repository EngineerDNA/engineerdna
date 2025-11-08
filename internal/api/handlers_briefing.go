package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/config"
)

// Briefing Handlers

func (s *Server) handleBriefingWeekly(w http.ResponseWriter, r *http.Request) {
	// Extract team ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/briefing/weekly/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Team ID required", http.StatusBadRequest)
		return
	}

	teamID := parts[0]

	// Check for operation after team ID
	if len(parts) > 1 {
		operation := parts[1]
		switch operation {
		case "generate":
			if r.Method == http.MethodPost {
				s.generateWeeklyBriefing(w, r, teamID)
				return
			}
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		default:
			http.Error(w, "Unknown operation", http.StatusBadRequest)
			return
		}
	}

	// GET /api/briefing/weekly/{team_id}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.getWeeklyBriefing(w, r, teamID)
}

func (s *Server) getWeeklyBriefing(w http.ResponseWriter, r *http.Request, teamID string) {
	// Parse week_start from query or default to current week
	weekStartStr := r.URL.Query().Get("week_start")
	var weekStart time.Time

	if weekStartStr == "" {
		weekStart = getCurrentWeekStart()
	} else {
		var err error
		weekStart, err = time.Parse("2006-01-02", weekStartStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid week_start format (expected YYYY-MM-DD)", err)
			return
		}
		weekStart = weekStart.UTC()
	}

	// Validate team exists
	team, err := s.teamStore.GetTeam(teamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to validate team", err)
		return
	}
	if team == nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	// Get or generate briefing
	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultTimeout)
	defer cancel()
	briefing, err := s.briefingService.GetWeeklyBriefing(ctx, teamID, weekStart)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get briefing", err)
		return
	}

	// If no briefing exists, return 404
	if briefing == nil {
		http.Error(w, "No briefing available for this week. Click 'Generate Briefing' to create one.", http.StatusNotFound)
		return
	}

	// Transform response to include additional fields expected by frontend
	weekEnd := briefing.WeekStart.AddDate(0, 0, 7)

	response := map[string]interface{}{
		"id":              briefing.ID,
		"team_id":         briefing.TeamID,
		"team_name":       team.Name,
		"week_start":      briefing.WeekStart.Format(time.RFC3339),
		"week_end":        weekEnd.Format(time.RFC3339),
		"tldr":            briefing.TLDR,
		"key_metrics":     briefing.KeyMetrics,
		"needs_attention": briefing.NeedsAttention,
		"insights":        briefing.Insights,
		"trending_up":     briefing.TrendingUp,
		"trending_down":   briefing.TrendingDown,
		"talking_points":  briefing.TalkingPoints,
		"generated_at":    briefing.GeneratedAt.Format(time.RFC3339),
		"last_updated":    briefing.CreatedAt.Format(time.RFC3339),
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) generateWeeklyBriefing(w http.ResponseWriter, r *http.Request, teamID string) {
	var req struct {
		WeekStart string `json:"week_start"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Parse week_start from body or default to current week
	var weekStart time.Time
	if req.WeekStart == "" {
		weekStart = getCurrentWeekStart()
	} else {
		var err error
		weekStart, err = time.Parse("2006-01-02", req.WeekStart)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid week_start format (expected YYYY-MM-DD)", err)
			return
		}
		weekStart = weekStart.UTC()
	}

	// Validate team exists
	team, err := s.teamStore.GetTeam(teamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to validate team", err)
		return
	}
	if team == nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	// Force regenerate briefing
	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultTimeout)
	defer cancel()
	briefing, err := s.briefingService.GenerateWeeklyBriefing(ctx, teamID, weekStart)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate briefing", err)
		return
	}

	// Transform response to include additional fields expected by frontend
	weekEnd := briefing.WeekStart.AddDate(0, 0, 7)

	response := map[string]interface{}{
		"id":              briefing.ID,
		"team_id":         briefing.TeamID,
		"team_name":       team.Name,
		"week_start":      briefing.WeekStart.Format(time.RFC3339),
		"week_end":        weekEnd.Format(time.RFC3339),
		"tldr":            briefing.TLDR,
		"key_metrics":     briefing.KeyMetrics,
		"needs_attention": briefing.NeedsAttention,
		"insights":        briefing.Insights,
		"trending_up":     briefing.TrendingUp,
		"trending_down":   briefing.TrendingDown,
		"talking_points":  briefing.TalkingPoints,
		"generated_at":    briefing.GeneratedAt.Format(time.RFC3339),
		"last_updated":    briefing.CreatedAt.Format(time.RFC3339),
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) handleBriefingHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract team ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/briefing/history/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Team ID required", http.StatusBadRequest)
		return
	}

	teamID := parts[0]

	// Validate team exists
	team, err := s.teamStore.GetTeam(teamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to validate team", err)
		return
	}
	if team == nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	limitStr := query.Get("limit")
	limit := 10 // Default limit

	if limitStr != "" {
		// Try parsing as integer for limit
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get briefing history
	briefings, err := s.briefingStore.ListBriefings(teamID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get briefing history", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"team_id":   teamID,
		"briefings": briefings,
		"count":     len(briefings),
	})
}

func (s *Server) handleBriefingOrg(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse week_start from query or default to current week
	weekStartStr := r.URL.Query().Get("week_start")
	var weekStart time.Time

	if weekStartStr == "" {
		weekStart = getCurrentWeekStart()
	} else {
		var err error
		weekStart, err = time.Parse("2006-01-02", weekStartStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid week_start format (expected YYYY-MM-DD)", err)
			return
		}
		weekStart = weekStart.UTC()
	}

	// Placeholder: Get all teams and aggregate briefings
	// This is a placeholder for org-wide briefing aggregation
	// Future implementation will fetch briefings for all teams and combine them
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"week_start": weekStart.Format("2006-01-02"),
		"message":    "Organization-wide briefing aggregation is not yet implemented",
		"teams":      []interface{}{},
	})
}

// getCurrentWeekStart returns the start of the current week (Sunday) in UTC
func getCurrentWeekStart() time.Time {
	now := time.Now().UTC()
	weekday := int(now.Weekday())
	return now.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
}
