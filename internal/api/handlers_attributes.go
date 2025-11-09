package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// handleAttributes handles GET and POST for entity attributes
func (s *Server) handleAttributes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleAttributesGet(w, r)
	case http.MethodPost:
		s.handleAttributesPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAttributesGet retrieves entity attributes with filters
// Query params: entity_type, entity_id, attribute_name, as_of, limit, offset
func (s *Server) handleAttributesGet(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filters := make(map[string]interface{})

	if entityType := r.URL.Query().Get("entity_type"); entityType != "" {
		filters["entity_type"] = entityType
	}

	if entityID := r.URL.Query().Get("entity_id"); entityID != "" {
		filters["entity_id"] = entityID
	}

	if attributeName := r.URL.Query().Get("attribute_name"); attributeName != "" {
		filters["attribute_name"] = attributeName
	}

	// Parse as_of timestamp (for temporal queries)
	if asOfStr := r.URL.Query().Get("as_of"); asOfStr != "" {
		asOf, err := time.Parse(time.RFC3339, asOfStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid as_of timestamp", err)
			return
		}
		filters["as_of"] = asOf
	}

	// Parse pagination
	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = db.DefaultQueryLimit
	}
	if limit > db.MaxQueryLimit {
		limit = db.MaxQueryLimit
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	// Query attributes
	attributes, err := s.attributeStore.List(filters, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to query attributes", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"attributes": attributes,
		"count":      len(attributes),
		"limit":      limit,
		"offset":     offset,
	})
}

// handleAttributesPost creates a new entity attribute
func (s *Server) handleAttributesPost(w http.ResponseWriter, r *http.Request) {
	var attr models.EntityAttribute
	if err := json.NewDecoder(r.Body).Decode(&attr); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if attr.EntityType == "" {
		respondError(w, http.StatusBadRequest, "entity_type is required", nil)
		return
	}
	if attr.EntityID == "" {
		respondError(w, http.StatusBadRequest, "entity_id is required", nil)
		return
	}
	if attr.AttributeName == "" {
		respondError(w, http.StatusBadRequest, "attribute_name is required", nil)
		return
	}
	if attr.Value == "" {
		respondError(w, http.StatusBadRequest, "value is required", nil)
		return
	}
	if attr.ValueType == "" {
		respondError(w, http.StatusBadRequest, "value_type is required", nil)
		return
	}
	if attr.Source == "" {
		respondError(w, http.StatusBadRequest, "source is required", nil)
		return
	}

	// Set valid_from to now if not provided
	if attr.ValidFrom.IsZero() {
		attr.ValidFrom = time.Now().UTC()
	}

	// Validate entity_type
	validEntityTypes := map[string]bool{
		"team":     true,
		"engineer": true,
		"org":      true,
	}
	if !validEntityTypes[attr.EntityType] {
		respondError(w, http.StatusBadRequest, "invalid entity_type (must be: team, engineer, org)", nil)
		return
	}

	// Validate value_type
	validValueTypes := map[string]bool{
		"string":  true,
		"number":  true,
		"boolean": true,
		"json":    true,
	}
	if !validValueTypes[attr.ValueType] {
		respondError(w, http.StatusBadRequest, "invalid value_type (must be: string, number, boolean, json)", nil)
		return
	}

	// Create attribute
	if err := s.attributeStore.Create(&attr); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create attribute", err)
		return
	}

	respondJSON(w, http.StatusCreated, attr)
}
