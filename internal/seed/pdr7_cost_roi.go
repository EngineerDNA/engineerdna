package seed

import (
	"fmt"
	"log"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/internal/utils"
)

// SeedCostROI creates cost/ROI analysis data
func SeedCostROI(data *SeedData) (featureIDs []string, roiCount int) {
	fmt.Println("\n13. Creating cost/ROI analysis data...")

	// Cost Configuration
	costConfig := &models.CostConfiguration{
		EntityType:    "org",
		MonthlyCost:   12500.0,
		Currency:      "USD",
		EffectiveFrom: data.Now.AddDate(-1, 0, 0),
		Notes:         "Average fully-loaded cost per engineer",
	}
	if err := data.CostROIStore.CreateCostConfiguration(costConfig); err != nil {
		log.Printf("Warning: Failed to create cost configuration: %v", err)
	}
	fmt.Printf("  Created cost configuration\n")

	// Feature Values
	features := []struct {
		name       string
		valueType  string
		value      float64
		confidence string
	}{
		{"Authentication System", "arr_impact", 250000, "validated"},
		{"Dashboard Analytics", "efficiency_gain", 100000, "estimated"},
		{"API Gateway", "customer_acquisition", 500000, "validated"},
		{"Mobile App", "arr_impact", 750000, "estimated"},
		{"Notification System", "retention_improvement", 150000, "estimated"},
		{"Admin Panel", "efficiency_gain", 75000, "actual"},
		{"Search Feature", "customer_acquisition", 200000, "estimated"},
		{"Export Functionality", "arr_impact", 100000, "validated"},
	}

	featureIDs = make([]string, 0)
	for _, f := range features {
		feature := &models.FeatureValue{
			FeatureName:        f.name,
			FeatureDescription: fmt.Sprintf("Implementation of %s", f.name),
			ValueType:          f.valueType,
			ValueAmount:        &f.value,
			ValueCurrency:      "USD",
			ConfidenceLevel:    f.confidence,
			Source:             "manual",
			TimePeriod:         "Q4 2024",
		}
		if err := data.CostROIStore.CreateFeatureValue(feature); err != nil {
			log.Printf("Warning: Failed to create feature value: %v", err)
		} else {
			featureIDs = append(featureIDs, feature.ID)
		}
	}
	fmt.Printf("  Created %d feature values\n", len(featureIDs))

	// Feature Work Items
	workItemCount := 0
	for i, featureID := range featureIDs[:5] {
		for j := 0; j < 3; j++ {
			workItem := &models.FeatureWorkItem{
				FeatureID:    featureID,
				WorkItemType: "story",
				WorkItemID:   fmt.Sprintf("STORY-%d", i*3+j+1),
				StoryPoints:  utils.PtrInt(5 + j*2),
				ActualHours:  utils.PtrFloat64(float64(20 + j*10)),
				EngineerID:   &data.EngineerIDs[i],
				TeamID:       &data.BackendTeam.ID,
				CompletedAt:  utils.PtrTime(data.Now.AddDate(0, 0, -(30 - i*7 - j))),
			}
			if err := data.CostROIStore.CreateFeatureWorkItem(workItem); err != nil {
				log.Printf("Warning: Failed to create feature work item: %v", err)
			} else {
				workItemCount++
			}
		}
	}
	fmt.Printf("  Created %d feature work items\n", workItemCount)

	// Feature Costs
	for i, featureID := range featureIDs[:5] {
		cost := &models.FeatureCost{
			FeatureID:         featureID,
			TotalStoryPoints:  15 + i*5,
			TotalHours:        120.0 + float64(i*40),
			TotalCost:         25000.0 + float64(i*5000),
			CostBreakdown:     `{"labor": 20000, "infrastructure": 5000}`,
			ComputationMethod: "story_points",
			ComputedAt:        data.Now,
		}
		if err := data.CostROIStore.CreateFeatureCost(cost); err != nil {
			log.Printf("Warning: Failed to create feature cost: %v", err)
		}
	}
	fmt.Printf("  Created feature costs\n")

	// ROI Calculations
	for i, featureID := range featureIDs[:5] {
		roi := &models.ROICalculation{
			FeatureID:     featureID,
			Investment:    25000.0 + float64(i*5000),
			ReturnValue:   utils.PtrFloat64(features[i].value),
			ROIPercentage: utils.PtrFloat64((features[i].value - 25000.0 - float64(i*5000)) / (25000.0 + float64(i*5000)) * 100),
			PaybackMonths: utils.PtrFloat64(3.0 + float64(i)*0.5),
			Confidence:    features[i].confidence,
			CalculatedAt:  data.Now,
		}
		if err := data.CostROIStore.CreateROICalculation(roi); err != nil {
			log.Printf("Warning: Failed to create ROI calculation: %v", err)
		} else {
			roiCount++
		}
	}
	fmt.Printf("  Created %d ROI calculations\n", roiCount)

	return featureIDs, roiCount
}

// SeedManagerContext creates manager context data
func SeedManagerContext(data *SeedData) (noteCount, responseCount int) {
	fmt.Println("\n14. Creating manager context data...")

	// Manager Notes
	notes := []struct {
		subjectType string
		subjectID   string
		noteType    string
		title       string
		content     string
		mood        string
	}{
		{"engineer", data.EngineerIDs[1], "1on1", "Marcus - Q1 Check-in", "Strong technical performance. Interested in leading more initiatives. Consider for tech lead role.", "positive"},
		{"engineer", data.EngineerIDs[2], "performance", "Elena - Code Review Excellence", "Consistently provides high-quality code reviews. Mentoring David effectively.", "positive"},
		{"engineer", data.EngineerIDs[3], "context", "David - Learning Progress", "Making good progress on backend skills. Needs more exposure to system design.", "neutral"},
		{"engineer", data.EngineerIDs[6], "1on1", "Maya - Workload Discussion", "Concerned about workload. Discussed prioritization and delegation strategies.", "concerned"},
		{"team", data.BackendTeam.ID, "context", "Backend Team - Sprint 5 Retrospective", "Team velocity improving. Tech debt paydown showing results.", "positive"},
		{"team", data.FrontendTeam.ID, "performance", "Frontend Team - Q1 Performance", "Solid delivery but PR review time needs improvement.", "neutral"},
	}

	for _, n := range notes {
		note := &models.ManagerNote{
			ManagerID:   data.EngineerIDs[0],
			SubjectType: n.subjectType,
			SubjectID:   n.subjectID,
			NoteType:    n.noteType,
			Title:       n.title,
			Content:     n.content,
			Visibility:  "private",
			Mood:        n.mood,
		}
		if err := data.ContextStore.CreateManagerNote(note); err != nil {
			log.Printf("Warning: Failed to create manager note: %v", err)
		} else {
			noteCount++
		}
	}
	fmt.Printf("  Created %d manager notes\n", noteCount)

	// Sentiment Surveys
	survey := &models.SentimentSurvey{
		SurveyType:  "quarterly",
		Title:       "Q1 2025 Team Sentiment Survey",
		Description: "Quarterly pulse check on team morale and satisfaction",
		Questions: []models.SurveyQuestion{
			{ID: "q1", Type: "scale", Question: "How satisfied are you with your work?"},
			{ID: "q2", Type: "scale", Question: "How is team collaboration?"},
			{ID: "q3", Type: "scale", Question: "Do you feel supported?"},
		},
		TargetAudience: "all",
		Anonymous:      true,
		Active:         true,
		CreatedBy:      data.EngineerIDs[0],
		ExpiresAt:      utils.PtrTime(data.Now.AddDate(0, 0, 30)),
	}
	if err := data.ContextStore.CreateSentimentSurvey(survey); err != nil {
		log.Printf("Warning: Failed to create sentiment survey: %v", err)
	}
	fmt.Printf("  Created sentiment survey\n")

	// Survey Responses
	for i, engID := range data.EngineerIDs {
		response := &models.SurveyResponse{
			SurveyID:     survey.ID,
			RespondentID: &engID,
			Responses: map[string]interface{}{
				"q1": 4 + i%2,
				"q2": 4 + i%2,
				"q3": 3 + i%3,
			},
			MoodRating:   utils.PtrInt(4 + i%2),
			TextFeedback: "Overall positive experience. Team collaboration is strong.",
			SubmittedAt:  data.Now.AddDate(0, 0, -i),
		}
		if err := data.ContextStore.CreateSurveyResponse(response); err != nil {
			log.Printf("Warning: Failed to create survey response: %v", err)
		} else {
			responseCount++
		}
	}
	fmt.Printf("  Created %d survey responses\n", responseCount)

	return noteCount, responseCount
}
