package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
)

// ReportStatus describes the outcome of a single upload attempt.
type ReportStatus string

const (
	ReportSuccess ReportStatus = "success"
	ReportFailed  ReportStatus = "failed"
	ReportSkipped ReportStatus = "skipped" // disabled or no endpoint configured
)

// ReportRecord logs a single upload attempt.
type ReportRecord struct {
	Timestamp  time.Time    `json:"timestamp"`
	Status     ReportStatus `json:"status"`
	StatusCode int          `json:"statusCode,omitempty"`
	Message    string       `json:"message,omitempty"`
	DurationMS int64        `json:"durationMs,omitempty"`
}

// ReporterConfig is re-exported from sysconfig for convenience.
type ReporterConfig = sysconfig.TelemetryReporterConfig

const (
	maxHistorySize  = 50
	reportTimeout   = 15 * time.Second
	maxRetries      = 1
	retryBackoff    = 3 * time.Second
	defaultInterval = 24 * time.Hour
	minInterval     = 1 * time.Hour
	configTimeout   = 3 * time.Second
	maxDrainBytes   = 4096 // cap response body drain to prevent memory pressure
)

// Reporter periodically uploads telemetry snapshots to a remote endpoint.
// Run must be called separately by the server bootstrap (in a goroutine).
type Reporter struct {
	svc    *Service
	client *http.Client

	mu      sync.Mutex
	history []ReportRecord
}

// NewReporter creates a Reporter bound to the given telemetry Service.
func NewReporter(svc *Service) *Reporter {
	return &Reporter{
		svc: svc,
		client: &http.Client{
			Timeout: reportTimeout,
		},
		history: make([]ReportRecord, 0, maxHistorySize),
	}
}

// Run starts the periodic reporting loop. It blocks until ctx is cancelled.
func (r *Reporter) Run(ctx context.Context) {
	// Wait a bit after startup before first report to let metrics accumulate.
	select {
	case <-ctx.Done():
		return
	case <-time.After(2 * time.Minute):
	}

	for {
		cfg := r.readConfigSafe(ctx)
		interval := cfg.Interval
		if interval < minInterval {
			interval = defaultInterval
		}

		if cfg.Enabled && cfg.Endpoint != "" {
			r.doReport(ctx, cfg)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}

// ReportNow triggers an immediate upload attempt. It uses a bounded context
// (max 45s) to avoid blocking the HTTP handler indefinitely.
func (r *Reporter) ReportNow(ctx context.Context) ReportRecord {
	cfg := r.readConfigSafe(ctx)
	if !cfg.Enabled {
		rec := ReportRecord{
			Timestamp: time.Now().UTC(),
			Status:    ReportSkipped,
			Message:   "telemetry reporting is disabled",
		}
		r.pushHistory(rec)
		return rec
	}
	if cfg.Endpoint == "" {
		rec := ReportRecord{
			Timestamp: time.Now().UTC(),
			Status:    ReportSkipped,
			Message:   "no telemetry endpoint configured",
		}
		r.pushHistory(rec)
		return rec
	}

	// Bound the total time for the handler to avoid blocking the Fiber worker.
	bounded, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	return r.doReport(bounded, cfg)
}

// History returns a copy of the recent upload history (newest first).
func (r *Reporter) History() []ReportRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]ReportRecord, len(r.history))
	copy(result, r.history)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// doReport builds a snapshot and uploads it, with a single retry.
// The config is passed in to avoid redundant DB reads.
func (r *Reporter) doReport(ctx context.Context, cfg ReporterConfig) ReportRecord {
	if err := validateEndpointURL(cfg.Endpoint); err != nil {
		rec := ReportRecord{
			Timestamp: time.Now().UTC(),
			Status:    ReportFailed,
			Message:   err.Error(),
		}
		r.pushHistory(rec)
		return rec
	}

	snap := r.svc.FullSnapshot(ctx)
	body, err := json.Marshal(snap)
	if err != nil {
		rec := ReportRecord{
			Timestamp: time.Now().UTC(),
			Status:    ReportFailed,
			Message:   fmt.Sprintf("marshal error: %v", err),
		}
		r.pushHistory(rec)
		return rec
	}

	var lastRec ReportRecord
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				cancelled := ReportRecord{
					Timestamp: time.Now().UTC(),
					Status:    ReportFailed,
					Message:   "context cancelled during retry",
				}
				r.pushHistory(cancelled)
				return cancelled
			case <-time.After(retryBackoff):
			}
		}

		lastRec = r.upload(ctx, cfg.Endpoint, body)
		if lastRec.Status == ReportSuccess {
			r.pushHistory(lastRec)
			log.Printf("[telemetry] report sent to %s (%dms)", cfg.Endpoint, lastRec.DurationMS)
			return lastRec
		}
	}

	r.pushHistory(lastRec)
	log.Printf("[telemetry] report failed after %d attempts: %s", maxRetries+1, lastRec.Message)
	return lastRec
}

// upload performs a single HTTP POST.
func (r *Reporter) upload(ctx context.Context, endpoint string, body []byte) ReportRecord {
	start := time.Now()
	rec := ReportRecord{
		Timestamp: start.UTC(),
		Status:    ReportFailed,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		rec.Message = fmt.Sprintf("create request: %v", err)
		rec.DurationMS = time.Since(start).Milliseconds()
		return rec
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "grtblog-telemetry/1.0")

	resp, err := r.client.Do(req)
	rec.DurationMS = time.Since(start).Milliseconds()
	if err != nil {
		rec.Message = fmt.Sprintf("http error: %v", err)
		return rec
	}
	defer resp.Body.Close()
	// Drain body (capped) to allow connection reuse.
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxDrainBytes))

	rec.StatusCode = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		rec.Status = ReportSuccess
		rec.Message = "ok"
	} else {
		rec.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	return rec
}

// readConfigSafe reads config with a bounded timeout to prevent stalling.
func (r *Reporter) readConfigSafe(ctx context.Context) ReporterConfig {
	if r.svc == nil || r.svc.sysCfg == nil {
		return ReporterConfig{}
	}
	timedCtx, cancel := context.WithTimeout(ctx, configTimeout)
	defer cancel()
	return r.svc.sysCfg.TelemetryReporterConfig(timedCtx)
}

// pushHistory appends a record to the ring buffer.
func (r *Reporter) pushHistory(rec ReportRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.history) >= maxHistorySize {
		copy(r.history, r.history[1:])
		r.history[len(r.history)-1] = rec
	} else {
		r.history = append(r.history, rec)
	}
}

// --- endpoint URL validation (SSRF prevention) ---

// validateEndpointURL ensures the endpoint is an HTTPS URL that does not
// target private, loopback, or link-local addresses.
func validateEndpointURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("telemetry endpoint must use https, got %q", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("telemetry endpoint has no host")
	}

	// Reject obviously private/loopback hostnames.
	lower := strings.ToLower(host)
	if lower == "localhost" || lower == "metadata.google.internal" {
		return fmt.Errorf("telemetry endpoint must not target %s", host)
	}

	// If the host is an IP literal, check for private ranges.
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("telemetry endpoint must not target private/loopback address %s", host)
		}
	}

	return nil
}
