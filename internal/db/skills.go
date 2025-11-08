package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// SkillsStore handles skill storage operations
type SkillsStore struct {
	db *sql.DB
}

// NewSkillsStore creates a new skills store
func NewSkillsStore(db *sql.DB) *SkillsStore {
	return &SkillsStore{db: db}
}

// Skill CRUD operations

// CreateSkill creates a new skill
func (s *SkillsStore) CreateSkill(skill *models.Skill) error {
	if skill.ID == "" {
		skill.ID = uuid.New().String()
	}
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO skills (
			id, name, category, subcategory, description,
			measurement_criteria, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		skill.ID, skill.Name, skill.Category, skill.Subcategory,
		skill.Description, skill.MeasurementCriteria, skill.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create skill: %w", err)
	}

	return nil
}

// GetSkill retrieves a skill by ID
func (s *SkillsStore) GetSkill(id string) (*models.Skill, error) {
	var skill models.Skill
	var subcategory, description, criteria sql.NullString

	err := s.db.QueryRow(`
		SELECT id, name, category, subcategory, description,
		       measurement_criteria, created_at
		FROM skills
		WHERE id = ?
	`, id).Scan(
		&skill.ID, &skill.Name, &skill.Category, &subcategory,
		&description, &criteria, &skill.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get skill: %w", err)
	}

	if subcategory.Valid {
		skill.Subcategory = subcategory.String
	}
	if description.Valid {
		skill.Description = description.String
	}
	if criteria.Valid {
		skill.MeasurementCriteria = criteria.String
	}

	return &skill, nil
}

// ListSkills retrieves all skills with optional filtering
func (s *SkillsStore) ListSkills(category string, limit, offset int) ([]*models.Skill, int, error) {
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

	if category != "" {
		whereClause += " AND category = ?"
		args = append(args, category)
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM skills" + whereClause
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count skills: %w", err)
	}

	// Query with pagination
	query := `
		SELECT id, name, category, subcategory, description,
		       measurement_criteria, created_at
		FROM skills` + whereClause + `
		ORDER BY category, name
		LIMIT ? OFFSET ?
	`

	paginatedArgs := append(args, limit, offset)
	rows, err := s.db.Query(query, paginatedArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list skills: %w", err)
	}
	defer rows.Close()

	var skills []*models.Skill
	for rows.Next() {
		var skill models.Skill
		var subcategory, description, criteria sql.NullString

		err := rows.Scan(
			&skill.ID, &skill.Name, &skill.Category, &subcategory,
			&description, &criteria, &skill.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan skill: %w", err)
		}

		if subcategory.Valid {
			skill.Subcategory = subcategory.String
		}
		if description.Valid {
			skill.Description = description.String
		}
		if criteria.Valid {
			skill.MeasurementCriteria = criteria.String
		}

		skills = append(skills, &skill)
	}

	return skills, total, rows.Err()
}

// EngineerSkill operations

// CreateEngineerSkill creates a new engineer skill record
func (s *SkillsStore) CreateEngineerSkill(es *models.EngineerSkill) error {
	if es.ID == "" {
		es.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if es.CreatedAt.IsZero() {
		es.CreatedAt = now
	}
	es.UpdatedAt = now

	_, err := s.db.Exec(`
		INSERT INTO engineer_skills (
			id, engineer_id, skill_id, level_score, previous_level_score,
			trajectory, last_evaluated, evidence_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		es.ID, es.EngineerID, es.SkillID, es.LevelScore, es.PreviousLevelScore,
		es.Trajectory, es.LastEvaluated.UTC(), es.EvidenceCount,
		es.CreatedAt, es.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create engineer skill: %w", err)
	}

	return nil
}

// GetEngineerSkill retrieves a specific engineer skill
func (s *SkillsStore) GetEngineerSkill(engineerID, skillID string) (*models.EngineerSkill, error) {
	var es models.EngineerSkill
	var previousScore sql.NullInt64

	err := s.db.QueryRow(`
		SELECT id, engineer_id, skill_id, level_score, previous_level_score,
		       trajectory, last_evaluated, evidence_count, created_at, updated_at
		FROM engineer_skills
		WHERE engineer_id = ? AND skill_id = ?
	`, engineerID, skillID).Scan(
		&es.ID, &es.EngineerID, &es.SkillID, &es.LevelScore, &previousScore,
		&es.Trajectory, &es.LastEvaluated, &es.EvidenceCount,
		&es.CreatedAt, &es.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get engineer skill: %w", err)
	}

	if previousScore.Valid {
		score := int(previousScore.Int64)
		es.PreviousLevelScore = &score
	}

	return &es, nil
}

// ListEngineerSkills retrieves all skills for an engineer
func (s *SkillsStore) ListEngineerSkills(engineerID string, limit, offset int) ([]*models.EngineerSkill, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = MaxQueryLimit
	}

	// Count total engineer skills
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM engineer_skills WHERE engineer_id = ?", engineerID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count engineer skills: %w", err)
	}

	// Query with pagination
	rows, err := s.db.Query(`
		SELECT id, engineer_id, skill_id, level_score, previous_level_score,
		       trajectory, last_evaluated, evidence_count, created_at, updated_at
		FROM engineer_skills
		WHERE engineer_id = ?
		ORDER BY level_score DESC
		LIMIT ? OFFSET ?
	`, engineerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list engineer skills: %w", err)
	}
	defer rows.Close()

	var skills []*models.EngineerSkill
	for rows.Next() {
		var es models.EngineerSkill
		var previousScore sql.NullInt64

		err := rows.Scan(
			&es.ID, &es.EngineerID, &es.SkillID, &es.LevelScore, &previousScore,
			&es.Trajectory, &es.LastEvaluated, &es.EvidenceCount,
			&es.CreatedAt, &es.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan engineer skill: %w", err)
		}

		if previousScore.Valid {
			score := int(previousScore.Int64)
			es.PreviousLevelScore = &score
		}

		skills = append(skills, &es)
	}

	return skills, total, rows.Err()
}

// UpdateEngineerSkill updates an engineer skill
func (s *SkillsStore) UpdateEngineerSkill(es *models.EngineerSkill) error {
	es.UpdatedAt = time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE engineer_skills
		SET level_score = ?, previous_level_score = ?, trajectory = ?,
		    last_evaluated = ?, evidence_count = ?, updated_at = ?
		WHERE id = ?
	`,
		es.LevelScore, es.PreviousLevelScore, es.Trajectory,
		es.LastEvaluated.UTC(), es.EvidenceCount, es.UpdatedAt, es.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update engineer skill: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("engineer skill not found: %s", es.ID)
	}

	return nil
}

// SkillEvidence operations

// CreateSkillEvidence creates a new skill evidence record
func (s *SkillsStore) CreateSkillEvidence(evidence *models.SkillEvidence) error {
	if evidence.ID == "" {
		evidence.ID = uuid.New().String()
	}
	if evidence.DetectedAt.IsZero() {
		evidence.DetectedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO skill_evidence (
			id, engineer_id, skill_id, evidence_type, evidence_source,
			strength, context, detected_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		evidence.ID, evidence.EngineerID, evidence.SkillID,
		evidence.EvidenceType, evidence.EvidenceSource,
		evidence.Strength, evidence.Context, evidence.DetectedAt.UTC(),
	)

	if err != nil {
		return fmt.Errorf("failed to create skill evidence: %w", err)
	}

	return nil
}

// ListSkillEvidence retrieves evidence for a skill
func (s *SkillsStore) ListSkillEvidence(engineerID, skillID string, limit, offset int) ([]*models.SkillEvidence, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = MaxQueryLimit
	}

	// Count total evidence
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM skill_evidence WHERE engineer_id = ? AND skill_id = ?", engineerID, skillID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count skill evidence: %w", err)
	}

	// Query with pagination
	query := `
		SELECT id, engineer_id, skill_id, evidence_type, evidence_source,
		       strength, context, detected_at
		FROM skill_evidence
		WHERE engineer_id = ? AND skill_id = ?
		ORDER BY detected_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, engineerID, skillID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list skill evidence: %w", err)
	}
	defer rows.Close()

	var evidences []*models.SkillEvidence
	for rows.Next() {
		var ev models.SkillEvidence
		var source, context sql.NullString

		err := rows.Scan(
			&ev.ID, &ev.EngineerID, &ev.SkillID, &ev.EvidenceType, &source,
			&ev.Strength, &context, &ev.DetectedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan skill evidence: %w", err)
		}

		if source.Valid {
			ev.EvidenceSource = source.String
		}
		if context.Valid {
			ev.Context = context.String
		}

		evidences = append(evidences, &ev)
	}

	return evidences, total, rows.Err()
}

// GetAllSkillEvidence retrieves all evidence for an engineer
func (s *SkillsStore) GetAllSkillEvidence(engineerID string, since time.Time, limit int) ([]*models.SkillEvidence, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = MaxQueryLimit
	}

	rows, err := s.db.Query(`
		SELECT id, engineer_id, skill_id, evidence_type, evidence_source,
		       strength, context, detected_at
		FROM skill_evidence
		WHERE engineer_id = ? AND detected_at >= ?
		ORDER BY detected_at DESC
		LIMIT ?
	`, engineerID, since.UTC(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill evidence: %w", err)
	}
	defer rows.Close()

	var evidences []*models.SkillEvidence
	for rows.Next() {
		var ev models.SkillEvidence
		var source, context sql.NullString

		err := rows.Scan(
			&ev.ID, &ev.EngineerID, &ev.SkillID, &ev.EvidenceType, &source,
			&ev.Strength, &context, &ev.DetectedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan skill evidence: %w", err)
		}

		if source.Valid {
			ev.EvidenceSource = source.String
		}
		if context.Valid {
			ev.Context = context.String
		}

		evidences = append(evidences, &ev)
	}

	return evidences, rows.Err()
}

// SkillProgression operations

// CreateSkillProgression creates a progression history record
func (s *SkillsStore) CreateSkillProgression(prog *models.SkillProgression) error {
	if prog.ID == "" {
		prog.ID = uuid.New().String()
	}
	if prog.EvaluatedAt.IsZero() {
		prog.EvaluatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO skill_progression (
			id, engineer_id, skill_id, previous_score, new_score,
			change_reason, evaluated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		prog.ID, prog.EngineerID, prog.SkillID, prog.PreviousScore,
		prog.NewScore, prog.ChangeReason, prog.EvaluatedAt.UTC(),
	)

	if err != nil {
		return fmt.Errorf("failed to create skill progression: %w", err)
	}

	return nil
}

// GetSkillProgression retrieves progression history for a skill
func (s *SkillsStore) GetSkillProgression(engineerID, skillID string, limit, offset int) ([]*models.SkillProgression, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 20
	}
	if limit > 1000 {
		limit = MaxQueryLimit
	}

	// Count total progression records
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM skill_progression WHERE engineer_id = ? AND skill_id = ?", engineerID, skillID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count skill progression: %w", err)
	}

	// Query with pagination
	query := `
		SELECT id, engineer_id, skill_id, previous_score, new_score,
		       change_reason, evaluated_at
		FROM skill_progression
		WHERE engineer_id = ? AND skill_id = ?
		ORDER BY evaluated_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, engineerID, skillID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get skill progression: %w", err)
	}
	defer rows.Close()

	var progressions []*models.SkillProgression
	for rows.Next() {
		var prog models.SkillProgression
		var reason sql.NullString

		err := rows.Scan(
			&prog.ID, &prog.EngineerID, &prog.SkillID, &prog.PreviousScore,
			&prog.NewScore, &reason, &prog.EvaluatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan skill progression: %w", err)
		}

		if reason.Valid {
			prog.ChangeReason = reason.String
		}

		progressions = append(progressions, &prog)
	}

	return progressions, total, rows.Err()
}

// SkillGoal operations

// CreateSkillGoal creates a new skill goal
func (s *SkillsStore) CreateSkillGoal(goal *models.SkillGoal) error {
	if goal.ID == "" {
		goal.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if goal.CreatedAt.IsZero() {
		goal.CreatedAt = now
	}
	goal.UpdatedAt = now

	_, err := s.db.Exec(`
		INSERT INTO skill_goals (
			id, engineer_id, skill_id, goal_id, current_level, target_level,
			target_date, milestones, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		goal.ID, goal.EngineerID, goal.SkillID, goal.GoalID,
		goal.CurrentLevel, goal.TargetLevel, goal.TargetDate,
		goal.Milestones, goal.Status, goal.CreatedAt, goal.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create skill goal: %w", err)
	}

	return nil
}

// GetSkillGoals retrieves skill goals for an engineer
func (s *SkillsStore) GetSkillGoals(engineerID string, status string) ([]*models.SkillGoal, error) {
	query := `
		SELECT id, engineer_id, skill_id, goal_id, current_level, target_level,
		       target_date, milestones, status, created_at, updated_at
		FROM skill_goals
		WHERE engineer_id = ?
	`
	args := []interface{}{engineerID}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill goals: %w", err)
	}
	defer rows.Close()

	var goals []*models.SkillGoal
	for rows.Next() {
		var goal models.SkillGoal
		var goalID, milestones sql.NullString
		var targetDate sql.NullTime

		err := rows.Scan(
			&goal.ID, &goal.EngineerID, &goal.SkillID, &goalID,
			&goal.CurrentLevel, &goal.TargetLevel, &targetDate,
			&milestones, &goal.Status, &goal.CreatedAt, &goal.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan skill goal: %w", err)
		}

		if goalID.Valid {
			goal.GoalID = &goalID.String
		}
		if targetDate.Valid {
			goal.TargetDate = &targetDate.Time
		}
		if milestones.Valid {
			goal.Milestones = milestones.String
		}

		goals = append(goals, &goal)
	}

	return goals, rows.Err()
}

// UpdateSkillGoal updates a skill goal
func (s *SkillsStore) UpdateSkillGoal(goal *models.SkillGoal) error {
	goal.UpdatedAt = time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE skill_goals
		SET current_level = ?, target_level = ?, target_date = ?,
		    milestones = ?, status = ?, updated_at = ?
		WHERE id = ?
	`,
		goal.CurrentLevel, goal.TargetLevel, goal.TargetDate,
		goal.Milestones, goal.Status, goal.UpdatedAt, goal.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update skill goal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("skill goal not found: %s", goal.ID)
	}

	return nil
}
