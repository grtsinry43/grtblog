package backup

import (
	"context"
	"testing"
	"time"

	backupdomain "github.com/grtsinry43/grtblog-v2/server/internal/domain/backup"
)

type scheduleTestRepository struct {
	backupdomain.Repository
	schedule *backupdomain.Schedule
}

func (r *scheduleTestRepository) GetSchedule(context.Context) (*backupdomain.Schedule, error) {
	copy := *r.schedule
	return &copy, nil
}

func (r *scheduleTestRepository) SaveSchedule(_ context.Context, schedule *backupdomain.Schedule) error {
	copy := *schedule
	r.schedule = &copy
	return nil
}

func TestUpdateScheduleKeepsNextRunWhenOnlyRetentionChanges(t *testing.T) {
	t.Parallel()
	next := time.Now().UTC().Add(12 * time.Hour)
	repo := &scheduleTestRepository{schedule: &backupdomain.Schedule{
		Enabled: true, IntervalHours: 24, RetentionCount: 7, NextRunAt: &next,
	}}
	svc := &Service{repo: repo}

	updated, err := svc.UpdateSchedule(context.Background(), true, 24, 12)
	if err != nil {
		t.Fatal(err)
	}
	if updated.NextRunAt == nil || !updated.NextRunAt.Equal(next) {
		t.Fatalf("retention-only update changed next run: want %v, got %v", next, updated.NextRunAt)
	}
	if updated.RetentionCount != 12 {
		t.Fatalf("expected retention 12, got %d", updated.RetentionCount)
	}
}

func TestUpdateScheduleValidatesBounds(t *testing.T) {
	t.Parallel()
	svc := &Service{}
	if _, err := svc.UpdateSchedule(context.Background(), true, 0, 7); err == nil {
		t.Fatal("expected interval validation error")
	}
	if _, err := svc.UpdateSchedule(context.Background(), true, 24, 101); err == nil {
		t.Fatal("expected retention validation error")
	}
}
