package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Cost Configuration Handlers

func (s *Server) handleCostConfiguration(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listCostConfiguration(w, r)
	case http.MethodPut:
		s.updateCostConfiguration(w, r)
	case http.MethodPost:
		s.createCostConfiguration(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listCostConfiguration(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")

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

	// Get cost attributes from entity_attributes (PDR-9 schema)
	filters := make(map[string]interface{})
	if entityType != "" {
		filters["entity_type"] = entityType
	}
	filters["attribute_name"] = "monthly_cost"
	filters["as_of"] = time.Now().UTC()

	attrs, err := s.attributeStore.List(filters, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list cost attributes", err)
		return
	}

	// Return entity attributes directly
	respondPaginated(w, attrs, len(attrs), limit, offset)
}

func (s *Server) createCostConfiguration(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EntityType    string  `json:"entity_type"`
		EntityID      string  `json:"entity_id"`
		MonthlyCost   float64 `json:"monthly_cost"`
		Currency      string  `json:"currency"`
		EffectiveFrom string  `json:"effective_from"`
		EffectiveTo   *string `json:"effective_to"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.EntityType == "" || req.EntityID == "" || req.MonthlyCost <= 0 {
		respondError(w, http.StatusBadRequest, "entity_type, entity_id, and monthly_cost are required", nil)
		return
	}

	effectiveFrom, err := time.Parse("2006-01-02", req.EffectiveFrom)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid effective_from format (expected YYYY-MM-DD)", err)
		return
	}

	var validUntil *time.Time
	if req.EffectiveTo != nil {
		t, err := time.Parse("2006-01-02", *req.EffectiveTo)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid effective_to format", err)
			return
		}
		validUntil = &t
	}

	// Create entity attribute directly (PDR-9 schema)
	attr := &models.EntityAttribute{
		EntityType:    req.EntityType,
		EntityID:      req.EntityID,
		AttributeName: "monthly_cost",
		Value:         fmt.Sprintf("%.2f", req.MonthlyCost),
		ValueType:     "currency",
		ValidFrom:     effectiveFrom.UTC(),
		ValidUntil:    validUntil,
		Source:        "user_input",
	}

	if err := s.attributeStore.Create(attr); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create cost attribute", err)
		return
	}

	// Return entity attribute directly
	respondJSON(w, http.StatusCreated, attr)
}

func (s *Server) updateCostConfiguration(w http.ResponseWriter, r *http.Request) {
	// Similar to create but updates existing
	respondError(w, http.StatusNotImplemented, "Update not implemented", nil)
}

func (s *Server) handleTeamCost(w http.ResponseWriter, r *http.Request) {
	// Extract team ID from path
	teamID := strings.TrimPrefix(r.URL.Path, "/api/cost/team/")
	if teamID == "" {
		http.Error(w, "Team ID required", http.StatusBadRequest)
		return
	}

	// Get team cost from entity_attributes table (PDR-9 migration)
	// This could be a direct team cost attribute OR sum of member costs
	costAttr, err := s.attributeStore.GetCurrentAttributeValue("team", teamID, "monthly_cost")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get team cost", err)
		return
	}

	var monthlyCost float64
	if costAttr != nil {
		// Team has a direct cost attribute
		if costStr, ok := costAttr.(string); ok {
			fmt.Sscanf(costStr, "%f", &monthlyCost)
		}
	} else {
		// Fall back to calculating from service (which might sum engineer costs)
		cost, err := s.costService.CalculateTeamCost(teamID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to calculate team cost", err)
			return
		}
		monthlyCost = cost
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"team_id":      teamID,
		"monthly_cost": monthlyCost,
		"currency":     "USD",
	})
}

func (s *Server) handleEngineerCost(w http.ResponseWriter, r *http.Request) {
	engineerID := strings.TrimPrefix(r.URL.Path, "/api/cost/engineer/")
	if engineerID == "" {
		http.Error(w, "Engineer ID required", http.StatusBadRequest)
		return
	}

	// Get cost from entity_attributes (PDR-9 schema)
	costAttr, err := s.attributeStore.GetEngineerCost(engineerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get engineer cost", err)
		return
	}

	if costAttr == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"entity_type":    "engineer",
			"entity_id":      engineerID,
			"attribute_name": "monthly_cost",
			"attributes":     []interface{}{},
		})
		return
	}

	// Return entity attribute directly
	respondJSON(w, http.StatusOK, costAttr)
}

// Feature Value Handlers

func (s *Server) handleFeatures(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listFeatures(w, r)
	case http.MethodPost:
		s.createFeature(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listFeatures(w http.ResponseWriter, r *http.Request) {
	timePeriod := r.URL.Query().Get("time_period")

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

	features, total, err := s.costROIStore.ListFeatureValues(timePeriod, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list features", err)
		return
	}

	respondPaginated(w, features, total, limit, offset)
}

func (s *Server) createFeature(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FeatureName        string   `json:"feature_name"`
		FeatureDescription string   `json:"feature_description"`
		ValueType          string   `json:"value_type"`
		ValueAmount        *float64 `json:"value_amount"`
		ValueCurrency      string   `json:"value_currency"`
		ConfidenceLevel    string   `json:"confidence_level"`
		Source             string   `json:"source"`
		SourceReference    string   `json:"source_reference"`
		TimePeriod         string   `json:"time_period"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.FeatureName == "" || req.ValueType == "" {
		respondError(w, http.StatusBadRequest, "feature_name and value_type are required", nil)
		return
	}

	feature := &models.FeatureValue{
		FeatureName:        req.FeatureName,
		FeatureDescription: req.FeatureDescription,
		ValueType:          req.ValueType,
		ValueAmount:        req.ValueAmount,
		ValueCurrency:      req.ValueCurrency,
		ConfidenceLevel:    req.ConfidenceLevel,
		Source:             req.Source,
		SourceReference:    req.SourceReference,
		TimePeriod:         req.TimePeriod,
	}

	if feature.ValueCurrency == "" {
		feature.ValueCurrency = "USD"
	}
	if feature.ConfidenceLevel == "" {
		feature.ConfidenceLevel = "estimated"
	}

	if err := s.costROIStore.CreateFeatureValue(feature); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create feature", err)
		return
	}

	respondJSON(w, http.StatusCreated, feature)
}

func (s *Server) handleFeatureByID(w http.ResponseWriter, r *http.Request) {
	// Extract feature ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/features/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Feature ID required", http.StatusBadRequest)
		return
	}

	featureID := parts[0]

	// Route based on operation
	if len(parts) > 1 {
		operation := parts[1]
		switch operation {
		case "cost":
			s.handleFeatureCost(w, r, featureID)
		case "work-items":
			s.handleFeatureWorkItems(w, r, featureID)
		case "roi":
			s.handleFeatureROI(w, r, featureID)
		default:
			http.Error(w, "Unknown operation", http.StatusBadRequest)
		}
	} else {
		// Single feature operations
		switch r.Method {
		case http.MethodGet:
			s.getFeature(w, r, featureID)
		case http.MethodPut:
			s.updateFeature(w, r, featureID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (s *Server) getFeature(w http.ResponseWriter, r *http.Request, featureID string) {
	feature, err := s.costROIStore.GetFeatureValue(featureID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get feature", err)
		return
	}
	if feature == nil {
		http.Error(w, "Feature not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, feature)
}

func (s *Server) updateFeature(w http.ResponseWriter, r *http.Request, featureID string) {
	feature, err := s.costROIStore.GetFeatureValue(featureID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get feature", err)
		return
	}
	if feature == nil {
		http.Error(w, "Feature not found", http.StatusNotFound)
		return
	}

	var req struct {
		ValueAmount     *float64 `json:"value_amount"`
		ConfidenceLevel *string  `json:"confidence_level"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.ValueAmount != nil {
		feature.ValueAmount = req.ValueAmount
	}
	if req.ConfidenceLevel != nil {
		feature.ConfidenceLevel = *req.ConfidenceLevel
	}

	if err := s.costROIStore.UpdateFeatureValue(feature); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update feature", err)
		return
	}

	respondJSON(w, http.StatusOK, feature)
}

func (s *Server) handleFeatureCost(w http.ResponseWriter, r *http.Request, featureID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get feature
	feature, err := s.costROIStore.GetFeatureValue(featureID)
	if err != nil || feature == nil {
		respondError(w, http.StatusNotFound, "Feature not found", err)
		return
	}

	// Calculate cost
	cost, err := s.costService.CalculateFeatureCost(feature)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to calculate feature cost", err)
		return
	}

	respondJSON(w, http.StatusOK, cost)
}

func (s *Server) handleFeatureWorkItems(w http.ResponseWriter, r *http.Request, featureID string) {
	switch r.Method {
	case http.MethodGet:
		workItems, err := s.costROIStore.GetFeatureWorkItems(featureID, 100)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to get work items", err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"work_items": workItems})

	case http.MethodPost:
		var req struct {
			WorkItemType string   `json:"work_item_type"`
			WorkItemID   string   `json:"work_item_id"`
			StoryPoints  *int     `json:"story_points"`
			ActualHours  *float64 `json:"actual_hours"`
			EngineerID   *string  `json:"engineer_id"`
			TeamID       *string  `json:"team_id"`
		}

		if err := decodeAndValidateJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		if req.WorkItemType == "" || req.WorkItemID == "" {
			respondError(w, http.StatusBadRequest, "work_item_type and work_item_id are required", nil)
			return
		}

		item := &models.FeatureWorkItem{
			FeatureID:    featureID,
			WorkItemType: req.WorkItemType,
			WorkItemID:   req.WorkItemID,
			StoryPoints:  req.StoryPoints,
			ActualHours:  req.ActualHours,
			EngineerID:   req.EngineerID,
			TeamID:       req.TeamID,
		}

		if err := s.costROIStore.CreateFeatureWorkItem(item); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create work item", err)
			return
		}

		respondJSON(w, http.StatusCreated, item)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFeatureROI(w http.ResponseWriter, r *http.Request, featureID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Calculate ROI
	roi, err := s.roiService.CalculateFeatureROI(featureID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to calculate ROI", err)
		return
	}

	respondJSON(w, http.StatusOK, roi)
}

// ROI Report Handler

func (s *Server) handleROIReport(w http.ResponseWriter, r *http.Request) {
	timePeriod := r.URL.Query().Get("time_period")
	if timePeriod == "" {
		respondError(w, http.StatusBadRequest, "time_period parameter required", nil)
		return
	}

	summary, err := s.roiService.GenerateROIReport(timePeriod)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate ROI report", err)
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

// Investment Breakdown Handler

func (s *Server) handleInvestmentBreakdown(w http.ResponseWriter, r *http.Request) {
	timePeriod := r.URL.Query().Get("time_period")
	if timePeriod == "" {
		respondError(w, http.StatusBadRequest, "time_period parameter required", nil)
		return
	}

	teamID := r.URL.Query().Get("team_id")
	var teamIDPtr *string
	if teamID != "" {
		teamIDPtr = &teamID
	}

	breakdown, err := s.costAnalyzer.AnalyzeTimeAllocation(timePeriod, teamIDPtr)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to analyze investment", err)
		return
	}

	respondJSON(w, http.StatusOK, breakdown)
}
