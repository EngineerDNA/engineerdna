package seed

import (
	"fmt"
	"log"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

// SeedBriefings generates weekly briefings
func SeedBriefings(data *SeedData) error {
	fmt.Println("\n7. Generating weekly briefings...")
	for weekOffset := 1; weekOffset <= 3; weekOffset++ {
		weekStart := data.Now.AddDate(0, 0, -7*weekOffset).Truncate(24 * time.Hour)

		briefing := &models.WeeklyBriefing{
			TeamID:    data.EngineeringTeam.ID,
			WeekStart: weekStart,
			TLDR:      fmt.Sprintf("Week of %s: Strong delivery with %d PRs merged, %d issues closed. Team velocity tracking above target.", weekStart.Format("Jan 2"), 30+(weekOffset*2), 15+weekOffset),
			KeyMetrics: []models.MetricSummary{
				{Name: "Team Score", Value: fmt.Sprintf("%.1f", 95.0-float64(weekOffset)*2), ChangePercent: -2.0, Direction: "down"},
				{Name: "PRs Merged", Value: fmt.Sprintf("%d", 30+(weekOffset*2)), ChangePercent: 5.0, Direction: "up"},
				{Name: "Issues Closed", Value: fmt.Sprintf("%d", 15+weekOffset), ChangePercent: 3.0, Direction: "up"},
			},
			NeedsAttention: []models.AttentionItem{
				{
					EngineerID:     data.EngineerIDs[6], // Maya Patel (Junior)
					EngineerName:   "Maya Patel",
					Severity:       "warning",
					Issue:          "Slower PR cycle time than peers",
					Evidence:       []string{"Average 4.5 days vs team average of 2.1 days", "3 PRs currently in review > 3 days"},
					SuggestedTopic: "Discuss code review process and identify blockers",
				},
			},
			Insights: []models.AIInsight{
				{
					Observation:    fmt.Sprintf("Backend team velocity increased %d%% this week", 5+(weekOffset*2)),
					Context:        "Marcus and Elena shipped major refactoring, unlocking faster feature development",
					Recommendation: "Consider documenting the new patterns for other teams",
				},
				{
					Observation:    "Frontend team showing strong collaboration metrics",
					Context:        "Code review participation up 25%, average review depth improved",
					Recommendation: "Recognize team collaboration in next all-hands",
				},
			},
			TrendingUp:    []string{"Marcus Johnson", "Elena Rodriguez", "Priya Sharma"},
			TrendingDown:  []string{"Maya Patel", "Nina Okafor"},
			TalkingPoints: fmt.Sprintf("1. Celebrate backend refactoring completion\n2. Check in with Maya on PR blockers\n3. Platform team capacity planning for Q%d", (weekOffset%4)+1),
			GeneratedAt:   weekStart.Add(24 * 6 * time.Hour), // Saturday morning
		}

		if err := data.BriefingStore.CreateBriefing(briefing); err != nil {
			log.Printf("Warning: Failed to create briefing: %v", err)
		}
		fmt.Printf("  Created briefing for week of %s\n", weekStart.Format("Jan 2"))
	}

	return nil
}
