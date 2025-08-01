package database

import (
	"context"
	"time"

	"github.com/andrasnagy-data/timelog/internal/shared/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// NewPgxPool creates a PostgreSQL connection pool with production-ready settings.
// It configures connection limits, timeouts, and lifetimes optimized for web applications.
// Pool settings: max 10 connections, min 5 connections, 1-hour max lifetime, 30-min idle timeout.
func NewPgxPool(cfg *config.Config, logger zerolog.Logger) (*pgxpool.Pool, error) {
	logger.Debug().Str("DATABASE_URL", cfg.DatabaseURL).Msg("Initializing database connection pool")

	config, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to parse database URL")
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	logger.Debug().
		Int32("max_conns", config.MaxConns).
		Int32("min_conns", config.MinConns).
		Dur("max_conns_lifetime", config.MaxConnLifetime).
		Dur("max_conns_idletime", config.MaxConnIdleTime).
		Msg("Database connection pool configuration")

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create database connection pool")
		return nil, err
	}

	logger.Debug().Msg("Database connection pool created successfully")
	return pool, nil
}
