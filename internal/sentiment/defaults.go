package sentiment

import "github.com/engineerdna/engineerdna/internal/models"

// GetDefaultPulseQuestions returns the default weekly pulse survey questions
func GetDefaultPulseQuestions() []models.SurveyQuestion {
	return []models.SurveyQuestion{
		{
			ID:       "mood",
			Type:     "scale",
			Question: "How are you feeling this week?",
			Scale: []models.SurveyScaleOption{
				{Value: 5, Label: "Great"},
				{Value: 4, Label: "Good"},
				{Value: 3, Label: "Okay"},
				{Value: 2, Label: "Not great"},
				{Value: 1, Label: "Struggling"},
			},
		},
		{
			ID:       "workload",
			Type:     "scale",
			Question: "How does your workload feel?",
			Scale: []models.SurveyScaleOption{
				{Value: 1, Label: "Too light"},
				{Value: 2, Label: "Just right"},
				{Value: 3, Label: "Too heavy"},
			},
		},
		{
			ID:       "work_life_balance",
			Type:     "scale",
			Question: "How is your work-life balance?",
			Scale: []models.SurveyScaleOption{
				{Value: 5, Label: "Excellent"},
				{Value: 4, Label: "Good"},
				{Value: 3, Label: "Okay"},
				{Value: 2, Label: "Struggling"},
				{Value: 1, Label: "Poor"},
			},
		},
		{
			ID:       "blocked",
			Type:     "choice",
			Question: "Are you blocked on anything?",
			Choices:  []string{"No", "Yes - technical", "Yes - organizational", "Yes - external dependency"},
		},
		{
			ID:       "feedback",
			Type:     "text",
			Question: "What's working well? What's frustrating? (Optional)",
		},
	}
}

// GetQuarterlyQuestions returns quarterly engagement survey questions
func GetQuarterlyQuestions() []models.SurveyQuestion {
	return []models.SurveyQuestion{
		{
			ID:       "engagement",
			Type:     "scale",
			Question: "How engaged do you feel in your work?",
			Scale: []models.SurveyScaleOption{
				{Value: 5, Label: "Highly engaged"},
				{Value: 4, Label: "Engaged"},
				{Value: 3, Label: "Neutral"},
				{Value: 2, Label: "Somewhat disengaged"},
				{Value: 1, Label: "Disengaged"},
			},
		},
		{
			ID:       "growth",
			Type:     "scale",
			Question: "Do you feel you're growing in your role?",
			Scale: []models.SurveyScaleOption{
				{Value: 5, Label: "Strongly agree"},
				{Value: 4, Label: "Agree"},
				{Value: 3, Label: "Neutral"},
				{Value: 2, Label: "Disagree"},
				{Value: 1, Label: "Strongly disagree"},
			},
		},
		{
			ID:       "tools",
			Type:     "scale",
			Question: "Do you have the tools and resources you need?",
			Scale: []models.SurveyScaleOption{
				{Value: 5, Label: "Strongly agree"},
				{Value: 4, Label: "Agree"},
				{Value: 3, Label: "Neutral"},
				{Value: 2, Label: "Disagree"},
				{Value: 1, Label: "Strongly disagree"},
			},
		},
		{
			ID:       "team_collaboration",
			Type:     "scale",
			Question: "How would you rate team collaboration?",
			Scale: []models.SurveyScaleOption{
				{Value: 5, Label: "Excellent"},
				{Value: 4, Label: "Good"},
				{Value: 3, Label: "Okay"},
				{Value: 2, Label: "Needs improvement"},
				{Value: 1, Label: "Poor"},
			},
		},
		{
			ID:       "manager_support",
			Type:     "scale",
			Question: "Do you feel supported by your manager?",
			Scale: []models.SurveyScaleOption{
				{Value: 5, Label: "Strongly agree"},
				{Value: 4, Label: "Agree"},
				{Value: 3, Label: "Neutral"},
				{Value: 2, Label: "Disagree"},
				{Value: 1, Label: "Strongly disagree"},
			},
		},
		{
			ID:       "recommend",
			Type:     "scale",
			Question: "Would you recommend this company to a friend?",
			Scale: []models.SurveyScaleOption{
				{Value: 10, Label: "Definitely"},
				{Value: 9, Label: "Very likely"},
				{Value: 8, Label: "Likely"},
				{Value: 7, Label: "Somewhat likely"},
				{Value: 6, Label: "Neutral"},
				{Value: 5, Label: "Somewhat unlikely"},
				{Value: 4, Label: "Unlikely"},
				{Value: 3, Label: "Very unlikely"},
				{Value: 2, Label: "Definitely not"},
				{Value: 1, Label: "Absolutely not"},
			},
		},
		{
			ID:       "comments",
			Type:     "text",
			Question: "Any additional feedback or suggestions?",
		},
	}
}
