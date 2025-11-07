package models

import "time"

// Role represents an engineering role with performance expectations
type Role struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	TargetScore  int                `json:"target_score"`
	Expectations map[string]float64 `json:"expectations"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// ScoringWeights represents global weights for each scoring component
type ScoringWeights struct {
	ID                  string    `json:"id"`
	ThroughputWeight    float64   `json:"throughput_weight"`
	QualityWeight       float64   `json:"quality_weight"`
	SpeedWeight         float64   `json:"speed_weight"`
	CollaborationWeight float64   `json:"collaboration_weight"`
	ImpactWeight        float64   `json:"impact_weight"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// RawMetrics represents actual metric values for an engineer
type RawMetrics struct {
	ThroughputPRsPerWeek      float64 `json:"throughput_prs_per_week"`
	ThroughputStoryPoints     float64 `json:"throughput_story_points"`
	QualityBugRate            float64 `json:"quality_bug_rate"`
	QualityReworkRate         float64 `json:"quality_rework_rate"`
	SpeedCycleTimeDays        float64 `json:"speed_cycle_time_days"`
	SpeedTimeToFirstReview    float64 `json:"speed_time_to_first_review"`
	CollaborationReviewsGiven float64 `json:"collaboration_reviews_given"`
	CollaborationReviewDepth  float64 `json:"collaboration_review_depth"`
	ImpactServicesTouched     float64 `json:"impact_services_touched"`
}

// ComponentScores represents calculated scores for each component
type ComponentScores struct {
	Throughput    float64 `json:"throughput"`
	Quality       float64 `json:"quality"`
	Speed         float64 `json:"speed"`
	Collaboration float64 `json:"collaboration"`
	Impact        float64 `json:"impact"`
}

// PerformanceScore represents a calculated performance score
type PerformanceScore struct {
	ID                 string    `json:"id"`
	EngineerID         string    `json:"engineer_id"`
	WeekStart          time.Time `json:"week_start"`
	TotalScore         float64   `json:"total_score"`
	ThroughputScore    float64   `json:"throughput_score"`
	QualityScore       float64   `json:"quality_score"`
	SpeedScore         float64   `json:"speed_score"`
	CollaborationScore float64   `json:"collaboration_score"`
	ImpactScore        float64   `json:"impact_score"`
	RawMetrics         string    `json:"raw_metrics"`
	BurnoutRisk        string    `json:"burnout_risk,omitempty"` // JSON-encoded BurnoutRisk
	CreatedAt          time.Time `json:"created_at"`
}

// PromotionSignal represents a detected promotion signal for an engineer
type PromotionSignal struct {
	ID            string     `json:"id"`
	EngineerID    string     `json:"engineer_id"`
	SignalType    string     `json:"signal_type"`
	DetectedAt    time.Time  `json:"detected_at"`
	WeeksDuration int        `json:"weeks_duration"`
	AvgScore      float64    `json:"avg_score"`
	Dismissed     bool       `json:"dismissed"`
	DismissedAt   *time.Time `json:"dismissed_at,omitempty"`
	Notes         string     `json:"notes,omitempty"`
}

// PromotionCandidate represents an engineer with an active promotion signal
type PromotionCandidate struct {
	Engineer    Engineer        `json:"engineer"`
	CurrentRole *Role           `json:"current_role,omitempty"`
	NextRole    *Role           `json:"next_role,omitempty"`
	Signal      PromotionSignal `json:"signal"`
}

// RedFlag represents a burnout warning indicator
type RedFlag struct {
	Type     string `json:"type"`     // "late_night_work", "weekend_work", "quality_decline", "review_decline"
	Severity string `json:"severity"` // "low", "medium", "high"
	Evidence string `json:"evidence"` // Human-readable description with data
	Count    int    `json:"count"`    // Number of occurrences
}

// BurnoutRisk represents detected burnout risk indicators for an engineer
type BurnoutRisk struct {
	EngineerID     string    `json:"engineer_id"`
	WeekStart      time.Time `json:"week_start"`
	Score          float64   `json:"score"`
	RedFlags       []RedFlag `json:"red_flags"`
	RiskLevel      string    `json:"risk_level"` // "none", "low", "medium", "high"
	Recommendation string    `json:"recommendation"`
	DetectedAt     time.Time `json:"detected_at"`
}
