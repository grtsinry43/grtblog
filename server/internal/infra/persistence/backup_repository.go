package persistence

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	backupdomain "github.com/grtsinry43/grtblog-v2/server/internal/domain/backup"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type BackupRepository struct{ db *gorm.DB }

func NewBackupRepository(db *gorm.DB) *BackupRepository { return &BackupRepository{db: db} }

func (r *BackupRepository) Create(ctx context.Context, item *backupdomain.Record) error {
	rec := backupRecordToModel(*item)
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *BackupRepository) Update(ctx context.Context, item *backupdomain.Record) error {
	rec := backupRecordToModel(*item)
	return r.db.WithContext(ctx).Save(&rec).Error
}

func (r *BackupRepository) Get(ctx context.Context, id string) (*backupdomain.Record, error) {
	var rec model.BackupRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, backupdomain.ErrNotFound
		}
		return nil, err
	}
	item := backupRecordFromModel(rec)
	return &item, nil
}

func (r *BackupRepository) List(ctx context.Context) ([]backupdomain.Record, error) {
	var records []model.BackupRecord
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]backupdomain.Record, len(records))
	for i, rec := range records {
		items[i] = backupRecordFromModel(rec)
	}
	return items, nil
}

func (r *BackupRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.BackupRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return backupdomain.ErrNotFound
	}
	return nil
}

func (r *BackupRepository) MarkInterrupted(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&model.BackupRecord{}).
		Where("status IN ?", []string{string(backupdomain.StatusQueued), string(backupdomain.StatusRunning)}).
		Updates(map[string]any{
			"status": string(backupdomain.StatusFailed), "stage": "interrupted",
			"error_message": "备份任务因服务重启而中断", "completed_at": time.Now().UTC(),
		}).Error
}

func (r *BackupRepository) CreateTicket(ctx context.Context, ticket backupdomain.DownloadTicket) error {
	rec := model.BackupDownloadTicket{TokenHash: ticket.TokenHash, BackupID: ticket.BackupID, ExpiresAt: ticket.ExpiresAt, CreatedAt: ticket.CreatedAt}
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *BackupRepository) ResolveTicket(ctx context.Context, tokenHash string) (*backupdomain.Record, error) {
	var rec model.BackupRecord
	err := r.db.WithContext(ctx).Table("backup_ops.backup_record AS b").
		Select("b.*").
		Joins("JOIN backup_ops.download_ticket AS t ON t.backup_id = b.id").
		Where("t.token_hash = ? AND t.expires_at > ? AND b.status = ?", tokenHash, time.Now().UTC(), backupdomain.StatusCompleted).
		First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, backupdomain.ErrInvalidTicket
		}
		return nil, err
	}
	item := backupRecordFromModel(rec)
	return &item, nil
}

func (r *BackupRepository) DeleteExpiredTickets(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at <= ?", time.Now().UTC()).Delete(&model.BackupDownloadTicket{}).Error
}

func backupRecordToModel(item backupdomain.Record) model.BackupRecord {
	return model.BackupRecord{
		ID: item.ID, Filename: item.Filename, Status: string(item.Status), Stage: item.Stage,
		TriggerType: item.TriggerType, SizeBytes: item.SizeBytes, SHA256: item.SHA256,
		AppVersion: item.AppVersion, MigrationVersion: item.MigrationVersion,
		DBServerVersion: item.DBServerVersion, SiteName: item.SiteName, SiteURL: item.SiteURL,
		UploadFileCount: item.UploadFileCount, ErrorMessage: item.ErrorMessage, Pinned: item.Pinned,
		CreatedAt: item.CreatedAt, StartedAt: item.StartedAt, CompletedAt: item.CompletedAt,
	}
}

func backupRecordFromModel(rec model.BackupRecord) backupdomain.Record {
	return backupdomain.Record{
		ID: rec.ID, Filename: rec.Filename, Status: backupdomain.Status(rec.Status), Stage: rec.Stage,
		TriggerType: rec.TriggerType, SizeBytes: rec.SizeBytes, SHA256: rec.SHA256,
		AppVersion: rec.AppVersion, MigrationVersion: rec.MigrationVersion,
		DBServerVersion: rec.DBServerVersion, SiteName: rec.SiteName, SiteURL: rec.SiteURL,
		UploadFileCount: rec.UploadFileCount, ErrorMessage: rec.ErrorMessage, Pinned: rec.Pinned,
		CreatedAt: rec.CreatedAt, StartedAt: rec.StartedAt, CompletedAt: rec.CompletedAt,
	}
}
