package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
)

var ErrInvalidPreferences = errors.New("invalid telemetry preferences")

// Preferences exposes the effective endpoint. UsingDefaultEndpoint tells the
// client whether Endpoint came from deployment config rather than sysconfig.
type Preferences struct {
	Enabled              bool   `json:"enabled"`
	Endpoint             string `json:"endpoint"`
	Interval             string `json:"interval"`
	UsingDefaultEndpoint bool   `json:"usingDefaultEndpoint"`
}

type UpdatePreferencesInput struct {
	Enabled  *bool   `json:"enabled"`
	Endpoint *string `json:"endpoint"`
	Interval *string `json:"interval"`
}

func (s *Service) Preferences(ctx context.Context) Preferences {
	if s == nil || s.sysCfg == nil {
		return Preferences{}
	}
	cfg := s.sysCfg.TelemetryReporterConfig(ctx)
	effective := strings.TrimSpace(cfg.Endpoint)
	usingDefault := effective == ""
	if usingDefault {
		effective = strings.TrimSpace(s.defaultEndpoint)
	}
	return Preferences{Enabled: cfg.Enabled, Endpoint: effective, Interval: cfg.Interval.String(), UsingDefaultEndpoint: usingDefault}
}

func (s *Service) UpdatePreferences(ctx context.Context, input UpdatePreferencesInput) (Preferences, error) {
	if s == nil || s.sysCfg == nil {
		return Preferences{}, errors.New("telemetry sysconfig is unavailable")
	}
	current := s.sysCfg.TelemetryReporterConfig(ctx)
	endpoint := current.Endpoint
	if input.Endpoint != nil {
		endpoint = strings.TrimSpace(*input.Endpoint)
		if endpoint != "" {
			if err := validateEndpointURL(endpoint); err != nil {
				return Preferences{}, fmt.Errorf("%w: %v", ErrInvalidPreferences, err)
			}
		}
	}
	interval := current.Interval
	if input.Interval != nil {
		var err error
		interval, err = parsePreferenceInterval(*input.Interval)
		if err != nil {
			return Preferences{}, err
		}
	}

	enabled := current.Enabled
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	enabledRaw := json.RawMessage("false")
	if enabled {
		enabledRaw = json.RawMessage("true")
	}
	endpointJSON, _ := json.Marshal(endpoint)
	intervalJSON, _ := json.Marshal(interval.String())
	endpointRaw := json.RawMessage(endpointJSON)
	intervalRaw := json.RawMessage(intervalJSON)
	boolType, stringType := "bool", "string"
	_, err := s.sysCfg.UpdateConfigs(ctx, []sysconfig.UpdateItem{
		{Key: "telemetry.enabled", Value: &enabledRaw, ValueType: &boolType},
		{Key: "telemetry.endpoint", Value: &endpointRaw, ValueType: &stringType},
		{Key: "telemetry.interval", Value: &intervalRaw, ValueType: &stringType},
	})
	if err != nil {
		return Preferences{}, err
	}
	return s.Preferences(ctx), nil
}

func parsePreferenceInterval(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("%w: interval is required", ErrInvalidPreferences)
	}
	interval, err := time.ParseDuration(raw)
	if err != nil || interval < minInterval {
		return 0, fmt.Errorf("%w: interval must be a duration of at least %s", ErrInvalidPreferences, minInterval)
	}
	return interval, nil
}
