package api

import (
	"net/http"
	"strings"

	"github.com/engineerdna/engineerdna/internal/models"
)

// Manager Notes Handlers

func (s *Server) handleNotes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listNotes(w, r)
	case http.MethodPost:
		s.createNote(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listNotes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	managerID := query.Get("manager_id")
	subjectType := query.Get("subject_type")
	subjectID := query.Get("subject_id")

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

	visibility := []string{"private", "shared_with_subject", "team", "org"}

	notes, total, err := s.contextStore.ListManagerNotes(managerID, subjectType, subjectID, visibility, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list notes", err)
		return
	}

	respondPaginated(w, notes, total, limit, offset)
}

func (s *Server) createNote(w http.ResponseWriter, r *http.Request) {
	var note models.ManagerNote
	if err := decodeAndValidateJSON(r, &note); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if note.ManagerID == "" || note.SubjectType == "" || note.SubjectID == "" || note.Content == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	if err := s.contextStore.CreateManagerNote(&note); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create note", err)
		return
	}

	respondJSON(w, http.StatusCreated, note)
}

func (s *Server) handleNoteByID(w http.ResponseWriter, r *http.Request) {
	noteID := strings.TrimPrefix(r.URL.Path, "/api/notes/")

	switch r.Method {
	case http.MethodGet:
		s.getNote(w, r, noteID)
	case http.MethodPut:
		s.updateNote(w, r, noteID)
	case http.MethodDelete:
		s.deleteNote(w, r, noteID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getNote(w http.ResponseWriter, r *http.Request, noteID string) {
	note, err := s.contextStore.GetManagerNote(noteID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get note", err)
		return
	}
	if note == nil {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, note)
}

func (s *Server) updateNote(w http.ResponseWriter, r *http.Request, noteID string) {
	var note models.ManagerNote
	if err := decodeAndValidateJSON(r, &note); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	note.ID = noteID
	if err := s.contextStore.UpdateManagerNote(&note); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update note", err)
		return
	}

	respondJSON(w, http.StatusOK, note)
}

func (s *Server) deleteNote(w http.ResponseWriter, r *http.Request, noteID string) {
	if err := s.contextStore.DeleteManagerNote(noteID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete note", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Context Annotations Handlers

func (s *Server) handleContextAnnotations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listAnnotations(w, r)
	case http.MethodPost:
		s.createAnnotation(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listAnnotations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	entityType := query.Get("entity_type")
	entityID := query.Get("entity_id")

	if entityType == "" || entityID == "" {
		respondError(w, http.StatusBadRequest, "entity_type and entity_id required", nil)
		return
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

	annotations, total, err := s.contextStore.GetContextAnnotations(entityType, entityID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get annotations", err)
		return
	}

	respondPaginated(w, annotations, total, limit, offset)
}

func (s *Server) createAnnotation(w http.ResponseWriter, r *http.Request) {
	var annotation models.ContextAnnotation
	if err := decodeAndValidateJSON(r, &annotation); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.contextStore.CreateContextAnnotation(&annotation); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create annotation", err)
		return
	}

	respondJSON(w, http.StatusCreated, annotation)
}

// Team Context Handlers

func (s *Server) handleTeamContext(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/teams/"), "/")
	if len(parts) < 2 || parts[1] != "context" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	teamID := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.getTeamContext(w, r, teamID)
	case http.MethodPost:
		s.createTeamContext(w, r, teamID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getTeamContext(w http.ResponseWriter, r *http.Request, teamID string) {
	timePeriod := r.URL.Query().Get("time_period")

	contexts, err := s.contextStore.GetTeamContext(teamID, timePeriod)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get team context", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"contexts": contexts,
	})
}

func (s *Server) createTeamContext(w http.ResponseWriter, r *http.Request, teamID string) {
	var ctx models.TeamContext
	if err := decodeAndValidateJSON(r, &ctx); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	ctx.TeamID = teamID

	if err := s.contextStore.CreateTeamContext(&ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create team context", err)
		return
	}

	respondJSON(w, http.StatusCreated, ctx)
}

// Sentiment Survey Handlers

func (s *Server) handleSurveys(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listSurveys(w, r)
	case http.MethodPost:
		s.createSurvey(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listSurveys(w http.ResponseWriter, r *http.Request) {
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

	surveys, total, err := s.contextStore.ListActiveSurveys(limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list surveys", err)
		return
	}

	respondPaginated(w, surveys, total, limit, offset)
}

func (s *Server) createSurvey(w http.ResponseWriter, r *http.Request) {
	var survey models.SentimentSurvey
	if err := decodeAndValidateJSON(r, &survey); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.contextStore.CreateSentimentSurvey(&survey); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create survey", err)
		return
	}

	respondJSON(w, http.StatusCreated, survey)
}

func (s *Server) handleSurveyByID(w http.ResponseWriter, r *http.Request) {
	surveyID := strings.TrimPrefix(r.URL.Path, "/api/surveys/")
	if strings.Contains(surveyID, "/") {
		// Handle sub-routes
		s.handleSurveySubRoutes(w, r, surveyID)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	survey, err := s.contextStore.GetSentimentSurvey(surveyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get survey", err)
		return
	}
	if survey == nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, survey)
}

func (s *Server) handleSurveySubRoutes(w http.ResponseWriter, r *http.Request, path string) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	surveyID := parts[0]
	action := parts[1]

	switch action {
	case "responses":
		if r.Method == http.MethodPost {
			s.submitSurveyResponse(w, r, surveyID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "results":
		if r.Method == http.MethodGet {
			s.getSurveyResults(w, r, surveyID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}

func (s *Server) submitSurveyResponse(w http.ResponseWriter, r *http.Request, surveyID string) {
	var response models.SurveyResponse
	if err := decodeAndValidateJSON(r, &response); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	response.SurveyID = surveyID

	if err := s.contextStore.CreateSurveyResponse(&response); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to submit response", err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{
		"status": "submitted",
		"id":     response.ID,
	})
}

func (s *Server) getSurveyResults(w http.ResponseWriter, r *http.Request, surveyID string) {
	results, err := s.sentimentService.AnalyzeSurveyResults(surveyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to analyze survey", err)
		return
	}

	respondJSON(w, http.StatusOK, results)
}

// Sentiment Analysis Handlers

func (s *Server) handleSentimentTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teamID := strings.TrimPrefix(r.URL.Path, "/api/sentiment/team/")
	timePeriod := r.URL.Query().Get("time_period")

	if timePeriod == "" {
		timePeriod = "current"
	}

	summary, err := s.sentimentService.GenerateSentimentReport("team", teamID, timePeriod)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get team sentiment", err)
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

func (s *Server) handleSentimentEngineer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	engineerID := strings.TrimPrefix(r.URL.Path, "/api/sentiment/engineer/")
	timePeriod := r.URL.Query().Get("time_period")

	if timePeriod == "" {
		timePeriod = "current"
	}

	summary, err := s.sentimentService.GenerateSentimentReport("engineer", engineerID, timePeriod)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get engineer sentiment", err)
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

// Engineer Context Handler

func (s *Server) handleEngineerContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/engineers/"), "/")
	if len(parts) < 2 || parts[1] != "context" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	engineerID := parts[0]

	context, err := s.contextService.GetEngineerContext(engineerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get engineer context", err)
		return
	}
	if context == nil {
		http.Error(w, "Engineer not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, context)
}

func (s *Server) handleEngineerNotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/engineers/"), "/")
	if len(parts) < 2 || parts[1] != "notes" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	engineerID := parts[0]

	notes, _, err := s.contextStore.ListManagerNotes("", "engineer", engineerID, []string{"private", "shared_with_subject", "team", "org"}, 100, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get notes", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"notes": notes,
	})
}
