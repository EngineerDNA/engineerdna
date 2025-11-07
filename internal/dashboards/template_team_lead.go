package dashboards

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// createTeamLeadTemplate creates the Team Lead dashboard template
func createTeamLeadTemplate() (*models.Dashboard, error) {
	layout := models.DashboardLayout{
		Widgets: []models.Widget{
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "Team Score",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "team_score",
					"entity_type": "team",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 0, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "Team PRs",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "pr_volume",
					"entity_type": "team",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 3, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "Active Alerts",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "active_alerts",
					"entity_type": "team",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 6, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "Goals On Track",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "active_goals",
					"entity_type": "team",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 9, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "timeseries",
				Title:      "Team Performance Trend",
				DataSource: "/api/metrics/timeseries",
				QueryParams: map[string]string{
					"metric":      "team_score",
					"entity_type": "team",
					"period":      "week",
					"count":       "12",
				},
				VisualizationConfig: models.VisualizationConfig{
					PrimaryColor: "#3b82f6",
					ShowLegend:   true,
					YAxisLabel:   "Team Score",
				},
				Position: models.WidgetPosition{X: 0, Y: 2, W: 8, H: 4},
			},
			{
				ID:         uuid.New().String(),
				Type:       "bar",
				Title:      "Engineer Performance",
				DataSource: "/api/metrics/compare",
				QueryParams: map[string]string{
					"metric":   "engineer_score",
					"group_by": "engineer",
					"period":   "week",
				},
				VisualizationConfig: models.VisualizationConfig{
					PrimaryColor: "#10b981",
				},
				Position: models.WidgetPosition{X: 8, Y: 2, W: 4, H: 4},
			},
			{
				ID:         uuid.New().String(),
				Type:       "table",
				Title:      "Team Members",
				DataSource: "/api/engineers/performance",
				QueryParams: map[string]string{
					"period":  "week",
					"sort_by": "score",
				},
				VisualizationConfig: models.VisualizationConfig{
					Columns: []models.TableColumn{
						{Key: "name", Label: "Name", Sortable: true},
						{Key: "score", Label: "Score", Sortable: true, Format: "number"},
						{Key: "pr_count", Label: "PRs", Sortable: true, Format: "number"},
					},
				},
				Position: models.WidgetPosition{X: 0, Y: 6, W: 12, H: 4},
			},
		},
	}

	layoutJSON, err := json.Marshal(layout)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Team Lead template layout: %w", err)
	}

	return &models.Dashboard{
		ID:          uuid.New().String(),
		Name:        "Team Lead",
		Description: "Dashboard for team leads to monitor team performance, individual contributors, and goals",
		Persona:     "team_lead",
		IsTemplate:  true,
		IsSystem:    true,
		Layout:      string(layoutJSON),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}
