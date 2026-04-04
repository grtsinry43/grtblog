package telemetry

import (
	"sort"
	"sync"
	"time"
)

// ErrorKind distinguishes error origins.
type ErrorKind string

const (
	KindBiz       ErrorKind = "biz"       // *response.AppError
	KindHTTP      ErrorKind = "http"      // *fiber.Error
	KindNotFound  ErrorKind = "not_found" // domain sentinel
	KindUnhandled ErrorKind = "unhandled" // unknown error
	KindPanic     ErrorKind = "panic"     // recovered panic
)

// maxDigests caps the number of unique fingerprints to prevent unbounded memory growth
// (e.g. an attacker sending requests to many distinct paths).
const maxDigests = 10000

// evictionInterval throttles the lazy eviction sweep so it runs at most once per minute.
const evictionInterval = time.Minute

// ErrorRecord is a single captured error occurrence (internal, never exported raw).
type ErrorRecord struct {
	Kind    ErrorKind
	BizCode string // e.g. "SERVER_ERROR", "NOT_FOUND", "" for panics
	Location string // normalised stack or handler hint
	Message  string // will be sanitised by Record()
}

// ErrorDigest is the aggregated, de-duplicated form keyed by fingerprint.
type ErrorDigest struct {
	Fingerprint   string    `json:"fingerprint"`
	Kind          ErrorKind `json:"kind"`
	BizCode       string    `json:"bizCode,omitempty"`
	Location      string    `json:"location"`
	SampleMessage string    `json:"sampleMessage"`
	Count         int64     `json:"count"`
	FirstSeen     time.Time `json:"firstSeen"`
	LastSeen      time.Time `json:"lastSeen"`
}

// Collector aggregates errors in memory with a configurable retention window.
// It is safe for concurrent use.
type Collector struct {
	mu            sync.Mutex
	retention     time.Duration
	digests       map[string]*ErrorDigest // fingerprint → digest
	now           func() time.Time       // injectable clock for testing
	lastEviction  time.Time
}

// NewCollector creates a Collector that retains error digests for the given
// duration. Expired digests are lazily evicted at most once per minute.
func NewCollector(retention time.Duration) *Collector {
	if retention <= 0 {
		retention = 24 * time.Hour
	}
	return &Collector{
		retention: retention,
		digests:   make(map[string]*ErrorDigest),
		now:       time.Now,
	}
}

// Record ingests a single error event, sanitises it, and merges it into the
// corresponding digest bucket.
func (c *Collector) Record(rec ErrorRecord) {
	if c == nil {
		return
	}

	rec.Message = SanitiseMessage(rec.Message)
	fp := Fingerprint(string(rec.Kind)+":"+rec.BizCode, rec.Location)

	now := c.now()
	c.mu.Lock()
	defer c.mu.Unlock()

	// Throttled lazy eviction: sweep at most once per evictionInterval.
	if now.Sub(c.lastEviction) >= evictionInterval {
		cutoff := now.Add(-c.retention)
		for k, d := range c.digests {
			if d.LastSeen.Before(cutoff) {
				delete(c.digests, k)
			}
		}
		c.lastEviction = now
	}

	d, ok := c.digests[fp]
	if ok {
		d.Count++
		d.LastSeen = now
		if rec.Message != "" {
			d.SampleMessage = truncate(rec.Message, 256)
		}
		return
	}

	// Guard against unbounded map growth.
	if len(c.digests) >= maxDigests {
		return
	}

	c.digests[fp] = &ErrorDigest{
		Fingerprint:   fp,
		Kind:          rec.Kind,
		BizCode:       rec.BizCode,
		Location:      rec.Location,
		SampleMessage: truncate(rec.Message, 256),
		Count:         1,
		FirstSeen:     now,
		LastSeen:      now,
	}
}

// Snapshot returns a copy of all current digests, sorted by count descending.
// The returned slice is safe to read without holding the lock.
func (c *Collector) Snapshot() []ErrorDigest {
	if c == nil {
		return nil
	}

	now := c.now()
	cutoff := now.Add(-c.retention)

	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]ErrorDigest, 0, len(c.digests))
	for _, d := range c.digests {
		if d.LastSeen.Before(cutoff) {
			continue
		}
		cp := *d
		result = append(result, cp)
	}

	// Sort: highest count first; ties broken by most recent.
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return result[i].LastSeen.After(result[j].LastSeen)
	})

	return result
}

// Reset clears all collected digests.
func (c *Collector) Reset() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.digests = make(map[string]*ErrorDigest)
}

// Stats returns quick summary numbers without copying all digests.
func (c *Collector) Stats() (uniqueErrors int, totalCount int64) {
	if c == nil {
		return 0, 0
	}

	now := c.now()
	cutoff := now.Add(-c.retention)

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, d := range c.digests {
		if d.LastSeen.Before(cutoff) {
			continue
		}
		uniqueErrors++
		totalCount += d.Count
	}
	return
}

// truncate limits s to maxLen runes, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
