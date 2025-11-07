package skills

import (
	"fmt"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Detector analyzes events and detects skill evidence
type Detector struct {
	skillsStore *db.SkillsStore
	eventStore  *db.EventStore
}

// NewDetector creates a new skill detector
func NewDetector(skillsStore *db.SkillsStore, eventStore *db.EventStore) *Detector {
	return &Detector{
		skillsStore: skillsStore,
		eventStore:  eventStore,
	}
}

// AnalyzeEventForSkills extracts all skill evidence from an event
func (d *Detector) AnalyzeEventForSkills(event *models.Event, engineerID string) ([]*models.SkillEvidence, error) {
	var evidences []*models.SkillEvidence

	// Event data is already a map
	eventData := event.Data

	// Detect different skills based on event type
	switch event.Type {
	case "pull_request", "pull_request_merged":
		evidences = append(evidences, d.detectPRSkills(event, eventData, engineerID)...)
	case "code_review", "review_comment":
		evidences = append(evidences, d.detectReviewSkills(event, eventData, engineerID)...)
	case "commit":
		evidences = append(evidences, d.detectCommitSkills(event, eventData, engineerID)...)
	case "design_doc", "documentation":
		evidences = append(evidences, d.detectDocumentationSkills(event, eventData, engineerID)...)
	}

	// Store evidences
	for _, evidence := range evidences {
		if err := d.skillsStore.CreateSkillEvidence(evidence); err != nil {
			return nil, fmt.Errorf("failed to store skill evidence: %w", err)
		}
	}

	return evidences, nil
}

// detectPRSkills detects skills from pull requests
func (d *Detector) detectPRSkills(event *models.Event, data map[string]interface{}, engineerID string) []*models.SkillEvidence {
	var evidences []*models.SkillEvidence

	// System Design - multi-service PRs
	if services, ok := data["services_touched"].([]interface{}); ok && len(services) >= 3 {
		evidences = append(evidences, &models.SkillEvidence{
			EngineerID:     engineerID,
			SkillID:        "system_design",
			EvidenceType:   "pr_complexity",
			EvidenceSource: event.ID,
			Strength:       0.3 + float64(len(services))*0.05, // Max 0.5
			DetectedAt:     time.Now().UTC(),
		})
	}

	// Code Quality - low bug rate, high test coverage
	if additions, ok := data["additions"].(float64); ok {
		testCoverage := 0.0
		if coverage, ok := data["test_coverage"].(float64); ok {
			testCoverage = coverage
		}

		// Higher test coverage = higher code quality evidence
		if testCoverage > 0.8 {
			evidences = append(evidences, &models.SkillEvidence{
				EngineerID:     engineerID,
				SkillID:        "code_quality",
				EvidenceType:   "test_coverage",
				EvidenceSource: event.ID,
				Strength:       testCoverage * 0.5,
				DetectedAt:     time.Now().UTC(),
			})
		}

		// Testing skill - test-to-code ratio
		if testAdditions, ok := data["test_additions"].(float64); ok && additions > 0 {
			testRatio := testAdditions / additions
			if testRatio > 0.3 {
				evidences = append(evidences, &models.SkillEvidence{
					EngineerID:     engineerID,
					SkillID:        "testing",
					EvidenceType:   "test_to_code_ratio",
					EvidenceSource: event.ID,
					Strength:       min(testRatio, 1.0) * 0.6,
					DetectedAt:     time.Now().UTC(),
				})
			}
		}
	}

	// Performance - performance-related PRs
	if title, ok := data["title"].(string); ok {
		perfKeywords := []string{"performance", "optimize", "faster", "speed", "cache", "benchmark"}
		for _, keyword := range perfKeywords {
			if strings.Contains(strings.ToLower(title), keyword) {
				evidences = append(evidences, &models.SkillEvidence{
					EngineerID:     engineerID,
					SkillID:        "performance",
					EvidenceType:   "perf_improvement",
					EvidenceSource: event.ID,
					Strength:       0.4,
					DetectedAt:     time.Now().UTC(),
				})
				break
			}
		}
	}

	// DevOps - infrastructure changes
	if files, ok := data["files"].([]interface{}); ok {
		infraFiles := 0
		for _, file := range files {
			if filename, ok := file.(string); ok {
				if isInfraFile(filename) {
					infraFiles++
				}
			}
		}
		if infraFiles > 0 {
			evidences = append(evidences, &models.SkillEvidence{
				EngineerID:     engineerID,
				SkillID:        "devops",
				EvidenceType:   "infra_changes",
				EvidenceSource: event.ID,
				Strength:       min(float64(infraFiles)*0.15, 0.5),
				DetectedAt:     time.Now().UTC(),
			})
		}
	}

	return evidences
}

// detectReviewSkills detects skills from code reviews
func (d *Detector) detectReviewSkills(event *models.Event, data map[string]interface{}, engineerID string) []*models.SkillEvidence {
	var evidences []*models.SkillEvidence

	// Code Review skill
	commentCount := 0
	if comments, ok := data["comments"].(float64); ok {
		commentCount = int(comments)
	}

	strength := 0.2
	if commentCount > 5 {
		strength = 0.4
	}
	if commentCount > 10 {
		strength = 0.6
	}

	evidences = append(evidences, &models.SkillEvidence{
		EngineerID:     engineerID,
		SkillID:        "code_review",
		EvidenceType:   "review_depth",
		EvidenceSource: event.ID,
		Strength:       strength,
		DetectedAt:     time.Now().UTC(),
	})

	// Mentoring - detailed, constructive reviews
	if body, ok := data["body"].(string); ok {
		mentoringKeywords := []string{"consider", "suggestion", "pattern", "best practice", "alternative"}
		mentoringCount := 0
		for _, keyword := range mentoringKeywords {
			if strings.Contains(strings.ToLower(body), keyword) {
				mentoringCount++
			}
		}
		if mentoringCount >= 2 {
			evidences = append(evidences, &models.SkillEvidence{
				EngineerID:     engineerID,
				SkillID:        "mentoring",
				EvidenceType:   "review_depth",
				EvidenceSource: event.ID,
				Strength:       min(float64(mentoringCount)*0.15, 0.5),
				DetectedAt:     time.Now().UTC(),
			})
		}
	}

	return evidences
}

// detectCommitSkills detects skills from commits
func (d *Detector) detectCommitSkills(event *models.Event, data map[string]interface{}, engineerID string) []*models.SkillEvidence {
	var evidences []*models.SkillEvidence

	// Communication - clear commit messages
	if message, ok := data["message"].(string); ok && len(message) > 50 {
		// Good commit messages show communication skill
		evidences = append(evidences, &models.SkillEvidence{
			EngineerID:     engineerID,
			SkillID:        "communication",
			EvidenceType:   "commit_message",
			EvidenceSource: event.ID,
			Strength:       0.1,
			DetectedAt:     time.Now().UTC(),
		})
	}

	return evidences
}

// detectDocumentationSkills detects skills from documentation
func (d *Detector) detectDocumentationSkills(event *models.Event, data map[string]interface{}, engineerID string) []*models.SkillEvidence {
	var evidences []*models.SkillEvidence

	docType := "documentation"
	if dtype, ok := data["doc_type"].(string); ok {
		docType = dtype
	}

	// Technical Writing
	evidences = append(evidences, &models.SkillEvidence{
		EngineerID:     engineerID,
		SkillID:        "technical_writing",
		EvidenceType:   "docs_authored",
		EvidenceSource: event.ID,
		Strength:       0.5,
		DetectedAt:     time.Now().UTC(),
	})

	// System Design - design documents
	if docType == "design_doc" || docType == "architecture" {
		evidences = append(evidences, &models.SkillEvidence{
			EngineerID:     engineerID,
			SkillID:        "system_design",
			EvidenceType:   "design_doc",
			EvidenceSource: event.ID,
			Strength:       0.6,
			DetectedAt:     time.Now().UTC(),
		})
	}

	return evidences
}

// DetectCodeQualitySkill calculates code quality from metrics
func (d *Detector) DetectCodeQualitySkill(engineerID string, timeframe time.Duration) (*models.SkillEvidence, error) {
	endDate := time.Now().UTC()
	startDate := endDate.Add(-timeframe)

	// Query events for engineer in timeframe
	filters := map[string]interface{}{
		"engineer_id": engineerID,
		"start_date":  startDate,
		"end_date":    endDate,
	}

	events, err := d.eventStore.List(filters, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for engineer %s in timeframe", engineerID)
	}

	// Calculate code quality metrics
	var totalPRs, highQualityPRs, totalBugs int
	var totalTestCoverage float64
	var testCoverageCount int

	for _, event := range events {
		switch event.Type {
		case "pull_request", "pull_request_merged":
			totalPRs++

			// Check for quality indicators
			if coverage, ok := event.Data["test_coverage"].(float64); ok && coverage > 0 {
				totalTestCoverage += coverage
				testCoverageCount++
				if coverage > 0.8 {
					highQualityPRs++
				}
			}

		case "bug", "issue":
			if severity, ok := event.Data["severity"].(string); ok {
				if severity == "critical" || severity == "high" {
					totalBugs++
				}
			}
		}
	}

	// Calculate strength based on metrics
	strength := 0.5 // Base strength

	// High test coverage increases strength
	if testCoverageCount > 0 {
		avgCoverage := totalTestCoverage / float64(testCoverageCount)
		strength += avgCoverage * 0.3
	}

	// High quality PR ratio increases strength
	if totalPRs > 0 {
		qualityRatio := float64(highQualityPRs) / float64(totalPRs)
		strength += qualityRatio * 0.2
	}

	// Bug rate decreases strength
	if totalPRs > 0 {
		bugRate := float64(totalBugs) / float64(totalPRs)
		strength -= bugRate * 0.3
	}

	// Clamp to 0-1 range
	if strength > 1.0 {
		strength = 1.0
	}
	if strength < 0.0 {
		strength = 0.0
	}

	return &models.SkillEvidence{
		EngineerID:   engineerID,
		SkillID:      "code_quality",
		EvidenceType: "metrics",
		Strength:     strength,
		DetectedAt:   endDate,
	}, nil
}

// Helper functions

func isInfraFile(filename string) bool {
	infraPatterns := []string{
		"Dockerfile", "docker-compose", ".yml", ".yaml",
		"terraform", ".tf", "kubernetes", "k8s",
		".github/workflows", "Makefile", "deploy",
	}
	for _, pattern := range infraPatterns {
		if strings.Contains(filename, pattern) {
			return true
		}
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
