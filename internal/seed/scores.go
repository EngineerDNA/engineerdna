package seed

import (
	"fmt"
	"log"

	"github.com/google/uuid"
)

// Base scores by role level
const (
	BaseScoreJunior = 70
	BaseScoreMid    = 90
	BaseScoreSenior = 100
	BaseScoreStaff  = 110
)

// SeedScores generates performance scores for engineers and teams
func SeedScores(data *SeedData) error {
	fmt.Println("\n6. Generating performance scores...")
	scoreCounter := 0
	for weekOffset := 0; weekOffset < 8; weekOffset++ {
		weekStart := data.Now.AddDate(0, 0, -7*weekOffset).Truncate(24 * 3600000000000) // 24 hours

		for i, engineerID := range data.EngineerIDs {
			eng := data.EngineerMap[engineerID]

			// Base score from role
			var baseScore float64
			if eng.Role == "Junior Engineer" {
				baseScore = BaseScoreJunior
			} else if eng.Role == "Mid-Level Engineer" {
				baseScore = BaseScoreMid
			} else if eng.Role == "Senior Engineer" {
				baseScore = BaseScoreSenior
			} else if eng.Role == "Staff Engineer" {
				baseScore = BaseScoreStaff
			}

			// Add realistic variation
			// Some engineers improving, some declining, some stable
			var trendFactor float64
			if i%3 == 0 {
				// Improving trend
				trendFactor = float64(weekOffset) * 1.5
			} else if i%3 == 1 {
				// Declining trend
				trendFactor = -float64(weekOffset) * 1.0
			} else {
				// Stable with noise
				trendFactor = float64(weekOffset%2)*2 - 1
			}

			totalScore := baseScore + trendFactor + float64((i*7)%10-5)

			// Component scores (realistic distribution)
			throughputScore := totalScore * 0.3
			qualityScore := totalScore * 0.25
			speedScore := totalScore * 0.2
			collaborationScore := totalScore * 0.15
			impactScore := totalScore * 0.1

			// Store scores in metric_values (new schema)
			dimensions := fmt.Sprintf(`{"engineer_id":"%s"}`, engineerID)

			// Total score
			_, err := data.Database.DB.Exec(`
				INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, uuid.New().String(), "engineer_total_score", "scoring_system", weekStart.Format("2006-01-02"), "weekly", totalScore, "score", dimensions, data.Now.Format("2006-01-02T15:04:05Z"))

			if err != nil {
				log.Printf("Warning: Failed to create total score: %v", err)
			}

			// Component scores
			componentScores := map[string]float64{
				"engineer_throughput_score":    throughputScore,
				"engineer_quality_score":       qualityScore,
				"engineer_speed_score":         speedScore,
				"engineer_collaboration_score": collaborationScore,
				"engineer_impact_score":        impactScore,
			}

			for metricName, value := range componentScores {
				_, err := data.Database.DB.Exec(`
					INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
				`, uuid.New().String(), metricName, "scoring_system", weekStart.Format("2006-01-02"), "weekly", value, "score", dimensions, data.Now.Format("2006-01-02T15:04:05Z"))

				if err != nil {
					log.Printf("Warning: Failed to create %s: %v", metricName, err)
				}
			}

			scoreCounter++
		}
	}
	fmt.Printf("  Created %d performance scores (8 weeks x %d engineers)\n", scoreCounter, len(data.EngineerIDs))

	// 6b. Aggregate Team Performance Scores
	fmt.Println("\n6b. Aggregating team performance scores...")
	teamScoreCounter := 0

	// Get all child teams (not parent Engineering team)
	childTeams := []struct {
		id   string
		name string
	}{
		{data.BackendTeam.ID, "Backend"},
		{data.FrontendTeam.ID, "Frontend"},
		{data.PlatformTeam.ID, "Platform"},
	}

	for weekOffset := 0; weekOffset < 8; weekOffset++ {
		weekStart := data.Now.AddDate(0, 0, -7*weekOffset).Truncate(24 * 3600000000000) // 24 hours

		for _, team := range childTeams {
			// Get team member IDs
			var memberIDs []string
			for engID, eng := range data.EngineerMap {
				if eng.TeamID == team.id {
					memberIDs = append(memberIDs, engID)
				}
			}

			if len(memberIDs) == 0 {
				continue
			}

			// Aggregate scores from team members (from metric_values)
			var totalScore float64
			var count int

			for _, memberID := range memberIDs {
				dimensions := fmt.Sprintf(`{"engineer_id":"%s"}`, memberID)
				timestamp := weekStart.Format("2006-01-02")

				// Get total score
				var score float64
				err := data.Database.DB.QueryRow(`
					SELECT value FROM metric_values
					WHERE metric_name = 'engineer_total_score'
					AND dimensions = ? AND timestamp = ?
				`, dimensions, timestamp).Scan(&score)

				if err == nil {
					totalScore += score
					count++
				}
			}

			if count > 0 {
				// Store team average scores in metric_values
				avgTotalScore := totalScore / float64(count)
				teamDimensions := fmt.Sprintf(`{"team_id":"%s","member_count":%d}`, team.id, count)

				_, err := data.Database.DB.Exec(`
					INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
				`, uuid.New().String(), "team_total_score", "scoring_system", weekStart.Format("2006-01-02"), "weekly", avgTotalScore, "score", teamDimensions, data.Now.Format("2006-01-02T15:04:05Z"))

				if err != nil {
					log.Printf("Warning: Failed to create team score: %v", err)
				} else {
					teamScoreCounter++
				}
			}
		}
	}
	fmt.Printf("  Created %d team performance scores (8 weeks x %d teams)\n", teamScoreCounter, len(childTeams))

	return nil
}
