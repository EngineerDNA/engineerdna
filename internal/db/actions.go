package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

type ActionsStore struct {
	db *sql.DB
}

func NewActionsStore(db *sql.DB) *ActionsStore {
	return &ActionsStore{db: db}
}

// Recommendations CRUD

func (s *ActionsStore) CreateRecommendation(rec *models.Recommendation) error {
	rec.ID = uuid.New().String()
	rec.CreatedAt = time.Now().UTC()
	if rec.Status == "" {
		rec.Status = "pending"
	}

	suggestedActionsJSON, _ := json.Marshal(rec.SuggestedActions)

	_, err := s.db.Exec(`
		INSERT INTO recommendations (
			id, source_type, source_id, recommendation_type, priority,
			subject_type, subject_id, title, description, suggested_actions,
			context, assigned_to, created_at, expires_at, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rec.ID, rec.SourceType, rec.SourceID, rec.RecommendationType, rec.Priority,
		rec.SubjectType, rec.SubjectID, rec.Title, rec.Description, string(suggestedActionsJSON),
		rec.Context, rec.AssignedTo, rec.CreatedAt, rec.ExpiresAt, rec.Status)

	if err != nil {
		return fmt.Errorf("failed to create recommendation: %w", err)
	}

	return nil
}

func (s *ActionsStore) GetRecommendation(id string) (*models.Recommendation, error) {
	var rec models.Recommendation
	var sourceID, subjectID, suggestedActionsJSON, context, assignedTo sql.NullString
	var expiresAt sql.NullString

	err := s.db.QueryRow(`
		SELECT id, source_type, source_id, recommendation_type, priority,
		       subject_type, subject_id, title, description, suggested_actions,
		       context, assigned_to, created_at, expires_at, status
		FROM recommendations
		WHERE id = ?
	`, id).Scan(&rec.ID, &rec.SourceType, &sourceID, &rec.RecommendationType, &rec.Priority,
		&rec.SubjectType, &subjectID, &rec.Title, &rec.Description, &suggestedActionsJSON,
		&context, &assignedTo, &rec.CreatedAt, &expiresAt, &rec.Status)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendation: %w", err)
	}

	if sourceID.Valid {
		sid := sourceID.String
		rec.SourceID = &sid
	}
	if subjectID.Valid {
		suid := subjectID.String
		rec.SubjectID = &suid
	}
	if suggestedActionsJSON.Valid {
		json.Unmarshal([]byte(suggestedActionsJSON.String), &rec.SuggestedActions)
	}
	if context.Valid {
		rec.Context = context.String
	}
	if assignedTo.Valid {
		at := assignedTo.String
		rec.AssignedTo = &at
	}
	if expiresAt.Valid {
		t, _ := time.Parse(time.RFC3339, expiresAt.String)
		rec.ExpiresAt = &t
	}

	return &rec, nil
}

func (s *ActionsStore) ListRecommendations(filters map[string]string, limit, offset int) ([]*models.Recommendation, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > MaxQueryLimit {
		limit = MaxQueryLimit
	}

	// Build WHERE clause
	whereClause := " WHERE 1=1"
	args := []interface{}{}

	if assignedTo, ok := filters["assigned_to"]; ok && assignedTo != "" {
		whereClause += " AND assigned_to = ?"
		args = append(args, assignedTo)
	}
	if status, ok := filters["status"]; ok && status != "" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}
	if subjectType, ok := filters["subject_type"]; ok && subjectType != "" {
		whereClause += " AND subject_type = ?"
		args = append(args, subjectType)
	}
	if subjectID, ok := filters["subject_id"]; ok && subjectID != "" {
		whereClause += " AND subject_id = ?"
		args = append(args, subjectID)
	}
	if sourceType, ok := filters["source_type"]; ok && sourceType != "" {
		whereClause += " AND source_type = ?"
		args = append(args, sourceType)
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM recommendations" + whereClause
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count recommendations: %w", err)
	}

	// Query with pagination
	query := `
		SELECT id, source_type, source_id, recommendation_type, priority,
		       subject_type, subject_id, title, description, suggested_actions,
		       context, assigned_to, created_at, expires_at, status
		FROM recommendations` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	paginatedArgs := append(args, limit, offset)

	rows, err := s.db.Query(query, paginatedArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list recommendations: %w", err)
	}
	defer rows.Close()

	var recommendations []*models.Recommendation
	for rows.Next() {
		var rec models.Recommendation
		var sourceID, subjectID, suggestedActionsJSON, context, assignedTo sql.NullString
		var expiresAt sql.NullString

		err := rows.Scan(&rec.ID, &rec.SourceType, &sourceID, &rec.RecommendationType, &rec.Priority,
			&rec.SubjectType, &subjectID, &rec.Title, &rec.Description, &suggestedActionsJSON,
			&context, &assignedTo, &rec.CreatedAt, &expiresAt, &rec.Status)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan recommendation: %w", err)
		}

		if sourceID.Valid {
			sid := sourceID.String
			rec.SourceID = &sid
		}
		if subjectID.Valid {
			suid := subjectID.String
			rec.SubjectID = &suid
		}
		if suggestedActionsJSON.Valid {
			json.Unmarshal([]byte(suggestedActionsJSON.String), &rec.SuggestedActions)
		}
		if context.Valid {
			rec.Context = context.String
		}
		if assignedTo.Valid {
			at := assignedTo.String
			rec.AssignedTo = &at
		}
		if expiresAt.Valid {
			t, _ := time.Parse(time.RFC3339, expiresAt.String)
			rec.ExpiresAt = &t
		}

		recommendations = append(recommendations, &rec)
	}

	return recommendations, total, rows.Err()
}

func (s *ActionsStore) UpdateRecommendationStatus(id, status, changedBy, reason string) error {
	result, err := s.db.Exec(`
		UPDATE recommendations
		SET status = ?
		WHERE id = ?
	`, status, id)

	if err != nil {
		return fmt.Errorf("failed to update recommendation status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("recommendation not found: %s", id)
	}

	// Log history
	history := &models.RecommendationHistory{
		RecommendationID: id,
		StatusChange:     status,
		Reason:           reason,
	}
	if changedBy != "" {
		history.ChangedBy = &changedBy
	}
	s.CreateRecommendationHistory(history)

	return nil
}

func (s *ActionsStore) DeleteRecommendation(id string) error {
	result, err := s.db.Exec(`DELETE FROM recommendations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete recommendation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("recommendation not found: %s", id)
	}

	return nil
}

// Actions CRUD

func (s *ActionsStore) CreateAction(action *models.Action) error {
	action.ID = uuid.New().String()
	action.TakenAt = time.Now().UTC()

	_, err := s.db.Exec(`
		INSERT INTO actions (
			id, recommendation_id, action_type, subject_type, subject_id,
			title, description, taken_by, taken_at, evidence,
			expected_outcome, follow_up_date
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, action.ID, action.RecommendationID, action.ActionType, action.SubjectType, action.SubjectID,
		action.Title, action.Description, action.TakenBy, action.TakenAt, action.Evidence,
		action.ExpectedOutcome, action.FollowUpDate)

	if err != nil {
		return fmt.Errorf("failed to create action: %w", err)
	}

	// If linked to recommendation, update recommendation status
	if action.RecommendationID != nil {
		s.UpdateRecommendationStatus(*action.RecommendationID, "in_progress", action.TakenBy, "Action taken")
	}

	return nil
}

func (s *ActionsStore) GetAction(id string) (*models.Action, error) {
	var action models.Action
	var recommendationID, subjectID, evidence, expectedOutcome sql.NullString
	var followUpDate sql.NullString

	err := s.db.QueryRow(`
		SELECT id, recommendation_id, action_type, subject_type, subject_id,
		       title, description, taken_by, taken_at, evidence,
		       expected_outcome, follow_up_date
		FROM actions
		WHERE id = ?
	`, id).Scan(&action.ID, &recommendationID, &action.ActionType, &action.SubjectType, &subjectID,
		&action.Title, &action.Description, &action.TakenBy, &action.TakenAt, &evidence,
		&expectedOutcome, &followUpDate)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get action: %w", err)
	}

	if recommendationID.Valid {
		rid := recommendationID.String
		action.RecommendationID = &rid
	}
	if subjectID.Valid {
		sid := subjectID.String
		action.SubjectID = &sid
	}
	if evidence.Valid {
		action.Evidence = evidence.String
	}
	if expectedOutcome.Valid {
		action.ExpectedOutcome = expectedOutcome.String
	}
	if followUpDate.Valid {
		t, _ := time.Parse(time.RFC3339, followUpDate.String)
		action.FollowUpDate = &t
	}

	return &action, nil
}

func (s *ActionsStore) ListActions(filters map[string]string, limit int) ([]*models.Action, error) {
	query := `
		SELECT id, recommendation_id, action_type, subject_type, subject_id,
		       title, description, taken_by, taken_at, evidence,
		       expected_outcome, follow_up_date
		FROM actions
		WHERE 1=1
	`
	args := []interface{}{}

	if takenBy, ok := filters["taken_by"]; ok && takenBy != "" {
		query += " AND taken_by = ?"
		args = append(args, takenBy)
	}
	if recommendationID, ok := filters["recommendation_id"]; ok && recommendationID != "" {
		query += " AND recommendation_id = ?"
		args = append(args, recommendationID)
	}
	if subjectType, ok := filters["subject_type"]; ok && subjectType != "" {
		query += " AND subject_type = ?"
		args = append(args, subjectType)
	}
	if subjectID, ok := filters["subject_id"]; ok && subjectID != "" {
		query += " AND subject_id = ?"
		args = append(args, subjectID)
	}

	query += " ORDER BY taken_at DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list actions: %w", err)
	}
	defer rows.Close()

	var actions []*models.Action
	for rows.Next() {
		var action models.Action
		var recommendationID, subjectID, evidence, expectedOutcome sql.NullString
		var followUpDate sql.NullString

		err := rows.Scan(&action.ID, &recommendationID, &action.ActionType, &action.SubjectType, &subjectID,
			&action.Title, &action.Description, &action.TakenBy, &action.TakenAt, &evidence,
			&expectedOutcome, &followUpDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}

		if recommendationID.Valid {
			rid := recommendationID.String
			action.RecommendationID = &rid
		}
		if subjectID.Valid {
			sid := subjectID.String
			action.SubjectID = &sid
		}
		if evidence.Valid {
			action.Evidence = evidence.String
		}
		if expectedOutcome.Valid {
			action.ExpectedOutcome = expectedOutcome.String
		}
		if followUpDate.Valid {
			t, _ := time.Parse(time.RFC3339, followUpDate.String)
			action.FollowUpDate = &t
		}

		actions = append(actions, &action)
	}

	return actions, rows.Err()
}

// Action Outcomes

func (s *ActionsStore) CreateActionOutcome(outcome *models.ActionOutcome) error {
	outcome.ID = uuid.New().String()
	outcome.MeasuredAt = time.Now().UTC()

	_, err := s.db.Exec(`
		INSERT INTO action_outcomes (
			id, action_id, outcome_type, measured_metric, before_value,
			after_value, change_percentage, time_to_impact_days,
			effectiveness, notes, measured_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, outcome.ID, outcome.ActionID, outcome.OutcomeType, outcome.MeasuredMetric, outcome.BeforeValue,
		outcome.AfterValue, outcome.ChangePercentage, outcome.TimeToImpactDays,
		outcome.Effectiveness, outcome.Notes, outcome.MeasuredAt)

	if err != nil {
		return fmt.Errorf("failed to create action outcome: %w", err)
	}

	return nil
}

func (s *ActionsStore) GetActionOutcomes(actionID string) ([]*models.ActionOutcome, error) {
	rows, err := s.db.Query(`
		SELECT id, action_id, outcome_type, measured_metric, before_value,
		       after_value, change_percentage, time_to_impact_days,
		       effectiveness, notes, measured_at
		FROM action_outcomes
		WHERE action_id = ?
		ORDER BY measured_at DESC
	`, actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get action outcomes: %w", err)
	}
	defer rows.Close()

	var outcomes []*models.ActionOutcome
	for rows.Next() {
		var outcome models.ActionOutcome
		var beforeValue, afterValue, changePercentage sql.NullFloat64
		var timeToImpactDays sql.NullInt64
		var notes sql.NullString

		err := rows.Scan(&outcome.ID, &outcome.ActionID, &outcome.OutcomeType, &outcome.MeasuredMetric, &beforeValue,
			&afterValue, &changePercentage, &timeToImpactDays,
			&outcome.Effectiveness, &notes, &outcome.MeasuredAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action outcome: %w", err)
		}

		if beforeValue.Valid {
			bv := beforeValue.Float64
			outcome.BeforeValue = &bv
		}
		if afterValue.Valid {
			av := afterValue.Float64
			outcome.AfterValue = &av
		}
		if changePercentage.Valid {
			cp := changePercentage.Float64
			outcome.ChangePercentage = &cp
		}
		if timeToImpactDays.Valid {
			ti := int(timeToImpactDays.Int64)
			outcome.TimeToImpactDays = &ti
		}
		if notes.Valid {
			outcome.Notes = notes.String
		}

		outcomes = append(outcomes, &outcome)
	}

	return outcomes, rows.Err()
}

// Recommendation History

func (s *ActionsStore) CreateRecommendationHistory(history *models.RecommendationHistory) error {
	history.ID = uuid.New().String()
	history.ChangedAt = time.Now().UTC()

	_, err := s.db.Exec(`
		INSERT INTO recommendation_history (
			id, recommendation_id, status_change, changed_by, reason, changed_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`, history.ID, history.RecommendationID, history.StatusChange, history.ChangedBy,
		history.Reason, history.ChangedAt)

	if err != nil {
		return fmt.Errorf("failed to create recommendation history: %w", err)
	}

	return nil
}

func (s *ActionsStore) GetRecommendationHistory(recommendationID string) ([]*models.RecommendationHistory, error) {
	rows, err := s.db.Query(`
		SELECT id, recommendation_id, status_change, changed_by, reason, changed_at
		FROM recommendation_history
		WHERE recommendation_id = ?
		ORDER BY changed_at DESC
	`, recommendationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendation history: %w", err)
	}
	defer rows.Close()

	var history []*models.RecommendationHistory
	for rows.Next() {
		var h models.RecommendationHistory
		var changedBy, reason sql.NullString

		err := rows.Scan(&h.ID, &h.RecommendationID, &h.StatusChange, &changedBy, &reason, &h.ChangedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recommendation history: %w", err)
		}

		if changedBy.Valid {
			cb := changedBy.String
			h.ChangedBy = &cb
		}
		if reason.Valid {
			h.Reason = reason.String
		}

		history = append(history, &h)
	}

	return history, rows.Err()
}

// Follow-Ups

func (s *ActionsStore) CreateFollowUp(followUp *models.FollowUp) error {
	followUp.ID = uuid.New().String()

	_, err := s.db.Exec(`
		INSERT INTO follow_ups (
			id, action_id, follow_up_date, follow_up_type, description, completed, completed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, followUp.ID, followUp.ActionID, followUp.FollowUpDate, followUp.FollowUpType,
		followUp.Description, followUp.Completed, followUp.CompletedAt)

	if err != nil {
		return fmt.Errorf("failed to create follow-up: %w", err)
	}

	return nil
}

func (s *ActionsStore) GetFollowUpsDue(dueDate time.Time) ([]*models.FollowUp, error) {
	rows, err := s.db.Query(`
		SELECT id, action_id, follow_up_date, follow_up_type, description, completed, completed_at
		FROM follow_ups
		WHERE follow_up_date <= ? AND completed = false
		ORDER BY follow_up_date
	`, dueDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get due follow-ups: %w", err)
	}
	defer rows.Close()

	var followUps []*models.FollowUp
	for rows.Next() {
		var followUp models.FollowUp
		var description sql.NullString
		var completedAt sql.NullString

		err := rows.Scan(&followUp.ID, &followUp.ActionID, &followUp.FollowUpDate,
			&followUp.FollowUpType, &description, &followUp.Completed, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follow-up: %w", err)
		}

		if description.Valid {
			followUp.Description = description.String
		}
		if completedAt.Valid {
			t, _ := time.Parse(time.RFC3339, completedAt.String)
			followUp.CompletedAt = &t
		}

		followUps = append(followUps, &followUp)
	}

	return followUps, rows.Err()
}

func (s *ActionsStore) CompleteFollowUp(id string) error {
	now := time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE follow_ups
		SET completed = true, completed_at = ?
		WHERE id = ?
	`, now, id)

	if err != nil {
		return fmt.Errorf("failed to complete follow-up: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("follow-up not found: %s", id)
	}

	return nil
}
