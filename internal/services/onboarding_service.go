package services

import (
	"fmt"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// OnboardingService handles user onboarding workflow
type OnboardingService struct {
	settingsStore  *db.SettingsStore
	templateStore  *db.DashboardTemplateStore
	dashboardStore *db.DashboardStore
}

// NewOnboardingService creates a new onboarding service
func NewOnboardingService(
	settingsStore *db.SettingsStore,
	templateStore *db.DashboardTemplateStore,
	dashboardStore *db.DashboardStore,
) *OnboardingService {
	return &OnboardingService{
		settingsStore:  settingsStore,
		templateStore:  templateStore,
		dashboardStore: dashboardStore,
	}
}

// OnboardingStatus represents the current onboarding state
type OnboardingStatus struct {
	Completed     bool   `json:"completed"`
	Role          string `json:"role"`
	HasDashboards bool   `json:"has_dashboards"`
}

// GetOnboardingStatus returns the current onboarding status
func (s *OnboardingService) GetOnboardingStatus() (*OnboardingStatus, error) {
	settings, err := s.settingsStore.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// Check if user has any non-template dashboards (user-created dashboards only)
	// Get all dashboards and filter out templates
	allDashboards, _, err := s.dashboardStore.ListDashboards("", false, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list dashboards: %w", err)
	}

	// Count non-template dashboards
	userDashboardCount := 0
	for _, d := range allDashboards {
		if !d.IsTemplate {
			userDashboardCount++
		}
	}

	return &OnboardingStatus{
		Completed:     settings.OnboardingCompleted,
		Role:          settings.Role,
		HasDashboards: userDashboardCount > 0,
	}, nil
}

// SetUserRole sets the user's role and creates default dashboards from templates
func (s *OnboardingService) SetUserRole(role string) error {
	// Validate role
	validRoles := map[string]bool{
		"ic":       true,
		"manager":  true,
		"director": true,
		"admin":    true,
	}
	if !validRoles[role] {
		return fmt.Errorf("invalid role: %s", role)
	}

	// Update role in settings
	if err := s.settingsStore.UpdateRole(role); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	// Get templates for this role
	templates, err := s.templateStore.ListTemplates(role)
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	if len(templates) == 0 {
		return fmt.Errorf("no templates found for role: %s", role)
	}

	// Create dashboards from templates
	var primaryDashboardID *string
	for i, template := range templates {
		dashboard := &models.Dashboard{
			Name:        template.Name,
			Description: template.Description,
			Persona:     template.Role,
			IsTemplate:  false,
			IsSystem:    false,
			Layout:      template.Layout,
			Filters:     "", // Empty filters for new dashboards
		}

		if err := s.dashboardStore.CreateDashboard(dashboard); err != nil {
			return fmt.Errorf("failed to create dashboard from template %s: %w", template.Name, err)
		}

		// Set the first dashboard as primary
		if i == 0 {
			primaryDashboardID = &dashboard.ID
		}
	}

	// Set primary dashboard
	if primaryDashboardID != nil {
		if err := s.settingsStore.UpdatePrimaryDashboard(primaryDashboardID); err != nil {
			return fmt.Errorf("failed to set primary dashboard: %w", err)
		}
	}

	return nil
}

// CompleteOnboarding marks onboarding as completed
func (s *OnboardingService) CompleteOnboarding() error {
	if err := s.settingsStore.UpdateOnboardingCompleted(true); err != nil {
		return fmt.Errorf("failed to mark onboarding complete: %w", err)
	}
	return nil
}
