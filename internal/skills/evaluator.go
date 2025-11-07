package skills

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Evaluator calculates skill scores and trajectories
type Evaluator struct {
	skillsStore *db.SkillsStore
}

// NewEvaluator creates a new skill evaluator
func NewEvaluator(skillsStore *db.SkillsStore) *Evaluator {
	return &Evaluator{
		skillsStore: skillsStore,
	}
}

// EvaluateEngineerSkills recalculates all skill scores for an engineer
func (e *Evaluator) EvaluateEngineerSkills(engineerID string) error {
	// Get all skills
	allSkills, _, err := e.skillsStore.ListSkills("", 0, 0)
	if err != nil {
		return fmt.Errorf("failed to list skills: %w", err)
	}

	now := time.Now().UTC()
	ninetyDaysAgo := now.AddDate(0, 0, -90)

	// Evaluate each skill
	for _, skill := range allSkills {
		// Get evidence for this skill
		evidences, _, err := e.skillsStore.ListSkillEvidence(engineerID, skill.ID, 0, 0)
		if err != nil {
			return fmt.Errorf("failed to get evidence for skill %s: %w", skill.ID, err)
		}

		// Filter to recent evidence (last 90 days)
		var recentEvidences []*models.SkillEvidence
		for _, ev := range evidences {
			if ev.DetectedAt.After(ninetyDaysAgo) {
				recentEvidences = append(recentEvidences, ev)
			}
		}

		if len(recentEvidences) == 0 {
			// No recent evidence, skip this skill
			continue
		}

		// Calculate new score
		newScore := e.CalculateSkillScore(engineerID, skill.ID, recentEvidences)

		// Get existing engineer skill record
		existing, err := e.skillsStore.GetEngineerSkill(engineerID, skill.ID)
		if err != nil {
			return fmt.Errorf("failed to get engineer skill: %w", err)
		}

		if existing == nil {
			// Create new record
			engineerSkill := &models.EngineerSkill{
				EngineerID:    engineerID,
				SkillID:       skill.ID,
				LevelScore:    newScore,
				Trajectory:    "stable",
				LastEvaluated: now,
				EvidenceCount: len(recentEvidences),
			}
			if err := e.skillsStore.CreateEngineerSkill(engineerSkill); err != nil {
				return fmt.Errorf("failed to create engineer skill: %w", err)
			}
		} else {
			// Update existing record
			previousScore := existing.LevelScore
			existing.PreviousLevelScore = &previousScore
			existing.LevelScore = newScore
			existing.Trajectory = e.DetermineSkillTrajectory(engineerID, skill.ID)
			existing.LastEvaluated = now
			existing.EvidenceCount = len(recentEvidences)

			if err := e.skillsStore.UpdateEngineerSkill(existing); err != nil {
				return fmt.Errorf("failed to update engineer skill: %w", err)
			}

			// Record progression if score changed significantly
			if math.Abs(float64(newScore-previousScore)) >= 5 {
				progression := &models.SkillProgression{
					EngineerID:    engineerID,
					SkillID:       skill.ID,
					PreviousScore: previousScore,
					NewScore:      newScore,
					ChangeReason:  "evidence_accumulated",
					EvaluatedAt:   now,
				}
				if err := e.skillsStore.CreateSkillProgression(progression); err != nil {
					return fmt.Errorf("failed to create progression: %w", err)
				}
			}
		}
	}

	return nil
}

// CalculateSkillScore aggregates evidence into a 0-100 score
func (e *Evaluator) CalculateSkillScore(engineerID, skillID string, evidences []*models.SkillEvidence) int {
	if len(evidences) == 0 {
		return 0
	}

	totalScore := 0.0
	totalWeight := 0.0

	now := time.Now().UTC()

	for _, ev := range evidences {
		// Calculate age-based weight (newer = more weight)
		ageDays := now.Sub(ev.DetectedAt).Hours() / 24
		weight := math.Exp(-ageDays / 90.0) // 90-day half-life

		totalScore += ev.Strength * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0
	}

	avgScore := totalScore / totalWeight

	// Scale to 0-100, with evidence count bonus
	baseScore := avgScore * 100

	// Evidence count bonus (more evidence = more confidence)
	evidenceBonus := math.Min(float64(len(evidences))*2, 20) // Max +20 points

	finalScore := int(baseScore + evidenceBonus)

	// Cap at 100
	if finalScore > 100 {
		finalScore = 100
	}

	return finalScore
}

// DetermineSkillTrajectory analyzes trend (improving/stable/declining)
func (e *Evaluator) DetermineSkillTrajectory(engineerID, skillID string) string {
	// Get recent progression history
	progressions, _, err := e.skillsStore.GetSkillProgression(engineerID, skillID, 5, 0)
	if err != nil || len(progressions) < 2 {
		return "stable"
	}

	// Calculate trend over last few evaluations
	totalChange := 0
	for _, prog := range progressions {
		totalChange += (prog.NewScore - prog.PreviousScore)
	}

	avgChange := float64(totalChange) / float64(len(progressions))

	if avgChange > 2 {
		return "improving"
	} else if avgChange < -2 {
		return "declining"
	}
	return "stable"
}

// CompareToBaseline compares engineer's skill to role baseline
func (e *Evaluator) CompareToBaseline(engineerID string, role string) (map[string]int, error) {
	// Role baselines (expected skill levels for each role)
	baselines := map[string]map[string]int{
		"junior": {
			"code_quality":  40,
			"testing":       30,
			"communication": 30,
		},
		"mid": {
			"code_quality":  60,
			"testing":       50,
			"system_design": 40,
			"communication": 50,
			"code_review":   50,
		},
		"senior": {
			"code_quality":      80,
			"testing":           70,
			"system_design":     70,
			"communication":     70,
			"code_review":       70,
			"mentoring":         60,
			"technical_writing": 60,
		},
		"staff": {
			"code_quality":       90,
			"testing":            80,
			"system_design":      85,
			"communication":      85,
			"code_review":        85,
			"mentoring":          80,
			"technical_writing":  75,
			"project_management": 70,
		},
	}

	baseline, ok := baselines[role]
	if !ok {
		return nil, fmt.Errorf("unknown role: %s", role)
	}

	// Get engineer's current skills
	engineerSkills, _, err := e.skillsStore.ListEngineerSkills(engineerID, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get engineer skills: %w", err)
	}

	skillMap := make(map[string]int)
	for _, es := range engineerSkills {
		skillMap[es.SkillID] = es.LevelScore
	}

	// Calculate gaps
	gaps := make(map[string]int)
	for skillID, expectedLevel := range baseline {
		currentLevel := skillMap[skillID]
		gap := expectedLevel - currentLevel
		if gap > 0 {
			gaps[skillID] = gap
		}
	}

	return gaps, nil
}

// IdentifySkillGaps finds skills below expected level for role
func (e *Evaluator) IdentifySkillGaps(engineerID string, role string) ([]*models.SkillGap, error) {
	gaps, err := e.CompareToBaseline(engineerID, role)
	if err != nil {
		return nil, err
	}

	var skillGaps []*models.SkillGap
	for skillID, gap := range gaps {
		skill, err := e.skillsStore.GetSkill(skillID)
		if err != nil || skill == nil {
			continue
		}

		// Get current level
		engineerSkill, err := e.skillsStore.GetEngineerSkill(engineerID, skillID)
		currentLevel := 0
		if err == nil && engineerSkill != nil {
			currentLevel = engineerSkill.LevelScore
		}

		priority := "low"
		if gap > 30 {
			priority = "high"
		} else if gap > 15 {
			priority = "medium"
		}

		skillGaps = append(skillGaps, &models.SkillGap{
			SkillID:       skillID,
			SkillName:     skill.Name,
			CurrentLevel:  currentLevel,
			ExpectedLevel: currentLevel + gap,
			Gap:           gap,
			Priority:      priority,
		})
	}

	// Sort by gap (highest first)
	sort.Slice(skillGaps, func(i, j int) bool {
		return skillGaps[i].Gap > skillGaps[j].Gap
	})

	return skillGaps, nil
}

// GenerateSkillRecommendations suggests focus areas
func (e *Evaluator) GenerateSkillRecommendations(engineerID string, role string) ([]string, error) {
	gaps, err := e.IdentifySkillGaps(engineerID, role)
	if err != nil {
		return nil, err
	}

	var recommendations []string

	// Focus on high-priority gaps first
	for _, gap := range gaps {
		if gap.Priority == "high" {
			recommendations = append(recommendations, fmt.Sprintf(
				"Focus on improving %s (current: %d, target: %d)",
				gap.SkillName, gap.CurrentLevel, gap.ExpectedLevel,
			))
		}
	}

	// Add general recommendations
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Continue developing your existing skills")
		recommendations = append(recommendations, "Consider mentoring junior engineers to build leadership skills")
	}

	return recommendations, nil
}

// GetPeerComparison compares engineer to peers (anonymized)
func (e *Evaluator) GetPeerComparison(engineerID string, skillID string, peerEngineerIDs []string) (*models.PeerSkillComparison, error) {
	// Get skill info
	skill, err := e.skillsStore.GetSkill(skillID)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill: %w", err)
	}
	if skill == nil {
		return nil, fmt.Errorf("skill not found: %s", skillID)
	}

	// Get engineer's score
	engineerSkill, err := e.skillsStore.GetEngineerSkill(engineerID, skillID)
	yourScore := 0
	if err == nil && engineerSkill != nil {
		yourScore = engineerSkill.LevelScore
	}

	// Get peer scores
	var peerScores []int
	for _, peerID := range peerEngineerIDs {
		if peerID == engineerID {
			continue // Skip self
		}
		peerSkill, err := e.skillsStore.GetEngineerSkill(peerID, skillID)
		if err == nil && peerSkill != nil {
			peerScores = append(peerScores, peerSkill.LevelScore)
		}
	}

	if len(peerScores) == 0 {
		return nil, fmt.Errorf("no peer data available")
	}

	// Calculate statistics
	sum := 0
	for _, score := range peerScores {
		sum += score
	}
	peerAverage := float64(sum) / float64(len(peerScores))

	// Calculate median
	sort.Ints(peerScores)
	peerMedian := float64(peerScores[len(peerScores)/2])

	// Calculate percentile
	rank := 0
	for _, score := range peerScores {
		if yourScore > score {
			rank++
		}
	}
	percentile := int(float64(rank) / float64(len(peerScores)) * 100)

	return &models.PeerSkillComparison{
		SkillID:     skillID,
		SkillName:   skill.Name,
		YourScore:   yourScore,
		PeerAverage: peerAverage,
		PeerMedian:  peerMedian,
		Percentile:  percentile,
		SampleSize:  len(peerScores),
	}, nil
}
