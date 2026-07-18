package backup

import "time"

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type Record struct {
	ID               string     `json:"id"`
	Filename         string     `json:"filename"`
	Status           Status     `json:"status"`
	Stage            string     `json:"stage"`
	TriggerType      string     `json:"triggerType"`
	SizeBytes        int64      `json:"sizeBytes"`
	SHA256           string     `json:"sha256,omitempty"`
	AppVersion       string     `json:"appVersion,omitempty"`
	MigrationVersion int64      `json:"migrationVersion"`
	DBServerVersion  string     `json:"dbServerVersion,omitempty"`
	SiteName         string     `json:"siteName,omitempty"`
	SiteURL          string     `json:"siteUrl,omitempty"`
	UploadFileCount  int64      `json:"uploadFileCount"`
	ErrorMessage     string     `json:"errorMessage,omitempty"`
	Pinned           bool       `json:"pinned"`
	CreatedAt        time.Time  `json:"createdAt"`
	StartedAt        *time.Time `json:"startedAt,omitempty"`
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
}

type DownloadTicket struct {
	TokenHash string
	BackupID  string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Schedule struct {
	Enabled        bool       `json:"enabled"`
	IntervalHours  int        `json:"intervalHours"`
	RetentionCount int        `json:"retentionCount"`
	NextRunAt      *time.Time `json:"nextRunAt,omitempty"`
	LastRunAt      *time.Time `json:"lastRunAt,omitempty"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}
