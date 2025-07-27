package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type (

	// HealthSrvc handles business logic for health check functionality
	HealthSrvc struct {
		logger zerolog.Logger
		pool   *pgxpool.Pool
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

func NewHealthSrvc(logger zerolog.Logger, pool *pgxpool.Pool) *HealthSrvc {
	return &HealthSrvc{
		logger: logger.With().Str("component", "health").Logger(),
		pool:   pool,
	}
}

func (s *HealthSrvc) healthCheck(ctx context.Context) HealthResponse {
	var (
		res  int
		dbOk bool
	)
	err := s.pool.QueryRow(ctx, "SELECT 1").Scan(&res)

	dbOk = err == nil && res == 1
	if dbOk {
		return HealthResponse{
			Status:    "serving",
			Timestamp: time.Now(),
			Database:  dbOk,
		}
	} else {
		return HealthResponse{
			Status:    "not serving",
			Timestamp: time.Now(),
			Database:  dbOk,
		}
	}
}

func NewHealthHandler(srvc *HealthSrvc) *HealthHandler {
	return &HealthHandler{
		srvc: srvc,
	}
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response := h.srvc.healthCheck(ctx)

	w.Header().Set("Content-Type", "application/json")

	if response.Database {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode health check response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().Str("status", response.Status).Bool("database", response.Database).Msg("Health check completed")
}
