package telemetry

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCollector_RecordAndSnapshot(t *testing.T) {
	c := NewCollector(time.Hour)

	// Record the same error type 3 times.
	for i := 0; i < 3; i++ {
		c.Record(ErrorRecord{
			Kind:    KindBiz,
			BizCode: "SERVER_ERROR",
			Location: "internal/app/content.(*Service).GenerateHTML",
			Message:  "template execution failed",
		})
	}

	// Record a different error once.
	c.Record(ErrorRecord{
		Kind:    KindPanic,
		Location: "internal/http/handler.(*CommentHandler).Create",
		Message:  "nil pointer dereference",
	})

	snap := c.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 unique digests, got %d", len(snap))
	}

	// First should be the one with count=3 (sorted by count desc).
	if snap[0].Count != 3 {
		t.Errorf("expected first digest count=3, got %d", snap[0].Count)
	}
	if snap[0].Kind != KindBiz {
		t.Errorf("expected first digest kind=biz, got %s", snap[0].Kind)
	}
	if snap[1].Count != 1 {
		t.Errorf("expected second digest count=1, got %d", snap[1].Count)
	}
	if snap[1].Kind != KindPanic {
		t.Errorf("expected second digest kind=panic, got %s", snap[1].Kind)
	}
}

func TestCollector_Retention(t *testing.T) {
	c := NewCollector(time.Minute)
	past := time.Now().Add(-2 * time.Minute)
	c.now = func() time.Time { return past }

	c.Record(ErrorRecord{
		Kind:    KindBiz,
		BizCode: "OLD_ERROR",
		Message: "old error",
	})

	// Reset clock to "now" and force eviction by advancing past evictionInterval.
	c.now = time.Now

	// Recording a new event triggers lazy eviction.
	c.Record(ErrorRecord{
		Kind:    KindBiz,
		BizCode: "NEW_ERROR",
		Message: "new error",
	})

	snap := c.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 digest after eviction, got %d", len(snap))
	}
	if snap[0].BizCode != "NEW_ERROR" {
		t.Errorf("expected remaining digest to be NEW_ERROR, got %s", snap[0].BizCode)
	}
}

func TestCollector_Stats(t *testing.T) {
	c := NewCollector(time.Hour)

	for i := 0; i < 5; i++ {
		c.Record(ErrorRecord{Kind: KindBiz, BizCode: "A", Location: "loc-a"})
	}
	c.Record(ErrorRecord{Kind: KindPanic, BizCode: "", Location: "loc-b"})

	unique, total := c.Stats()
	if unique != 2 {
		t.Errorf("expected 2 unique errors, got %d", unique)
	}
	if total != 6 {
		t.Errorf("expected 6 total count, got %d", total)
	}
}

func TestCollector_NilSafe(t *testing.T) {
	var c *Collector
	// All methods should be no-ops on nil receiver.
	c.Record(ErrorRecord{Kind: KindBiz})
	c.Reset()
	snap := c.Snapshot()
	if snap != nil {
		t.Errorf("expected nil snapshot, got %v", snap)
	}
	unique, total := c.Stats()
	if unique != 0 || total != 0 {
		t.Errorf("expected 0/0 stats, got %d/%d", unique, total)
	}
}

func TestCollector_MessageSanitised(t *testing.T) {
	c := NewCollector(time.Hour)
	c.Record(ErrorRecord{
		Kind:    KindBiz,
		BizCode: "SERVER_ERROR",
		Message: "failed for user@example.com with token eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.abc",
	})

	snap := c.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 digest, got %d", len(snap))
	}
	msg := snap[0].SampleMessage
	if msg == "" {
		t.Fatal("sample message should not be empty")
	}
	for _, pii := range []string{"user@example.com", "eyJ"} {
		if strings.Contains(msg, pii) {
			t.Errorf("sample message should not contain PII %q, got: %s", pii, msg)
		}
	}
}

func TestCollector_MaxDigests(t *testing.T) {
	c := NewCollector(time.Hour)

	// Fill to capacity with unique fingerprints.
	for i := 0; i < maxDigests+100; i++ {
		c.Record(ErrorRecord{
			Kind:    KindBiz,
			BizCode: "ERR",
			Location: strings.Repeat("x", i), // unique location per iteration
		})
	}

	snap := c.Snapshot()
	if len(snap) > maxDigests {
		t.Errorf("expected at most %d digests, got %d", maxDigests, len(snap))
	}
}

func TestCollector_ConcurrentAccess(t *testing.T) {
	c := NewCollector(time.Hour)
	var wg sync.WaitGroup

	// Concurrent writes.
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Record(ErrorRecord{
				Kind:    KindBiz,
				BizCode: "CONCURRENT",
				Location: "test",
				Message:  "concurrent error",
			})
		}(i)
	}

	// Concurrent reads.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Snapshot()
			_, _ = c.Stats()
		}()
	}

	wg.Wait()

	snap := c.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 digest, got %d", len(snap))
	}
	if snap[0].Count != 100 {
		t.Errorf("expected count=100, got %d", snap[0].Count)
	}
}
