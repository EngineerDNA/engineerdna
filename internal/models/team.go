package models

import "time"

// Team represents an engineering team
type Team struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	ParentTeamID *string   `json:"parent_team_id,omitempty"`
	ManagerID    *string   `json:"manager_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TeamMembership represents a member's association with a team
type TeamMembership struct {
	ID       string     `json:"id"`
	TeamID   string     `json:"team_id"`
	MemberID string     `json:"member_id"`
	Role     string     `json:"role,omitempty"`
	JoinedAt time.Time  `json:"joined_at"`
	LeftAt   *time.Time `json:"left_at,omitempty"`
}

// TeamPerformanceScore represents aggregated performance scores for a team
// NOTE: This is a computed view, not persisted. Data stored in metric_values table.
type TeamPerformanceScore struct {
	ID                 string    `json:"id"`
	TeamID             string    `json:"team_id"`
	WeekStart          time.Time `json:"week_start"`
	TotalScore         float64   `json:"total_score"`
	MemberCount        int       `json:"member_count"`
	ThroughputScore    float64   `json:"throughput_score"`
	QualityScore       float64   `json:"quality_score"`
	SpeedScore         float64   `json:"speed_score"`
	CollaborationScore float64   `json:"collaboration_score"`
	ImpactScore        float64   `json:"impact_score"`
	CreatedAt          time.Time `json:"created_at"`
}

// TeamWithMembers represents a team with its active members
type TeamWithMembers struct {
	Team    Team       `json:"team"`
	Members []Engineer `json:"members"`
}

// TeamScorecard represents a team's performance scorecard
type TeamScorecard struct {
	Team          Team      `json:"team"`
	Score         float64   `json:"score"`
	PrevWeekScore float64   `json:"prev_week_score"`
	ScoreChange   float64   `json:"score_change"`
	MemberCount   int       `json:"member_count"`
	WeekStart     time.Time `json:"week_start"`
}

// TeamHierarchyNode represents a node in the team hierarchy tree
type TeamHierarchyNode struct {
	Team     Team                 `json:"team"`
	Manager  *Engineer            `json:"manager,omitempty"`
	Children []*TeamHierarchyNode `json:"children,omitempty"`
}

// OrgScorecard represents organization-wide aggregated metrics
type OrgScorecard struct {
	TotalEngineers        int     `json:"total_engineers"`
	TotalScore            float64 `json:"total_score"`
	ScoreChange           float64 `json:"score_change"`
	VelocityPRsPerWeek    float64 `json:"velocity_prs_per_week"`
	VelocityChange        float64 `json:"velocity_change"`
	CycleTimeDays         float64 `json:"cycle_time_days"`
	CycleTimeChange       float64 `json:"cycle_time_change"`
	TeamsNeedingAttention int     `json:"teams_needing_attention"`
}
