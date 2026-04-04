package telemetry

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestService_FullSnapshot_NilDeps(t *testing.T) {
	// Service with only a collector and no other dependencies should
	// produce a valid snapshot without panicking.
	collector := NewCollector(time.Hour)
	collector.Record(ErrorRecord{
		Kind:    KindBiz,
		BizCode: "SERVER_ERROR",
		Location: "test.Handler",
		Message: "something broke",
	})
	collector.Record(ErrorRecord{
		Kind:    KindPanic,
		Location: "test.PanicHandler",
		Message: "nil pointer",
	})

	svc := NewService(collector, nil, nil, nil, nil, nil, "")
	snap := svc.FullSnapshot(context.Background())

	if snap == nil {
		t.Fatal("FullSnapshot returned nil")
	}
	if snap.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should not be zero")
	}
	if snap.Instance.Version == "" {
		t.Error("Version should not be empty")
	}
	if snap.Instance.InstanceID == "" {
		t.Error("InstanceID should not be empty")
	}
	if snap.Instance.DeployMode == "" {
		t.Error("DeployMode should not be empty")
	}
	if snap.Summary.UniqueErrors != 1 {
		t.Errorf("expected 1 unique error, got %d", snap.Summary.UniqueErrors)
	}
	if snap.Summary.UniquePanics != 1 {
		t.Errorf("expected 1 unique panic, got %d", snap.Summary.UniquePanics)
	}
	// Metrics should be zero-valued but present.
	if snap.Metrics.Content.ArticlesTotal != 0 {
		t.Errorf("expected 0 articles with nil DB, got %d", snap.Metrics.Content.ArticlesTotal)
	}
	if snap.Metrics.Traffic.Window != "" {
		t.Errorf("expected empty traffic window with nil HTTPStats, got %q", snap.Metrics.Traffic.Window)
	}
}

func TestService_FullSnapshot_NilCollector(t *testing.T) {
	// FullSnapshot must degrade gracefully when collector is nil.
	svc := NewService(nil, nil, nil, nil, nil, nil, "")
	snap := svc.FullSnapshot(context.Background())

	if snap == nil {
		t.Fatal("FullSnapshot should not return nil with nil collector")
	}
	if snap.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should not be zero even with nil collector")
	}
	if snap.Summary.UniqueErrors != 0 || snap.Summary.UniquePanics != 0 {
		t.Error("expected zero errors/panics with nil collector")
	}
}

func TestService_DetectDeployMode(t *testing.T) {
	mode := detectDeployMode()
	// In a test environment it should be "binary" (no Docker markers).
	if mode != "binary" {
		t.Logf("detectDeployMode returned %q (may be running in Docker)", mode)
	}
}

func TestService_SetWSManager_Concurrent(t *testing.T) {
	svc := NewService(NewCollector(time.Hour), nil, nil, nil, nil, nil, "")

	// Should not panic on nil.
	svc.SetWSManager(nil)

	// Nil service should not panic.
	var nilSvc *Service
	nilSvc.SetWSManager(nil)

	// Concurrent SetWSManager + FullSnapshot should not race.
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			svc.SetWSManager(nil) // Store nil is a no-op
		}()
		go func() {
			defer wg.Done()
			_ = svc.FullSnapshot(context.Background())
		}()
	}
	wg.Wait()
}
