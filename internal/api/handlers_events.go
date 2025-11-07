package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// handleEvents routes event-related requests
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listEvents(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listEvents retrieves a filtered and paginated list of events
func (s *Server) listEvents(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	filters := make(map[string]interface{})

	if since := query.Get("since"); since != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err == nil {
			filters["since"] = t
		}
	}
	if eventType := query.Get("type"); eventType != "" {
		filters["type"] = eventType
	}
	if source := query.Get("source"); source != "" {
		filters["source"] = source
	}
	if actor := query.Get("actor"); actor != "" {
		filters["actor"] = actor
	}

	limit := 50
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			// Cap limit at 1000 to prevent excessive memory usage
			if limit > 1000 {
				limit = 1000
			}
		}
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	events, err := s.eventStore.List(filters, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list events", err)
		return
	}

	total, err := s.eventStore.Count(filters)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to count events", err)
		return
	}

	respondPaginated(w, events, total, limit, offset)
}

// handleEventByID retrieves a single event by its ID
func (s *Server) handleEventByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/events/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Event ID required", http.StatusBadRequest)
		return
	}

	id := parts[0]

	event, err := s.eventStore.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get event", err)
		return
	}

	if event == nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, event)
}
