package skills

// Skill strength multipliers and thresholds for skill detection.
// These constants define how strongly various activities demonstrate skill proficiency.
const (
	// DefaultSkillStrength is the standard strength value for skill evidence
	// when detected from various activities like code reviews, testing, or design work.
	// Value of 0.6 indicates moderate confidence in skill demonstration.
	DefaultSkillStrength = 0.6

	// TestCoverageMultiplier is the multiplier applied to test-to-code ratio
	// to calculate testing skill strength. Higher test coverage indicates
	// stronger testing skills.
	TestCoverageMultiplier = 0.6

	// CodeReviewMultiplier is the multiplier for code review activity strength.
	// Applied to review participation to calculate code review skill level.
	CodeReviewMultiplier = 0.6

	// MaxSkillStrength is the maximum possible strength value for skill evidence.
	// Represents 100% confidence in skill demonstration.
	MaxSkillStrength = 1.0
)
