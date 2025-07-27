package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/hlog"
)

type (
	// HealthSrvc handles business logic for health check functionality
	HealthSrvc struct {
		pool *pgxpool.Pool
	}

	// HealthResponse represents the response structure for health check endpoint
	HealthResponse struct {
		Status    string    `json:"status"`
		Timestamp time.Time `json:"timestamp"`
		Database  bool      `json:"database"`
	}
)

func NewHealthHandler(srvc *HealthSrvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := hlog.FromRequest(r)

		response := srvc.check(ctx)

		w.Header().Set("Content-Type", "application/json")

		if response.Database {
			logger.Debug().Msg("Database healthcheck ok")
			w.WriteHeader(http.StatusOK)
		} else {
			logger.Error().Msg("Database healthcheck failed")
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error().Err(err).Msg("Failed to encode health check response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func NewHealthSrvc(pool *pgxpool.Pool) *HealthSrvc {
	return &HealthSrvc{pool: pool}
}

func (s *HealthSrvc) check(ctx context.Context) HealthResponse {
	var (
		res  int
		dbOk bool
	)
	err := s.pool.QueryRow(ctx, "SELECT 1").Scan(&res)

	now := time.Now().UTC()

	dbOk = err == nil && res == 1
	if dbOk {
		return HealthResponse{
			Status:    "serving",
			Timestamp: now,
			Database:  dbOk,
		}
	} else {
		return HealthResponse{
			Status:    "not serving",
			Timestamp: now,
			Database:  dbOk,
		}
	}
}
