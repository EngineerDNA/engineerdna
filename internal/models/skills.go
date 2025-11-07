package models

import "time"

// Skill represents a skill in the skill taxonomy
type Skill struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Category            string    `json:"category"`              // 'technical', 'leadership', 'communication'
	Subcategory         string    `json:"subcategory,omitempty"` // 'system_design', 'code_quality', etc.
	Description         string    `json:"description,omitempty"`
	MeasurementCriteria string    `json:"measurement_criteria,omitempty"` // JSON
	CreatedAt           time.Time `json:"created_at"`
}

// EngineerSkill represents a skill level for a specific engineer
type EngineerSkill struct {
	ID                 string    `json:"id"`
	EngineerID         string    `json:"engineer_id"`
	SkillID            string    `json:"skill_id"`
	LevelScore         int       `json:"level_score"` // 0-100
	PreviousLevelScore *int      `json:"previous_level_score,omitempty"`
	Trajectory         string    `json:"trajectory"` // 'improving', 'stable', 'declining'
	LastEvaluated      time.Time `json:"last_evaluated"`
	EvidenceCount      int       `json:"evidence_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// SkillEvidence represents proof of skill usage/growth
type SkillEvidence struct {
	ID             string    `json:"id"`
	EngineerID     string    `json:"engineer_id"`
	SkillID        string    `json:"skill_id"`
	EvidenceType   string    `json:"evidence_type"`             // 'pr_complexity', 'review_depth', etc.
	EvidenceSource string    `json:"evidence_source,omitempty"` // event_id, pr_id, etc.
	Strength       float64   `json:"strength"`                  // 0.0-1.0
	Context        string    `json:"context,omitempty"`         // JSON
	DetectedAt     time.Time `json:"detected_at"`
}

// SkillProgression tracks historical changes in skill levels
type SkillProgression struct {
	ID            string    `json:"id"`
	EngineerID    string    `json:"engineer_id"`
	SkillID       string    `json:"skill_id"`
	PreviousScore int       `json:"previous_score"`
	NewScore      int       `json:"new_score"`
	ChangeReason  string    `json:"change_reason,omitempty"` // 'evidence_accumulated', etc.
	EvaluatedAt   time.Time `json:"evaluated_at"`
}

// SkillGoal represents a skill development goal
type SkillGoal struct {
	ID           string     `json:"id"`
	EngineerID   string     `json:"engineer_id"`
	SkillID      string     `json:"skill_id"`
	GoalID       *string    `json:"goal_id,omitempty"` // optional link to formal goal
	CurrentLevel int        `json:"current_level"`
	TargetLevel  int        `json:"target_level"`
	TargetDate   *time.Time `json:"target_date,omitempty"`
	Milestones   string     `json:"milestones,omitempty"` // JSON array
	Status       string     `json:"status"`               // 'active', 'completed', 'abandoned'
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// SkillSummary provides a computed summary of an engineer's skills
type SkillSummary struct {
	EngineerID       string            `json:"engineer_id"`
	TotalSkills      int               `json:"total_skills"`
	SkillsByCategory map[string]int    `json:"skills_by_category"`
	AverageScore     float64           `json:"average_score"`
	ImprovingSkills  int               `json:"improving_skills"`
	StableSkills     int               `json:"stable_skills"`
	DecliningSkills  int               `json:"declining_skills"`
	TopSkills        []*SkillWithScore `json:"top_skills,omitempty"`
	SkillGaps        []*SkillWithScore `json:"skill_gaps,omitempty"`
}

// SkillWithScore combines skill info with score
type SkillWithScore struct {
	Skill      *Skill `json:"skill"`
	LevelScore int    `json:"level_score"`
	Trajectory string `json:"trajectory"`
}

// SkillProfile provides complete skill information for an engineer
type SkillProfile struct {
	EngineerID     string           `json:"engineer_id"`
	EngineerName   string           `json:"engineer_name"`
	Skills         []*EngineerSkill `json:"skills"`
	Summary        *SkillSummary    `json:"summary"`
	RecentEvidence []*SkillEvidence `json:"recent_evidence,omitempty"`
	Goals          []*SkillGoal     `json:"goals,omitempty"`
}

// SkillDevelopmentPlan provides a recommended path to improve a skill
type SkillDevelopmentPlan struct {
	SkillID         string   `json:"skill_id"`
	SkillName       string   `json:"skill_name"`
	CurrentLevel    int      `json:"current_level"`
	TargetLevel     int      `json:"target_level"`
	EstimatedWeeks  int      `json:"estimated_weeks"`
	Recommendations []string `json:"recommendations"`
	RelatedSkills   []string `json:"related_skills,omitempty"`
	ResourceLinks   []string `json:"resource_links,omitempty"`
}

// SkillGapAnalysis identifies skills below expected level
type SkillGapAnalysis struct {
	EngineerID      string      `json:"engineer_id"`
	EngineerName    string      `json:"engineer_name"`
	Role            string      `json:"role,omitempty"`
	SkillGaps       []*SkillGap `json:"skill_gaps"`
	Recommendations []string    `json:"recommendations"`
}

// SkillGap represents a skill below expected level
type SkillGap struct {
	SkillID       string `json:"skill_id"`
	SkillName     string `json:"skill_name"`
	CurrentLevel  int    `json:"current_level"`
	ExpectedLevel int    `json:"expected_level"`
	Gap           int    `json:"gap"`
	Priority      string `json:"priority"` // 'high', 'medium', 'low'
}

// PeerSkillComparison provides anonymized peer comparison
type PeerSkillComparison struct {
	SkillID     string  `json:"skill_id"`
	SkillName   string  `json:"skill_name"`
	YourScore   int     `json:"your_score"`
	PeerAverage float64 `json:"peer_average"`
	PeerMedian  float64 `json:"peer_median"`
	Percentile  int     `json:"percentile"` // 0-100
	SampleSize  int     `json:"sample_size"`
}
