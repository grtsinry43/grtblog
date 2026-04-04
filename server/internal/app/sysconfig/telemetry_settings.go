package sysconfig

import (
	"context"
	"strconv"
	"strings"
	"time"
)

// TelemetryReporterConfig describes the remote telemetry upload settings.
type TelemetryReporterConfig struct {
	Enabled  bool
	Endpoint string
	Interval time.Duration
}

// TelemetryReporterConfig returns the current telemetry reporting configuration.
//
// Keys:
//   - telemetry.enabled:  bool — master switch (default false)
//   - telemetry.endpoint: string — HTTPS URL to POST snapshots to
//   - telemetry.interval: string — Go duration or "Nd" (default "24h")
func (s *Service) TelemetryReporterConfig(ctx context.Context) TelemetryReporterConfig {
	cfg := TelemetryReporterConfig{
		Enabled:  false,
		Endpoint: "",
		Interval: 24 * time.Hour,
	}

	keys := []string{"telemetry.enabled", "telemetry.endpoint", "telemetry.interval"}
	items, err := s.repo.List(ctx, keys)
	if err != nil {
		return cfg
	}
	lookup := makeLookup(items)

	cfg.Enabled = cfgParseBool(lookup["telemetry.enabled"], false)
	cfg.Endpoint = cfgParseString(lookup["telemetry.endpoint"], "")

	raw := cfgParseString(lookup["telemetry.interval"], "")
	if raw != "" {
		// Support "Nd" format (e.g. "7d" = 7 days).
		if strings.HasSuffix(raw, "d") {
			if days, parseErr := strconv.Atoi(strings.TrimSuffix(raw, "d")); parseErr == nil && days > 0 {
				cfg.Interval = time.Duration(days) * 24 * time.Hour
			}
		} else if d, parseErr := time.ParseDuration(raw); parseErr == nil && d > 0 {
			cfg.Interval = d
		}
	}

	return cfg
}
