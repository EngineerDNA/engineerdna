package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// PlanningStore handles sprint, epic, and story storage operations
type PlanningStore struct {
	db *sql.DB
}

// NewPlanningStore creates a new planning store
func NewPlanningStore(db *sql.DB) *PlanningStore {
	return &PlanningStore{db: db}
}

// Sprint CRUD operations

// CreateSprint creates a new sprint
func (s *PlanningStore) CreateSprint(sprint *models.Sprint) error {
	if sprint.ID == "" {
		sprint.ID = uuid.New().String()
	}
	if sprint.CreatedAt.IsZero() {
		sprint.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO sprints (
			id, name, start_date, end_date,
			committed_points, completed_points, team_capacity,
			status, team_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		sprint.ID, sprint.Name, sprint.StartDate.UTC(), sprint.EndDate.UTC(),
		sprint.CommittedPoints, sprint.CompletedPoints, sprint.TeamCapacity,
		sprint.Status, sprint.TeamID, sprint.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create sprint: %w", err)
	}

	return nil
}

// GetSprint retrieves a sprint by ID
func (s *PlanningStore) GetSprint(id string) (*models.Sprint, error) {
	var sprint models.Sprint
	var committedPoints, completedPoints, teamCapacity sql.NullInt64
	var status, teamID sql.NullString

	err := s.db.QueryRow(`
		SELECT id, name, start_date, end_date,
		       committed_points, completed_points, team_capacity,
		       status, team_id, created_at
		FROM sprints
		WHERE id = ?
	`, id).Scan(
		&sprint.ID, &sprint.Name, &sprint.StartDate, &sprint.EndDate,
		&committedPoints, &completedPoints, &teamCapacity,
		&status, &teamID, &sprint.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sprint: %w", err)
	}

	// Handle NULL values
	if committedPoints.Valid {
		sprint.CommittedPoints = int(committedPoints.Int64)
	}
	if completedPoints.Valid {
		sprint.CompletedPoints = int(completedPoints.Int64)
	}
	if teamCapacity.Valid {
		sprint.TeamCapacity = int(teamCapacity.Int64)
	}
	if status.Valid {
		sprint.Status = status.String
	}
	if teamID.Valid {
		sprint.TeamID = teamID.String
	}

	return &sprint, nil
}

// ListSprints retrieves all sprints, ordered by start date (newest first)
func (s *PlanningStore) ListSprints(limit int) ([]*models.Sprint, error) {
	// Always enforce limits
	if limit <= 0 {
		limit = 100 // Default
	}
	if limit > 1000 {
		limit = 1000 // Maximum
	}

	query := `
		SELECT id, name, start_date, end_date,
		       committed_points, completed_points, team_capacity,
		       status, team_id, created_at
		FROM sprints
		ORDER BY start_date DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query sprints: %w", err)
	}
	defer rows.Close()

	var sprints []*models.Sprint
	for rows.Next() {
		var sprint models.Sprint
		var committedPoints, completedPoints, teamCapacity sql.NullInt64
		var status, teamID sql.NullString

		err := rows.Scan(
			&sprint.ID, &sprint.Name, &sprint.StartDate, &sprint.EndDate,
			&committedPoints, &completedPoints, &teamCapacity,
			&status, &teamID, &sprint.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sprint: %w", err)
		}

		if committedPoints.Valid {
			sprint.CommittedPoints = int(committedPoints.Int64)
		}
		if completedPoints.Valid {
			sprint.CompletedPoints = int(completedPoints.Int64)
		}
		if teamCapacity.Valid {
			sprint.TeamCapacity = int(teamCapacity.Int64)
		}
		if status.Valid {
			sprint.Status = status.String
		}
		if teamID.Valid {
			sprint.TeamID = teamID.String
		}

		sprints = append(sprints, &sprint)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sprints: %w", err)
	}

	return sprints, nil
}

// UpdateSprint updates an existing sprint
func (s *PlanningStore) UpdateSprint(sprint *models.Sprint) error {
	result, err := s.db.Exec(`
		UPDATE sprints
		SET name = ?, start_date = ?, end_date = ?,
		    committed_points = ?, completed_points = ?, team_capacity = ?,
		    status = ?, team_id = ?
		WHERE id = ?
	`,
		sprint.Name, sprint.StartDate.UTC(), sprint.EndDate.UTC(),
		sprint.CommittedPoints, sprint.CompletedPoints, sprint.TeamCapacity,
		sprint.Status, sprint.TeamID, sprint.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update sprint: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("sprint not found: %s", sprint.ID)
	}

	return nil
}

// GetSprintsByTeam retrieves sprints for a specific team, ordered by start date
func (s *PlanningStore) GetSprintsByTeam(teamID string, limit int) ([]*models.Sprint, error) {
	// Always enforce limits
	if limit <= 0 {
		limit = 100 // Default
	}
	if limit > 1000 {
		limit = 1000 // Maximum
	}

	query := `
		SELECT id, name, start_date, end_date,
		       committed_points, completed_points, team_capacity,
		       status, team_id, created_at
		FROM sprints
		WHERE team_id = ?
		ORDER BY start_date DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, teamID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query sprints by team: %w", err)
	}
	defer rows.Close()

	var sprints []*models.Sprint
	for rows.Next() {
		var sprint models.Sprint
		var committedPoints, completedPoints, teamCapacity sql.NullInt64
		var status, teamID sql.NullString

		err := rows.Scan(
			&sprint.ID, &sprint.Name, &sprint.StartDate, &sprint.EndDate,
			&committedPoints, &completedPoints, &teamCapacity,
			&status, &teamID, &sprint.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sprint: %w", err)
		}

		if committedPoints.Valid {
			sprint.CommittedPoints = int(committedPoints.Int64)
		}
		if completedPoints.Valid {
			sprint.CompletedPoints = int(completedPoints.Int64)
		}
		if teamCapacity.Valid {
			sprint.TeamCapacity = int(teamCapacity.Int64)
		}
		if status.Valid {
			sprint.Status = status.String
		}
		if teamID.Valid {
			sprint.TeamID = teamID.String
		}

		sprints = append(sprints, &sprint)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sprints: %w", err)
	}

	return sprints, nil
}

// Epic CRUD operations

// CreateEpic creates a new epic
func (s *PlanningStore) CreateEpic(epic *models.Epic) error {
	if epic.ID == "" {
		epic.ID = uuid.New().String()
	}
	if epic.CreatedAt.IsZero() {
		epic.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO epics (
			id, jira_id, title, description,
			original_estimate_points, current_scope_points,
			created_at, started_at, completed_at,
			status, team_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		epic.ID, epic.JiraID, epic.Title, epic.Description,
		epic.OriginalEstimatePoints, epic.CurrentScopePoints,
		epic.CreatedAt, epic.StartedAt, epic.CompletedAt,
		epic.Status, epic.TeamID,
	)

	if err != nil {
		return fmt.Errorf("failed to create epic: %w", err)
	}

	return nil
}

// GetEpic retrieves an epic by ID
func (s *PlanningStore) GetEpic(id string) (*models.Epic, error) {
	var epic models.Epic
	var jiraID, description sql.NullString
	var originalEstimate, currentScope sql.NullInt64
	var startedAt, completedAt sql.NullTime
	var status, teamID sql.NullString

	err := s.db.QueryRow(`
		SELECT id, jira_id, title, description,
		       original_estimate_points, current_scope_points,
		       created_at, started_at, completed_at,
		       status, team_id
		FROM epics
		WHERE id = ?
	`, id).Scan(
		&epic.ID, &jiraID, &epic.Title, &description,
		&originalEstimate, &currentScope,
		&epic.CreatedAt, &startedAt, &completedAt,
		&status, &teamID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}

	// Handle NULL values
	if jiraID.Valid {
		epic.JiraID = &jiraID.String
	}
	if description.Valid {
		epic.Description = description.String
	}
	if originalEstimate.Valid {
		epic.OriginalEstimatePoints = int(originalEstimate.Int64)
	}
	if currentScope.Valid {
		epic.CurrentScopePoints = int(currentScope.Int64)
	}
	if startedAt.Valid {
		epic.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		epic.CompletedAt = &completedAt.Time
	}
	if status.Valid {
		epic.Status = status.String
	}
	if teamID.Valid {
		epic.TeamID = teamID.String
	}

	return &epic, nil
}

// ListEpics retrieves all epics, ordered by created date (newest first)
func (s *PlanningStore) ListEpics(limit int) ([]*models.Epic, error) {
	// Always enforce limits
	if limit <= 0 {
		limit = 100 // Default
	}
	if limit > 1000 {
		limit = 1000 // Maximum
	}

	query := `
		SELECT id, jira_id, title, description,
		       original_estimate_points, current_scope_points,
		       created_at, started_at, completed_at,
		       status, team_id
		FROM epics
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query epics: %w", err)
	}
	defer rows.Close()

	var epics []*models.Epic
	for rows.Next() {
		var epic models.Epic
		var jiraID, description sql.NullString
		var originalEstimate, currentScope sql.NullInt64
		var startedAt, completedAt sql.NullTime
		var status, teamID sql.NullString

		err := rows.Scan(
			&epic.ID, &jiraID, &epic.Title, &description,
			&originalEstimate, &currentScope,
			&epic.CreatedAt, &startedAt, &completedAt,
			&status, &teamID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan epic: %w", err)
		}

		if jiraID.Valid {
			epic.JiraID = &jiraID.String
		}
		if description.Valid {
			epic.Description = description.String
		}
		if originalEstimate.Valid {
			epic.OriginalEstimatePoints = int(originalEstimate.Int64)
		}
		if currentScope.Valid {
			epic.CurrentScopePoints = int(currentScope.Int64)
		}
		if startedAt.Valid {
			epic.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			epic.CompletedAt = &completedAt.Time
		}
		if status.Valid {
			epic.Status = status.String
		}
		if teamID.Valid {
			epic.TeamID = teamID.String
		}

		epics = append(epics, &epic)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating epics: %w", err)
	}

	return epics, nil
}

// UpdateEpic updates an existing epic
func (s *PlanningStore) UpdateEpic(epic *models.Epic) error {
	result, err := s.db.Exec(`
		UPDATE epics
		SET jira_id = ?, title = ?, description = ?,
		    original_estimate_points = ?, current_scope_points = ?,
		    started_at = ?, completed_at = ?,
		    status = ?, team_id = ?
		WHERE id = ?
	`,
		epic.JiraID, epic.Title, epic.Description,
		epic.OriginalEstimatePoints, epic.CurrentScopePoints,
		epic.StartedAt, epic.CompletedAt,
		epic.Status, epic.TeamID, epic.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update epic: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("epic not found: %s", epic.ID)
	}

	return nil
}

// Story CRUD operations

// CreateStory creates a new story
func (s *PlanningStore) CreateStory(story *models.Story) error {
	if story.ID == "" {
		story.ID = uuid.New().String()
	}
	if story.CreatedAt.IsZero() {
		story.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO stories (
			id, jira_id, epic_id, sprint_id, title,
			story_points, status,
			created_at, started_at, completed_at,
			actual_days_to_complete, assignee_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		story.ID, story.JiraID, story.EpicID, story.SprintID, story.Title,
		story.StoryPoints, story.Status,
		story.CreatedAt, story.StartedAt, story.CompletedAt,
		story.ActualDaysToComplete, story.AssigneeID,
	)

	if err != nil {
		return fmt.Errorf("failed to create story: %w", err)
	}

	return nil
}

// GetStory retrieves a story by ID
func (s *PlanningStore) GetStory(id string) (*models.Story, error) {
	var story models.Story
	var jiraID, epicID, sprintID sql.NullString
	var storyPoints sql.NullInt64
	var status sql.NullString
	var startedAt, completedAt sql.NullTime
	var actualDays sql.NullFloat64
	var assigneeID sql.NullString

	err := s.db.QueryRow(`
		SELECT id, jira_id, epic_id, sprint_id, title,
		       story_points, status,
		       created_at, started_at, completed_at,
		       actual_days_to_complete, assignee_id
		FROM stories
		WHERE id = ?
	`, id).Scan(
		&story.ID, &jiraID, &epicID, &sprintID, &story.Title,
		&storyPoints, &status,
		&story.CreatedAt, &startedAt, &completedAt,
		&actualDays, &assigneeID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get story: %w", err)
	}

	// Handle NULL values
	if jiraID.Valid {
		story.JiraID = &jiraID.String
	}
	if epicID.Valid {
		story.EpicID = &epicID.String
	}
	if sprintID.Valid {
		story.SprintID = &sprintID.String
	}
	if storyPoints.Valid {
		story.StoryPoints = int(storyPoints.Int64)
	}
	if status.Valid {
		story.Status = status.String
	}
	if startedAt.Valid {
		story.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		story.CompletedAt = &completedAt.Time
	}
	if actualDays.Valid {
		story.ActualDaysToComplete = &actualDays.Float64
	}
	if assigneeID.Valid {
		story.AssigneeID = &assigneeID.String
	}

	return &story, nil
}

// ListStories retrieves all stories, ordered by created date (newest first)
func (s *PlanningStore) ListStories(limit int) ([]*models.Story, error) {
	// Always enforce limits
	if limit <= 0 {
		limit = 100 // Default
	}
	if limit > 1000 {
		limit = 1000 // Maximum
	}

	query := `
		SELECT id, jira_id, epic_id, sprint_id, title,
		       story_points, status,
		       created_at, started_at, completed_at,
		       actual_days_to_complete, assignee_id
		FROM stories
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query stories: %w", err)
	}
	defer rows.Close()

	var stories []*models.Story
	for rows.Next() {
		var story models.Story
		var jiraID, epicID, sprintID sql.NullString
		var storyPoints sql.NullInt64
		var status sql.NullString
		var startedAt, completedAt sql.NullTime
		var actualDays sql.NullFloat64
		var assigneeID sql.NullString

		err := rows.Scan(
			&story.ID, &jiraID, &epicID, &sprintID, &story.Title,
			&storyPoints, &status,
			&story.CreatedAt, &startedAt, &completedAt,
			&actualDays, &assigneeID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan story: %w", err)
		}

		if jiraID.Valid {
			story.JiraID = &jiraID.String
		}
		if epicID.Valid {
			story.EpicID = &epicID.String
		}
		if sprintID.Valid {
			story.SprintID = &sprintID.String
		}
		if storyPoints.Valid {
			story.StoryPoints = int(storyPoints.Int64)
		}
		if status.Valid {
			story.Status = status.String
		}
		if startedAt.Valid {
			story.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			story.CompletedAt = &completedAt.Time
		}
		if actualDays.Valid {
			story.ActualDaysToComplete = &actualDays.Float64
		}
		if assigneeID.Valid {
			story.AssigneeID = &assigneeID.String
		}

		stories = append(stories, &story)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stories: %w", err)
	}

	return stories, nil
}

// UpdateStory updates an existing story
func (s *PlanningStore) UpdateStory(story *models.Story) error {
	result, err := s.db.Exec(`
		UPDATE stories
		SET jira_id = ?, epic_id = ?, sprint_id = ?, title = ?,
		    story_points = ?, status = ?,
		    started_at = ?, completed_at = ?,
		    actual_days_to_complete = ?, assignee_id = ?
		WHERE id = ?
	`,
		story.JiraID, story.EpicID, story.SprintID, story.Title,
		story.StoryPoints, story.Status,
		story.StartedAt, story.CompletedAt,
		story.ActualDaysToComplete, story.AssigneeID, story.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update story: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("story not found: %s", story.ID)
	}

	return nil
}

// GetStoriesBySprint retrieves all stories for a specific sprint
func (s *PlanningStore) GetStoriesBySprint(sprintID string) ([]*models.Story, error) {
	query := `
		SELECT id, jira_id, epic_id, sprint_id, title,
		       story_points, status,
		       created_at, started_at, completed_at,
		       actual_days_to_complete, assignee_id
		FROM stories
		WHERE sprint_id = ?
		ORDER BY created_at ASC
		LIMIT 2000
	`

	rows, err := s.db.Query(query, sprintID)
	if err != nil {
		return nil, fmt.Errorf("failed to query stories by sprint: %w", err)
	}
	defer rows.Close()

	var stories []*models.Story
	for rows.Next() {
		var story models.Story
		var jiraID, epicID, sprintIDVal sql.NullString
		var storyPoints sql.NullInt64
		var status sql.NullString
		var startedAt, completedAt sql.NullTime
		var actualDays sql.NullFloat64
		var assigneeID sql.NullString

		err := rows.Scan(
			&story.ID, &jiraID, &epicID, &sprintIDVal, &story.Title,
			&storyPoints, &status,
			&story.CreatedAt, &startedAt, &completedAt,
			&actualDays, &assigneeID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan story: %w", err)
		}

		if jiraID.Valid {
			story.JiraID = &jiraID.String
		}
		if epicID.Valid {
			story.EpicID = &epicID.String
		}
		if sprintIDVal.Valid {
			story.SprintID = &sprintIDVal.String
		}
		if storyPoints.Valid {
			story.StoryPoints = int(storyPoints.Int64)
		}
		if status.Valid {
			story.Status = status.String
		}
		if startedAt.Valid {
			story.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			story.CompletedAt = &completedAt.Time
		}
		if actualDays.Valid {
			story.ActualDaysToComplete = &actualDays.Float64
		}
		if assigneeID.Valid {
			story.AssigneeID = &assigneeID.String
		}

		stories = append(stories, &story)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stories: %w", err)
	}

	return stories, nil
}

// Capacity history operations

// CreateCapacityHistory stores capacity data for a week
func (s *PlanningStore) CreateCapacityHistory(history *models.CapacityHistory) error {
	if history.ID == "" {
		history.ID = uuid.New().String()
	}
	if history.CreatedAt.IsZero() {
		history.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO capacity_history (
			id, week_start, team_id, team_size,
			available_engineers, completed_points, meeting_hours,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		history.ID, history.WeekStart.UTC(), history.TeamID, history.TeamSize,
		history.AvailableEngineers, history.CompletedPoints, history.MeetingHours,
		history.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create capacity history: %w", err)
	}

	return nil
}

// GetCapacityHistory retrieves capacity history for a team
func (s *PlanningStore) GetCapacityHistory(teamID string, numWeeks int) ([]*models.CapacityHistory, error) {
	// Always enforce limits
	if numWeeks <= 0 {
		numWeeks = 12 // Default (12 weeks = 3 months)
	}
	if numWeeks > 52 {
		numWeeks = 52 // Maximum (1 year)
	}

	query := `
		SELECT id, week_start, team_id, team_size,
		       available_engineers, completed_points, meeting_hours,
		       created_at
		FROM capacity_history
		WHERE team_id = ?
		ORDER BY week_start DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, teamID, numWeeks)
	if err != nil {
		return nil, fmt.Errorf("failed to query capacity history: %w", err)
	}
	defer rows.Close()

	var history []*models.CapacityHistory
	for rows.Next() {
		var h models.CapacityHistory
		var teamSize sql.NullInt64
		var availableEngineers, meetingHours sql.NullFloat64
		var completedPoints sql.NullInt64

		err := rows.Scan(
			&h.ID, &h.WeekStart, &h.TeamID, &teamSize,
			&availableEngineers, &completedPoints, &meetingHours,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan capacity history: %w", err)
		}

		if teamSize.Valid {
			h.TeamSize = int(teamSize.Int64)
		}
		if availableEngineers.Valid {
			h.AvailableEngineers = availableEngineers.Float64
		}
		if completedPoints.Valid {
			h.CompletedPoints = int(completedPoints.Int64)
		}
		if meetingHours.Valid {
			h.MeetingHours = meetingHours.Float64
		}

		history = append(history, &h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating capacity history: %w", err)
	}

	return history, nil
}
