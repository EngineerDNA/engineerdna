package seed

import (
	"fmt"
	"log"

	"github.com/engineerdna/engineerdna/internal/models"
)

// SeedSkills creates skill development data
func SeedSkills(data *SeedData) (skillIDs []string, engineerSkillCount, evidenceCount int) {
	fmt.Println("\n12. Creating skill development data...")

	skills := []struct {
		name        string
		category    string
		subcategory string
	}{
		{"Go Programming", "technical", "backend"},
		{"React Development", "technical", "frontend"},
		{"PostgreSQL", "technical", "database"},
		{"Docker", "technical", "devops"},
		{"Kubernetes", "technical", "devops"},
		{"System Design", "technical", "architecture"},
		{"Code Review", "technical", "quality"},
		{"Mentoring", "leadership", "people"},
		{"Documentation", "communication", "writing"},
		{"API Design", "technical", "backend"},
		{"Testing", "technical", "quality"},
		{"Communication", "communication", "verbal"},
	}

	skillIDs = make([]string, 0)
	for _, s := range skills {
		skill := &models.Skill{
			Name:        s.name,
			Category:    s.category,
			Subcategory: s.subcategory,
			Description: fmt.Sprintf("Proficiency in %s", s.name),
		}
		if err := data.SkillsStore.CreateSkill(skill); err != nil {
			log.Printf("Warning: Failed to create skill: %v", err)
		} else {
			skillIDs = append(skillIDs, skill.ID)
		}
	}
	fmt.Printf("  Created %d skills\n", len(skillIDs))

	// Engineer Skills
	for _, engID := range data.EngineerIDs {
		numSkills := 3 + (len(engID) % 3)
		for i := 0; i < numSkills && i < len(skillIDs); i++ {
			skillIndex := (len(engID) + i) % len(skillIDs)
			levelScore := 50 + ((len(engID)+i)*7)%50

			engSkill := &models.EngineerSkill{
				EngineerID:    engID,
				SkillID:       skillIDs[skillIndex],
				LevelScore:    levelScore,
				Trajectory:    []string{"improving", "stable", "improving"}[i%3],
				LastEvaluated: data.Now.AddDate(0, 0, -7),
				EvidenceCount: 5 + i,
			}
			if err := data.SkillsStore.CreateEngineerSkill(engSkill); err != nil {
				log.Printf("Warning: Failed to create engineer skill: %v", err)
			} else {
				engineerSkillCount++
			}
		}
	}
	fmt.Printf("  Created %d engineer skills\n", engineerSkillCount)

	// Skill Evidence
	for i, engID := range data.EngineerIDs[:5] {
		for j := 0; j < 3; j++ {
			evidence := &models.SkillEvidence{
				EngineerID:     engID,
				SkillID:        skillIDs[j%len(skillIDs)],
				EvidenceType:   []string{"pr_complexity", "code_review", "design_doc"}[j%3],
				EvidenceSource: fmt.Sprintf("event_%d", i*3+j),
				Strength:       0.7 + float64(i)*0.05,
				Context:        `{"details": "Strong evidence of skill usage"}`,
				DetectedAt:     data.Now.AddDate(0, 0, -(i*7 + j)),
			}
			if err := data.SkillsStore.CreateSkillEvidence(evidence); err != nil {
				log.Printf("Warning: Failed to create skill evidence: %v", err)
			} else {
				evidenceCount++
			}
		}
	}
	fmt.Printf("  Created %d skill evidence entries\n", evidenceCount)

	return skillIDs, engineerSkillCount, evidenceCount
}
