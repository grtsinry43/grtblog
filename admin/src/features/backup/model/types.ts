export type BackupStatus = 'queued' | 'running' | 'completed' | 'failed'

export interface BackupRecord {
  id: string
  filename: string
  status: BackupStatus
  stage: string
  triggerType: string
  sizeBytes: number
  sha256?: string
  appVersion?: string
  migrationVersion: number
  dbServerVersion?: string
  siteName?: string
  siteUrl?: string
  uploadFileCount: number
  errorMessage?: string
  pinned: boolean
  createdAt: string
  startedAt?: string
  completedAt?: string
}

export interface BackupDownloadTicket {
  url: string
  expiresAt: string
}

export interface BackupSchedule {
  enabled: boolean
  intervalHours: number
  retentionCount: number
  nextRunAt?: string
  lastRunAt?: string
  updatedAt: string
}

export interface UpdateBackupScheduleRequest {
  enabled: boolean
  intervalHours: number
  retentionCount: number
}

export type RestoreState = 'idle' | 'pending_restart' | 'running' | 'succeeded' | 'failed'

export interface RestoreStatus {
  state: RestoreState
  requestId?: string
  backupId?: string
  archiveFilename?: string
  message?: string
  requestedAt?: string
  startedAt?: string
  completedAt?: string
}

export interface ImportRestoreResult {
  backup: BackupRecord
  restore: RestoreStatus
}
