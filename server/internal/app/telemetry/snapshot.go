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

// TelemetrySnapshot is the full payload that could be sent to a remote
// collection endpoint (P2+), or displayed in the admin dashboard for
// audit purposes.
type TelemetrySnapshot struct {
	GeneratedAt time.Time        `json:"generatedAt"`
	Instance    InstanceInfo     `json:"instance"`
	Errors      []ErrorDigest    `json:"errors"`
	Panics      []ErrorDigest    `json:"panics"`
	Summary     ErrorSummaryInfo `json:"summary"`
}

// InstanceInfo contains anonymous, non-PII environment metadata.
type InstanceInfo struct {
	InstanceID string `json:"instanceId"` // SHA-256 of hostname; not reversible
	Version    string `json:"version"`
	GoVersion  string `json:"goVersion"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
}

// ErrorSummaryInfo provides high-level aggregates.
type ErrorSummaryInfo struct {
	UniqueErrors int   `json:"uniqueErrors"`
	TotalErrors  int64 `json:"totalErrors"`
	UniquePanics int   `json:"uniquePanics"`
	TotalPanics  int64 `json:"totalPanics"`
}

// BuildSnapshot creates an audit-ready snapshot from the collector's current state.
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

// buildInstanceInfo gathers anonymous environment info.
func buildInstanceInfo() InstanceInfo {
	return InstanceInfo{
		InstanceID: anonymousInstanceID(),
		Version:    buildinfo.Version(),
		GoVersion:  runtime.Version(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
	}
}

// anonymousInstanceID generates a stable, non-reversible identifier based on the
// hostname. This allows correlating reports from the same instance over time
// without revealing the actual hostname.
func anonymousInstanceID() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	// Add a static salt so the hash isn't trivially rainbow-tabled.
	h := sha256.Sum256([]byte("grtblog-telemetry:" + hostname))
	return fmt.Sprintf("%x", h[:8]) // 16-char hex
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
			if i >= 20 { // cap display
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
