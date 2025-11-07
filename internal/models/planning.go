package models

import "time"

// Sprint represents a sprint for planning and tracking
type Sprint struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	CommittedPoints int       `json:"committed_points"`
	CompletedPoints int       `json:"completed_points"`
	TeamCapacity    int       `json:"team_capacity"`
	Status          string    `json:"status"`
	TeamID          string    `json:"team_id"`
	CreatedAt       time.Time `json:"created_at"`
}

// Epic represents a large initiative or feature
type Epic struct {
	ID                     string     `json:"id"`
	JiraID                 *string    `json:"jira_id,omitempty"`
	Title                  string     `json:"title"`
	Description            string     `json:"description,omitempty"`
	OriginalEstimatePoints int        `json:"original_estimate_points"`
	CurrentScopePoints     int        `json:"current_scope_points"`
	CreatedAt              time.Time  `json:"created_at"`
	StartedAt              *time.Time `json:"started_at,omitempty"`
	CompletedAt            *time.Time `json:"completed_at,omitempty"`
	Status                 string     `json:"status"`
	TeamID                 string     `json:"team_id"`
}

// Story represents a user story or task
type Story struct {
	ID                   string     `json:"id"`
	JiraID               *string    `json:"jira_id,omitempty"`
	EpicID               *string    `json:"epic_id,omitempty"`
	SprintID             *string    `json:"sprint_id,omitempty"`
	Title                string     `json:"title"`
	StoryPoints          int        `json:"story_points"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
	StartedAt            *time.Time `json:"started_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	ActualDaysToComplete *float64   `json:"actual_days_to_complete,omitempty"`
	AssigneeID           *string    `json:"assignee_id,omitempty"`
}

// CapacityHistory tracks team capacity and delivery over time
type CapacityHistory struct {
	ID                 string    `json:"id"`
	WeekStart          time.Time `json:"week_start"`
	TeamID             string    `json:"team_id"`
	TeamSize           int       `json:"team_size"`
	AvailableEngineers float64   `json:"available_engineers"`
	CompletedPoints    int       `json:"completed_points"`
	MeetingHours       float64   `json:"meeting_hours"`
	CreatedAt          time.Time `json:"created_at"`
}

// PlanningAlert represents a detected planning issue
type PlanningAlert struct {
	ID               string    `json:"id"`
	AlertType        string    `json:"alert_type"`
	EntityType       string    `json:"entity_type"`
	EntityID         string    `json:"entity_id"`
	Severity         string    `json:"severity"`
	Message          string    `json:"message"`
	AIRecommendation string    `json:"ai_recommendation,omitempty"`
	Dismissed        bool      `json:"dismissed"`
	CreatedAt        time.Time `json:"created_at"`
}

// SprintHealth represents the calculated health status of a sprint
type SprintHealth struct {
	Sprint          *Sprint `json:"sprint"`
	RiskLevel       string  `json:"risk_level"` // "on_track", "at_risk", "high_risk"
	PercentComplete float64 `json:"percent_complete"`
	DaysRemaining   int     `json:"days_remaining"`
	DaysElapsed     int     `json:"days_elapsed"`
	TotalDays       int     `json:"total_days"`
	HistoricalAvg   float64 `json:"historical_avg"`
	Overcommitment  float64 `json:"overcommitment"` // percentage over capacity
	Recommendation  string  `json:"recommendation"`
}

// VelocityTrend represents historical velocity analysis
type VelocityTrend struct {
	TeamID          string  `json:"team_id"`
	AveragePoints   float64 `json:"average_points"`
	Trend           string  `json:"trend"` // "increasing", "decreasing", "stable"
	Last12Sprints   []int   `json:"last_12_sprints"`
	StdDeviation    float64 `json:"std_deviation"`
	ConfidenceLevel string  `json:"confidence_level"` // "high", "medium", "low"
}

// TimelineEstimate represents an estimated timeline for a feature
type TimelineEstimate struct {
	FeatureName     string  `json:"feature_name"`
	EstimatedPoints int     `json:"estimated_points"`
	BestCase        int     `json:"best_case"`  // weeks
	Likely          int     `json:"likely"`     // weeks
	WorstCase       int     `json:"worst_case"` // weeks
	ConfidenceLevel string  `json:"confidence_level"`
	Recommendation  string  `json:"recommendation"`
	TeamID          string  `json:"team_id"`
	TeamVelocity    float64 `json:"team_velocity"`
}
