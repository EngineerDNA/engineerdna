package skills

import (
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Service handles skill business logic
type Service struct {
	skillsStore *db.SkillsStore
	eventStore  *db.EventStore
	detector    *Detector
	evaluator   *Evaluator
}

// NewService creates a new skill service
func NewService(skillsStore *db.SkillsStore, eventStore *db.EventStore) *Service {
	detector := NewDetector(skillsStore, eventStore)
	evaluator := NewEvaluator(skillsStore)

	return &Service{
		skillsStore: skillsStore,
		eventStore:  eventStore,
		detector:    detector,
		evaluator:   evaluator,
	}
}

// GetEngineerSkillProfile returns complete skill profile for an engineer
func (s *Service) GetEngineerSkillProfile(engineerID, engineerName string) (*models.SkillProfile, error) {
	// Get all skills for engineer
	skills, _, err := s.skillsStore.ListEngineerSkills(engineerID, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get engineer skills: %w", err)
	}

	// Get summary
	summary, err := s.GetSkillSummary(engineerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill summary: %w", err)
	}

	// Get recent evidence (last 30 days)
	thirtyDaysAgo := time.Now().UTC().AddDate(0, 0, -30)
	recentEvidence, err := s.skillsStore.GetAllSkillEvidence(engineerID, thirtyDaysAgo, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent evidence: %w", err)
	}

	// Get active goals
	goals, err := s.skillsStore.GetSkillGoals(engineerID, "active")
	if err != nil {
		return nil, fmt.Errorf("failed to get skill goals: %w", err)
	}

	return &models.SkillProfile{
		EngineerID:     engineerID,
		EngineerName:   engineerName,
		Skills:         skills,
		Summary:        summary,
		RecentEvidence: recentEvidence,
		Goals:          goals,
	}, nil
}

// GetSkillSummary generates summary statistics
func (s *Service) GetSkillSummary(engineerID string) (*models.SkillSummary, error) {
	skills, _, err := s.skillsStore.ListEngineerSkills(engineerID, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get engineer skills: %w", err)
	}

	summary := &models.SkillSummary{
		EngineerID:       engineerID,
		TotalSkills:      len(skills),
		SkillsByCategory: make(map[string]int),
	}

	totalScore := 0
	var topSkills []*models.SkillWithScore

	for _, es := range skills {
		// Get skill details
		skill, err := s.skillsStore.GetSkill(es.SkillID)
		if err != nil || skill == nil {
			continue
		}

		// Count by category
		summary.SkillsByCategory[skill.Category]++

		// Count trajectories
		switch es.Trajectory {
		case "improving":
			summary.ImprovingSkills++
		case "stable":
			summary.StableSkills++
		case "declining":
			summary.DecliningSkills++
		}

		// Sum scores
		totalScore += es.LevelScore

		// Track top skills
		topSkills = append(topSkills, &models.SkillWithScore{
			Skill:      skill,
			LevelScore: es.LevelScore,
			Trajectory: es.Trajectory,
		})
	}

	// Calculate average
	if len(skills) > 0 {
		summary.AverageScore = float64(totalScore) / float64(len(skills))
	}

	// Sort top skills by score (descending)
	for i := 0; i < len(topSkills); i++ {
		for j := i + 1; j < len(topSkills); j++ {
			if topSkills[j].LevelScore > topSkills[i].LevelScore {
				topSkills[i], topSkills[j] = topSkills[j], topSkills[i]
			}
		}
	}

	// Keep top 5
	if len(topSkills) > 5 {
		summary.TopSkills = topSkills[:5]
	} else {
		summary.TopSkills = topSkills
	}

	return summary, nil
}

// GetSkillDevelopmentPlan provides recommended path to improve a skill
func (s *Service) GetSkillDevelopmentPlan(engineerID, skillID string) (*models.SkillDevelopmentPlan, error) {
	// Get skill info
	skill, err := s.skillsStore.GetSkill(skillID)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill: %w", err)
	}
	if skill == nil {
		return nil, fmt.Errorf("skill not found: %s", skillID)
	}

	// Get current level
	engineerSkill, err := s.skillsStore.GetEngineerSkill(engineerID, skillID)
	currentLevel := 0
	if err == nil && engineerSkill != nil {
		currentLevel = engineerSkill.LevelScore
	}

	// Target level (aim for next tier)
	targetLevel := ((currentLevel / 20) + 1) * 20 // Round up to next 20
	if targetLevel > 100 {
		targetLevel = 100
	}

	// Estimate time (rough estimate: 1 week per 5 points)
	gap := targetLevel - currentLevel
	estimatedWeeks := gap / 5
	if estimatedWeeks < 1 {
		estimatedWeeks = 1
	}

	// Generate recommendations based on skill type
	recommendations := s.generateSkillRecommendations(skill)

	// Find related skills
	relatedSkills := s.findRelatedSkills(skill)

	return &models.SkillDevelopmentPlan{
		SkillID:         skillID,
		SkillName:       skill.Name,
		CurrentLevel:    currentLevel,
		TargetLevel:     targetLevel,
		EstimatedWeeks:  estimatedWeeks,
		Recommendations: recommendations,
		RelatedSkills:   relatedSkills,
	}, nil
}

// TrackSkillProgress returns historical progression
func (s *Service) TrackSkillProgress(engineerID, skillID string) ([]*models.SkillProgression, error) {
	progressions, _, err := s.skillsStore.GetSkillProgression(engineerID, skillID, 0, 0)
	return progressions, err
}

// ComparePeerSkills provides anonymized peer comparison
func (s *Service) ComparePeerSkills(engineerID string, peerEngineerIDs []string) ([]*models.PeerSkillComparison, error) {
	// Get all skills for engineer
	skills, _, err := s.skillsStore.ListEngineerSkills(engineerID, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get engineer skills: %w", err)
	}

	var comparisons []*models.PeerSkillComparison
	for _, es := range skills {
		comparison, err := s.evaluator.GetPeerComparison(engineerID, es.SkillID, peerEngineerIDs)
		if err != nil {
			continue // Skip if no peer data
		}
		comparisons = append(comparisons, comparison)
	}

	return comparisons, nil
}

// GetPromotionReadiness identifies skill gaps for promotion to next level
func (s *Service) GetPromotionReadiness(engineerID, currentRole, targetRole string) (*models.SkillGapAnalysis, error) {
	// Get skill gaps for target role
	gaps, err := s.evaluator.IdentifySkillGaps(engineerID, targetRole)
	if err != nil {
		return nil, fmt.Errorf("failed to identify skill gaps: %w", err)
	}

	// Generate recommendations
	recommendations, err := s.evaluator.GenerateSkillRecommendations(engineerID, targetRole)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	return &models.SkillGapAnalysis{
		EngineerID:      engineerID,
		Role:            targetRole,
		SkillGaps:       gaps,
		Recommendations: recommendations,
	}, nil
}

// ProcessEventForSkills analyzes an event and detects skill evidence
func (s *Service) ProcessEventForSkills(event *models.Event, engineerID string) error {
	_, err := s.detector.AnalyzeEventForSkills(event, engineerID)
	return err
}

// EvaluateAllSkills runs full skill evaluation for an engineer
func (s *Service) EvaluateAllSkills(engineerID string) error {
	return s.evaluator.EvaluateEngineerSkills(engineerID)
}

// Helper functions

func (s *Service) generateSkillRecommendations(skill *models.Skill) []string {
	// Skill-specific recommendations
	recommendations := map[string][]string{
		"system_design": {
			"Work on features that span multiple services",
			"Write design documents for new features",
			"Participate in architecture reviews",
			"Study distributed systems patterns",
		},
		"code_quality": {
			"Increase test coverage for your PRs",
			"Focus on reducing bug rates",
			"Apply linting and static analysis tools",
			"Refactor complex code for clarity",
		},
		"testing": {
			"Write tests alongside production code",
			"Improve test coverage to >80%",
			"Learn advanced testing patterns (mocking, fixtures)",
			"Add integration and e2e tests",
		},
		"mentoring": {
			"Pair program with junior engineers",
			"Provide detailed, constructive code reviews",
			"Lead knowledge sharing sessions",
			"Document best practices for the team",
		},
		"technical_writing": {
			"Write design documents for features",
			"Improve inline code documentation",
			"Contribute to team wiki/knowledge base",
			"Write technical blog posts",
		},
		"code_review": {
			"Review more PRs from team members",
			"Provide detailed, constructive feedback",
			"Catch bugs and suggest improvements",
			"Help enforce coding standards",
		},
	}

	if recs, ok := recommendations[skill.ID]; ok {
		return recs
	}

	// Generic recommendations
	return []string{
		fmt.Sprintf("Practice %s in your daily work", skill.Name),
		"Seek feedback from peers and mentors",
		"Study best practices and patterns",
	}
}

func (s *Service) findRelatedSkills(skill *models.Skill) []string {
	// Skills that commonly develop together
	relatedMap := map[string][]string{
		"system_design":      {"code_quality", "testing", "technical_writing"},
		"code_quality":       {"testing", "code_review"},
		"testing":            {"code_quality", "devops"},
		"mentoring":          {"code_review", "technical_writing", "communication"},
		"technical_writing":  {"communication", "system_design"},
		"code_review":        {"mentoring", "code_quality"},
		"project_management": {"communication", "system_design"},
	}

	if related, ok := relatedMap[skill.ID]; ok {
		return related
	}
	return []string{}
}
