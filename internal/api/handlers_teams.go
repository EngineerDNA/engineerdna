package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

// Teams Management Handlers

func (s *Server) handleTeams(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listTeams(w, r)
	case http.MethodPost:
		s.createTeam(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listTeams(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	teams, err := s.teamStore.ListTeams(limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list teams", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"teams": teams,
	})
}

func (s *Server) createTeam(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string  `json:"name"`
		ParentTeamID *string `json:"parent_team_id,omitempty"`
		ManagerID    *string `json:"manager_id,omitempty"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required", nil)
		return
	}

	// Validate parent team exists if provided
	if req.ParentTeamID != nil {
		parentTeam, err := s.teamStore.GetTeam(*req.ParentTeamID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to validate parent team", err)
			return
		}
		if parentTeam == nil {
			respondError(w, http.StatusBadRequest, "Parent team not found", nil)
			return
		}
	}

	// Validate manager exists if provided
	if req.ManagerID != nil {
		manager, err := s.identityService.GetEngineer(*req.ManagerID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to validate manager", err)
			return
		}
		if manager == nil {
			respondError(w, http.StatusBadRequest, "Manager not found", nil)
			return
		}
	}

	team, err := s.teamStore.CreateTeam(req.Name, req.ParentTeamID, req.ManagerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create team", err)
		return
	}

	respondJSON(w, http.StatusCreated, team)
}

func (s *Server) handleTeamByID(w http.ResponseWriter, r *http.Request) {
	// Extract team ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/teams/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Team ID required", http.StatusBadRequest)
		return
	}

	teamID := parts[0]

	// Route based on operation
	if len(parts) > 1 {
		operation := parts[1]
		switch operation {
		case "members":
			s.handleTeamMembers(w, r, teamID, parts[2:])
		case "performance":
			s.handleTeamPerformance(w, r, teamID, parts[2:])
		default:
			http.Error(w, "Unknown operation", http.StatusBadRequest)
		}
	} else {
		// Single team operations
		switch r.Method {
		case http.MethodGet:
			s.getTeam(w, r, teamID)
		case http.MethodPut:
			s.updateTeam(w, r, teamID)
		case http.MethodDelete:
			s.deleteTeam(w, r, teamID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (s *Server) getTeam(w http.ResponseWriter, r *http.Request, teamID string) {
	team, err := s.teamStore.GetTeam(teamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get team", err)
		return
	}

	if team == nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, team)
}

func (s *Server) updateTeam(w http.ResponseWriter, r *http.Request, teamID string) {
	var req struct {
		Name         string  `json:"name"`
		ParentTeamID *string `json:"parent_team_id"`
		ManagerID    *string `json:"manager_id"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required", nil)
		return
	}

	// Validate parent team exists if provided and prevent circular references
	if req.ParentTeamID != nil {
		if *req.ParentTeamID == teamID {
			respondError(w, http.StatusBadRequest, "Team cannot be its own parent", nil)
			return
		}

		parentTeam, err := s.teamStore.GetTeam(*req.ParentTeamID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to validate parent team", err)
			return
		}
		if parentTeam == nil {
			respondError(w, http.StatusBadRequest, "Parent team not found", nil)
			return
		}

		// Check if parent is a descendant of current team (would create cycle)
		isDescendant, err := s.teamStore.IsDescendant(teamID, *req.ParentTeamID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to validate team hierarchy", err)
			return
		}
		if isDescendant {
			respondError(w, http.StatusBadRequest, "Cannot create circular team hierarchy", nil)
			return
		}
	}

	// Validate manager exists if provided
	if req.ManagerID != nil {
		manager, err := s.identityService.GetEngineer(*req.ManagerID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to validate manager", err)
			return
		}
		if manager == nil {
			respondError(w, http.StatusBadRequest, "Manager not found", nil)
			return
		}
	}

	err := s.teamStore.UpdateTeam(teamID, req.Name, req.ParentTeamID, req.ManagerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update team", err)
		return
	}

	// Fetch updated team
	team, err := s.teamStore.GetTeam(teamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated team", err)
		return
	}

	respondJSON(w, http.StatusOK, team)
}

func (s *Server) deleteTeam(w http.ResponseWriter, r *http.Request, teamID string) {
	err := s.teamStore.DeleteTeam(teamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete team", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Team Hierarchy Handlers

func (s *Server) handleTeamsHierarchy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get root team ID from query parameter (optional)
	rootTeamID := r.URL.Query().Get("root")

	if rootTeamID == "" {
		// Return all top-level teams with their hierarchies
		limit, err := parseLimitParam(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
			return
		}

		teams, err := s.teamStore.ListTeams(limit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to list teams", err)
			return
		}

		// Build hierarchy for each top-level team
		var hierarchies []*models.TeamHierarchyNode
		for _, team := range teams {
			if team.ParentTeamID == nil {
				hierarchy, err := s.teamService.GetTeamHierarchy(team.ID)
				if err != nil {
					respondError(w, http.StatusInternalServerError, "Failed to build team hierarchy", err)
					return
				}
				hierarchies = append(hierarchies, hierarchy)
			}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"hierarchies": hierarchies,
		})
		return
	}

	// Return hierarchy for specific root team
	hierarchy, err := s.teamService.GetTeamHierarchy(rootTeamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get team hierarchy", err)
		return
	}

	respondJSON(w, http.StatusOK, hierarchy)
}

// Team Membership Handlers

func (s *Server) handleTeamMembers(w http.ResponseWriter, r *http.Request, teamID string, parts []string) {
	// Check if this is a specific member operation
	if len(parts) > 0 && parts[0] != "" {
		memberID := parts[0]
		if r.Method == http.MethodDelete {
			s.removeTeamMember(w, r, teamID, memberID)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Team members list/add operations
	switch r.Method {
	case http.MethodGet:
		s.getTeamMembers(w, r, teamID)
	case http.MethodPost:
		s.addTeamMember(w, r, teamID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getTeamMembers(w http.ResponseWriter, r *http.Request, teamID string) {
	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	members, err := s.teamStore.GetTeamMembers(teamID, includeInactive)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get team members", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"members": members,
	})
}

func (s *Server) addTeamMember(w http.ResponseWriter, r *http.Request, teamID string) {
	var req struct {
		MemberID string `json:"member_id"`
		Role     string `json:"role,omitempty"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.MemberID == "" {
		respondError(w, http.StatusBadRequest, "member_id is required", nil)
		return
	}

	// Validate team exists
	team, err := s.teamStore.GetTeam(teamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to validate team", err)
		return
	}
	if team == nil {
		respondError(w, http.StatusBadRequest, "Team not found", nil)
		return
	}

	// Validate member exists
	member, err := s.identityService.GetEngineer(req.MemberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to validate member", err)
		return
	}
	if member == nil {
		respondError(w, http.StatusBadRequest, "Member not found", nil)
		return
	}

	membership, err := s.teamStore.AddTeamMember(teamID, req.MemberID, req.Role)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to add team member", err)
		return
	}

	respondJSON(w, http.StatusCreated, membership)
}

func (s *Server) removeTeamMember(w http.ResponseWriter, r *http.Request, teamID, memberID string) {
	err := s.teamStore.RemoveTeamMember(teamID, memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to remove team member", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Team Performance Handlers

func (s *Server) handleTeamPerformance(w http.ResponseWriter, r *http.Request, teamID string, parts []string) {
	// Check if this is calculate operation
	if len(parts) > 0 && parts[0] == "calculate" {
		if r.Method == http.MethodPost {
			s.calculateTeamScore(w, r, teamID)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get team performance scores
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.getTeamPerformance(w, r, teamID)
}

func (s *Server) getTeamPerformance(w http.ResponseWriter, r *http.Request, teamID string) {
	// Get current week start (normalize to Sunday)
	now := time.Now().UTC()
	weekday := int(now.Weekday())
	currentWeekStart := now.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)

	// Get current week score
	currentScores, err := s.teamStore.GetTeamPerformanceScores(teamID, currentWeekStart, now, 1)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get current team performance scores", err)
		return
	}

	// Return zeros if no data for current week
	if len(currentScores) == 0 {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"team_id":               teamID,
			"week_start":            currentWeekStart.Format(time.RFC3339),
			"total_score":           0.0,
			"member_count":          0,
			"score_change":          0.0,
			"velocity_prs_per_week": 0.0,
			"velocity_change":       0.0,
			"cycle_time_days":       0.0,
			"cycle_time_change":     0.0,
		})
		return
	}

	current := currentScores[0]

	// Get previous week score for change calculation
	prevWeekStart := currentWeekStart.AddDate(0, 0, -7)
	prevWeekEnd := currentWeekStart.Add(-time.Second)
	prevScores, err := s.teamStore.GetTeamPerformanceScores(teamID, prevWeekStart, prevWeekEnd, 1)
	if err != nil {
		// Log error but continue with zero changes
		prevScores = nil
	}

	// Calculate changes from previous week
	var scoreChange, velocityChange, cycleTimeChange float64
	if len(prevScores) > 0 {
		prev := prevScores[0]
		scoreChange = current.TotalScore - prev.TotalScore
		velocityChange = current.ThroughputScore - prev.ThroughputScore
		cycleTimeChange = current.SpeedScore - prev.SpeedScore
	}

	// Return single object with calculated metrics
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"team_id":               current.TeamID,
		"week_start":            current.WeekStart.Format(time.RFC3339),
		"total_score":           current.TotalScore,
		"member_count":          current.MemberCount,
		"score_change":          scoreChange,
		"velocity_prs_per_week": current.ThroughputScore,
		"velocity_change":       velocityChange,
		"cycle_time_days":       current.SpeedScore,
		"cycle_time_change":     cycleTimeChange,
	})
}

func (s *Server) calculateTeamScore(w http.ResponseWriter, r *http.Request, teamID string) {
	var req struct {
		WeekStart string `json:"week_start"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
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

	// Calculate team score
	score, err := s.teamService.CalculateTeamScore(teamID, weekStart)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to calculate team score", err)
		return
	}

	if score == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"message": "No score calculated (team may have no active members or no data for this period)",
		})
		return
	}

	respondJSON(w, http.StatusOK, score)
}

// Organization Scorecard Handlers

func (s *Server) handleOrgScorecard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get week_start from query parameter
	query := r.URL.Query()
	weekStartStr := query.Get("week_start")

	// Default to current week if not specified
	weekStart := time.Now().UTC().Truncate(24 * time.Hour)
	if weekStartStr != "" {
		parsed, err := time.Parse("2006-01-02", weekStartStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid week_start format (expected YYYY-MM-DD)", err)
			return
		}
		weekStart = parsed
	}

	prevWeekStart := weekStart.AddDate(0, 0, -7)

	// Get all child teams (teams with non-null parent_id)
	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	teams, err := s.teamStore.ListTeams(limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list teams", err)
		return
	}

	var childTeams []*models.Team
	for _, team := range teams {
		if team.ParentTeamID != nil {
			childTeams = append(childTeams, team)
		}
	}

	if len(childTeams) == 0 {
		// No child teams, return zeros
		respondJSON(w, http.StatusOK, &models.OrgScorecard{})
		return
	}

	// Aggregate metrics across child teams
	var totalEngineers int
	var totalWeightedScore float64
	var totalThroughput float64
	var totalSpeed float64
	var teamsNeedingAttention int
	var teamsWithCurrentData int

	var prevTotalWeightedScore float64
	var prevTotalThroughput float64
	var prevTotalSpeed float64
	var prevTeamsWithData int

	for _, team := range childTeams {
		// Get current week score
		currentScores, err := s.teamStore.GetTeamPerformanceScores(team.ID, weekStart, weekStart, 1)
		if err != nil {
			continue
		}

		if len(currentScores) > 0 {
			score := currentScores[0]
			totalEngineers += score.MemberCount
			totalWeightedScore += score.TotalScore * float64(score.MemberCount)
			totalThroughput += score.ThroughputScore
			totalSpeed += score.SpeedScore
			teamsWithCurrentData++

			if score.TotalScore < 80 {
				teamsNeedingAttention++
			}
		}

		// Get previous week score for change calculation
		prevScores, err := s.teamStore.GetTeamPerformanceScores(team.ID, prevWeekStart, prevWeekStart, 1)
		if err != nil {
			continue
		}

		if len(prevScores) > 0 {
			prevScore := prevScores[0]
			prevTotalWeightedScore += prevScore.TotalScore * float64(prevScore.MemberCount)
			prevTotalThroughput += prevScore.ThroughputScore
			prevTotalSpeed += prevScore.SpeedScore
			prevTeamsWithData++
		}
	}

	// Calculate averages and changes
	var orgScore models.OrgScorecard
	orgScore.TotalEngineers = totalEngineers
	orgScore.TeamsNeedingAttention = teamsNeedingAttention

	if totalEngineers > 0 {
		orgScore.TotalScore = totalWeightedScore / float64(totalEngineers)
	}

	if teamsWithCurrentData > 0 {
		orgScore.VelocityPRsPerWeek = totalThroughput / float64(teamsWithCurrentData)
		orgScore.CycleTimeDays = totalSpeed / float64(teamsWithCurrentData)
	}

	// Calculate changes from previous week
	if prevTeamsWithData > 0 {
		prevAvgScore := prevTotalWeightedScore / float64(totalEngineers)
		prevAvgThroughput := prevTotalThroughput / float64(prevTeamsWithData)
		prevAvgSpeed := prevTotalSpeed / float64(prevTeamsWithData)

		orgScore.ScoreChange = orgScore.TotalScore - prevAvgScore
		orgScore.VelocityChange = orgScore.VelocityPRsPerWeek - prevAvgThroughput
		orgScore.CycleTimeChange = orgScore.CycleTimeDays - prevAvgSpeed
	}

	respondJSON(w, http.StatusOK, orgScore)
}
