import { request } from '@/services/http'

import type {
  BackupDownloadTicket,
  BackupRecord,
  BackupSchedule,
  ImportRestoreResult,
  RestoreStatus,
  UpdateBackupScheduleRequest,
} from '../model/types'

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

export function getBackupSchedule() {
  return request<BackupSchedule>(`${basePath}/schedule`)
}

export function updateBackupSchedule(payload: UpdateBackupScheduleRequest) {
  return request<BackupSchedule>(`${basePath}/schedule`, { method: 'PUT', body: payload })
}

export function setBackupPinned(id: string, pinned: boolean) {
  return request<{ id: string; pinned: boolean }>(`${basePath}/${encodeURIComponent(id)}/pin`, {
    method: 'PATCH',
    body: { pinned },
  })
}

export function getRestoreStatus() {
  return request<RestoreStatus>(`${basePath}/restore-status`)
}

export function requestBackupRestore(id: string, confirmation: string) {
  return request<RestoreStatus>(`${basePath}/${encodeURIComponent(id)}/restore`, {
    method: 'POST',
    body: { confirmation },
  })
}

export function uploadBackupForRestore(file: File, confirmation: string) {
  const body = new FormData()
  body.append('archive', file)
  body.append('confirmation', confirmation)
  return request<ImportRestoreResult>(`${basePath}/restore-upload`, { method: 'POST', body })
}
