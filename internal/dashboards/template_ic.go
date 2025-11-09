package dashboards

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// createICTemplate creates the Individual Contributor dashboard template
func createICTemplate() (*models.Dashboard, error) {
	layout := models.DashboardLayout{
		Widgets: []models.Widget{
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "My Score",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "engineer_score",
					"entity_type": "engineer",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 0, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "PRs This Week",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "pr_count",
					"entity_type": "engineer",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 3, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "Active Goals",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "active_goals",
					"entity_type": "engineer",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 6, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "status",
				Title:      "System Health",
				DataSource: "/api/metrics/health-status",
				Position:   models.WidgetPosition{X: 9, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "timeseries",
				Title:      "My Performance Trend",
				DataSource: "/api/metrics/timeseries",
				QueryParams: map[string]string{
					"metric":      "engineer_score",
					"entity_type": "engineer",
					"period":      "week",
					"count":       "12",
				},
				VisualizationConfig: models.VisualizationConfig{
					PrimaryColor: "#3b82f6",
					ShowLegend:   true,
					YAxisLabel:   "Score",
				},
				Position: models.WidgetPosition{X: 0, Y: 2, W: 6, H: 4},
			},
			{
				ID:         uuid.New().String(),
				Type:       "timeseries",
				Title:      "PR Volume Trend",
				DataSource: "/api/metrics/timeseries",
				QueryParams: map[string]string{
					"metric":      "pr_count",
					"entity_type": "engineer",
					"period":      "week",
					"count":       "12",
				},
				VisualizationConfig: models.VisualizationConfig{
					PrimaryColor: "#10b981",
					ShowLegend:   true,
					YAxisLabel:   "PRs",
				},
				Position: models.WidgetPosition{X: 6, Y: 2, W: 6, H: 4},
			},
			{
				ID:         uuid.New().String(),
				Type:       "feed",
				Title:      "My Recent Activity",
				DataSource: "/api/activity/feed",
				Position:   models.WidgetPosition{X: 0, Y: 6, W: 12, H: 3},
			},
		},
	}

	layoutJSON, err := json.Marshal(layout)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal IC template layout: %w", err)
	}

	return &models.Dashboard{
		ID:          uuid.New().String(),
		Name:        "Individual Contributor",
		Description: "Personal dashboard for individual contributors tracking their own performance, goals, and activity",
		Persona:     "ic",
		IsTemplate:  true,
		IsSystem:    true,
		Layout:      string(layoutJSON),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}
