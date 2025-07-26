package main

import (
	"os"
	"time"

	sentryzerolog "github.com/getsentry/sentry-go/zerolog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// NewLogger creates a new zerolog logger with pretty console output for development or JSON output for production, and returns an optional Sentry writer (nil if not production)
func NewLogger(config *Config) (zerolog.Logger, *sentryzerolog.Writer) {
	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		// Default to info level if parsing fails
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	if !config.IsEnvProd() {
		// Development: pretty console output
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
		return zerolog.New(consoleWriter).
			With().
			Timestamp().
			Caller().
			Logger(), nil
	}

	// Create Sentry writer using official integration (assumes Sentry client already initialized)
	sentryWriter, err := sentryzerolog.New(sentryzerolog.Config{
		Options: sentryzerolog.Options{
			Levels:          []zerolog.Level{zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel},
			WithBreadcrumbs: true,
			FlushTimeout:    3 * time.Second,
		},
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Sentry writer, using console only")
		// Default to development settings
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
		return zerolog.New(consoleWriter).
			With().
			Timestamp().
			Caller().
			Logger(), nil
	}

	log.Info().Msg("Zerolog Sentry writer initialized")

	// Production: JSON output to stderr + Sentry writer
	multiWriter := zerolog.MultiLevelWriter(os.Stderr, sentryWriter)

	return zerolog.New(multiWriter).
		With().
		Timestamp().
		Caller().
		Str("version", config.Version).
		Str("environment", config.Environment).
		Logger(), sentryWriter
}
