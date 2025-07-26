package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type (

	// HealthSrvc handles business logic for health check functionality
	HealthSrvc struct {
		logger zerolog.Logger
	}

	// HealthResponse represents the response structure for health check endpoint
	HealthResponse struct {
		Status    string    `json:"status"`
		Timestamp time.Time `json:"timestamp"`
		Database  bool      `json:"database"`
	}

	// HealthHandler handles HTTP requests for health check endpoint
	HealthHandler struct {
		srvc   *HealthSrvc
		logger zerolog.Logger
	}
)

func NewHealthSrvc(logger zerolog.Logger) *HealthSrvc {
	return &HealthSrvc{
		logger: logger.With().Str("component", "health").Logger(),
	}
}

func (s *HealthSrvc) healthCheck() HealthResponse {
	// TODO database check
	return HealthResponse{
		Status:    "serving",
		Timestamp: time.Now(),
		Database:  true,
	}
}

func NewHealthHandler(srvc *HealthSrvc) *HealthHandler {
	return &HealthHandler{
		srvc: srvc,
	}
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	response := h.srvc.healthCheck()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode health check response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().Str("status", response.Status).Msg("Health check completed")
}
