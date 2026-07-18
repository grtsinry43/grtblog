import { onActivated, onBeforeUnmount, onDeactivated, onMounted, shallowRef } from 'vue'

import {
  createBackup,
  createBackupDownloadTicket,
  deleteBackup,
  getBackupSchedule,
  getRestoreStatus,
  listBackups,
  setBackupPinned,
  updateBackupSchedule,
  requestBackupRestore,
  uploadBackupForRestore,
} from '../api/backup-api'

import type {
  BackupRecord,
  BackupSchedule,
  RestoreStatus,
  UpdateBackupScheduleRequest,
} from '../model/types'

export function useBackups() {
  const records = shallowRef<BackupRecord[]>([])
  const loading = shallowRef(false)
  const creating = shallowRef(false)
  const deletingId = shallowRef<string | null>(null)
  const schedule = shallowRef<BackupSchedule | null>(null)
  const scheduleLoading = shallowRef(false)
  const scheduleSaving = shallowRef(false)
  const restoreStatus = shallowRef<RestoreStatus | null>(null)
  const restoring = shallowRef(false)
  let timer: ReturnType<typeof setInterval> | undefined

  async function refresh(silent = false) {
    if (!silent) loading.value = true
    try {
      records.value = await listBackups()
    } finally {
      if (!silent) loading.value = false
    }
  }

  async function create() {
    creating.value = true
    try {
      const item = await createBackup()
      records.value = [item, ...records.value]
      return item
    } finally {
      creating.value = false
    }
  }

  async function remove(id: string) {
    deletingId.value = id
    try {
      await deleteBackup(id)
      records.value = records.value.filter((item) => item.id !== id)
    } finally {
      deletingId.value = null
    }
  }

  async function download(id: string) {
    const ticket = await createBackupDownloadTicket(id)
    window.location.assign(new URL(ticket.url, window.location.origin).toString())
  }

  async function loadSchedule() {
    scheduleLoading.value = true
    try {
      schedule.value = await getBackupSchedule()
    } finally {
      scheduleLoading.value = false
    }
  }

  async function saveSchedule(payload: UpdateBackupScheduleRequest) {
    scheduleSaving.value = true
    try {
      schedule.value = await updateBackupSchedule(payload)
      return schedule.value
    } finally {
      scheduleSaving.value = false
    }
  }

  async function setPinned(id: string, pinned: boolean) {
    await setBackupPinned(id, pinned)
    records.value = records.value.map((item) => (item.id === id ? { ...item, pinned } : item))
  }

  async function loadRestoreStatus() {
    restoreStatus.value = await getRestoreStatus()
  }

  async function restoreExisting(id: string, confirmation: string) {
    restoring.value = true
    try {
      restoreStatus.value = await requestBackupRestore(id, confirmation)
      return restoreStatus.value
    } finally {
      restoring.value = false
    }
  }

  async function restoreUpload(file: File, confirmation: string) {
    restoring.value = true
    try {
      const result = await uploadBackupForRestore(file, confirmation)
      records.value = [result.backup, ...records.value]
      restoreStatus.value = result.restore
      return result.restore
    } finally {
      restoring.value = false
    }
  }

  function startPolling() {
    stopPolling()
    timer = setInterval(() => {
      if (records.value.some((item) => item.status === 'queued' || item.status === 'running')) {
        void refresh(true)
      }
    }, 2000)
  }

  function stopPolling() {
    if (timer) clearInterval(timer)
    timer = undefined
  }

  onMounted(() => {
    void refresh()
    void loadSchedule()
    void loadRestoreStatus()
    startPolling()
  })
  onActivated(startPolling)
  onDeactivated(stopPolling)
  onBeforeUnmount(stopPolling)

  return {
    records,
    loading,
    creating,
    deletingId,
    schedule,
    scheduleLoading,
    scheduleSaving,
    restoreStatus,
    restoring,
    refresh,
    create,
    remove,
    download,
    loadSchedule,
    saveSchedule,
    setPinned,
    loadRestoreStatus,
    restoreExisting,
    restoreUpload,
  }
}
