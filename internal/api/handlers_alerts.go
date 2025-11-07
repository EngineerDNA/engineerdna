package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

// Alert Rules Handlers

func (s *Server) handleAlertRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listAlertRules(w, r)
	case http.MethodPost:
		s.createAlertRule(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listAlertRules(w http.ResponseWriter, r *http.Request) {
	enabledOnly := r.URL.Query().Get("enabled") == "true"

	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = 100
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	rules, total, err := s.alertsStore.ListAlertRules(enabledOnly, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list alert rules", err)
		return
	}

	respondPaginated(w, rules, total, limit, offset)
}

func (s *Server) createAlertRule(w http.ResponseWriter, r *http.Request) {
	var rule models.AlertRule

	if err := decodeAndValidateJSON(r, &rule); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if rule.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required", nil)
		return
	}

	if rule.AlertType == "" {
		respondError(w, http.StatusBadRequest, "alert_type is required", nil)
		return
	}

	if rule.Severity == "" {
		respondError(w, http.StatusBadRequest, "severity is required", nil)
		return
	}

	if err := s.alertsStore.CreateAlertRule(&rule); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create alert rule", err)
		return
	}

	respondJSON(w, http.StatusCreated, rule)
}

func (s *Server) handleAlertRuleByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/alerts/rules/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Alert rule ID required", http.StatusBadRequest)
		return
	}

	ruleID := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.getAlertRule(w, r, ruleID)
	case http.MethodPut:
		s.updateAlertRule(w, r, ruleID)
	case http.MethodDelete:
		s.deleteAlertRule(w, r, ruleID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getAlertRule(w http.ResponseWriter, r *http.Request, ruleID string) {
	rule, err := s.alertsStore.GetAlertRule(ruleID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get alert rule", err)
		return
	}

	if rule == nil {
		http.Error(w, "Alert rule not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, rule)
}

func (s *Server) updateAlertRule(w http.ResponseWriter, r *http.Request, ruleID string) {
	var rule models.AlertRule

	if err := decodeAndValidateJSON(r, &rule); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	rule.ID = ruleID

	if err := s.alertsStore.UpdateAlertRule(&rule); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update alert rule", err)
		return
	}

	// Fetch updated rule
	updated, err := s.alertsStore.GetAlertRule(ruleID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated alert rule", err)
		return
	}

	respondJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteAlertRule(w http.ResponseWriter, r *http.Request, ruleID string) {
	if err := s.alertsStore.DeleteAlertRule(ruleID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete alert rule", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Alert Instances Handlers

func (s *Server) handleAlertInstances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	filters := make(map[string]string)

	if ruleID := query.Get("rule_id"); ruleID != "" {
		filters["rule_id"] = ruleID
	}
	if severity := query.Get("severity"); severity != "" {
		filters["severity"] = severity
	}
	if entityType := query.Get("entity_type"); entityType != "" {
		filters["entity_type"] = entityType
	}
	if entityID := query.Get("entity_id"); entityID != "" {
		filters["entity_id"] = entityID
	}
	if status := query.Get("status"); status != "" {
		filters["status"] = status
	}

	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = 100
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	instances, total, err := s.alertsStore.ListAlertInstances(filters, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list alert instances", err)
		return
	}

	respondPaginated(w, instances, total, limit, offset)
}

func (s *Server) handleAlertInstanceByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/alerts/instances/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Alert instance ID required", http.StatusBadRequest)
		return
	}

	instanceID := parts[0]

	if r.Method == http.MethodGet {
		s.getAlertInstance(w, r, instanceID)
		return
	}

	if len(parts) < 2 {
		http.Error(w, "Action required", http.StatusBadRequest)
		return
	}

	action := parts[1]

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch action {
	case "acknowledge":
		s.acknowledgeAlert(w, r, instanceID)
	case "snooze":
		s.snoozeAlert(w, r, instanceID)
	case "dismiss":
		s.dismissAlert(w, r, instanceID)
	case "resolve":
		s.resolveAlert(w, r, instanceID)
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}

func (s *Server) getAlertInstance(w http.ResponseWriter, r *http.Request, instanceID string) {
	instance, err := s.alertsStore.GetAlertInstance(instanceID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get alert instance", err)
		return
	}

	if instance == nil {
		http.Error(w, "Alert instance not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, instance)
}

func (s *Server) acknowledgeAlert(w http.ResponseWriter, r *http.Request, instanceID string) {
	var req struct {
		AcknowledgedBy string `json:"acknowledged_by"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.AcknowledgedBy == "" {
		req.AcknowledgedBy = "user"
	}

	if err := s.alertsStore.AcknowledgeAlert(instanceID, req.AcknowledgedBy); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to acknowledge alert", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) snoozeAlert(w http.ResponseWriter, r *http.Request, instanceID string) {
	var req struct {
		Duration int `json:"duration"` // Duration in minutes
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Duration <= 0 {
		req.Duration = 60 // Default 1 hour
	}

	snoozedUntil := time.Now().UTC().Add(time.Duration(req.Duration) * time.Minute)

	if err := s.alertsStore.SnoozeAlert(instanceID, snoozedUntil); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to snooze alert", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "ok",
		"snoozed_until": snoozedUntil,
	})
}

func (s *Server) dismissAlert(w http.ResponseWriter, r *http.Request, instanceID string) {
	var req struct {
		DismissedBy string `json:"dismissed_by"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.DismissedBy == "" {
		req.DismissedBy = "user"
	}

	if err := s.alertsStore.DismissAlert(instanceID, req.DismissedBy); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to dismiss alert", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) resolveAlert(w http.ResponseWriter, r *http.Request, instanceID string) {
	if err := s.alertsStore.ResolveAlert(instanceID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to resolve alert", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Alert Channels Handlers

func (s *Server) handleAlertChannels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listAlertChannels(w, r)
	case http.MethodPost:
		s.createAlertChannel(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listAlertChannels(w http.ResponseWriter, r *http.Request) {
	ruleID := r.URL.Query().Get("rule_id")
	if ruleID == "" {
		respondError(w, http.StatusBadRequest, "rule_id query parameter is required", nil)
		return
	}

	// GetAlertChannels returns all channels for a rule (no pagination needed for this specific case)
	// But we still use respondPaginated for consistency
	channels, err := s.alertsStore.GetAlertChannels(ruleID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list alert channels", err)
		return
	}

	total := len(channels)
	respondPaginated(w, channels, total, total, 0)
}

func (s *Server) createAlertChannel(w http.ResponseWriter, r *http.Request) {
	var channel models.AlertChannel

	if err := decodeAndValidateJSON(r, &channel); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if channel.RuleID == "" {
		respondError(w, http.StatusBadRequest, "rule_id is required", nil)
		return
	}

	if channel.ChannelType == "" {
		respondError(w, http.StatusBadRequest, "channel_type is required", nil)
		return
	}

	if err := s.alertsStore.CreateAlertChannel(&channel); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create alert channel", err)
		return
	}

	respondJSON(w, http.StatusCreated, channel)
}

func (s *Server) handleAlertChannelByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/alerts/channels/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Alert channel ID required", http.StatusBadRequest)
		return
	}

	channelID := parts[0]

	switch r.Method {
	case http.MethodPut:
		s.updateAlertChannel(w, r, channelID)
	case http.MethodDelete:
		s.deleteAlertChannel(w, r, channelID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) updateAlertChannel(w http.ResponseWriter, r *http.Request, channelID string) {
	var channel models.AlertChannel

	if err := decodeAndValidateJSON(r, &channel); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	channel.ID = channelID

	if err := s.alertsStore.UpdateAlertChannel(&channel); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update alert channel", err)
		return
	}

	respondJSON(w, http.StatusOK, channel)
}

func (s *Server) deleteAlertChannel(w http.ResponseWriter, r *http.Request, channelID string) {
	if err := s.alertsStore.DeleteAlertChannel(channelID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete alert channel", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
