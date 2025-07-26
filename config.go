package main

import "github.com/caarlos0/env/v11"

// Config holds application configuration
type Config struct {
	Version     string `env:"VERSION" envDefault:"0.1.0"`
	Port        int    `env:"PORT" envDefault:"8080"`
	Environment string `env:"ENVIRONMENT" envDefault:"prod"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	SentryDSN   string `env:"SENTRY_DSN" envDefault:"https://dd4ddf4fbcf1a88df33c296eca5aeac6@o4509623498833920.ingest.de.sentry.io/4509731597910096"`
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
