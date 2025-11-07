package scoring

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

const (
	// MaxComponentScore caps component scores to prevent outliers
	MaxComponentScore = 150.0
)

// ScoringService handles performance score calculations
type ScoringService struct {
	db *sql.DB
}

// NewScoringService creates a new scoring service
func NewScoringService(db *sql.DB) *ScoringService {
	return &ScoringService{db: db}
}

// CalculateIndividualScore calculates performance score for an engineer for a given week
func (s *ScoringService) CalculateIndividualScore(engineerID string, weekStart time.Time) (*models.PerformanceScore, error) {
	// Normalize weekStart to beginning of day UTC
	weekStart = weekStart.UTC().Truncate(24 * time.Hour)
	weekEnd := weekStart.AddDate(0, 0, 7)

	// Fetch engineer's role
	role, err := s.getEngineerRole(engineerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get engineer role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("no role found for engineer %s", engineerID)
	}

	// Fetch scoring weights
	weights, err := s.getScoringWeights()
	if err != nil {
		return nil, fmt.Errorf("failed to get scoring weights: %w", err)
	}
	if weights == nil {
		return nil, fmt.Errorf("no scoring weights configured")
	}

	// Calculate raw metrics from events
	rawMetrics, err := s.calculateRawMetrics(engineerID, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate raw metrics: %w", err)
	}

	// Calculate component scores
	componentScores := s.calculateComponentScores(rawMetrics, role.Expectations)

	// Calculate normalized total score
	totalScore := s.normalizeScore(componentScores, weights)

	// Store raw metrics as JSON
	rawMetricsJSON, err := json.Marshal(rawMetrics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw metrics: %w", err)
	}

	// Detect burnout risks if score is high
	var burnoutRiskJSON string
	if totalScore > 120 {
		burnoutRisk, err := s.DetectBurnoutRisks(engineerID, weekStart, totalScore, rawMetrics)
		if err != nil {
			// Log error but don't fail the entire calculation
			fmt.Printf("Warning: failed to detect burnout risks: %v\n", err)
		} else if burnoutRisk != nil {
			burnoutJSON, err := json.Marshal(burnoutRisk)
			if err == nil {
				burnoutRiskJSON = string(burnoutJSON)
			}
		}
	}

	// Create performance score record
	score := &models.PerformanceScore{
		ID:                 uuid.New().String(),
		EngineerID:         engineerID,
		WeekStart:          weekStart,
		TotalScore:         totalScore,
		ThroughputScore:    componentScores.Throughput,
		QualityScore:       componentScores.Quality,
		SpeedScore:         componentScores.Speed,
		CollaborationScore: componentScores.Collaboration,
		ImpactScore:        componentScores.Impact,
		RawMetrics:         string(rawMetricsJSON),
		BurnoutRisk:        burnoutRiskJSON,
		CreatedAt:          time.Now().UTC(),
	}

	// Store in database
	err = s.storePerformanceScore(score)
	if err != nil {
		return nil, fmt.Errorf("failed to store performance score: %w", err)
	}

	return score, nil
}

// getEngineerRole retrieves the role for an engineer
func (s *ScoringService) getEngineerRole(engineerID string) (*models.Role, error) {
	// Query engineers table for role_id
	var roleID sql.NullString
	err := s.db.QueryRow(`SELECT role_id FROM engineers WHERE id = ?`, engineerID).Scan(&roleID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query engineer role: %w", err)
	}

	// If no role assigned, return nil
	if !roleID.Valid {
		return nil, nil
	}

	// Query roles table
	var role models.Role
	var expectationsJSON string
	err = s.db.QueryRow(`
		SELECT id, name, target_score, expectations, created_at, updated_at
		FROM roles WHERE id = ?
	`, roleID.String).Scan(&role.ID, &role.Name, &role.TargetScore, &expectationsJSON, &role.CreatedAt, &role.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query role: %w", err)
	}

	// Parse expectations JSON
	err = json.Unmarshal([]byte(expectationsJSON), &role.Expectations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal role expectations: %w", err)
	}

	return &role, nil
}

// getScoringWeights retrieves the current scoring weights
func (s *ScoringService) getScoringWeights() (*models.ScoringWeights, error) {
	var weights models.ScoringWeights

	// Get the most recent weights (there should typically be only one row)
	err := s.db.QueryRow(`
		SELECT id, throughput_weight, quality_weight, speed_weight, collaboration_weight, impact_weight, updated_at
		FROM scoring_weights
		ORDER BY updated_at DESC
		LIMIT 1
	`).Scan(&weights.ID, &weights.ThroughputWeight, &weights.QualityWeight, &weights.SpeedWeight, &weights.CollaborationWeight, &weights.ImpactWeight, &weights.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query scoring weights: %w", err)
	}

	return &weights, nil
}

// calculateRawMetrics calculates raw metrics from events for the given time period
func (s *ScoringService) calculateRawMetrics(engineerID string, weekStart, weekEnd time.Time) (*models.RawMetrics, error) {
	metrics := &models.RawMetrics{}

	// Throughput: PRs per week (count merged PRs)
	var prCount sql.NullInt64
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM events
		WHERE engineer_id = ?
		  AND type = 'pull_request'
		  AND json_extract(data, '$.action') = 'merged'
		  AND timestamp >= ?
		  AND timestamp < ?
	`, engineerID, weekStart, weekEnd).Scan(&prCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count PRs: %w", err)
	}
	if prCount.Valid {
		metrics.ThroughputPRsPerWeek = float64(prCount.Int64)
	}

	// Quality: Bug rate (bugs per PR)
	bugRate, err := s.calculateBugRate(engineerID, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate bug rate: %w", err)
	}
	metrics.QualityBugRate = bugRate

	// Quality: Rework rate (requires commit events with PR tracking - not yet implemented)
	metrics.QualityReworkRate = 0.0

	// Speed: Cycle time (average days from PR open to merge)
	cycleTime, err := s.calculateCycleTime(engineerID, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate cycle time: %w", err)
	}
	metrics.SpeedCycleTimeDays = cycleTime

	// Speed: Time to first review (requires pull_request_review events - not yet implemented)
	metrics.SpeedTimeToFirstReview = 0.0

	// Collaboration: Reviews given per week (requires pull_request_review events - not yet implemented)
	metrics.CollaborationReviewsGiven = 0.0

	// Collaboration: Review depth (requires pull_request_review events - not yet implemented)
	metrics.CollaborationReviewDepth = 0.0

	// Impact: Services touched (count distinct repositories)
	servicesTouched, err := s.calculateServicesTouched(engineerID, weekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate services touched: %w", err)
	}
	metrics.ImpactServicesTouched = servicesTouched

	return metrics, nil
}

// calculateBugRate calculates the bug rate (bugs per PR)
func (s *ScoringService) calculateBugRate(engineerID string, weekStart, weekEnd time.Time) (float64, error) {
	// Count merged PRs
	var prCount sql.NullInt64
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM events
		WHERE engineer_id = ?
		  AND type = 'pull_request'
		  AND json_extract(data, '$.action') = 'merged'
		  AND timestamp >= ?
		  AND timestamp < ?
	`, engineerID, weekStart, weekEnd).Scan(&prCount)
	if err != nil {
		return 0.0, fmt.Errorf("failed to count PRs: %w", err)
	}
	if !prCount.Valid || prCount.Int64 == 0 {
		return 0.0, nil // No PRs, no bug rate
	}

	// Count bug-related issues
	// Issues with "bug" label in their labels array
	var bugCount sql.NullInt64
	err = s.db.QueryRow(`
		SELECT COUNT(*)
		FROM events
		WHERE engineer_id = ?
		  AND type = 'issue'
		  AND json_extract(data, '$.action') = 'opened'
		  AND (
		    json_extract(data, '$.labels') LIKE '%bug%'
		    OR json_extract(data, '$.labels') LIKE '%Bug%'
		  )
		  AND timestamp >= ?
		  AND timestamp < ?
	`, engineerID, weekStart, weekEnd).Scan(&bugCount)
	if err != nil {
		return 0.0, fmt.Errorf("failed to count bugs: %w", err)
	}
	if !bugCount.Valid || bugCount.Int64 == 0 {
		return 0.0, nil // No bugs
	}

	// Calculate bugs per PR
	return float64(bugCount.Int64) / float64(prCount.Int64), nil
}

// calculateCycleTime calculates average cycle time in days (PR open to merge)
func (s *ScoringService) calculateCycleTime(engineerID string, weekStart, weekEnd time.Time) (float64, error) {
	// Get all merged PRs with their created_at and merged_at timestamps
	rows, err := s.db.Query(`
		SELECT
			json_extract(data, '$.created_at') as created_at,
			json_extract(data, '$.merged_at') as merged_at
		FROM events
		WHERE engineer_id = ?
		  AND type = 'pull_request'
		  AND json_extract(data, '$.action') = 'merged'
		  AND timestamp >= ?
		  AND timestamp < ?
		  AND json_extract(data, '$.created_at') IS NOT NULL
		  AND json_extract(data, '$.merged_at') IS NOT NULL
		LIMIT 10000
	`, engineerID, weekStart, weekEnd)
	if err != nil {
		return 0.0, fmt.Errorf("failed to query PRs: %w", err)
	}
	defer rows.Close()

	var totalDays float64
	var count int

	for rows.Next() {
		var createdAtStr, mergedAtStr string
		if err := rows.Scan(&createdAtStr, &mergedAtStr); err != nil {
			return 0.0, fmt.Errorf("failed to scan PR timestamps: %w", err)
		}

		// Parse timestamps
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			continue // Skip invalid timestamps
		}
		mergedAt, err := time.Parse(time.RFC3339, mergedAtStr)
		if err != nil {
			continue // Skip invalid timestamps
		}

		// Calculate duration in days
		duration := mergedAt.Sub(createdAt)
		days := duration.Hours() / 24.0
		totalDays += days
		count++
	}

	if err := rows.Err(); err != nil {
		return 0.0, fmt.Errorf("error iterating PRs: %w", err)
	}

	if count == 0 {
		return 0.0, nil // No PRs with valid timestamps
	}

	// Return average cycle time
	return totalDays / float64(count), nil
}

// calculateServicesTouched calculates distinct repositories touched in the last 30 days
func (s *ScoringService) calculateServicesTouched(engineerID string, weekStart time.Time) (float64, error) {
	// Look back 30 days from week start for monthly metric
	monthStart := weekStart.AddDate(0, 0, -30)

	var repoCount sql.NullInt64
	err := s.db.QueryRow(`
		SELECT COUNT(DISTINCT json_extract(data, '$.repo'))
		FROM events
		WHERE engineer_id = ?
		  AND timestamp >= ?
		  AND timestamp < ?
		  AND json_extract(data, '$.repo') IS NOT NULL
	`, engineerID, monthStart, weekStart.AddDate(0, 0, 7)).Scan(&repoCount)
	if err != nil {
		return 0.0, fmt.Errorf("failed to count services: %w", err)
	}

	if repoCount.Valid {
		return float64(repoCount.Int64), nil
	}
	return 0.0, nil
}

// calculateComponentScores calculates component scores based on actual vs baseline metrics
func (s *ScoringService) calculateComponentScores(rawMetrics *models.RawMetrics, baselines map[string]float64) *models.ComponentScores {
	scores := &models.ComponentScores{}

	// Throughput (higher is better)
	if baseline, ok := baselines["throughput_prs_per_week"]; ok && baseline > 0 {
		scores.Throughput = calculateComponentScore(rawMetrics.ThroughputPRsPerWeek, baseline, true)
	} else {
		scores.Throughput = 100.0 // Default to baseline if no expectation set
	}

	// Quality (lower is better for bug rate)
	// Average bug rate and rework rate scores
	var qualityScores []float64
	if baseline, ok := baselines["quality_bug_rate"]; ok && baseline > 0 {
		qualityScores = append(qualityScores, calculateComponentScore(rawMetrics.QualityBugRate, baseline, false))
	}
	// Rework rate not yet implemented (requires commit events)
	// if baseline, ok := baselines["quality_rework_rate"]; ok && baseline > 0 {
	//     qualityScores = append(qualityScores, calculateComponentScore(rawMetrics.QualityReworkRate, baseline, false))
	// }

	if len(qualityScores) > 0 {
		scores.Quality = average(qualityScores)
	} else {
		scores.Quality = 100.0 // Default to baseline if no quality metrics
	}

	// Speed (lower is better for cycle time and time to first review)
	var speedScores []float64
	if baseline, ok := baselines["speed_cycle_time_days"]; ok && baseline > 0 {
		speedScores = append(speedScores, calculateComponentScore(rawMetrics.SpeedCycleTimeDays, baseline, false))
	}
	// Time to first review not yet implemented (requires pull_request_review events)
	// if baseline, ok := baselines["speed_time_to_first_review"]; ok && baseline > 0 {
	//     speedScores = append(speedScores, calculateComponentScore(rawMetrics.SpeedTimeToFirstReview, baseline, false))
	// }

	if len(speedScores) > 0 {
		scores.Speed = average(speedScores)
	} else {
		scores.Speed = 100.0 // Default to baseline if no speed metrics
	}

	// Collaboration (higher is better for reviews given)
	// Not yet implemented (requires pull_request_review events)
	// For now, use placeholder
	scores.Collaboration = 100.0

	// Impact (higher is better for services touched)
	if baseline, ok := baselines["impact_services_touched"]; ok && baseline > 0 {
		scores.Impact = calculateComponentScore(rawMetrics.ImpactServicesTouched, baseline, true)
	} else {
		scores.Impact = 100.0 // Default to baseline if no expectation set
	}

	return scores
}

// average calculates the arithmetic mean of a slice of floats
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateComponentScore calculates a single component score with the given formula
func calculateComponentScore(actual, baseline float64, higherIsBetter bool) float64 {
	// Handle edge cases
	if baseline == 0 {
		// If baseline is 0, we can't calculate a meaningful score
		if actual == 0 {
			return 100.0 // Both are 0, consider it meeting baseline
		}
		// If actual > 0 but baseline is 0, we can't calculate ratio
		return 100.0 // Default to baseline score
	}

	var score float64
	if higherIsBetter {
		// For "higher is better" metrics (PRs, reviews given)
		score = (actual / baseline) * 100.0
	} else {
		// For "lower is better" metrics (bug rate, cycle time)
		if actual == 0 {
			// If actual is 0, that's perfect (no bugs, instant cycle time)
			score = MaxComponentScore
		} else {
			score = (baseline / actual) * 100.0
		}
	}

	// Cap at 150 to prevent outliers
	if score > MaxComponentScore {
		score = MaxComponentScore
	}

	// Ensure non-negative
	if score < 0 {
		score = 0
	}

	return score
}

// normalizeScore calculates the normalized total score based on component scores and weights
func (s *ScoringService) normalizeScore(componentScores *models.ComponentScores, weights *models.ScoringWeights) float64 {
	// Calculate weighted sum
	weightedSum := 0.0
	weightedSum += componentScores.Throughput * weights.ThroughputWeight
	weightedSum += componentScores.Quality * weights.QualityWeight
	weightedSum += componentScores.Speed * weights.SpeedWeight
	weightedSum += componentScores.Collaboration * weights.CollaborationWeight
	weightedSum += componentScores.Impact * weights.ImpactWeight

	// Calculate total weights
	totalWeights := weights.ThroughputWeight + weights.QualityWeight + weights.SpeedWeight + weights.CollaborationWeight + weights.ImpactWeight

	// Avoid division by zero
	if totalWeights == 0 {
		return 0.0
	}

	// Normalize to 100
	normalizedScore := (weightedSum / totalWeights) * 100.0

	// Round to 2 decimal places
	normalizedScore = math.Round(normalizedScore*100) / 100

	return normalizedScore
}

// storePerformanceScore stores a performance score in the database
func (s *ScoringService) storePerformanceScore(score *models.PerformanceScore) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO performance_scores (
			id, engineer_id, week_start, total_score,
			throughput_score, quality_score, speed_score, collaboration_score, impact_score,
			raw_metrics, burnout_risk, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, score.ID, score.EngineerID, score.WeekStart, score.TotalScore,
		score.ThroughputScore, score.QualityScore, score.SpeedScore, score.CollaborationScore, score.ImpactScore,
		score.RawMetrics, score.BurnoutRisk, score.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert performance score: %w", err)
	}

	return nil
}

// DetectBurnoutRisks analyzes work patterns to detect burnout indicators
func (s *ScoringService) DetectBurnoutRisks(engineerID string, weekStart time.Time, totalScore float64, currentMetrics *models.RawMetrics) (*models.BurnoutRisk, error) {
	weekEnd := weekStart.AddDate(0, 0, 7)
	redFlags := []models.RedFlag{}

	// 1. Check for late-night work (10pm-5am)
	lateNightCount, err := s.countLateNightCommits(engineerID, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to count late night commits: %w", err)
	}
	if lateNightCount >= 3 {
		severity := "medium"
		if lateNightCount >= 5 {
			severity = "high"
		}
		redFlags = append(redFlags, models.RedFlag{
			Type:     "late_night_work",
			Severity: severity,
			Evidence: fmt.Sprintf("%d commits between 10pm-5am", lateNightCount),
			Count:    lateNightCount,
		})
	}

	// 2. Check for weekend work
	weekendCount, err := s.countWeekendCommits(engineerID, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to count weekend commits: %w", err)
	}
	if weekendCount >= 2 {
		severity := "medium"
		if weekendCount >= 4 {
			severity = "high"
		}
		redFlags = append(redFlags, models.RedFlag{
			Type:     "weekend_work",
			Severity: severity,
			Evidence: fmt.Sprintf("%d commits on Saturday/Sunday", weekendCount),
			Count:    weekendCount,
		})
	}

	// 3. Check for quality decline (bug rate increasing)
	previousWeekStart := weekStart.AddDate(0, 0, -7)
	previousWeekEnd := previousWeekStart.AddDate(0, 0, 7)
	previousBugRate, err := s.calculateBugRate(engineerID, previousWeekStart, previousWeekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate previous bug rate: %w", err)
	}

	// Only flag if both weeks have data and current rate is 30% higher
	if currentMetrics.QualityBugRate > 0 && previousBugRate > 0 {
		increase := (currentMetrics.QualityBugRate - previousBugRate) / previousBugRate
		if increase > 0.3 { // 30% increase
			severity := "medium"
			if increase > 0.5 { // 50% increase
				severity = "high"
			}
			redFlags = append(redFlags, models.RedFlag{
				Type:     "quality_decline",
				Severity: severity,
				Evidence: fmt.Sprintf("Bug rate increased from %.2f to %.2f bugs/PR (%.0f%% increase)",
					previousBugRate, currentMetrics.QualityBugRate, increase*100),
				Count: int(increase * 100),
			})
		}
	}

	// 4. Check for review depth decline
	// Note: Review depth not yet implemented, placeholder for future
	// When pull_request_review events are available, implement this

	// Calculate risk level based on number of red flags
	riskLevel := "none"
	recommendation := "No burnout indicators detected. Sustainable high performance."

	if len(redFlags) >= 3 {
		riskLevel = "high"
		recommendation = "Immediate check-in recommended. Multiple burnout indicators detected. Suggested 1:1 topic: 'I noticed you're working unusual hours. What's driving this? Can we adjust scope?'"
	} else if len(redFlags) >= 2 {
		riskLevel = "medium"
		recommendation = "Monitor closely, consider scheduling a 1:1. Work patterns suggest increased stress or workload."
	} else if len(redFlags) >= 1 {
		riskLevel = "low"
		recommendation = "Keep an eye on work patterns. Single indicator detected but not yet concerning."
	}

	// If no red flags, return nil to indicate no burnout risk
	if len(redFlags) == 0 {
		return nil, nil
	}

	return &models.BurnoutRisk{
		EngineerID:     engineerID,
		WeekStart:      weekStart,
		Score:          totalScore,
		RedFlags:       redFlags,
		RiskLevel:      riskLevel,
		Recommendation: recommendation,
		DetectedAt:     time.Now().UTC(),
	}, nil
}

// countLateNightCommits counts commits between 10pm and 5am
func (s *ScoringService) countLateNightCommits(engineerID string, weekStart, weekEnd time.Time) (int, error) {
	// Query for commits (we'll need to look at event timestamps)
	// For GitHub events, the timestamp field contains the commit time
	rows, err := s.db.Query(`
		SELECT timestamp
		FROM events
		WHERE engineer_id = ?
		  AND (type = 'push' OR type = 'commit')
		  AND timestamp >= ?
		  AND timestamp < ?
		LIMIT 10000
	`, engineerID, weekStart, weekEnd)
	if err != nil {
		return 0, fmt.Errorf("failed to query commits: %w", err)
	}
	defer rows.Close()

	lateNightCount := 0
	for rows.Next() {
		var timestamp time.Time
		if err := rows.Scan(&timestamp); err != nil {
			return 0, fmt.Errorf("failed to scan timestamp: %w", err)
		}

		// Check if commit hour is between 22:00 (10pm) and 05:00 (5am)
		hour := timestamp.Hour()
		if hour >= 22 || hour < 5 {
			lateNightCount++
		}
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("error iterating commits: %w", err)
	}

	return lateNightCount, nil
}

// countWeekendCommits counts commits on Saturday and Sunday
func (s *ScoringService) countWeekendCommits(engineerID string, weekStart, weekEnd time.Time) (int, error) {
	rows, err := s.db.Query(`
		SELECT timestamp
		FROM events
		WHERE engineer_id = ?
		  AND (type = 'push' OR type = 'commit')
		  AND timestamp >= ?
		  AND timestamp < ?
		LIMIT 10000
	`, engineerID, weekStart, weekEnd)
	if err != nil {
		return 0, fmt.Errorf("failed to query commits: %w", err)
	}
	defer rows.Close()

	weekendCount := 0
	for rows.Next() {
		var timestamp time.Time
		if err := rows.Scan(&timestamp); err != nil {
			return 0, fmt.Errorf("failed to scan timestamp: %w", err)
		}

		// Check if day is Saturday (6) or Sunday (0)
		weekday := timestamp.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			weekendCount++
		}
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("error iterating commits: %w", err)
	}

	return weekendCount, nil
}
