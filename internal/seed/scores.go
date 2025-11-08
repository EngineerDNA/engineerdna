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

			scoreID := uuid.New().String()
			_, err := data.Database.DB.Exec(`
				INSERT INTO performance_scores (
					id, engineer_id, week_start, total_score,
					throughput_score, quality_score, speed_score,
					collaboration_score, impact_score, raw_metrics, created_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, scoreID, engineerID, weekStart, totalScore,
				throughputScore, qualityScore, speedScore,
				collaborationScore, impactScore, "{}", data.Now)

			if err != nil {
				log.Printf("Warning: Failed to create performance score: %v", err)
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

			// Aggregate scores from team members
			var totalScore, throughputScore, qualityScore, speedScore, collaborationScore, impactScore float64
			var count int

			for _, memberID := range memberIDs {
				var score, tp, qual, spd, collab, imp float64
				err := data.Database.DB.QueryRow(`
					SELECT total_score, throughput_score, quality_score, speed_score,
					       collaboration_score, impact_score
					FROM performance_scores
					WHERE engineer_id = ? AND week_start = ?
				`, memberID, weekStart).Scan(&score, &tp, &qual, &spd, &collab, &imp)

				if err == nil {
					totalScore += score
					throughputScore += tp
					qualityScore += qual
					speedScore += spd
					collaborationScore += collab
					impactScore += imp
					count++
				}
			}

			if count > 0 {
				// Calculate team averages
				teamScoreID := uuid.New().String()
				_, err := data.Database.DB.Exec(`
					INSERT INTO team_performance_scores (
						id, team_id, week_start, total_score,
						throughput_score, quality_score, speed_score,
						collaboration_score, impact_score, member_count, created_at
					) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				`, teamScoreID, team.id, weekStart, totalScore/float64(count),
					throughputScore/float64(count), qualityScore/float64(count), speedScore/float64(count),
					collaborationScore/float64(count), impactScore/float64(count), count, data.Now)

				if err != nil {
					log.Printf("Warning: Failed to create team performance score: %v", err)
				} else {
					teamScoreCounter++
				}
			}
		}
	}
	fmt.Printf("  Created %d team performance scores (8 weeks x %d teams)\n", teamScoreCounter, len(childTeams))

	return nil
}
