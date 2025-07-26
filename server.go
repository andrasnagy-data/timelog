package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	sentryzerolog "github.com/getsentry/sentry-go/zerolog"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"go.uber.org/fx"
)

type (
	// Server represents the HTTP server with all dependencies
	Server struct {
		config        *Config
		logger        zerolog.Logger
		server        *http.Server
		healthHandler *HealthHandler
		sentryWriter  *sentryzerolog.Writer
	}

	params struct {
		fx.In

		Config        *Config
		HealthHandler *HealthHandler
		Logger        zerolog.Logger
		SentryWriter  *sentryzerolog.Writer
	}
)

// buildMiddleware creates the middleware chain
func buildMiddleware(handler http.Handler, config *Config, logger zerolog.Logger) http.Handler {
	handler = cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})(handler)

	if config.IsEnvProd() {
		handler = hlog.RequestIDHandler("req_id", "Request-Id")(handler)
		handler = hlog.NewHandler(logger)(handler)
		handler = hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("request")
		})(handler)

		// Sentry recovery middleware
		sentryHandler := sentryhttp.New(sentryhttp.Options{
			Repanic: false,
		})
		handler = sentryHandler.Handle(handler)
	}

	return handler
}

func NewServer(p params) *Server {
	if p.Config.IsEnvProd() {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              p.Config.SentryDSN,
			Environment:      p.Config.Environment,
			Release:          p.Config.Version,
			AttachStacktrace: true,
			SendDefaultPII:   true,
			EnableTracing:    true,
			TracesSampler: sentry.TracesSampler(func(ctx sentry.SamplingContext) float64 {
				if ctx.Parent != nil && ctx.Parent.Sampled != sentry.SampledUndefined {
					if ctx.Parent.Sampled.Bool() {
						// TODO inherit parent's sr
						return 1.0
					}
					return 0.0
				}

				if ctx.Span.Name == "GET /health" || ctx.Span.Name == "GET /metrics" {
					return 0.0
				}
				return 1.0
			}),
		})
		if err != nil {
			p.Logger.Error().Err(err).Msg("Failed to initialize Sentry")
		} else {
			p.Logger.Debug().Str("environment", p.Config.Environment).Msg("Sentry initialized")
		}
	}

	mux := http.NewServeMux()
	// Middleware
	handler := buildMiddleware(mux, p.Config, p.Logger)
	// Routes
	mux.HandleFunc("GET /health", p.HealthHandler.Handle)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", p.Config.Port),
		Handler: handler,
	}

	return &Server{
		config:        p.Config,
		healthHandler: p.HealthHandler,
		logger:        p.Logger.With().Str("component", "server").Logger(),
		server:        server,
		sentryWriter:  p.SentryWriter,
	}
}

func (s *Server) Start(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			s.logger.Info().
				Str("addr", s.server.Addr).
				Str("environment", s.config.Environment).
				Bool("sentry_enabled", s.config.IsEnvProd()).
				Msg("Starting HTTP server")
			go func() {
				if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					s.logger.Fatal().Err(err).Msg("Server failed to start")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			s.logger.Info().Msg("Stopping HTTP server")

			if err := s.server.Shutdown(shutdownCtx); err != nil {
				s.logger.Error().Err(err).Msg("Error during server shutdown")
			}

			if s.config.IsEnvProd() {
				s.logger.Info().Msg("Flushing Sentry client and writer")
				if s.sentryWriter != nil {
					s.sentryWriter.Close()
				}
				sentry.Flush(2 * time.Second)
			}

			return nil
		},
	})
}
