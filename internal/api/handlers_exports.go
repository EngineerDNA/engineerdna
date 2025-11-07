package api

import (
	"net/http"

	sdk "github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

// handleExports routes export-related requests
func (s *Server) handleExports(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listExports(w, r)
	case http.MethodPost:
		s.createExport(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listExports retrieves export history
func (s *Server) listExports(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}
	if limit == 0 {
		limit = 100 // Default limit
	}

	offset, err := parseOffsetParam(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
		return
	}

	exports, err := s.exportsStore.ListExports(limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve exports", err)
		return
	}

	total, err := s.exportsStore.CountExports()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to count exports", err)
		return
	}

	respondPaginated(w, exports, total, limit, offset)
}

// createExport exports data to a destination plugin
func (s *Server) createExport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Plugin   string                 `json:"plugin"`
		DataType string                 `json:"data_type"`
		Data     map[string]interface{} `json:"data"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if req.Plugin == "" || req.DataType == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields: plugin, data_type", nil)
		return
	}

	result, err := s.executor.ExportToDestination(req.Plugin, sdk.ExportParams{
		DataType: req.DataType,
		Data:     req.Data,
	}, false)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Export failed", err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}
