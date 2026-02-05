package metrics

import (
	"sync"
	"time"
)

type HTTPWindowSnapshot struct {
	Window       time.Duration `json:"window"`
	Requests     int64         `json:"requests"`
	Errors       int64         `json:"errors"`
	ErrorRate    float64       `json:"errorRate"`
	RPS          float64       `json:"rps"`
	P95LatencyMS float64       `json:"p95LatencyMs"`
	Status1xx    int64         `json:"status1xx"`
	Status2xx    int64         `json:"status2xx"`
	Status3xx    int64         `json:"status3xx"`
	Status4xx    int64         `json:"status4xx"`
	Status5xx    int64         `json:"status5xx"`
}

type HTTPStats struct {
	mu        sync.Mutex
	retention time.Duration
	now       func() time.Time
	buckets   map[int64]*httpMinuteBucket
}

type httpMinuteBucket struct {
	requests int64
	errors   int64
	s1xx     int64
	s2xx     int64
	s3xx     int64
	s4xx     int64
	s5xx     int64
	latBins  []int64
}

var httpLatencyBoundariesMS = []int64{5, 10, 25, 50, 100, 200, 500, 1000, 2000, 5000}

func NewHTTPStats(retention time.Duration) *HTTPStats {
	if retention <= 0 {
		retention = 2 * time.Hour
	}
	return &HTTPStats{
		retention: retention,
		now:       time.Now,
		buckets:   make(map[int64]*httpMinuteBucket),
	}
}

func (s *HTTPStats) Record(status int, latency time.Duration) {
	if s == nil {
		return
	}
	now := s.now().UTC()
	minute := now.Truncate(time.Minute).Unix()
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket := s.buckets[minute]
	if bucket == nil {
		bucket = &httpMinuteBucket{latBins: make([]int64, len(httpLatencyBoundariesMS)+1)}
		s.buckets[minute] = bucket
	}
	bucket.requests++
	if status >= 500 {
		bucket.errors++
	}
	switch {
	case status >= 100 && status < 200:
		bucket.s1xx++
	case status >= 200 && status < 300:
		bucket.s2xx++
	case status >= 300 && status < 400:
		bucket.s3xx++
	case status >= 400 && status < 500:
		bucket.s4xx++
	case status >= 500:
		bucket.s5xx++
	}

	ms := latency.Milliseconds()
	idx := len(httpLatencyBoundariesMS)
	for i, bound := range httpLatencyBoundariesMS {
		if ms <= bound {
			idx = i
			break
		}
	}
	bucket.latBins[idx]++

	expireBefore := now.Add(-s.retention).Truncate(time.Minute).Unix()
	for key := range s.buckets {
		if key < expireBefore {
			delete(s.buckets, key)
		}
	}
}

func (s *HTTPStats) Snapshot(window time.Duration) HTTPWindowSnapshot {
	if s == nil {
		return HTTPWindowSnapshot{Window: window}
	}
	if window <= 0 {
		window = 5 * time.Minute
	}
	nowMinute := s.now().UTC().Truncate(time.Minute).Unix()
	minMinute := s.now().UTC().Add(-window).Truncate(time.Minute).Unix()

	s.mu.Lock()
	defer s.mu.Unlock()

	out := HTTPWindowSnapshot{Window: window}
	latBins := make([]int64, len(httpLatencyBoundariesMS)+1)
	for minute, bucket := range s.buckets {
		if minute < minMinute || minute > nowMinute {
			continue
		}
		out.Requests += bucket.requests
		out.Errors += bucket.errors
		out.Status1xx += bucket.s1xx
		out.Status2xx += bucket.s2xx
		out.Status3xx += bucket.s3xx
		out.Status4xx += bucket.s4xx
		out.Status5xx += bucket.s5xx
		for i := range latBins {
			latBins[i] += bucket.latBins[i]
		}
	}
	if out.Requests > 0 {
		out.ErrorRate = float64(out.Errors) / float64(out.Requests)
	}
	seconds := window.Seconds()
	if seconds > 0 {
		out.RPS = float64(out.Requests) / seconds
	}
	out.P95LatencyMS = percentileFromBins(latBins, httpLatencyBoundariesMS, 0.95)
	return out
}

func percentileFromBins(counts []int64, bounds []int64, q float64) float64 {
	if len(counts) == 0 {
		return 0
	}
	var total int64
	for _, count := range counts {
		total += count
	}
	if total == 0 {
		return 0
	}
	target := int64(float64(total) * q)
	if target <= 0 {
		target = 1
	}
	var seen int64
	for i, count := range counts {
		seen += count
		if seen >= target {
			if i >= len(bounds) {
				return float64(bounds[len(bounds)-1])
			}
			return float64(bounds[i])
		}
	}
	return float64(bounds[len(bounds)-1])
}
