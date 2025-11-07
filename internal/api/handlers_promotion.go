package api

import (
	"net/http"
	"strings"
)

// Promotion Signals Handlers

func (s *Server) handlePromotionSignals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listPromotionSignals(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listPromotionSignals(w http.ResponseWriter, r *http.Request) {
	signals, err := s.promotionService.GetPromotionSignals()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get promotion signals", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"signals": signals,
	})
}

func (s *Server) handlePromotionSignalsByEngineer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract engineer ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/promotion/signals/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Engineer ID required", http.StatusBadRequest)
		return
	}

	engineerID := parts[0]

	signals, err := s.promotionService.GetPromotionSignalsByEngineer(engineerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get promotion signals", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"engineer_id": engineerID,
		"signals":     signals,
	})
}

func (s *Server) handlePromotionSignalDismiss(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract signal ID from path (/api/promotion/dismiss/{signal_id})
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/promotion/dismiss/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Signal ID required", http.StatusBadRequest)
		return
	}

	signalID := parts[0]

	var req struct {
		Notes string `json:"notes"`
	}

	if err := decodeAndValidateJSON(r, &req); err != nil {
		// Notes are optional, so an empty body is okay
		req.Notes = ""
	}

	err := s.promotionService.DismissPromotionSignal(signalID, req.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to dismiss promotion signal", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Promotion signal dismissed",
	})
}

func (s *Server) handlePromotionDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	signals, err := s.promotionService.DetectPromotionSignals()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to detect promotion signals", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"detected": len(signals),
		"signals":  signals,
	})
}
