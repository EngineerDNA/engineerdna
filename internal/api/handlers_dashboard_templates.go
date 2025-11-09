package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/engineerdna/engineerdna/internal/models"
)

// Dashboard Template Handlers

// handleDashboardTemplates handles GET /api/dashboard-templates
func (s *Server) handleDashboardTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role := r.URL.Query().Get("role")

	// Validate role if provided
	if role != "" {
		validRoles := map[string]bool{
			"ic":       true,
			"manager":  true,
			"director": true,
			"admin":    true,
		}
		if !validRoles[role] {
			respondError(w, http.StatusBadRequest, "Invalid role parameter", nil)
			return
		}
	}

	templates, err := s.templateStore.ListTemplates(role)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list templates", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
	})
}

// handleDashboardTemplateByID handles GET /api/dashboard-templates/:id
func (s *Server) handleDashboardTemplateByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondError(w, http.StatusBadRequest, "Invalid path", nil)
		return
	}
	id := parts[3]

	template, err := s.templateStore.GetTemplate(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get template", err)
		return
	}
	if template == nil {
		respondError(w, http.StatusNotFound, "Template not found", nil)
		return
	}

	respondJSON(w, http.StatusOK, template)
}

// handleCreateDashboardFromTemplate handles POST /api/dashboards/from-template/:templateId
func (s *Server) handleCreateDashboardFromTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract template ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondError(w, http.StatusBadRequest, "Invalid path", nil)
		return
	}
	templateID := parts[4]

	// Get template
	template, err := s.templateStore.GetTemplate(templateID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get template", err)
		return
	}
	if template == nil {
		respondError(w, http.StatusNotFound, "Template not found", nil)
		return
	}

	// Create dashboard from template
	dashboard := &models.Dashboard{
		Name:        template.Name,
		Description: template.Description,
		Persona:     template.Role,
		IsTemplate:  false,
		IsSystem:    false,
		Layout:      template.Layout,
		Filters:     "",
	}

	if err := s.dashboardStore.CreateDashboard(dashboard); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create dashboard", err)
		return
	}

	respondJSON(w, http.StatusCreated, dashboard)
}

// handleSettingsRole handles GET and PUT /api/settings/role
func (s *Server) handleSettingsRole(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getSettingsRole(w, r)
	case http.MethodPut:
		s.updateSettingsRole(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getSettingsRole(w http.ResponseWriter, r *http.Request) {
	settings, err := s.settingsStore.GetSettings()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get settings", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"role": settings.Role,
	})
}

func (s *Server) updateSettingsRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Role string `json:"role"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate role
	validRoles := map[string]bool{
		"ic":       true,
		"manager":  true,
		"director": true,
		"admin":    true,
	}
	if !validRoles[req.Role] {
		respondError(w, http.StatusBadRequest, "Invalid role. Must be one of: ic, manager, director, admin", nil)
		return
	}

	// Use onboarding service to set role and create dashboards
	if s.onboardingService != nil {
		if err := s.onboardingService.SetUserRole(req.Role); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to set role", err)
			return
		}
	} else {
		// Fallback to just updating role if onboarding service not available
		if err := s.settingsStore.UpdateRole(req.Role); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to update role", err)
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"role":    req.Role,
	})
}

// handleSettingsPrimaryDashboard handles GET and PUT /api/settings/primary-dashboard
func (s *Server) handleSettingsPrimaryDashboard(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getSettingsPrimaryDashboard(w, r)
	case http.MethodPut:
		s.updateSettingsPrimaryDashboard(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getSettingsPrimaryDashboard(w http.ResponseWriter, r *http.Request) {
	settings, err := s.settingsStore.GetSettings()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get settings", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"primary_dashboard_id": settings.PrimaryDashboardID,
	})
}

func (s *Server) updateSettingsPrimaryDashboard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DashboardID *string `json:"dashboard_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Verify dashboard exists if provided
	if req.DashboardID != nil && *req.DashboardID != "" {
		dashboard, err := s.dashboardStore.GetDashboard(*req.DashboardID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to verify dashboard", err)
			return
		}
		if dashboard == nil {
			respondError(w, http.StatusNotFound, "Dashboard not found", nil)
			return
		}
	}

	if err := s.settingsStore.UpdatePrimaryDashboard(req.DashboardID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update primary dashboard", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":              true,
		"primary_dashboard_id": req.DashboardID,
	})
}
