package dashboards

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// createDirectorTemplate creates the Director dashboard template
func createDirectorTemplate() (*models.Dashboard, error) {
	layout := models.DashboardLayout{
		Widgets: []models.Widget{
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "Org Score",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "team_score",
					"entity_type": "org",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 0, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "Total PRs",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "pr_volume",
					"entity_type": "org",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 3, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "Critical Alerts",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "active_alerts",
					"entity_type": "org",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 6, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "status",
				Title:      "Org Health",
				DataSource: "/api/metrics/health-status",
				Position:   models.WidgetPosition{X: 9, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "bar",
				Title:      "Team Comparison",
				DataSource: "/api/metrics/compare",
				QueryParams: map[string]string{
					"metric":   "team_score",
					"group_by": "team",
					"period":   "week",
				},
				VisualizationConfig: models.VisualizationConfig{
					PrimaryColor: "#3b82f6",
				},
				Position: models.WidgetPosition{X: 0, Y: 2, W: 6, H: 4},
			},
			{
				ID:         uuid.New().String(),
				Type:       "bar",
				Title:      "Team PR Volume",
				DataSource: "/api/metrics/compare",
				QueryParams: map[string]string{
					"metric":   "pr_volume",
					"group_by": "team",
					"period":   "week",
				},
				VisualizationConfig: models.VisualizationConfig{
					PrimaryColor: "#10b981",
				},
				Position: models.WidgetPosition{X: 6, Y: 2, W: 6, H: 4},
			},
			{
				ID:         uuid.New().String(),
				Type:       "timeseries",
				Title:      "Organization Performance Trend",
				DataSource: "/api/metrics/timeseries",
				QueryParams: map[string]string{
					"metric":      "team_score",
					"entity_type": "org",
					"period":      "week",
					"count":       "12",
				},
				VisualizationConfig: models.VisualizationConfig{
					PrimaryColor: "#8b5cf6",
					ShowLegend:   true,
					YAxisLabel:   "Org Score",
				},
				Position: models.WidgetPosition{X: 0, Y: 6, W: 12, H: 4},
			},
		},
	}

	layoutJSON, err := json.Marshal(layout)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Director template layout: %w", err)
	}

	return &models.Dashboard{
		ID:          uuid.New().String(),
		Name:        "Director",
		Description: "Executive dashboard for directors to view organization-wide metrics, team comparisons, and trends",
		Persona:     "director",
		IsTemplate:  true,
		IsSystem:    true,
		Layout:      string(layoutJSON),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}
