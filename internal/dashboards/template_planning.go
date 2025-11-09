package dashboards

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// createPlanningTemplate creates the Planning dashboard template
func createPlanningTemplate() (*models.Dashboard, error) {
	layout := models.DashboardLayout{
		Widgets: []models.Widget{
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "Sprint Velocity",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "pr_volume",
					"entity_type": "team",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 0, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "Active Goals",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "active_goals",
					"entity_type": "team",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 3, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "Goals At Risk",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "goals_at_risk",
					"entity_type": "team",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 6, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "number",
				Title:      "Team Capacity",
				DataSource: "/api/metrics/aggregate",
				QueryParams: map[string]string{
					"metric":      "team_capacity",
					"entity_type": "team",
					"period":      "week",
				},
				Position: models.WidgetPosition{X: 9, Y: 0, W: 3, H: 2},
			},
			{
				ID:         uuid.New().String(),
				Type:       "timeseries",
				Title:      "Velocity Trend",
				DataSource: "/api/metrics/timeseries",
				QueryParams: map[string]string{
					"metric":      "pr_volume",
					"entity_type": "team",
					"period":      "week",
					"count":       "12",
				},
				VisualizationConfig: models.VisualizationConfig{
					PrimaryColor: "#10b981",
					ShowLegend:   true,
					YAxisLabel:   "PRs per Week",
				},
				Position: models.WidgetPosition{X: 0, Y: 2, W: 6, H: 4},
			},
			{
				ID:         uuid.New().String(),
				Type:       "bar",
				Title:      "Goal Progress",
				DataSource: "/api/metrics/compare",
				QueryParams: map[string]string{
					"metric":   "goal_progress",
					"group_by": "goal",
					"period":   "week",
				},
				VisualizationConfig: models.VisualizationConfig{
					PrimaryColor: "#3b82f6",
				},
				Position: models.WidgetPosition{X: 6, Y: 2, W: 6, H: 4},
			},
			{
				ID:         uuid.New().String(),
				Type:       "table",
				Title:      "Sprint Burndown",
				DataSource: "/api/planning/sprint-burndown",
				VisualizationConfig: models.VisualizationConfig{
					Columns: []models.TableColumn{
						{Key: "date", Label: "Date", Sortable: true, Format: "date"},
						{Key: "remaining", Label: "Remaining", Sortable: true, Format: "number"},
						{Key: "completed", Label: "Completed", Sortable: true, Format: "number"},
						{Key: "ideal", Label: "Ideal", Sortable: false, Format: "number"},
					},
				},
				Position: models.WidgetPosition{X: 0, Y: 6, W: 12, H: 4},
			},
		},
	}

	layoutJSON, err := json.Marshal(layout)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Planning template layout: %w", err)
	}

	return &models.Dashboard{
		ID:          uuid.New().String(),
		Name:        "Planning",
		Description: "Dashboard for sprint planning, goal tracking, and capacity management",
		Persona:     "planning",
		IsTemplate:  true,
		IsSystem:    true,
		Layout:      string(layoutJSON),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}
