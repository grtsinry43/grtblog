import { request } from '@/services/http'

import type { BackupDownloadTicket, BackupRecord } from '../model/types'

const basePath = '/admin/backups'

export function listBackups() {
  return request<BackupRecord[]>(basePath)
}

export function createBackup() {
  return request<BackupRecord>(basePath, { method: 'POST' })
}

export function deleteBackup(id: string) {
  return request<{ id: string }>(`${basePath}/${encodeURIComponent(id)}`, { method: 'DELETE' })
}

export function createBackupDownloadTicket(id: string) {
  return request<BackupDownloadTicket>(`${basePath}/${encodeURIComponent(id)}/download-ticket`, {
    method: 'POST',
  })
}
