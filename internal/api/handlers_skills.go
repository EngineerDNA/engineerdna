package api

import (
	"net/http"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

// Skills Handlers

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listSkills(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listSkills(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

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

	skills, total, err := s.skillsStore.ListSkills(category, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list skills", err)
		return
	}

	respondPaginated(w, skills, total, limit, offset)
}

// Engineer Skills Handlers

func (s *Server) handleEngineerSkills(w http.ResponseWriter, r *http.Request, engineerID string) {
	switch r.Method {
	case http.MethodGet:
		s.getEngineerSkillProfile(w, r, engineerID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getEngineerSkillProfile(w http.ResponseWriter, r *http.Request, engineerID string) {
	// Get engineer name (would normally come from engineer service)
	engineerName := "Engineer" // Placeholder

	profile, err := s.skillsService.GetEngineerSkillProfile(engineerID, engineerName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get skill profile", err)
		return
	}

	respondJSON(w, http.StatusOK, profile)
}

func (s *Server) handleEngineerSkillByID(w http.ResponseWriter, r *http.Request, engineerID, skillID string) {
	switch r.Method {
	case http.MethodGet:
		s.getEngineerSkill(w, r, engineerID, skillID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getEngineerSkill(w http.ResponseWriter, r *http.Request, engineerID, skillID string) {
	// Get engineer skill
	engineerSkill, err := s.skillsStore.GetEngineerSkill(engineerID, skillID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get engineer skill", err)
		return
	}

	if engineerSkill == nil {
		http.Error(w, "Skill not found for engineer", http.StatusNotFound)
		return
	}

	// Get skill details
	skill, err := s.skillsStore.GetSkill(skillID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get skill details", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"engineer_skill": engineerSkill,
		"skill":          skill,
	})
}

// Skill Evidence Handlers

func (s *Server) handleSkillEvidence(w http.ResponseWriter, r *http.Request, engineerID, skillID string) {
	switch r.Method {
	case http.MethodGet:
		s.getSkillEvidence(w, r, engineerID, skillID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getSkillEvidence(w http.ResponseWriter, r *http.Request, engineerID, skillID string) {
	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = 50
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	evidences, total, err := s.skillsStore.ListSkillEvidence(engineerID, skillID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get skill evidence", err)
		return
	}

	respondPaginated(w, evidences, total, limit, offset)
}

// Skill Progression Handlers

func (s *Server) handleSkillProgression(w http.ResponseWriter, r *http.Request, engineerID, skillID string) {
	switch r.Method {
	case http.MethodGet:
		s.getSkillProgression(w, r, engineerID, skillID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getSkillProgression(w http.ResponseWriter, r *http.Request, engineerID, skillID string) {
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

	progressions, total, err := s.skillsStore.GetSkillProgression(engineerID, skillID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get skill progression", err)
		return
	}

	respondPaginated(w, progressions, total, limit, offset)
}

// Skill Goals Handlers

func (s *Server) handleSkillGoals(w http.ResponseWriter, r *http.Request, engineerID string) {
	switch r.Method {
	case http.MethodGet:
		s.getSkillGoals(w, r, engineerID)
	case http.MethodPost:
		s.createSkillGoal(w, r, engineerID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getSkillGoals(w http.ResponseWriter, r *http.Request, engineerID string) {
	status := r.URL.Query().Get("status")

	goals, err := s.skillsStore.GetSkillGoals(engineerID, status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get skill goals", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"goals": goals,
	})
}

func (s *Server) createSkillGoal(w http.ResponseWriter, r *http.Request, engineerID string) {
	var req struct {
		SkillID      string  `json:"skill_id"`
		GoalID       *string `json:"goal_id"`
		CurrentLevel int     `json:"current_level"`
		TargetLevel  int     `json:"target_level"`
		TargetDate   *string `json:"target_date"`
		Milestones   string  `json:"milestones"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.SkillID == "" {
		respondError(w, http.StatusBadRequest, "skill_id is required", nil)
		return
	}

	goal := &models.SkillGoal{
		EngineerID:   engineerID,
		SkillID:      req.SkillID,
		GoalID:       req.GoalID,
		CurrentLevel: req.CurrentLevel,
		TargetLevel:  req.TargetLevel,
		Milestones:   req.Milestones,
		Status:       "active",
	}

	if req.TargetDate != nil {
		targetDate, err := time.Parse("2006-01-02", *req.TargetDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid target_date format (expected YYYY-MM-DD)", err)
			return
		}
		goal.TargetDate = &targetDate
	}

	if err := s.skillsStore.CreateSkillGoal(goal); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create skill goal", err)
		return
	}

	respondJSON(w, http.StatusCreated, goal)
}

// Skill Analysis Handlers

func (s *Server) handleSkillGaps(w http.ResponseWriter, r *http.Request, engineerID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role := r.URL.Query().Get("role")
	if role == "" {
		role = "mid" // Default to mid-level
	}

	targetRole := r.URL.Query().Get("target_role")
	if targetRole == "" {
		targetRole = role
	}

	analysis, err := s.skillsService.GetPromotionReadiness(engineerID, role, targetRole)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get skill gaps", err)
		return
	}

	respondJSON(w, http.StatusOK, analysis)
}

func (s *Server) handleSkillRecommendations(w http.ResponseWriter, r *http.Request, engineerID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	skillID := r.URL.Query().Get("skill_id")
	if skillID == "" {
		respondError(w, http.StatusBadRequest, "skill_id is required", nil)
		return
	}

	plan, err := s.skillsService.GetSkillDevelopmentPlan(engineerID, skillID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get development plan", err)
		return
	}

	respondJSON(w, http.StatusOK, plan)
}

func (s *Server) handleSkillPeerComparison(w http.ResponseWriter, r *http.Request, engineerID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get peer engineer IDs (would normally come from team membership)
	// For now, this would need to be provided or derived from team data
	var peerEngineerIDs []string

	comparisons, err := s.skillsService.ComparePeerSkills(engineerID, peerEngineerIDs)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get peer comparison", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"comparisons": comparisons,
	})
}

// Skill Evaluation Handler

func (s *Server) handleSkillEvaluate(w http.ResponseWriter, r *http.Request, engineerID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.skillsService.EvaluateAllSkills(engineerID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to evaluate skills", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Skills evaluated successfully",
	})
}

// Routing helper for engineer skill operations
// Called from handlers_engineers.go handleEngineerOperations
func (s *Server) handleEngineerSkillOperations(w http.ResponseWriter, r *http.Request, engineerID string, parts []string) {
	if len(parts) < 3 {
		// /api/engineers/:id/skills - list all skills
		s.handleEngineerSkills(w, r, engineerID)
		return
	}

	operation := parts[2]

	switch operation {
	case "gaps":
		s.handleSkillGaps(w, r, engineerID)
	case "recommendations":
		s.handleSkillRecommendations(w, r, engineerID)
	case "peer-comparison":
		s.handleSkillPeerComparison(w, r, engineerID)
	case "evaluate":
		s.handleSkillEvaluate(w, r, engineerID)
	case "goals":
		s.handleSkillGoals(w, r, engineerID)
	default:
		// /api/engineers/:id/skills/:skill_id
		skillID := operation
		if len(parts) > 3 {
			subOp := parts[3]
			switch subOp {
			case "evidence":
				s.handleSkillEvidence(w, r, engineerID, skillID)
			case "progression":
				s.handleSkillProgression(w, r, engineerID, skillID)
			default:
				http.Error(w, "Unknown operation", http.StatusBadRequest)
			}
		} else {
			s.handleEngineerSkillByID(w, r, engineerID, skillID)
		}
	}
}
