package backup

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	mediaapp "github.com/grtsinry43/grtblog-v2/server/internal/app/media"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/buildinfo"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	backupdomain "github.com/grtsinry43/grtblog-v2/server/internal/domain/backup"
)

type Service struct {
	cfg       config.BackupConfig
	db        *gorm.DB
	repo      backupdomain.Repository
	dumper    PostgresDumper
	sysConfig *sysconfig.Service
	mediaGate *mediaapp.MutationGate
	rootCtx   context.Context
	mu        sync.Mutex
}

func NewService(rootCtx context.Context, cfg config.BackupConfig, db *gorm.DB, repo backupdomain.Repository, dumper PostgresDumper, sysConfig *sysconfig.Service, mediaGate *mediaapp.MutationGate) *Service {
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	return &Service{cfg: cfg, db: db, repo: repo, dumper: dumper, sysConfig: sysConfig, mediaGate: mediaGate, rootCtx: rootCtx}
}

func (s *Service) Initialize(ctx context.Context) error {
	if err := os.MkdirAll(s.cfg.RootDir, 0o700); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(s.cfg.RootDir, ".work"), 0o700); err != nil {
		return err
	}
	if err := s.repo.MarkInterrupted(ctx); err != nil {
		return err
	}
	return s.repo.DeleteExpiredTickets(ctx)
}

func (s *Service) CreateManual(ctx context.Context) (*backupdomain.Record, error) {
	if !s.mu.TryLock() {
		return nil, backupdomain.ErrBackupRunning
	}
	return s.createLocked(ctx, "manual")
}

func (s *Service) createLocked(ctx context.Context, triggerType string) (*backupdomain.Record, error) {
	now := time.Now().UTC()
	id := uuid.NewString()
	item := &backupdomain.Record{
		ID: id, Filename: fmt.Sprintf("grtblog-backup-%s-%s.tar.gz", now.Format("20060102T150405Z"), id[:8]),
		Status: backupdomain.StatusQueued, Stage: "queued", TriggerType: triggerType, CreatedAt: now,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		s.mu.Unlock()
		return nil, err
	}
	go func() {
		defer s.mu.Unlock()
		timeout := s.cfg.CommandTimeout
		if timeout <= 0 {
			timeout = 30 * time.Minute
		}
		jobCtx, cancel := context.WithTimeout(s.rootCtx, timeout)
		defer cancel()
		s.run(jobCtx, item)
	}()
	return item, nil
}

func (s *Service) RunScheduler(ctx context.Context) {
	s.runScheduledIfDue(ctx)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runScheduledIfDue(ctx)
		}
	}
}

func (s *Service) runScheduledIfDue(ctx context.Context) {
	if !s.mu.TryLock() {
		return
	}
	claimed, _, err := s.repo.TryClaimSchedule(ctx, time.Now().UTC())
	if err != nil {
		s.mu.Unlock()
		log.Printf("[backup] claim scheduled run failed: %v", err)
		return
	}
	if !claimed {
		s.mu.Unlock()
		return
	}
	if _, err := s.createLocked(ctx, "scheduled"); err != nil {
		log.Printf("[backup] create scheduled backup failed: %v", err)
	}
}

func (s *Service) run(ctx context.Context, item *backupdomain.Record) {
	now := time.Now().UTC()
	item.Status, item.Stage, item.StartedAt = backupdomain.StatusRunning, "preparing", &now
	if err := s.repo.Update(context.Background(), item); err != nil {
		return
	}
	fail := func(err error) {
		completed := time.Now().UTC()
		item.Status, item.Stage, item.CompletedAt = backupdomain.StatusFailed, "failed", &completed
		item.ErrorMessage = err.Error()
		_ = s.repo.Update(context.Background(), item)
	}

	workDir := filepath.Join(s.cfg.RootDir, ".work", item.ID)
	if err := os.MkdirAll(workDir, 0o700); err != nil {
		fail(err)
		return
	}
	defer os.RemoveAll(workDir)
	dumpPath := filepath.Join(workDir, "site.dump")
	uploadsPath := filepath.Join(workDir, "uploads")

	info, _ := s.siteInfo(ctx)
	item.SiteName, item.SiteURL = info["website_name"], info["public_url"]
	item.AppVersion = buildinfo.Version()
	if err := s.loadDatabaseMetadata(ctx, item); err != nil {
		fail(err)
		return
	}

	sqlDB, err := s.db.DB()
	if err != nil {
		fail(err)
		return
	}
	var tx *sql.Tx
	var snapshot string
	item.Stage = "snapshotting_uploads"
	_ = s.repo.Update(context.Background(), item)
	checksums := make(map[string]string)
	var uploadCount int64
	if err := s.mediaGate.WithSnapshot(func() error {
		var beginErr error
		tx, beginErr = sqlDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
		if beginErr != nil {
			return beginErr
		}
		if snapshotErr := tx.QueryRowContext(ctx, "SELECT pg_export_snapshot()").Scan(&snapshot); snapshotErr != nil {
			_ = tx.Rollback()
			tx = nil
			return fmt.Errorf("export database snapshot: %w", snapshotErr)
		}
		var copyErr error
		checksums, uploadCount, copyErr = snapshotUploads(ctx, s.cfg.UploadDir, uploadsPath)
		if copyErr != nil {
			_ = tx.Rollback()
			tx = nil
		}
		return copyErr
	}); err != nil {
		fail(fmt.Errorf("snapshot uploads: %w", err))
		return
	}
	defer tx.Rollback()
	item.UploadFileCount = uploadCount
	item.Stage = "dumping_database"
	_ = s.repo.Update(context.Background(), item)
	pgDumpVersion, err := s.dumper.Dump(ctx, snapshot, dumpPath)
	if err != nil {
		fail(err)
		return
	}
	_ = tx.Rollback()

	manifest := Manifest{
		FormatVersion: ArchiveFormatVersion, BackupID: item.ID, CreatedAt: item.CreatedAt,
		AppVersion: item.AppVersion, MigrationVersion: item.MigrationVersion,
		DBServerVersion: item.DBServerVersion, PGDumpVersion: pgDumpVersion,
		SiteName: item.SiteName, SiteURL: item.SiteURL, UploadFileCount: uploadCount,
		ContainsSensitive: true, Checksums: checksums,
	}
	item.Stage = "packing_archive"
	_ = s.repo.Update(context.Background(), item)
	tempArchive := filepath.Join(s.cfg.RootDir, ".work", item.ID+".tar.gz")
	if err := writeArchive(ctx, tempArchive, dumpPath, uploadsPath, manifest); err != nil {
		fail(fmt.Errorf("pack backup archive: %w", err))
		return
	}
	finalPath := s.archivePath(item.Filename)
	if err := os.Rename(tempArchive, finalPath); err != nil {
		fail(fmt.Errorf("publish backup archive: %w", err))
		return
	}
	stat, err := os.Stat(finalPath)
	if err != nil {
		fail(err)
		return
	}
	archiveSum, err := hashPath(finalPath)
	if err != nil {
		fail(err)
		return
	}
	completed := time.Now().UTC()
	item.Status, item.Stage, item.CompletedAt = backupdomain.StatusCompleted, "completed", &completed
	item.SizeBytes, item.SHA256, item.ErrorMessage = stat.Size(), archiveSum, ""
	_ = s.repo.Update(context.Background(), item)
	if item.TriggerType == "scheduled" {
		if err := s.pruneScheduled(context.Background()); err != nil {
			log.Printf("[backup] retention cleanup failed: %v", err)
		}
	}
}

func (s *Service) List(ctx context.Context) ([]backupdomain.Record, error) { return s.repo.List(ctx) }
func (s *Service) Get(ctx context.Context, id string) (*backupdomain.Record, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item.Status == backupdomain.StatusRunning || item.Status == backupdomain.StatusQueued {
		return backupdomain.ErrBackupRunning
	}
	if err := os.Remove(s.archivePath(item.Filename)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) IssueDownloadTicket(ctx context.Context, id string) (string, time.Time, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return "", time.Time{}, err
	}
	if item.Status != backupdomain.StatusCompleted {
		return "", time.Time{}, errors.New("backup is not ready for download")
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	expires := time.Now().UTC().Add(s.cfg.TicketTTL)
	if err := s.repo.CreateTicket(ctx, backupdomain.DownloadTicket{TokenHash: tokenHash(token), BackupID: id, ExpiresAt: expires, CreatedAt: time.Now().UTC()}); err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

func (s *Service) ResolveDownload(ctx context.Context, token string) (*backupdomain.Record, string, error) {
	item, err := s.repo.ResolveTicket(ctx, tokenHash(token))
	if err != nil {
		return nil, "", err
	}
	path := s.archivePath(item.Filename)
	if _, err := os.Stat(path); err != nil {
		return nil, "", err
	}
	return item, path, nil
}

func (s *Service) GetSchedule(ctx context.Context) (*backupdomain.Schedule, error) {
	return s.repo.GetSchedule(ctx)
}

func (s *Service) UpdateSchedule(ctx context.Context, enabled bool, intervalHours, retentionCount int) (*backupdomain.Schedule, error) {
	if intervalHours < 1 || intervalHours > 8760 {
		return nil, errors.New("backup interval must be between 1 and 8760 hours")
	}
	if retentionCount < 1 || retentionCount > 100 {
		return nil, errors.New("backup retention must be between 1 and 100")
	}
	current, err := s.repo.GetSchedule(ctx)
	if err != nil {
		return nil, err
	}
	wasEnabled := current.Enabled
	previousInterval := current.IntervalHours
	current.Enabled = enabled
	current.IntervalHours = intervalHours
	current.RetentionCount = retentionCount
	if enabled && (!wasEnabled || previousInterval != intervalHours || current.NextRunAt == nil) {
		next := time.Now().UTC().Add(time.Duration(intervalHours) * time.Hour)
		current.NextRunAt = &next
	} else if !enabled {
		current.NextRunAt = nil
	}
	if err := s.repo.SaveSchedule(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) SetPinned(ctx context.Context, id string, pinned bool) error {
	return s.repo.SetPinned(ctx, id, pinned)
}

func (s *Service) pruneScheduled(ctx context.Context) error {
	schedule, err := s.repo.GetSchedule(ctx)
	if err != nil {
		return err
	}
	items, err := s.repo.List(ctx)
	if err != nil {
		return err
	}
	retained := 0
	for i := range items {
		item := items[i]
		if item.TriggerType != "scheduled" || item.Status != backupdomain.StatusCompleted || item.Pinned {
			continue
		}
		retained++
		if retained <= schedule.RetentionCount {
			continue
		}
		if err := s.Delete(ctx, item.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) archivePath(filename string) string {
	return filepath.Join(s.cfg.RootDir, filepath.Base(filename))
}

func (s *Service) siteInfo(ctx context.Context) (map[string]string, error) {
	if s.sysConfig == nil {
		return map[string]string{}, nil
	}
	return s.sysConfig.WebsiteInfo(ctx)
}

func (s *Service) loadDatabaseMetadata(ctx context.Context, item *backupdomain.Record) error {
	if err := s.db.WithContext(ctx).Raw("SELECT version()").Scan(&item.DBServerVersion).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Raw("SELECT COALESCE(MAX(version_id), 0) FROM goose_db_version WHERE is_applied = TRUE").Scan(&item.MigrationVersion).Error
}

func tokenHash(token string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(h[:])
}
