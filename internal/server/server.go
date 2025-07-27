package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	sentryzerolog "github.com/getsentry/sentry-go/zerolog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/fx"

	_ "github.com/andrasnagy-data/timelog/api"
	"github.com/andrasnagy-data/timelog/internal/shared/config"
)

type (
	// Server represents the HTTP server with all dependencies
	Server struct {
		server         *http.Server
		config         *config.Config
		logger         zerolog.Logger
		healthHandler  http.HandlerFunc
		sentryWriter   *sentryzerolog.Writer
		ActivityRouter chi.Router
	}

	params struct {
		fx.In

		Config         *config.Config
		Logger         zerolog.Logger
		HealthHandler  http.HandlerFunc
		SentryWriter   *sentryzerolog.Writer
		ActivityRouter chi.Router
	}
)

func NewServer(p params) *Server {
	r := chi.NewRouter()

	if p.Config.IsEnvProd() {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              p.Config.SentryDSN,
			Environment:      p.Config.Environment,
			Release:          p.Config.Version,
			AttachStacktrace: true,
			SendDefaultPII:   true,
			EnableTracing:    true,
			TracesSampler: sentry.TracesSampler(func(ctx sentry.SamplingContext) float64 {
				//TODO get parent's sampling rate if it exists

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

		sentryHandler := sentryhttp.New(sentryhttp.Options{})

		// Recovery middleware
		// Recover only in prod
		r.Use(sentryHandler.Handle)
	}

	// Middleware
	r.Use(hlog.NewHandler(p.Logger))
	r.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("url", r.URL.Path).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("HTTP request")
	}))
	r.Use(hlog.RequestIDHandler("req_id", "Request-Id"))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Get("/health", p.HealthHandler)

	r.Mount("/activities", p.ActivityRouter)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", p.Config.Port),
		Handler: r,
	}

	return &Server{
		config:        p.Config,
		healthHandler: p.HealthHandler,
		logger:        p.Logger,
		server:        server,
		sentryWriter:  p.SentryWriter,
	}
}

func (s *Server) Start(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: s.start,
		OnStop:  s.stop,
	})
}

// start starts the HTTP server
func (s *Server) start(_ context.Context) error {
	s.logger.Info().
		Str("addr", s.server.Addr).
		Str("environment", s.config.Environment).
		Bool("sentry_enabled", s.config.IsEnvProd()).
		Msg("Starting HTTP server")
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error().Err(err).Msg("Server failed to start")
		}
	}()

	s.logger.Info().Msg("HTTP server started")
	return nil
}

// stop gracefully shuts down the HTTP server
func (s *Server) stop(ctx context.Context) error {
	// Create timeout context for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s.logger.Info().Msg("Shutting down HTTP server...")

	if s.config.IsEnvProd() {
		s.logger.Info().Msg("Flushing Sentry client and writer")
		if s.sentryWriter != nil {
			s.sentryWriter.Close()
		}
		sentry.Flush(2 * time.Second)
	}

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error().Err(err).Msg("Error during server shutdown")
		return err
	}

	s.logger.Info().Msg("HTTP server shutdown completed")
	return nil
}
