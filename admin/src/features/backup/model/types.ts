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
