package dashboards

import (
	"fmt"
	"log"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// CreateDefaultTemplates creates the 4 default dashboard templates
func CreateDefaultTemplates(store *db.DashboardStore) error {
	// Check if templates already exist
	existing, _, err := store.ListDashboards("", true, 10, 0)
	if err != nil {
		return fmt.Errorf("failed to check existing templates: %w", err)
	}

	if len(existing) > 0 {
		log.Printf("Dashboard templates already exist (%d found), skipping creation", len(existing))
		return nil
	}

	log.Println("Creating default dashboard templates...")

	icTemplate, err := createICTemplate()
	if err != nil {
		return fmt.Errorf("failed to generate IC template: %w", err)
	}

	teamLeadTemplate, err := createTeamLeadTemplate()
	if err != nil {
		return fmt.Errorf("failed to generate Team Lead template: %w", err)
	}

	directorTemplate, err := createDirectorTemplate()
	if err != nil {
		return fmt.Errorf("failed to generate Director template: %w", err)
	}

	planningTemplate, err := createPlanningTemplate()
	if err != nil {
		return fmt.Errorf("failed to generate Planning template: %w", err)
	}

	templates := []*models.Dashboard{
		icTemplate,
		teamLeadTemplate,
		directorTemplate,
		planningTemplate,
	}

	for _, template := range templates {
		if err := store.CreateDashboard(template); err != nil {
			return fmt.Errorf("failed to create template %s: %w", template.Name, err)
		}
		log.Printf("Created template: %s", template.Name)
	}

	log.Printf("Successfully created %d dashboard templates", len(templates))
	return nil
}
