package team

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// Service handles team management and aggregation
type Service struct {
	db *sql.DB
}

// NewService creates a new team service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// CalculateTeamScore calculates the aggregated performance score for a team
func (s *Service) CalculateTeamScore(teamID string, weekStart time.Time) (*models.TeamPerformanceScore, error) {
	// Normalize weekStart to beginning of day UTC
	weekStart = weekStart.UTC().Truncate(24 * time.Hour)

	// Get active members for this team
	members, err := s.getActiveMembers(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active members: %w", err)
	}

	if len(members) == 0 {
		return nil, nil // No score for empty teams
	}

	// Collect individual scores for active members
	var totalScore float64
	var throughputScore float64
	var qualityScore float64
	var speedScore float64
	var collaborationScore float64
	var impactScore float64
	var validScoreCount int

	for _, member := range members {
		score, err := s.getIndividualScore(member.MemberID, weekStart)
		if err != nil {
			// Log but continue - member might not have events this week
			continue
		}
		if score == nil {
			// No score for this member this week
			continue
		}

		totalScore += score.TotalScore
		throughputScore += score.ThroughputScore
		qualityScore += score.QualityScore
		speedScore += score.SpeedScore
		collaborationScore += score.CollaborationScore
		impactScore += score.ImpactScore
		validScoreCount++
	}

	// If no members have scores, return nil
	if validScoreCount == 0 {
		return nil, nil
	}

	// Calculate averages
	count := float64(validScoreCount)
	teamScore := &models.TeamPerformanceScore{
		ID:                 uuid.New().String(),
		TeamID:             teamID,
		WeekStart:          weekStart,
		TotalScore:         totalScore / count,
		MemberCount:        len(members), // Total active members, not just those with scores
		ThroughputScore:    throughputScore / count,
		QualityScore:       qualityScore / count,
		SpeedScore:         speedScore / count,
		CollaborationScore: collaborationScore / count,
		ImpactScore:        impactScore / count,
		CreatedAt:          time.Now().UTC(),
	}

	// Store in database
	err = s.storeTeamScore(teamScore)
	if err != nil {
		return nil, fmt.Errorf("failed to store team score: %w", err)
	}

	return teamScore, nil
}

// GetTeamHierarchy retrieves the team hierarchy starting from the given root team
func (s *Service) GetTeamHierarchy(rootTeamID string) (*models.TeamHierarchyNode, error) {
	// Get root team
	team, err := s.getTeamByID(rootTeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get root team: %w", err)
	}
	if team == nil {
		return nil, fmt.Errorf("team not found: %s", rootTeamID)
	}

	// Build hierarchy recursively
	// Track visited teams to prevent infinite loops from circular references
	visited := make(map[string]bool)
	return s.buildTeamHierarchy(team, visited)
}

// GetOrgWideScores retrieves scores for all top-level teams
func (s *Service) GetOrgWideScores(weekStart time.Time) ([]*models.TeamScorecard, error) {
	// Normalize weekStart to beginning of day UTC
	weekStart = weekStart.UTC().Truncate(24 * time.Hour)
	prevWeekStart := weekStart.AddDate(0, 0, -7)

	// Get all top-level teams (parent_team_id IS NULL)
	teams, err := s.getTopLevelTeams()
	if err != nil {
		return nil, fmt.Errorf("failed to get top-level teams: %w", err)
	}

	var scorecards []*models.TeamScorecard
	for _, team := range teams {
		// Get current week score
		currentScore, err := s.getStoredTeamScore(team.ID, weekStart)
		if err != nil {
			return nil, fmt.Errorf("failed to get current score for team %s: %w", team.ID, err)
		}

		// Get previous week score
		prevScore, err := s.getStoredTeamScore(team.ID, prevWeekStart)
		if err != nil {
			return nil, fmt.Errorf("failed to get previous score for team %s: %w", team.ID, err)
		}

		// Calculate score and change
		var score, prevScoreVal, change float64
		var memberCount int

		if currentScore != nil {
			score = currentScore.TotalScore
			memberCount = currentScore.MemberCount
		}

		if prevScore != nil {
			prevScoreVal = prevScore.TotalScore
		}

		change = score - prevScoreVal

		scorecard := &models.TeamScorecard{
			Team:          *team,
			Score:         score,
			PrevWeekScore: prevScoreVal,
			ScoreChange:   change,
			MemberCount:   memberCount,
			WeekStart:     weekStart,
		}

		scorecards = append(scorecards, scorecard)
	}

	return scorecards, nil
}

// getActiveMembers retrieves active members for a team
func (s *Service) getActiveMembers(teamID string) ([]*models.TeamMembership, error) {
	rows, err := s.db.Query(`
		SELECT id, team_id, member_id, role, joined_at, left_at
		FROM team_membership
		WHERE team_id = ?
		  AND left_at IS NULL
		LIMIT 1000
	`, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
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

// getIndividualScore retrieves an individual's performance score for a given week
func (s *Service) getIndividualScore(engineerID string, weekStart time.Time) (*models.PerformanceScore, error) {
	var score models.PerformanceScore

	err := s.db.QueryRow(`
		SELECT id, engineer_id, week_start, total_score,
		       throughput_score, quality_score, speed_score, collaboration_score, impact_score,
		       raw_metrics, created_at
		FROM performance_scores
		WHERE engineer_id = ?
		  AND week_start = ?
	`, engineerID, weekStart).Scan(
		&score.ID, &score.EngineerID, &score.WeekStart, &score.TotalScore,
		&score.ThroughputScore, &score.QualityScore, &score.SpeedScore, &score.CollaborationScore, &score.ImpactScore,
		&score.RawMetrics, &score.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query performance score: %w", err)
	}

	return &score, nil
}

// storeTeamScore stores a team performance score in the database
func (s *Service) storeTeamScore(score *models.TeamPerformanceScore) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO team_performance_scores (
			id, team_id, week_start, total_score, member_count,
			throughput_score, quality_score, speed_score, collaboration_score, impact_score,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, score.ID, score.TeamID, score.WeekStart, score.TotalScore, score.MemberCount,
		score.ThroughputScore, score.QualityScore, score.SpeedScore, score.CollaborationScore, score.ImpactScore,
		score.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert team performance score: %w", err)
	}

	return nil
}

// getTeamByID retrieves a team by ID
func (s *Service) getTeamByID(teamID string) (*models.Team, error) {
	var team models.Team
	var parentTeamID, managerID sql.NullString

	err := s.db.QueryRow(`
		SELECT id, name, parent_team_id, manager_id, created_at, updated_at
		FROM teams
		WHERE id = ?
	`, teamID).Scan(&team.ID, &team.Name, &parentTeamID, &managerID, &team.CreatedAt, &team.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query team: %w", err)
	}

	if parentTeamID.Valid {
		team.ParentTeamID = &parentTeamID.String
	}
	if managerID.Valid {
		team.ManagerID = &managerID.String
	}

	return &team, nil
}

// getTopLevelTeams retrieves all teams with no parent (top-level teams)
func (s *Service) getTopLevelTeams() ([]*models.Team, error) {
	rows, err := s.db.Query(`
		SELECT id, name, parent_team_id, manager_id, created_at, updated_at
		FROM teams
		WHERE parent_team_id IS NULL
		ORDER BY name
		LIMIT 100
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query top-level teams: %w", err)
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

// getStoredTeamScore retrieves a stored team performance score
func (s *Service) getStoredTeamScore(teamID string, weekStart time.Time) (*models.TeamPerformanceScore, error) {
	var score models.TeamPerformanceScore
	var totalScore, throughputScore, qualityScore, speedScore, collaborationScore, impactScore sql.NullFloat64
	var memberCount sql.NullInt64

	err := s.db.QueryRow(`
		SELECT id, team_id, week_start, total_score, member_count,
		       throughput_score, quality_score, speed_score, collaboration_score, impact_score,
		       created_at
		FROM team_performance_scores
		WHERE team_id = ?
		  AND week_start = ?
	`, teamID, weekStart).Scan(
		&score.ID, &score.TeamID, &score.WeekStart, &totalScore, &memberCount,
		&throughputScore, &qualityScore, &speedScore, &collaborationScore, &impactScore,
		&score.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query team performance score: %w", err)
	}

	// Handle NULL values
	if totalScore.Valid {
		score.TotalScore = totalScore.Float64
	}
	if memberCount.Valid {
		score.MemberCount = int(memberCount.Int64)
	}
	if throughputScore.Valid {
		score.ThroughputScore = throughputScore.Float64
	}
	if qualityScore.Valid {
		score.QualityScore = qualityScore.Float64
	}
	if speedScore.Valid {
		score.SpeedScore = speedScore.Float64
	}
	if collaborationScore.Valid {
		score.CollaborationScore = collaborationScore.Float64
	}
	if impactScore.Valid {
		score.ImpactScore = impactScore.Float64
	}

	return &score, nil
}

// buildTeamHierarchy recursively builds the team hierarchy tree
func (s *Service) buildTeamHierarchy(team *models.Team, visited map[string]bool) (*models.TeamHierarchyNode, error) {
	// Check for circular reference
	if visited[team.ID] {
		return nil, fmt.Errorf("circular reference detected in team hierarchy: %s", team.ID)
	}
	visited[team.ID] = true

	node := &models.TeamHierarchyNode{
		Team: *team,
	}

	// Get manager if present
	if team.ManagerID != nil {
		manager, err := s.getEngineerByID(*team.ManagerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get manager: %w", err)
		}
		node.Manager = manager
	}

	// Get child teams
	children, err := s.getChildTeams(team.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get child teams: %w", err)
	}

	// Recursively build child nodes
	for _, child := range children {
		childNode, err := s.buildTeamHierarchy(child, visited)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

// getChildTeams retrieves all teams that have the given team as parent
func (s *Service) getChildTeams(parentTeamID string) ([]*models.Team, error) {
	rows, err := s.db.Query(`
		SELECT id, name, parent_team_id, manager_id, created_at, updated_at
		FROM teams
		WHERE parent_team_id = ?
		ORDER BY name
		LIMIT 1000
	`, parentTeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to query child teams: %w", err)
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		var team models.Team
		var parentID, managerID sql.NullString

		err := rows.Scan(&team.ID, &team.Name, &parentID, &managerID, &team.CreatedAt, &team.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}

		if parentID.Valid {
			team.ParentTeamID = &parentID.String
		}
		if managerID.Valid {
			team.ManagerID = &managerID.String
		}

		teams = append(teams, &team)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating child teams: %w", err)
	}

	return teams, nil
}

// getEngineerByID retrieves an engineer by ID
func (s *Service) getEngineerByID(engineerID string) (*models.Engineer, error) {
	var engineer models.Engineer
	var email, manager sql.NullString
	var identifiersJSON string

	err := s.db.QueryRow(`
		SELECT id, name, email, manager, identifiers, active, created_at, updated_at
		FROM engineers
		WHERE id = ?
	`, engineerID).Scan(&engineer.ID, &engineer.Name, &email, &manager, &identifiersJSON, &engineer.Active, &engineer.CreatedAt, &engineer.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query engineer: %w", err)
	}

	if email.Valid {
		engineer.Email = email.String
	}
	if manager.Valid {
		engineer.Manager = manager.String
	}

	// Note: identifiers JSON parsing omitted for simplicity in this query
	// In production, you'd unmarshal the JSON into engineer.Identifiers

	return &engineer, nil
}
