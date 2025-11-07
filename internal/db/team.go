package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

type TeamStore struct {
	db *sql.DB
}

func NewTeamStore(db *sql.DB) *TeamStore {
	return &TeamStore{db: db}
}

// Team CRUD operations

func (s *TeamStore) CreateTeam(name string, parentTeamID, managerID *string) (*models.Team, error) {
	team := &models.Team{
		ID:           uuid.New().String(),
		Name:         name,
		ParentTeamID: parentTeamID,
		ManagerID:    managerID,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	_, err := s.db.Exec(`
		INSERT INTO teams (id, name, parent_team_id, manager_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, team.ID, team.Name, team.ParentTeamID, team.ManagerID, team.CreatedAt, team.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return team, nil
}

func (s *TeamStore) GetTeam(id string) (*models.Team, error) {
	var team models.Team
	var parentTeamID, managerID sql.NullString

	err := s.db.QueryRow(`
		SELECT id, name, parent_team_id, manager_id, created_at, updated_at
		FROM teams
		WHERE id = ?
	`, id).Scan(&team.ID, &team.Name, &parentTeamID, &managerID, &team.CreatedAt, &team.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	if parentTeamID.Valid {
		team.ParentTeamID = &parentTeamID.String
	}
	if managerID.Valid {
		team.ManagerID = &managerID.String
	}

	return &team, nil
}

func (s *TeamStore) ListTeams(limit int) ([]*models.Team, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	rows, err := s.db.Query(`
		SELECT id, name, parent_team_id, manager_id, created_at, updated_at
		FROM teams
		ORDER BY name
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		var team models.Team
		var parentTeamID, managerID sql.NullString

		err := rows.Scan(&team.ID, &team.Name, &parentTeamID, &managerID, &team.CreatedAt, &team.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}

		if parentTeamID.Valid {
			team.ParentTeamID = &parentTeamID.String
		}
		if managerID.Valid {
			team.ManagerID = &managerID.String
		}

		teams = append(teams, &team)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating teams: %w", err)
	}

	return teams, nil
}

func (s *TeamStore) UpdateTeam(id, name string, parentTeamID, managerID *string) error {
	result, err := s.db.Exec(`
		UPDATE teams
		SET name = ?, parent_team_id = ?, manager_id = ?, updated_at = ?
		WHERE id = ?
	`, name, parentTeamID, managerID, time.Now().UTC(), id)

	if err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("team not found: %s", id)
	}

	return nil
}

func (s *TeamStore) DeleteTeam(id string) error {
	result, err := s.db.Exec(`DELETE FROM teams WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("team not found: %s", id)
	}

	return nil
}

// Team Membership operations

func (s *TeamStore) AddTeamMember(teamID, memberID, role string) (*models.TeamMembership, error) {
	membership := &models.TeamMembership{
		ID:       uuid.New().String(),
		TeamID:   teamID,
		MemberID: memberID,
		Role:     role,
		JoinedAt: time.Now().UTC(),
		LeftAt:   nil,
	}

	_, err := s.db.Exec(`
		INSERT INTO team_membership (id, team_id, member_id, role, joined_at, left_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, membership.ID, membership.TeamID, membership.MemberID, membership.Role, membership.JoinedAt, membership.LeftAt)

	if err != nil {
		return nil, fmt.Errorf("failed to add team member: %w", err)
	}

	return membership, nil
}

func (s *TeamStore) RemoveTeamMember(teamID, memberID string) error {
	now := time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE team_membership
		SET left_at = ?
		WHERE team_id = ? AND member_id = ? AND left_at IS NULL
	`, now, teamID, memberID)

	if err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("active team membership not found")
	}

	return nil
}

func (s *TeamStore) GetTeamMembers(teamID string, includeInactive bool) ([]*models.TeamMembership, error) {
	query := `
		SELECT id, team_id, member_id, role, joined_at, left_at
		FROM team_membership
		WHERE team_id = ?
	`

	if !includeInactive {
		query += ` AND left_at IS NULL`
	}

	query += ` ORDER BY joined_at DESC LIMIT 1000`

	rows, err := s.db.Query(query, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	defer rows.Close()

	var members []*models.TeamMembership
	for rows.Next() {
		var member models.TeamMembership
		var role sql.NullString
		var leftAt sql.NullTime

		err := rows.Scan(&member.ID, &member.TeamID, &member.MemberID, &role, &member.JoinedAt, &leftAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}

		if role.Valid {
			member.Role = role.String
		}
		if leftAt.Valid {
			member.LeftAt = &leftAt.Time
		}

		members = append(members, &member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team members: %w", err)
	}

	return members, nil
}

func (s *TeamStore) GetTeamWithMembers(teamID string) (*models.TeamWithMembers, error) {
	team, err := s.GetTeam(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	if team == nil {
		return nil, nil
	}

	// Get active members
	memberships, err := s.GetTeamMembers(teamID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	// Get engineer details for each membership
	var members []models.Engineer
	for _, membership := range memberships {
		var engineer models.Engineer
		var email, manager sql.NullString
		var identifiersJSON string

		err := s.db.QueryRow(`
			SELECT id, canonical_name, email, manager, identifiers, active, created_at, updated_at
			FROM engineers
			WHERE id = ?
		`, membership.MemberID).Scan(&engineer.ID, &engineer.Name, &email, &manager, &identifiersJSON, &engineer.Active, &engineer.CreatedAt, &engineer.UpdatedAt)

		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get engineer: %w", err)
		}

		if err == nil {
			if email.Valid {
				engineer.Email = email.String
			}
			if manager.Valid {
				engineer.Manager = manager.String
			}
			members = append(members, engineer)
		}
	}

	return &models.TeamWithMembers{
		Team:    *team,
		Members: members,
	}, nil
}

// GetTeamPerformanceScores retrieves performance scores for a team within a date range
// limit: max number of weeks to return (default 104 weeks = 2 years)
func (s *TeamStore) GetTeamPerformanceScores(teamID string, startDate, endDate time.Time, limit int) ([]*models.TeamPerformanceScore, error) {
	// Apply default and reasonable max limit
	if limit <= 0 {
		limit = 104 // 2 years of weekly data
	}
	if limit > 520 {
		limit = 520 // Max 10 years
	}

	rows, err := s.db.Query(`
		SELECT id, team_id, week_start, total_score, member_count,
		       throughput_score, quality_score, speed_score, collaboration_score, impact_score,
		       created_at
		FROM team_performance_scores
		WHERE team_id = ?
		  AND week_start >= ?
		  AND week_start <= ?
		ORDER BY week_start ASC
		LIMIT ?
	`, teamID, startDate, endDate, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to get team performance scores: %w", err)
	}
	defer rows.Close()

	var scores []*models.TeamPerformanceScore
	for rows.Next() {
		var score models.TeamPerformanceScore

		err := rows.Scan(
			&score.ID, &score.TeamID, &score.WeekStart, &score.TotalScore, &score.MemberCount,
			&score.ThroughputScore, &score.QualityScore, &score.SpeedScore,
			&score.CollaborationScore, &score.ImpactScore, &score.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team performance score: %w", err)
		}

		scores = append(scores, &score)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team performance scores: %w", err)
	}

	return scores, nil
}

// IsDescendant checks if potentialDescendant is a descendant of ancestor
func (s *TeamStore) IsDescendant(ancestor, potentialDescendant string) (bool, error) {
	if ancestor == potentialDescendant {
		return true, nil
	}

	// Walk up the tree from potentialDescendant
	currentID := potentialDescendant
	visited := make(map[string]bool)

	for currentID != "" {
		if visited[currentID] {
			return false, fmt.Errorf("circular reference detected in team hierarchy")
		}
		visited[currentID] = true

		var parentID sql.NullString
		err := s.db.QueryRow(`
			SELECT parent_team_id FROM teams WHERE id = ?
		`, currentID).Scan(&parentID)

		if err == sql.ErrNoRows {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("failed to check parent: %w", err)
		}

		if !parentID.Valid {
			return false, nil
		}

		if parentID.String == ancestor {
			return true, nil
		}

		currentID = parentID.String
	}

	return false, nil
}
