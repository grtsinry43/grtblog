package backup

import "time"

const ArchiveFormatVersion = 1

type Manifest struct {
	FormatVersion     int               `json:"formatVersion"`
	BackupID          string            `json:"backupId"`
	CreatedAt         time.Time         `json:"createdAt"`
	AppVersion        string            `json:"appVersion"`
	MigrationVersion  int64             `json:"migrationVersion"`
	DBServerVersion   string            `json:"dbServerVersion"`
	PGDumpVersion     string            `json:"pgDumpVersion"`
	SiteName          string            `json:"siteName"`
	SiteURL           string            `json:"siteUrl"`
	UploadFileCount   int64             `json:"uploadFileCount"`
	ContainsSensitive bool              `json:"containsSensitive"`
	Checksums         map[string]string `json:"checksums"`
}
