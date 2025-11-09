package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

type ContextStore struct {
	db *sql.DB
}

func NewContextStore(db *sql.DB) *ContextStore {
	return &ContextStore{db: db}
}

// Manager Notes CRUD

func (s *ContextStore) CreateManagerNote(note *models.ManagerNote) error {
	note.ID = uuid.New().String()
	note.CreatedAt = time.Now().UTC()
	note.UpdatedAt = time.Now().UTC()

	tagsJSON, _ := json.Marshal(note.Tags)
	actionItemsJSON, _ := json.Marshal(note.ActionItems)
	linkedEventsJSON, _ := json.Marshal(note.LinkedEvents)

	_, err := s.db.Exec(`
		INSERT INTO manager_notes (
			id, manager_id, subject_type, subject_id, note_type, title, content,
			visibility, tags, mood, action_items, linked_events, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, note.ID, note.ManagerID, note.SubjectType, note.SubjectID, note.NoteType, note.Title,
		note.Content, note.Visibility, string(tagsJSON), note.Mood,
		string(actionItemsJSON), string(linkedEventsJSON), note.CreatedAt, note.UpdatedAt)

	return err
}

func (s *ContextStore) GetManagerNote(id string) (*models.ManagerNote, error) {
	var note models.ManagerNote
	var tagsJSON, actionItemsJSON, linkedEventsJSON string
	var title, mood sql.NullString

	err := s.db.QueryRow(`
		SELECT id, manager_id, subject_type, subject_id, note_type, title, content,
		       visibility, tags, mood, action_items, linked_events, created_at, updated_at
		FROM manager_notes WHERE id = ?
	`, id).Scan(&note.ID, &note.ManagerID, &note.SubjectType, &note.SubjectID, &note.NoteType,
		&title, &note.Content, &note.Visibility, &tagsJSON, &mood,
		&actionItemsJSON, &linkedEventsJSON, &note.CreatedAt, &note.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if title.Valid {
		note.Title = title.String
	}
	if mood.Valid {
		note.Mood = mood.String
	}

	json.Unmarshal([]byte(tagsJSON), &note.Tags)
	json.Unmarshal([]byte(actionItemsJSON), &note.ActionItems)
	json.Unmarshal([]byte(linkedEventsJSON), &note.LinkedEvents)

	return &note, nil
}

func (s *ContextStore) ListManagerNotes(managerID, subjectType, subjectID string, visibility []string, limit, offset int) ([]*models.ManagerNote, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = MaxQueryLimit
	}

	// Build WHERE clause
	whereClause := " WHERE 1=1"
	args := []interface{}{}

	if managerID != "" {
		whereClause += " AND manager_id = ?"
		args = append(args, managerID)
	}
	if subjectType != "" {
		whereClause += " AND subject_type = ?"
		args = append(args, subjectType)
	}
	if subjectID != "" {
		whereClause += " AND subject_id = ?"
		args = append(args, subjectID)
	}
	if len(visibility) > 0 {
		whereClause += " AND visibility IN ("
		for i, v := range visibility {
			if i > 0 {
				whereClause += ","
			}
			whereClause += "?"
			args = append(args, v)
		}
		whereClause += ")"
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM manager_notes" + whereClause
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query with pagination
	query := `
		SELECT id, manager_id, subject_type, subject_id, note_type, title, content,
		       visibility, tags, mood, action_items, linked_events, created_at, updated_at
		FROM manager_notes` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	paginatedArgs := append(args, limit, offset)
	rows, err := s.db.Query(query, paginatedArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notes []*models.ManagerNote
	for rows.Next() {
		var note models.ManagerNote
		var tagsJSON, actionItemsJSON, linkedEventsJSON string
		var title, mood sql.NullString

		err := rows.Scan(&note.ID, &note.ManagerID, &note.SubjectType, &note.SubjectID,
			&note.NoteType, &title, &note.Content, &note.Visibility, &tagsJSON, &mood,
			&actionItemsJSON, &linkedEventsJSON, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}

		if title.Valid {
			note.Title = title.String
		}
		if mood.Valid {
			note.Mood = mood.String
		}

		json.Unmarshal([]byte(tagsJSON), &note.Tags)
		json.Unmarshal([]byte(actionItemsJSON), &note.ActionItems)
		json.Unmarshal([]byte(linkedEventsJSON), &note.LinkedEvents)

		notes = append(notes, &note)
	}

	return notes, total, rows.Err()
}

func (s *ContextStore) UpdateManagerNote(note *models.ManagerNote) error {
	note.UpdatedAt = time.Now().UTC()

	tagsJSON, _ := json.Marshal(note.Tags)
	actionItemsJSON, _ := json.Marshal(note.ActionItems)
	linkedEventsJSON, _ := json.Marshal(note.LinkedEvents)

	result, err := s.db.Exec(`
		UPDATE manager_notes SET
			title = ?, content = ?, visibility = ?, tags = ?, mood = ?,
			action_items = ?, linked_events = ?, updated_at = ?
		WHERE id = ?
	`, note.Title, note.Content, note.Visibility, string(tagsJSON), note.Mood,
		string(actionItemsJSON), string(linkedEventsJSON), note.UpdatedAt, note.ID)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("note not found: %s", note.ID)
	}

	return nil
}

func (s *ContextStore) DeleteManagerNote(id string) error {
	result, err := s.db.Exec(`DELETE FROM manager_notes WHERE id = ?`, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("note not found: %s", id)
	}

	return nil
}

// Context Annotations CRUD

func (s *ContextStore) CreateContextAnnotation(annotation *models.ContextAnnotation) error {
	annotation.ID = uuid.New().String()
	annotation.CreatedAt = time.Now().UTC()

	_, err := s.db.Exec(`
		INSERT INTO context_annotations (
			id, entity_type, entity_id, annotation_type, content, author_id, visibility, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, annotation.ID, annotation.EntityType, annotation.EntityID, annotation.AnnotationType,
		annotation.Content, annotation.AuthorID, annotation.Visibility, annotation.CreatedAt)

	return err
}

func (s *ContextStore) GetContextAnnotations(entityType, entityID string, limit, offset int) ([]*models.ContextAnnotation, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = MaxQueryLimit
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM context_annotations WHERE entity_type = ? AND entity_id = ?"
	var total int
	err := s.db.QueryRow(countQuery, entityType, entityID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query with pagination
	query := `
		SELECT id, entity_type, entity_id, annotation_type, content, author_id, visibility, created_at
		FROM context_annotations WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, entityType, entityID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var annotations []*models.ContextAnnotation
	for rows.Next() {
		var ann models.ContextAnnotation
		err := rows.Scan(&ann.ID, &ann.EntityType, &ann.EntityID, &ann.AnnotationType,
			&ann.Content, &ann.AuthorID, &ann.Visibility, &ann.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		annotations = append(annotations, &ann)
	}

	return annotations, total, rows.Err()
}

// Team Context CRUD

func (s *ContextStore) CreateTeamContext(ctx *models.TeamContext) error {
	ctx.ID = uuid.New().String()
	ctx.CreatedAt = time.Now().UTC()

	_, err := s.db.Exec(`
		INSERT INTO team_context (
			id, team_id, context_type, time_period, title, description, impact, author_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ctx.ID, ctx.TeamID, ctx.ContextType, ctx.TimePeriod, ctx.Title,
		ctx.Description, ctx.Impact, ctx.AuthorID, ctx.CreatedAt)

	return err
}

func (s *ContextStore) GetTeamContext(teamID, timePeriod string) ([]*models.TeamContext, error) {
	query := `
		SELECT id, team_id, context_type, time_period, title, description, impact, author_id, created_at
		FROM team_context WHERE team_id = ?
	`
	args := []interface{}{teamID}

	if timePeriod != "" {
		query += " AND time_period = ?"
		args = append(args, timePeriod)
	}

	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contexts []*models.TeamContext
	for rows.Next() {
		var ctx models.TeamContext
		var impact sql.NullString
		err := rows.Scan(&ctx.ID, &ctx.TeamID, &ctx.ContextType, &ctx.TimePeriod,
			&ctx.Title, &ctx.Description, &impact, &ctx.AuthorID, &ctx.CreatedAt)
		if err != nil {
			return nil, err
		}
		if impact.Valid {
			ctx.Impact = impact.String
		}
		contexts = append(contexts, &ctx)
	}

	return contexts, rows.Err()
}

// Sentiment Surveys CRUD

func (s *ContextStore) CreateSentimentSurvey(survey *models.SentimentSurvey) error {
	survey.ID = uuid.New().String()
	survey.CreatedAt = time.Now().UTC()

	questionsJSON, _ := json.Marshal(survey.Questions)

	_, err := s.db.Exec(`
		INSERT INTO sentiment_surveys (
			id, survey_type, title, description, questions, target_audience,
			anonymous, active, created_by, created_at, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, survey.ID, survey.SurveyType, survey.Title, survey.Description, string(questionsJSON),
		survey.TargetAudience, survey.Anonymous, survey.Active, survey.CreatedBy,
		survey.CreatedAt, survey.ExpiresAt)

	return err
}

func (s *ContextStore) GetSentimentSurvey(id string) (*models.SentimentSurvey, error) {
	var survey models.SentimentSurvey
	var questionsJSON string
	var description sql.NullString
	var expiresAt sql.NullTime

	err := s.db.QueryRow(`
		SELECT id, survey_type, title, description, questions, target_audience,
		       anonymous, active, created_by, created_at, expires_at
		FROM sentiment_surveys WHERE id = ?
	`, id).Scan(&survey.ID, &survey.SurveyType, &survey.Title, &description, &questionsJSON,
		&survey.TargetAudience, &survey.Anonymous, &survey.Active, &survey.CreatedBy,
		&survey.CreatedAt, &expiresAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if description.Valid {
		survey.Description = description.String
	}
	if expiresAt.Valid {
		survey.ExpiresAt = &expiresAt.Time
	}

	json.Unmarshal([]byte(questionsJSON), &survey.Questions)

	return &survey, nil
}

func (s *ContextStore) ListActiveSurveys(limit, offset int) ([]*models.SentimentSurvey, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = MaxQueryLimit
	}

	// Count total matching records
	countQuery := `
		SELECT COUNT(*) FROM sentiment_surveys
		WHERE active = 1 AND (expires_at IS NULL OR expires_at > ?)
	`
	var total int
	err := s.db.QueryRow(countQuery, time.Now().UTC()).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query with pagination
	rows, err := s.db.Query(`
		SELECT id, survey_type, title, description, questions, target_audience,
		       anonymous, active, created_by, created_at, expires_at
		FROM sentiment_surveys
		WHERE active = 1 AND (expires_at IS NULL OR expires_at > ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, time.Now().UTC(), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var surveys []*models.SentimentSurvey
	for rows.Next() {
		var survey models.SentimentSurvey
		var questionsJSON string
		var description sql.NullString
		var expiresAt sql.NullTime

		err := rows.Scan(&survey.ID, &survey.SurveyType, &survey.Title, &description,
			&questionsJSON, &survey.TargetAudience, &survey.Anonymous, &survey.Active,
			&survey.CreatedBy, &survey.CreatedAt, &expiresAt)
		if err != nil {
			return nil, 0, err
		}

		if description.Valid {
			survey.Description = description.String
		}
		if expiresAt.Valid {
			survey.ExpiresAt = &expiresAt.Time
		}

		json.Unmarshal([]byte(questionsJSON), &survey.Questions)
		surveys = append(surveys, &survey)
	}

	return surveys, total, rows.Err()
}

// Survey Responses CRUD

func (s *ContextStore) CreateSurveyResponse(response *models.SurveyResponse) error {
	response.ID = uuid.New().String()
	response.SubmittedAt = time.Now().UTC()

	responsesJSON, _ := json.Marshal(response.Responses)

	_, err := s.db.Exec(`
		INSERT INTO survey_responses (
			id, survey_id, respondent_id, responses, mood_rating, text_feedback, submitted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, response.ID, response.SurveyID, response.RespondentID, string(responsesJSON),
		response.MoodRating, response.TextFeedback, response.SubmittedAt)

	return err
}

func (s *ContextStore) GetSurveyResponses(surveyID string) ([]*models.SurveyResponse, error) {
	rows, err := s.db.Query(`
		SELECT id, survey_id, respondent_id, responses, mood_rating, text_feedback, submitted_at
		FROM survey_responses WHERE survey_id = ?
		ORDER BY submitted_at DESC
		LIMIT 10000
	`, surveyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responses []*models.SurveyResponse
	for rows.Next() {
		var resp models.SurveyResponse
		var responsesJSON string
		var respondentID sql.NullString
		var moodRating sql.NullInt64
		var textFeedback sql.NullString

		err := rows.Scan(&resp.ID, &resp.SurveyID, &respondentID, &responsesJSON,
			&moodRating, &textFeedback, &resp.SubmittedAt)
		if err != nil {
			return nil, err
		}

		if respondentID.Valid {
			s := respondentID.String
			resp.RespondentID = &s
		}
		if moodRating.Valid {
			r := int(moodRating.Int64)
			resp.MoodRating = &r
		}
		if textFeedback.Valid {
			resp.TextFeedback = textFeedback.String
		}

		json.Unmarshal([]byte(responsesJSON), &resp.Responses)
		responses = append(responses, &resp)
	}

	return responses, rows.Err()
}

// Sentiment Analysis CRUD

func (s *ContextStore) CreateSentimentAnalysis(analysis *models.SentimentAnalysis) error {
	analysis.ID = uuid.New().String()
	analysis.ComputedAt = time.Now().UTC()

	detailsJSON, _ := json.Marshal(analysis.Details)

	_, err := s.db.Exec(`
		INSERT INTO sentiment_analysis (
			id, entity_type, entity_id, time_period, source, sentiment_score,
			confidence, sample_size, details, computed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, analysis.ID, analysis.EntityType, analysis.EntityID, analysis.TimePeriod,
		analysis.Source, analysis.SentimentScore, analysis.Confidence,
		analysis.SampleSize, string(detailsJSON), analysis.ComputedAt)

	return err
}

func (s *ContextStore) GetSentimentAnalysis(entityType, entityID, timePeriod string) ([]*models.SentimentAnalysis, error) {
	query := `
		SELECT id, entity_type, entity_id, time_period, source, sentiment_score,
		       confidence, sample_size, details, computed_at
		FROM sentiment_analysis WHERE entity_type = ?
	`
	args := []interface{}{entityType}

	if entityID != "" {
		query += " AND entity_id = ?"
		args = append(args, entityID)
	}
	if timePeriod != "" {
		query += " AND time_period = ?"
		args = append(args, timePeriod)
	}

	query += " ORDER BY computed_at DESC LIMIT 1000"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analyses []*models.SentimentAnalysis
	for rows.Next() {
		var analysis models.SentimentAnalysis
		var detailsJSON string
		var entityID sql.NullString
		var sampleSize sql.NullInt64

		err := rows.Scan(&analysis.ID, &analysis.EntityType, &entityID, &analysis.TimePeriod,
			&analysis.Source, &analysis.SentimentScore, &analysis.Confidence,
			&sampleSize, &detailsJSON, &analysis.ComputedAt)
		if err != nil {
			return nil, err
		}

		if entityID.Valid {
			analysis.EntityID = entityID.String
		}
		if sampleSize.Valid {
			analysis.SampleSize = int(sampleSize.Int64)
		}

		json.Unmarshal([]byte(detailsJSON), &analysis.Details)
		analyses = append(analyses, &analysis)
	}

	return analyses, rows.Err()
}
