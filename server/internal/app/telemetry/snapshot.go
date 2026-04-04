package telemetry

import (
	"crypto/sha256"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/buildinfo"
)

// ---------------------------------------------------------------------------
// P0: Error-only snapshot (backward compatible)
// ---------------------------------------------------------------------------

// TelemetrySnapshot is the P0 error-only payload.
// Kept for backward compatibility with the existing admin endpoint.
type TelemetrySnapshot struct {
	GeneratedAt time.Time        `json:"generatedAt"`
	Instance    InstanceInfo     `json:"instance"`
	Errors      []ErrorDigest    `json:"errors"`
	Panics      []ErrorDigest    `json:"panics"`
	Summary     ErrorSummaryInfo `json:"summary"`
}

// BuildSnapshot creates an error-only snapshot (P0 compat).
func BuildSnapshot(c *Collector) TelemetrySnapshot {
	all := c.Snapshot()

	var errors, panics []ErrorDigest
	for _, d := range all {
		if d.Kind == KindPanic {
			panics = append(panics, d)
		} else {
			errors = append(errors, d)
		}
	}

	var totalErrors, totalPanics int64
	for _, d := range errors {
		totalErrors += d.Count
	}
	for _, d := range panics {
		totalPanics += d.Count
	}

	return TelemetrySnapshot{
		GeneratedAt: c.now().UTC(),
		Instance:    buildInstanceInfo(),
		Errors:      errors,
		Panics:      panics,
		Summary: ErrorSummaryInfo{
			UniqueErrors: len(errors),
			TotalErrors:  totalErrors,
			UniquePanics: len(panics),
			TotalPanics:  totalPanics,
		},
	}
}

// ---------------------------------------------------------------------------
// P1: Full snapshot (environment + metrics + errors)
// ---------------------------------------------------------------------------

// FullTelemetrySnapshot is the comprehensive payload including environment
// info, feature flags, runtime metrics, and error digests.
type FullTelemetrySnapshot struct {
	GeneratedAt time.Time        `json:"generatedAt"`
	Instance    InstanceInfo     `json:"instance"`
	Metrics     RuntimeMetrics   `json:"metrics"`
	Errors      []ErrorDigest    `json:"errors"`
	Panics      []ErrorDigest    `json:"panics"`
	Summary     ErrorSummaryInfo `json:"summary"`
}

// InstanceInfo contains anonymous, non-PII environment metadata.
type InstanceInfo struct {
	InstanceID     string       `json:"instanceId"`
	Version        string       `json:"version"`
	GoVersion      string       `json:"goVersion"`
	OS             string       `json:"os"`
	Arch           string       `json:"arch"`
	UptimeSeconds  int64        `json:"uptimeSeconds,omitempty"`
	DeployMode     string       `json:"deployMode,omitempty"`     // "docker" | "binary" | "unknown"
	Features       FeatureFlags `json:"features,omitempty"`
}

// FeatureFlags captures which optional features are enabled.
// Only booleans; no secrets, keys, or PII.
type FeatureFlags struct {
	FederationEnabled  bool `json:"federationEnabled"`
	ActivityPubEnabled bool `json:"activityPubEnabled"`
	CommentsDisabled   bool `json:"commentsDisabled"`
	EmailEnabled       bool `json:"emailEnabled"`
	TurnstileEnabled   bool `json:"turnstileEnabled"`
}

// RuntimeMetrics contains aggregated operational metrics.
type RuntimeMetrics struct {
	Content    ContentMetrics    `json:"content"`
	Traffic    TrafficMetrics    `json:"traffic"`
	ISR        ISRMetrics        `json:"isr"`
	Federation FederationMetrics `json:"federation"`
	Realtime   RealtimeMetrics   `json:"realtime"`
}

// ContentMetrics reports anonymous content volume.
type ContentMetrics struct {
	ArticlesTotal    int64 `json:"articlesTotal"`
	MomentsTotal     int64 `json:"momentsTotal"`
	CommentsTotal    int64 `json:"commentsTotal"`
	FriendLinksTotal int64 `json:"friendLinksTotal"`
}

// TrafficMetrics reports aggregated request statistics.
type TrafficMetrics struct {
	Window       string  `json:"window"`
	RequestTotal int64   `json:"requestTotal"`
	ErrorRate5xx float64 `json:"errorRate5xx"`
	P95LatencyMS float64 `json:"p95LatencyMs"`
}

// ISRMetrics reports rendering pipeline statistics.
type ISRMetrics struct {
	RenderTotal   int64   `json:"renderTotal"`
	RenderSuccess int64   `json:"renderSuccess"`
	RenderFailed  int64   `json:"renderFailed"`
	AvgRenderMS   float64 `json:"avgRenderMs"`
	P95RenderMS   float64 `json:"p95RenderMs"`
}

// FederationMetrics reports federation delivery statistics (24h window).
type FederationMetrics struct {
	OutboundTotal    int64 `json:"outboundTotal"`
	OutboundFailures int64 `json:"outboundFailures"`
	ActiveInstances  int64 `json:"activeInstances"`
}

// RealtimeMetrics reports WebSocket connection statistics.
type RealtimeMetrics struct {
	WSConnectionsCurrent int64 `json:"wsConnectionsCurrent"`
	WSRooms              int   `json:"wsRooms"`
	BroadcastTotal       int64 `json:"broadcastTotal"`
}

// ErrorSummaryInfo provides high-level error aggregates.
type ErrorSummaryInfo struct {
	UniqueErrors int   `json:"uniqueErrors"`
	TotalErrors  int64 `json:"totalErrors"`
	UniquePanics int   `json:"uniquePanics"`
	TotalPanics  int64 `json:"totalPanics"`
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildInstanceInfo gathers anonymous environment info (base fields only).
func buildInstanceInfo() InstanceInfo {
	return InstanceInfo{
		InstanceID: anonymousInstanceID(),
		Version:    buildinfo.Version(),
		GoVersion:  runtime.Version(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
	}
}

// anonymousInstanceID generates a stable, non-reversible identifier based on
// the hostname.
func anonymousInstanceID() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	h := sha256.Sum256([]byte("grtblog-telemetry:" + hostname))
	return fmt.Sprintf("%x", h[:8]) // 16-char hex
}

// detectDeployMode infers whether the process is running inside Docker.
func detectDeployMode() string {
	// /.dockerenv is created by Docker; /run/.containerenv by Podman.
	for _, marker := range []string{"/.dockerenv", "/run/.containerenv"} {
		if _, err := os.Stat(marker); err == nil {
			return "docker"
		}
	}
	return "binary"
}

// FormatSnapshotText produces a human-readable summary for CLI / log output.
func FormatSnapshotText(snap TelemetrySnapshot) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Telemetry Snapshot (%s) ===\n", snap.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "Instance: %s  Version: %s  Go: %s  OS: %s/%s\n",
		snap.Instance.InstanceID, snap.Instance.Version,
		snap.Instance.GoVersion, snap.Instance.OS, snap.Instance.Arch)
	fmt.Fprintf(&b, "Errors: %d unique / %d total   Panics: %d unique / %d total\n\n",
		snap.Summary.UniqueErrors, snap.Summary.TotalErrors,
		snap.Summary.UniquePanics, snap.Summary.TotalPanics)

	if len(snap.Errors) > 0 {
		b.WriteString("── Errors ──\n")
		for i, d := range snap.Errors {
			if i >= 20 {
				fmt.Fprintf(&b, "  ... and %d more\n", len(snap.Errors)-20)
				break
			}
			fmt.Fprintf(&b, "  [%s] %s  count=%d  biz=%s\n    %s\n",
				d.Fingerprint, d.Location, d.Count, d.BizCode, d.SampleMessage)
		}
		b.WriteString("\n")
	}

	if len(snap.Panics) > 0 {
		b.WriteString("── Panics ──\n")
		for i, d := range snap.Panics {
			if i >= 10 {
				fmt.Fprintf(&b, "  ... and %d more\n", len(snap.Panics)-10)
				break
			}
			fmt.Fprintf(&b, "  [%s] %s  count=%d\n    %s\n",
				d.Fingerprint, d.Location, d.Count, d.SampleMessage)
		}
	}

	return b.String()
}
