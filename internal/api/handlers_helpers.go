package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	apperrors "github.com/engineerdna/engineerdna/internal/errors"
)

// respondJSON writes a JSON response with the given status code
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response with the given status code
func respondError(w http.ResponseWriter, status int, message string, err error) {
	// Log full error details to stdout (visible to user since localhost)
	log.Printf("Error: %s: %v", message, err)

	// Check if error is already a structured AppError
	if appErr, ok := err.(*apperrors.AppError); ok {
		respondJSON(w, status, appErr)
		return
	}

	// Convert to a generic internal error if not structured
	appErr := apperrors.NewInternalError(message, err)
	respondJSON(w, status, appErr)
}

// validatePluginName checks if a plugin name is safe (prevents path traversal)
func validatePluginName(name string) error {
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid plugin name: must contain only letters, numbers, hyphens, and underscores")
	}
	return nil
}

// validateContentType checks if the request has the correct Content-Type
func validateContentType(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" && !strings.HasPrefix(contentType, "application/json;") {
		return fmt.Errorf("Content-Type must be application/json, got: %s", contentType)
	}
	return nil
}

// decodeAndValidateJSON decodes JSON from request body with size limit
func decodeAndValidateJSON(r *http.Request, v interface{}) error {
	// Content-Type validation
	if err := validateContentType(r); err != nil {
		return err
	}

	// Decode with limits
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Reject extra fields for security

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// parseLimitParam extracts and validates the limit query parameter
// Returns 0 if not provided (will use default in store layer)
// Returns error if invalid or negative
func parseLimitParam(r *http.Request) (int, error) {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return 0, nil // Use default in store
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return 0, fmt.Errorf("invalid limit parameter: must be an integer")
	}

	if limit < 0 {
		return 0, fmt.Errorf("invalid limit parameter: must be non-negative")
	}

	return limit, nil
}

// parseOffsetParam extracts and validates the offset query parameter
// Returns 0 if not provided (start from beginning)
// Returns error if invalid or negative
func parseOffsetParam(r *http.Request) (int, error) {
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		return 0, nil // Start from beginning
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return 0, fmt.Errorf("invalid offset parameter: must be an integer")
	}

	if offset < 0 {
		return 0, fmt.Errorf("invalid offset parameter: must be non-negative")
	}

	return offset, nil
}

// PaginationMeta contains pagination metadata for list responses
type PaginationMeta struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"hasMore"`
}

// PaginatedResponse is the standard response format for paginated list endpoints
type PaginatedResponse struct {
	Items      interface{}     `json:"items"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

// respondPaginated writes a paginated JSON response with standard metadata
// Use this for all endpoints that support limit/offset pagination
// For non-paginated endpoints, use respondJSON directly with just the items
func respondPaginated(w http.ResponseWriter, items interface{}, total, limit, offset int) {
	hasMore := offset+limit < total
	response := PaginatedResponse{
		Items: items,
		Pagination: &PaginationMeta{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: hasMore,
		},
	}
	respondJSON(w, http.StatusOK, response)
}
