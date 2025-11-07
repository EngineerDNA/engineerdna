package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Dashboard Handlers

// parseDashboardLayout parses layout JSON and returns widgets array
func parseDashboardLayout(layoutJSON string) ([]models.Widget, error) {
	if layoutJSON == "" {
		return []models.Widget{}, nil
	}

	var layout models.DashboardLayout
	if err := json.Unmarshal([]byte(layoutJSON), &layout); err != nil {
		return nil, err
	}
	return layout.Widgets, nil
}

// enrichDashboardResponse adds widgets field to dashboard response
func enrichDashboardResponse(dashboard *models.Dashboard) map[string]interface{} {
	widgets, err := parseDashboardLayout(dashboard.Layout)
	if err != nil {
		widgets = []models.Widget{}
	}

	return map[string]interface{}{
		"id":          dashboard.ID,
		"name":        dashboard.Name,
		"description": dashboard.Description,
		"persona":     dashboard.Persona,
		"is_template": dashboard.IsTemplate,
		"is_system":   dashboard.IsSystem,
		"layout":      dashboard.Layout,
		"widgets":     widgets,
		"filters":     dashboard.Filters,
		"created_at":  dashboard.CreatedAt,
		"updated_at":  dashboard.UpdatedAt,
	}
}

// handleDashboards handles GET (list) and POST (create) for dashboards
func (s *Server) handleDashboards(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listDashboards(w, r)
	case http.MethodPost:
		s.createDashboard(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listDashboards handles GET /api/dashboards - lists dashboards with filtering and pagination
func (s *Server) listDashboards(w http.ResponseWriter, r *http.Request) {
	persona := r.URL.Query().Get("persona")
	templatesOnly := r.URL.Query().Get("templates") == "true"

	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = db.DefaultQueryLimit
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	dashboards, total, err := s.dashboardStore.ListDashboards(persona, templatesOnly, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list dashboards", err)
		return
	}

	// Enrich dashboards with widgets
	enrichedDashboards := make([]interface{}, len(dashboards))
	for i, dashboard := range dashboards {
		enrichedDashboards[i] = enrichDashboardResponse(dashboard)
	}

	hasMore := offset+limit < total
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"dashboards": enrichedDashboards,
		"pagination": map[string]interface{}{
			"total":   total,
			"limit":   limit,
			"offset":  offset,
			"hasMore": hasMore,
		},
	})
}

// createDashboard handles POST /api/dashboards - creates new dashboard with validation
func (s *Server) createDashboard(w http.ResponseWriter, r *http.Request) {
	var dashboard models.Dashboard

	if err := decodeAndValidateJSON(r, &dashboard); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if dashboard.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required", nil)
		return
	}

	if dashboard.Layout == "" {
		respondError(w, http.StatusBadRequest, "layout is required", nil)
		return
	}

	// Validate layout is valid JSON with correct structure
	var layoutCheck models.DashboardLayout
	if err := json.Unmarshal([]byte(dashboard.Layout), &layoutCheck); err != nil {
		respondError(w, http.StatusBadRequest, "layout must be valid JSON", err)
		return
	}

	if err := s.dashboardStore.CreateDashboard(&dashboard); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create dashboard", err)
		return
	}

	respondJSON(w, http.StatusCreated, enrichDashboardResponse(&dashboard))
}

// handleDashboardByID handles GET (retrieve), PUT (update), DELETE (delete) for a specific dashboard
func (s *Server) handleDashboardByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/dashboards/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Dashboard ID required", http.StatusBadRequest)
		return
	}

	dashboardID := parts[0]

	// Check for clone operation
	if len(parts) == 2 && parts[1] == "clone" && r.Method == http.MethodPost {
		s.cloneDashboard(w, r, dashboardID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getDashboard(w, r, dashboardID)
	case http.MethodPut:
		s.updateDashboard(w, r, dashboardID)
	case http.MethodDelete:
		s.deleteDashboard(w, r, dashboardID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getDashboard handles GET /api/dashboards/:id - retrieves single dashboard by ID
func (s *Server) getDashboard(w http.ResponseWriter, r *http.Request, dashboardID string) {
	dashboard, err := s.dashboardStore.GetDashboard(dashboardID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get dashboard", err)
		return
	}

	if dashboard == nil {
		http.Error(w, "Dashboard not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, enrichDashboardResponse(dashboard))
}

// updateDashboard handles PUT /api/dashboards/:id - updates dashboard layout and settings
func (s *Server) updateDashboard(w http.ResponseWriter, r *http.Request, dashboardID string) {
	var dashboard models.Dashboard

	if err := decodeAndValidateJSON(r, &dashboard); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	dashboard.ID = dashboardID

	if dashboard.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required", nil)
		return
	}

	if dashboard.Layout == "" {
		respondError(w, http.StatusBadRequest, "layout is required", nil)
		return
	}

	// Validate layout is valid JSON with correct structure
	var layoutCheck models.DashboardLayout
	if err := json.Unmarshal([]byte(dashboard.Layout), &layoutCheck); err != nil {
		respondError(w, http.StatusBadRequest, "layout must be valid JSON", err)
		return
	}

	if err := s.dashboardStore.UpdateDashboard(&dashboard); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Dashboard not found", http.StatusNotFound)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update dashboard", err)
		return
	}

	// Fetch updated dashboard
	updated, err := s.dashboardStore.GetDashboard(dashboardID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated dashboard", err)
		return
	}

	respondJSON(w, http.StatusOK, enrichDashboardResponse(updated))
}

// deleteDashboard handles DELETE /api/dashboards/:id - deletes dashboard with system protection
func (s *Server) deleteDashboard(w http.ResponseWriter, r *http.Request, dashboardID string) {
	if err := s.dashboardStore.DeleteDashboard(dashboardID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Dashboard not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "system dashboard") {
			respondError(w, http.StatusForbidden, "Cannot delete system dashboard", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete dashboard", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// cloneDashboard handles POST /api/dashboards/:id/clone to duplicate a dashboard
func (s *Server) cloneDashboard(w http.ResponseWriter, r *http.Request, sourceID string) {
	var req struct {
		Name string `json:"name"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required", nil)
		return
	}

	cloned, err := s.dashboardStore.CloneDashboard(sourceID, req.Name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Source dashboard not found", http.StatusNotFound)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to clone dashboard", err)
		return
	}

	respondJSON(w, http.StatusCreated, enrichDashboardResponse(cloned))
}

// Metric Snapshot Handlers

// handleMetricSnapshots handles GET /api/metrics/snapshots to query pre-computed metric data
func (s *Server) handleMetricSnapshots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metricName := r.URL.Query().Get("metric")
	entityType := r.URL.Query().Get("entity_type")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if metricName == "" {
		respondError(w, http.StatusBadRequest, "metric parameter is required", nil)
		return
	}

	if entityType == "" {
		respondError(w, http.StatusBadRequest, "entity_type parameter is required", nil)
		return
	}

	if startDate == "" {
		respondError(w, http.StatusBadRequest, "start_date parameter is required", nil)
		return
	}

	if endDate == "" {
		respondError(w, http.StatusBadRequest, "end_date parameter is required", nil)
		return
	}

	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = db.DefaultQueryLimit
	}

	snapshots, err := s.dashboardStore.GetMetricSnapshots(metricName, entityType, startDate, endDate, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get metric snapshots", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"snapshots": snapshots,
		"count":     len(snapshots),
	})
}
