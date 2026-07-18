package model

import "time"

type BackupRecord struct {
	ID               string     `gorm:"column:id;primaryKey"`
	Filename         string     `gorm:"column:filename"`
	Status           string     `gorm:"column:status"`
	Stage            string     `gorm:"column:stage"`
	TriggerType      string     `gorm:"column:trigger_type"`
	SizeBytes        int64      `gorm:"column:size_bytes"`
	SHA256           string     `gorm:"column:sha256"`
	AppVersion       string     `gorm:"column:app_version"`
	MigrationVersion int64      `gorm:"column:migration_version"`
	DBServerVersion  string     `gorm:"column:db_server_version"`
	SiteName         string     `gorm:"column:site_name"`
	SiteURL          string     `gorm:"column:site_url"`
	UploadFileCount  int64      `gorm:"column:upload_file_count"`
	ErrorMessage     string     `gorm:"column:error_message"`
	Pinned           bool       `gorm:"column:pinned"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	StartedAt        *time.Time `gorm:"column:started_at"`
	CompletedAt      *time.Time `gorm:"column:completed_at"`
}

func (BackupRecord) TableName() string { return "backup_ops.backup_record" }

type BackupDownloadTicket struct {
	TokenHash string    `gorm:"column:token_hash;primaryKey"`
	BackupID  string    `gorm:"column:backup_id"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (BackupDownloadTicket) TableName() string { return "backup_ops.download_ticket" }

type BackupScheduleConfig struct {
	ID             int16      `gorm:"column:id;primaryKey"`
	Enabled        bool       `gorm:"column:enabled"`
	IntervalHours  int        `gorm:"column:interval_hours"`
	RetentionCount int        `gorm:"column:retention_count"`
	NextRunAt      *time.Time `gorm:"column:next_run_at"`
	LastRunAt      *time.Time `gorm:"column:last_run_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (BackupScheduleConfig) TableName() string { return "backup_ops.schedule_config" }
