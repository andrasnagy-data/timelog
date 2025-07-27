package config

import "github.com/caarlos0/env/v11"

// Config holds application configuration
type Config struct {
	Version     string `env:"VERSION"`
	Port        int    `env:"PORT"`
	Environment string `env:"ENVIRONMENT"`
	LogLevel    string `env:"LOG_LEVEL"`
	SentryDSN   string `env:"SENTRY_DSN"`
	DatabaseURL string `env:"DATABASE_URL"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) IsEnvProd() bool {
	if c.Environment == "prod" && c.SentryDSN != "" {
		return true
	}
	return false
}
