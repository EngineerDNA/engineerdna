package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/internal/scheduler"
	"github.com/google/uuid"
)

func (s *Server) handleExportScheduleOperations(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	runPattern := regexp.MustCompile(`^/api/exports/schedules/([^/]+)/run$`)
	if runPattern.MatchString(path) {
		s.handleExportScheduleRun(w, r)
		return
	}

	idPattern := regexp.MustCompile(`^/api/exports/schedules/([^/]+)$`)
	if idPattern.MatchString(path) {
		s.handleExportScheduleByID(w, r)
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func (s *Server) handleExportSchedules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listExportSchedules(w, r)
	case http.MethodPost:
		s.createExportSchedule(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleExportScheduleByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/exports/schedules/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Schedule ID required", http.StatusBadRequest)
		return
	}

	id := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.getExportSchedule(w, r, id)
	case http.MethodPut:
		s.updateExportSchedule(w, r, id)
	case http.MethodDelete:
		s.deleteExportSchedule(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleExportScheduleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pattern := regexp.MustCompile(`^/api/exports/schedules/([^/]+)/run$`)
	matches := pattern.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	id := matches[1]
	s.runExportScheduleNow(w, r, id)
}

func (s *Server) listExportSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := s.scheduleStore.List(1000)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list export schedules", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"schedules": schedules,
	})
}

func (s *Server) createExportSchedule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PluginName string `json:"plugin_name"`
		Frequency  string `json:"frequency"`
		DayOfWeek  *int   `json:"day_of_week,omitempty"`
		TimeOfDay  string `json:"time_of_day"`
		Enabled    *bool  `json:"enabled,omitempty"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.PluginName == "" || req.Frequency == "" || req.TimeOfDay == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields: plugin_name, frequency, time_of_day", nil)
		return
	}

	if !isValidFrequency(req.Frequency) {
		respondError(w, http.StatusBadRequest, "Invalid frequency: must be 'daily', 'weekly', or 'monthly'", nil)
		return
	}

	if req.Frequency == "weekly" && req.DayOfWeek == nil {
		respondError(w, http.StatusBadRequest, "day_of_week is required for weekly schedules", nil)
		return
	}

	if req.DayOfWeek != nil && (*req.DayOfWeek < 0 || *req.DayOfWeek > 6) {
		respondError(w, http.StatusBadRequest, "day_of_week must be between 0 (Sunday) and 6 (Saturday)", nil)
		return
	}

	if !isValidTimeOfDay(req.TimeOfDay) {
		respondError(w, http.StatusBadRequest, "Invalid time_of_day: must be in HH:MM format (24-hour)", nil)
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	now := time.Now().UTC()
	nextRun := scheduler.CalculateNextRun(req.Frequency, req.DayOfWeek, req.TimeOfDay, now)

	schedule := &models.ExportSchedule{
		ID:         uuid.New().String(),
		PluginName: req.PluginName,
		Frequency:  req.Frequency,
		DayOfWeek:  req.DayOfWeek,
		TimeOfDay:  req.TimeOfDay,
		Enabled:    enabled,
		NextRun:    nextRun,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.scheduleStore.Create(schedule); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create export schedule", err)
		return
	}

	respondJSON(w, http.StatusCreated, schedule)
}

func (s *Server) getExportSchedule(w http.ResponseWriter, r *http.Request, id string) {
	schedule, err := s.scheduleStore.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Export schedule not found", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get export schedule", err)
		}
		return
	}

	respondJSON(w, http.StatusOK, schedule)
}

func (s *Server) updateExportSchedule(w http.ResponseWriter, r *http.Request, id string) {
	schedule, err := s.scheduleStore.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Export schedule not found", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get export schedule", err)
		}
		return
	}

	var req struct {
		PluginName *string `json:"plugin_name,omitempty"`
		Frequency  *string `json:"frequency,omitempty"`
		DayOfWeek  *int    `json:"day_of_week,omitempty"`
		TimeOfDay  *string `json:"time_of_day,omitempty"`
		Enabled    *bool   `json:"enabled,omitempty"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	updated := false

	if req.PluginName != nil {
		schedule.PluginName = *req.PluginName
		updated = true
	}

	if req.Frequency != nil {
		if !isValidFrequency(*req.Frequency) {
			respondError(w, http.StatusBadRequest, "Invalid frequency: must be 'daily', 'weekly', or 'monthly'", nil)
			return
		}
		schedule.Frequency = *req.Frequency
		updated = true
	}

	if req.DayOfWeek != nil {
		if *req.DayOfWeek < 0 || *req.DayOfWeek > 6 {
			respondError(w, http.StatusBadRequest, "day_of_week must be between 0 (Sunday) and 6 (Saturday)", nil)
			return
		}
		schedule.DayOfWeek = req.DayOfWeek
		updated = true
	}

	if req.TimeOfDay != nil {
		if !isValidTimeOfDay(*req.TimeOfDay) {
			respondError(w, http.StatusBadRequest, "Invalid time_of_day: must be in HH:MM format (24-hour)", nil)
			return
		}
		schedule.TimeOfDay = *req.TimeOfDay
		updated = true
	}

	if req.Enabled != nil {
		schedule.Enabled = *req.Enabled
		updated = true
	}

	if schedule.Frequency == "weekly" && schedule.DayOfWeek == nil {
		respondError(w, http.StatusBadRequest, "day_of_week is required for weekly schedules", nil)
		return
	}

	if updated {
		now := time.Now().UTC()
		schedule.NextRun = scheduler.CalculateNextRun(schedule.Frequency, schedule.DayOfWeek, schedule.TimeOfDay, now)
	}

	if err := s.scheduleStore.Update(schedule); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update export schedule", err)
		return
	}

	respondJSON(w, http.StatusOK, schedule)
}

func (s *Server) deleteExportSchedule(w http.ResponseWriter, r *http.Request, id string) {
	if err := s.scheduleStore.Delete(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Export schedule not found", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to delete export schedule", err)
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Export schedule deleted successfully",
	})
}

func (s *Server) runExportScheduleNow(w http.ResponseWriter, r *http.Request, id string) {
	schedule, err := s.scheduleStore.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Export schedule not found", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get export schedule", err)
		}
		return
	}

	if s.scheduler == nil {
		respondError(w, http.StatusInternalServerError, "Scheduler not initialized", nil)
		return
	}

	go func() {
		if err := s.scheduler.ExecuteSchedule(schedule); err != nil {
			fmt.Printf("Failed to execute schedule %s: %v\n", id, err)
		}
	}()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Export schedule triggered successfully",
	})
}

func isValidFrequency(frequency string) bool {
	return frequency == "daily" || frequency == "weekly" || frequency == "monthly"
}

func isValidTimeOfDay(timeOfDay string) bool {
	_, err := time.Parse("15:04", timeOfDay)
	return err == nil
}
