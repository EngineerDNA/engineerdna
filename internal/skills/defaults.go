package skills

import (
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// DefaultSkills defines the standard skill taxonomy
var DefaultSkills = []models.Skill{
	// Technical Skills - System Design
	{
		ID:          "system_design",
		Name:        "System Design",
		Category:    "technical",
		Subcategory: "architecture",
		Description: "Ability to design scalable, maintainable systems across multiple services",
		MeasurementCriteria: `{
			"services_touched": "Number of services/repos worked in",
			"design_docs": "Architecture documents authored",
			"pr_complexity": "Complexity of changes (lines, services, dependencies)"
		}`,
	},
	{
		ID:          "code_quality",
		Name:        "Code Quality",
		Category:    "technical",
		Subcategory: "fundamentals",
		Description: "Writing clean, maintainable, bug-free code with high standards",
		MeasurementCriteria: `{
			"bug_rate": "Bugs per PR (lower is better)",
			"test_coverage": "Test coverage percentage",
			"review_feedback": "Quality themes in code reviews"
		}`,
	},
	{
		ID:          "testing",
		Name:        "Testing",
		Category:    "technical",
		Subcategory: "fundamentals",
		Description: "Writing comprehensive tests and maintaining test quality",
		MeasurementCriteria: `{
			"test_to_code_ratio": "Lines of test code vs production code",
			"coverage": "Test coverage percentage",
			"test_quality": "Tests written with PRs"
		}`,
	},
	{
		ID:          "devops",
		Name:        "DevOps",
		Category:    "technical",
		Subcategory: "infrastructure",
		Description: "Infrastructure, deployment, CI/CD, and operational excellence",
		MeasurementCriteria: `{
			"infra_changes": "Infrastructure code contributions",
			"ci_cd": "CI/CD pipeline improvements",
			"monitoring": "Monitoring and alerting setup"
		}`,
	},
	{
		ID:          "security",
		Name:        "Security",
		Category:    "technical",
		Subcategory: "infrastructure",
		Description: "Security awareness, vulnerability prevention, and secure coding",
		MeasurementCriteria: `{
			"security_reviews": "Security-related PRs",
			"vulnerabilities": "Security issues found/fixed",
			"secure_patterns": "Use of secure coding patterns"
		}`,
	},
	{
		ID:          "performance",
		Name:        "Performance Optimization",
		Category:    "technical",
		Subcategory: "fundamentals",
		Description: "Optimizing code performance, database queries, and system efficiency",
		MeasurementCriteria: `{
			"perf_improvements": "Performance-related PRs",
			"profiling": "Use of profiling tools",
			"benchmarks": "Performance benchmarks added"
		}`,
	},

	// Leadership Skills
	{
		ID:          "mentoring",
		Name:        "Mentoring",
		Category:    "leadership",
		Subcategory: "people",
		Description: "Ability to teach, guide, and grow other engineers",
		MeasurementCriteria: `{
			"pairing_sessions": "Pair programming frequency",
			"review_depth": "Depth and quality of code review feedback",
			"mentee_growth": "Growth of mentored engineers"
		}`,
	},
	{
		ID:          "technical_writing",
		Name:        "Technical Writing",
		Category:    "leadership",
		Subcategory: "communication",
		Description: "Writing clear documentation, design docs, and technical content",
		MeasurementCriteria: `{
			"docs_authored": "Documentation files authored",
			"design_docs": "Design documents written",
			"wiki_contributions": "Wiki/knowledge base contributions"
		}`,
	},
	{
		ID:          "code_review",
		Name:        "Code Review",
		Category:    "leadership",
		Subcategory: "technical",
		Description: "Providing thorough, constructive code reviews",
		MeasurementCriteria: `{
			"review_count": "Number of reviews completed",
			"review_depth": "Comments per review",
			"bugs_caught": "Issues caught before merge"
		}`,
	},
	{
		ID:          "project_management",
		Name:        "Project Management",
		Category:    "leadership",
		Subcategory: "delivery",
		Description: "Planning, coordinating, and delivering complex projects",
		MeasurementCriteria: `{
			"projects_led": "Projects led or coordinated",
			"on_time_delivery": "Projects delivered on time",
			"stakeholder_management": "Cross-team coordination"
		}`,
	},
	{
		ID:          "communication",
		Name:        "Communication",
		Category:    "leadership",
		Subcategory: "communication",
		Description: "Clear, effective communication with team and stakeholders",
		MeasurementCriteria: `{
			"pr_descriptions": "Quality of PR descriptions",
			"meeting_participation": "Active participation in discussions",
			"documentation": "Clarity of written communication"
		}`,
	},
	{
		ID:          "problem_solving",
		Name:        "Problem Solving",
		Category:    "technical",
		Subcategory: "fundamentals",
		Description: "Analytical thinking and creative problem-solving ability",
		MeasurementCriteria: `{
			"complex_issues": "Complex bugs/issues solved",
			"creative_solutions": "Novel approaches to problems",
			"root_cause_analysis": "Depth of problem investigation"
		}`,
	},
}

// CreateDefaultSkills populates the database with default skill taxonomy
func CreateDefaultSkills(store *db.SkillsStore) error {
	for _, skill := range DefaultSkills {
		// Check if skill already exists
		existing, err := store.GetSkill(skill.ID)
		if err != nil {
			return fmt.Errorf("failed to check existing skill %s: %w", skill.ID, err)
		}

		if existing != nil {
			// Skill already exists, skip
			continue
		}

		// Create new skill
		skillCopy := skill
		skillCopy.CreatedAt = time.Now().UTC()
		if err := store.CreateSkill(&skillCopy); err != nil {
			return fmt.Errorf("failed to create skill %s: %w", skill.ID, err)
		}
	}

	return nil
}

// GetSkillIDByName returns skill ID for a skill name (case-insensitive)
func GetSkillIDByName(name string) string {
	for _, skill := range DefaultSkills {
		if skill.Name == name {
			return skill.ID
		}
	}
	return ""
}

// GetSkillsByCategory returns all skills in a category
func GetSkillsByCategory(category string) []models.Skill {
	var skills []models.Skill
	for _, skill := range DefaultSkills {
		if skill.Category == category {
			skills = append(skills, skill)
		}
	}
	return skills
}
