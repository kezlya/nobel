// Package config loads runtime configuration from the environment.
package config

import "os"

// Config holds process-wide settings.
type Config struct {
	// Addr is the host:port the HTTP server binds to.
	Addr string
	// DatabaseURL is the Postgres connection string. Empty means "use the
	// in-memory store" — convenient for local development and tests.
	DatabaseURL string
}

// Load reads configuration from environment variables with sane defaults.
func Load() Config {
	return Config{
		Addr:        envOr("NOBEL_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
