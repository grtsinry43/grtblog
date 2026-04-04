package main

import (
	"os"
	"strings"
)

// Config holds all configuration for the telemetry collector.
type Config struct {
	Port           string // HTTP listen port (default "9090")
	DSN            string // PostgreSQL connection string
	SetupToken     string // One-time token for first Passkey registration (required if empty DB)
	WebAuthnRPID   string // Relying Party ID (domain, e.g. "telemetry.example.com")
	WebAuthnOrigin string // Allowed origin (e.g. "https://telemetry.example.com")
	GrafanaURL     string // Internal Grafana URL to reverse-proxy (e.g. "http://grafana:3000")
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() Config {
	return Config{
		Port:           getEnv("PORT", "9090"),
		DSN:            getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/grtblog_telemetry?sslmode=disable"),
		SetupToken:     strings.TrimSpace(getEnv("SETUP_TOKEN", "")),
		WebAuthnRPID:   getEnv("WEBAUTHN_RP_ID", "localhost"),
		WebAuthnOrigin: getEnv("WEBAUTHN_RP_ORIGIN", "http://localhost:9090"),
		GrafanaURL:     strings.TrimRight(getEnv("GRAFANA_URL", "http://localhost:3000"), "/"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
