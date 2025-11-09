package models

import "time"

// Goal represents a goal (OKR) for an individual, team, or org
type Goal struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	Description        string    `json:"description,omitempty"`
	GoalType           string    `json:"goal_type"`          // 'individual', 'team', 'org'
	OwnerType          string    `json:"owner_type"`         // 'engineer', 'team', 'org'
	OwnerID            *string   `json:"owner_id,omitempty"` // engineer_id, team_id, or null
	TimePeriod         string    `json:"time_period"`        // 'Q1 2025', 'Q4 2024', etc.
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	TrackingMethod     string    `json:"tracking_method"`            // 'manual', 'automatic', 'hybrid'
	SuccessCriteria    string    `json:"success_criteria,omitempty"` // JSON array
	Status             string    `json:"status"`                     // 'active', 'completed', 'at_risk', 'off_track', 'archived'
	ProgressPercentage int       `json:"progress_percentage"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// GoalMilestone represents a milestone within a goal
type GoalMilestone struct {
	ID           string     `json:"id"`
	GoalID       string     `json:"goal_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description,omitempty"`
	TargetValue  float64    `json:"target_value,omitempty"`
	CurrentValue float64    `json:"current_value"`
	Unit         string     `json:"unit,omitempty"` // 'features', 'PRs', 'bugs', etc.
	Completed    bool       `json:"completed"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// GoalProgressLog tracks changes to goal progress
type GoalProgressLog struct {
	ID            string    `json:"id"`
	GoalID        string    `json:"goal_id"`
	MilestoneID   *string   `json:"milestone_id,omitempty"`
	PreviousValue *float64  `json:"previous_value,omitempty"`
	NewValue      *float64  `json:"new_value,omitempty"`
	ChangeType    string    `json:"change_type"`        // 'manual_update', 'auto_detected', 'milestone_complete'
	Evidence      string    `json:"evidence,omitempty"` // JSON
	LoggedAt      time.Time `json:"logged_at"`
	LoggedBy      *string   `json:"logged_by,omitempty"` // engineer_id or 'system'
}

// GoalDependency represents dependencies between goals
type GoalDependency struct {
	ID              string    `json:"id"`
	GoalID          string    `json:"goal_id"`
	DependsOnGoalID string    `json:"depends_on_goal_id"`
	DependencyType  string    `json:"dependency_type"` // 'blocks', 'relates_to'
	Status          string    `json:"status"`          // 'active', 'resolved'
	CreatedAt       time.Time `json:"created_at"`
}

// GoalMetric links a goal to a quantitative metric
type GoalMetric struct {
	ID            string     `json:"id"`
	GoalID        string     `json:"goal_id"`
	MetricName    string     `json:"metric_name"` // 'cycle_time', 'pr_count', etc.
	TargetValue   float64    `json:"target_value"`
	CurrentValue  *float64   `json:"current_value,omitempty"`
	Operator      string     `json:"operator"` // 'decrease_to', 'increase_to', 'maintain'
	LastEvaluated *time.Time `json:"last_evaluated,omitempty"`
}

// GoalSummary provides a computed summary of goals
type GoalSummary struct {
	OwnerType       string  `json:"owner_type"`
	OwnerID         *string `json:"owner_id,omitempty"`
	TotalGoals      int     `json:"total_goals"`
	ActiveGoals     int     `json:"active_goals"`
	CompletedGoals  int     `json:"completed_goals"`
	AtRiskGoals     int     `json:"at_risk_goals"`
	OffTrackGoals   int     `json:"off_track_goals"`
	AverageProgress float64 `json:"average_progress"`
}

// GoalWithDetails includes milestones, metrics, and dependencies
type GoalWithDetails struct {
	Goal         *Goal              `json:"goal"`
	Milestones   []*GoalMilestone   `json:"milestones,omitempty"`
	Metrics      []*GoalMetric      `json:"metrics,omitempty"`
	Dependencies []*GoalDependency  `json:"dependencies,omitempty"`
	ProgressLogs []*GoalProgressLog `json:"progress_logs,omitempty"`
}
