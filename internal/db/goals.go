package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// GoalsStore handles goal storage operations
type GoalsStore struct {
	db *sql.DB
}

// NewGoalsStore creates a new goals store
func NewGoalsStore(db *sql.DB) *GoalsStore {
	return &GoalsStore{db: db}
}

// Goal CRUD operations

// CreateGoal creates a new goal
func (s *GoalsStore) CreateGoal(goal *models.Goal) error {
	if goal.ID == "" {
		goal.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if goal.CreatedAt.IsZero() {
		goal.CreatedAt = now
	}
	goal.UpdatedAt = now

	_, err := s.db.Exec(`
		INSERT INTO goals (
			id, title, description, goal_type, owner_type, owner_id,
			time_period, start_date, end_date, tracking_method,
			success_criteria, status, progress_percentage,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		goal.ID, goal.Title, goal.Description, goal.GoalType, goal.OwnerType, goal.OwnerID,
		goal.TimePeriod, goal.StartDate.UTC(), goal.EndDate.UTC(), goal.TrackingMethod,
		goal.SuccessCriteria, goal.Status, goal.ProgressPercentage,
		goal.CreatedAt, goal.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create goal: %w", err)
	}

	return nil
}

// GetGoal retrieves a goal by ID
func (s *GoalsStore) GetGoal(id string) (*models.Goal, error) {
	var goal models.Goal
	var ownerID, description, successCriteria sql.NullString

	err := s.db.QueryRow(`
		SELECT id, title, description, goal_type, owner_type, owner_id,
		       time_period, start_date, end_date, tracking_method,
		       success_criteria, status, progress_percentage,
		       created_at, updated_at
		FROM goals
		WHERE id = ?
	`, id).Scan(
		&goal.ID, &goal.Title, &description, &goal.GoalType, &goal.OwnerType, &ownerID,
		&goal.TimePeriod, &goal.StartDate, &goal.EndDate, &goal.TrackingMethod,
		&successCriteria, &goal.Status, &goal.ProgressPercentage,
		&goal.CreatedAt, &goal.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	if ownerID.Valid {
		goal.OwnerID = &ownerID.String
	}
	if description.Valid {
		goal.Description = description.String
	}
	if successCriteria.Valid {
		goal.SuccessCriteria = successCriteria.String
	}

	return &goal, nil
}

// ListGoals retrieves goals with optional filtering
func (s *GoalsStore) ListGoals(ownerType, ownerID, status, timePeriod string, limit, offset int) ([]*models.Goal, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Build WHERE clause for both count and data queries
	whereClause := " WHERE 1=1"
	args := []interface{}{}

	if ownerType != "" {
		whereClause += " AND owner_type = ?"
		args = append(args, ownerType)
	}
	if ownerID != "" {
		whereClause += " AND owner_id = ?"
		args = append(args, ownerID)
	}
	if status != "" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}
	if timePeriod != "" {
		whereClause += " AND time_period = ?"
		args = append(args, timePeriod)
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM goals" + whereClause
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count goals: %w", err)
	}

	// Query with pagination
	query := `
		SELECT id, title, description, goal_type, owner_type, owner_id,
		       time_period, start_date, end_date, tracking_method,
		       success_criteria, status, progress_percentage,
		       created_at, updated_at
		FROM goals` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	paginatedArgs := append(args, limit, offset)
	rows, err := s.db.Query(query, paginatedArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list goals: %w", err)
	}
	defer rows.Close()

	var goals []*models.Goal
	for rows.Next() {
		var goal models.Goal
		var ownerID, description, successCriteria sql.NullString

		err := rows.Scan(
			&goal.ID, &goal.Title, &description, &goal.GoalType, &goal.OwnerType, &ownerID,
			&goal.TimePeriod, &goal.StartDate, &goal.EndDate, &goal.TrackingMethod,
			&successCriteria, &goal.Status, &goal.ProgressPercentage,
			&goal.CreatedAt, &goal.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan goal: %w", err)
		}

		if ownerID.Valid {
			goal.OwnerID = &ownerID.String
		}
		if description.Valid {
			goal.Description = description.String
		}
		if successCriteria.Valid {
			goal.SuccessCriteria = successCriteria.String
		}

		goals = append(goals, &goal)
	}

	return goals, total, rows.Err()
}

// UpdateGoal updates an existing goal
func (s *GoalsStore) UpdateGoal(goal *models.Goal) error {
	goal.UpdatedAt = time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE goals
		SET title = ?, description = ?, goal_type = ?, owner_type = ?,
		    owner_id = ?, time_period = ?, start_date = ?, end_date = ?,
		    tracking_method = ?, success_criteria = ?, status = ?,
		    progress_percentage = ?, updated_at = ?
		WHERE id = ?
	`,
		goal.Title, goal.Description, goal.GoalType, goal.OwnerType,
		goal.OwnerID, goal.TimePeriod, goal.StartDate.UTC(), goal.EndDate.UTC(),
		goal.TrackingMethod, goal.SuccessCriteria, goal.Status,
		goal.ProgressPercentage, goal.UpdatedAt, goal.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update goal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("goal not found: %s", goal.ID)
	}

	return nil
}

// DeleteGoal deletes a goal
func (s *GoalsStore) DeleteGoal(id string) error {
	result, err := s.db.Exec("DELETE FROM goals WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete goal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("goal not found: %s", id)
	}

	return nil
}

// Milestone operations

// CreateMilestone creates a new milestone
func (s *GoalsStore) CreateMilestone(milestone *models.GoalMilestone) error {
	if milestone.ID == "" {
		milestone.ID = uuid.New().String()
	}
	if milestone.CreatedAt.IsZero() {
		milestone.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO goal_milestones (
			id, goal_id, title, description, target_value,
			current_value, unit, completed, completed_at,
			due_date, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		milestone.ID, milestone.GoalID, milestone.Title, milestone.Description,
		milestone.TargetValue, milestone.CurrentValue, milestone.Unit,
		milestone.Completed, milestone.CompletedAt, milestone.DueDate,
		milestone.CreatedAt,
	)

	return err
}

// GetMilestoneByID retrieves a single milestone by ID
func (s *GoalsStore) GetMilestoneByID(milestoneID string) (*models.GoalMilestone, error) {
	var m models.GoalMilestone
	var description, unit sql.NullString
	var targetValue, currentValue sql.NullFloat64
	var completedAt, dueDate sql.NullTime

	err := s.db.QueryRow(`
		SELECT id, goal_id, title, description, target_value,
		       current_value, unit, completed, completed_at,
		       due_date, created_at
		FROM goal_milestones
		WHERE id = ?
	`, milestoneID).Scan(
		&m.ID, &m.GoalID, &m.Title, &description, &targetValue,
		&currentValue, &unit, &m.Completed, &completedAt,
		&dueDate, &m.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("milestone not found: %s", milestoneID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}

	if description.Valid {
		m.Description = description.String
	}
	if targetValue.Valid {
		m.TargetValue = targetValue.Float64
	}
	if currentValue.Valid {
		m.CurrentValue = currentValue.Float64
	}
	if unit.Valid {
		m.Unit = unit.String
	}
	if completedAt.Valid {
		m.CompletedAt = &completedAt.Time
	}
	if dueDate.Valid {
		m.DueDate = &dueDate.Time
	}

	return &m, nil
}

// GetMilestones retrieves milestones for a goal
func (s *GoalsStore) GetMilestones(goalID string, limit, offset int) ([]*models.GoalMilestone, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Count total milestones for this goal
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM goal_milestones WHERE goal_id = ?", goalID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count milestones: %w", err)
	}

	// Query with pagination
	rows, err := s.db.Query(`
		SELECT id, goal_id, title, description, target_value,
		       current_value, unit, completed, completed_at,
		       due_date, created_at
		FROM goal_milestones
		WHERE goal_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`, goalID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*models.GoalMilestone
	for rows.Next() {
		var m models.GoalMilestone
		var description, unit sql.NullString
		var targetValue, currentValue sql.NullFloat64
		var completedAt, dueDate sql.NullTime

		err := rows.Scan(
			&m.ID, &m.GoalID, &m.Title, &description, &targetValue,
			&currentValue, &unit, &m.Completed, &completedAt,
			&dueDate, &m.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan milestone: %w", err)
		}

		if description.Valid {
			m.Description = description.String
		}
		if targetValue.Valid {
			m.TargetValue = targetValue.Float64
		}
		if currentValue.Valid {
			m.CurrentValue = currentValue.Float64
		}
		if unit.Valid {
			m.Unit = unit.String
		}
		if completedAt.Valid {
			m.CompletedAt = &completedAt.Time
		}
		if dueDate.Valid {
			m.DueDate = &dueDate.Time
		}

		milestones = append(milestones, &m)
	}

	return milestones, total, rows.Err()
}

// UpdateMilestone updates a milestone
func (s *GoalsStore) UpdateMilestone(milestone *models.GoalMilestone) error {
	_, err := s.db.Exec(`
		UPDATE goal_milestones
		SET title = ?, description = ?, target_value = ?,
		    current_value = ?, unit = ?, completed = ?,
		    completed_at = ?, due_date = ?
		WHERE id = ?
	`,
		milestone.Title, milestone.Description, milestone.TargetValue,
		milestone.CurrentValue, milestone.Unit, milestone.Completed,
		milestone.CompletedAt, milestone.DueDate, milestone.ID,
	)

	return err
}

// Progress log operations

// LogProgress creates a progress log entry
func (s *GoalsStore) LogProgress(log *models.GoalProgressLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.LoggedAt.IsZero() {
		log.LoggedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO goal_progress_logs (
			id, goal_id, milestone_id, previous_value, new_value,
			change_type, evidence, logged_at, logged_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		log.ID, log.GoalID, log.MilestoneID, log.PreviousValue, log.NewValue,
		log.ChangeType, log.Evidence, log.LoggedAt, log.LoggedBy,
	)

	return err
}

// GetProgressLogs retrieves progress logs for a goal
func (s *GoalsStore) GetProgressLogs(goalID string, limit, offset int) ([]*models.GoalProgressLog, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 20
	}
	if limit > 1000 {
		limit = 1000
	}

	// Count total progress logs for this goal
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM goal_progress_logs WHERE goal_id = ?", goalID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count progress logs: %w", err)
	}

	// Query with pagination
	query := `
		SELECT id, goal_id, milestone_id, previous_value, new_value,
		       change_type, evidence, logged_at, logged_by
		FROM goal_progress_logs
		WHERE goal_id = ?
		ORDER BY logged_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, goalID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get progress logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.GoalProgressLog
	for rows.Next() {
		var log models.GoalProgressLog
		var milestoneID, evidence, loggedBy sql.NullString
		var previousValue, newValue sql.NullFloat64

		err := rows.Scan(
			&log.ID, &log.GoalID, &milestoneID, &previousValue, &newValue,
			&log.ChangeType, &evidence, &log.LoggedAt, &loggedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan progress log: %w", err)
		}

		if milestoneID.Valid {
			log.MilestoneID = &milestoneID.String
		}
		if previousValue.Valid {
			log.PreviousValue = &previousValue.Float64
		}
		if newValue.Valid {
			log.NewValue = &newValue.Float64
		}
		if evidence.Valid {
			log.Evidence = evidence.String
		}
		if loggedBy.Valid {
			log.LoggedBy = &loggedBy.String
		}

		logs = append(logs, &log)
	}

	return logs, total, rows.Err()
}

// Dependency operations

// CreateDependency creates a goal dependency
func (s *GoalsStore) CreateDependency(dep *models.GoalDependency) error {
	if dep.ID == "" {
		dep.ID = uuid.New().String()
	}
	if dep.CreatedAt.IsZero() {
		dep.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO goal_dependencies (
			id, goal_id, depends_on_goal_id, dependency_type, status, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		dep.ID, dep.GoalID, dep.DependsOnGoalID, dep.DependencyType, dep.Status, dep.CreatedAt,
	)

	return err
}

// GetDependencies retrieves dependencies for a goal
func (s *GoalsStore) GetDependencies(goalID string, limit, offset int) ([]*models.GoalDependency, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Count total dependencies for this goal
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM goal_dependencies WHERE goal_id = ?", goalID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count dependencies: %w", err)
	}

	// Query with pagination
	rows, err := s.db.Query(`
		SELECT id, goal_id, depends_on_goal_id, dependency_type, status, created_at
		FROM goal_dependencies
		WHERE goal_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`, goalID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get dependencies: %w", err)
	}
	defer rows.Close()

	var deps []*models.GoalDependency
	for rows.Next() {
		var dep models.GoalDependency
		err := rows.Scan(
			&dep.ID, &dep.GoalID, &dep.DependsOnGoalID,
			&dep.DependencyType, &dep.Status, &dep.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan dependency: %w", err)
		}
		deps = append(deps, &dep)
	}

	return deps, total, rows.Err()
}

// Goal metric operations

// CreateGoalMetric creates a goal metric
func (s *GoalsStore) CreateGoalMetric(metric *models.GoalMetric) error {
	if metric.ID == "" {
		metric.ID = uuid.New().String()
	}

	_, err := s.db.Exec(`
		INSERT INTO goal_metrics (
			id, goal_id, metric_name, target_value, current_value,
			operator, last_evaluated
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		metric.ID, metric.GoalID, metric.MetricName, metric.TargetValue,
		metric.CurrentValue, metric.Operator, metric.LastEvaluated,
	)

	return err
}

// GetGoalMetrics retrieves metrics for a goal
func (s *GoalsStore) GetGoalMetrics(goalID string, limit, offset int) ([]*models.GoalMetric, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Count total metrics for this goal
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM goal_metrics WHERE goal_id = ?", goalID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count goal metrics: %w", err)
	}

	// Query with pagination
	rows, err := s.db.Query(`
		SELECT id, goal_id, metric_name, target_value, current_value,
		       operator, last_evaluated
		FROM goal_metrics
		WHERE goal_id = ?
		ORDER BY last_evaluated DESC
		LIMIT ? OFFSET ?
	`, goalID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get goal metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*models.GoalMetric
	for rows.Next() {
		var m models.GoalMetric
		var currentValue sql.NullFloat64
		var lastEvaluated sql.NullTime

		err := rows.Scan(
			&m.ID, &m.GoalID, &m.MetricName, &m.TargetValue, &currentValue,
			&m.Operator, &lastEvaluated,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan goal metric: %w", err)
		}

		if currentValue.Valid {
			m.CurrentValue = &currentValue.Float64
		}
		if lastEvaluated.Valid {
			m.LastEvaluated = &lastEvaluated.Time
		}

		metrics = append(metrics, &m)
	}

	return metrics, total, rows.Err()
}

// UpdateGoalMetric updates a goal metric
func (s *GoalsStore) UpdateGoalMetric(metric *models.GoalMetric) error {
	now := time.Now().UTC()
	metric.LastEvaluated = &now

	_, err := s.db.Exec(`
		UPDATE goal_metrics
		SET current_value = ?, last_evaluated = ?
		WHERE id = ?
	`,
		metric.CurrentValue, metric.LastEvaluated, metric.ID,
	)

	return err
}
