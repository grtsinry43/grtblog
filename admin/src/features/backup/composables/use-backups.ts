import { onActivated, onBeforeUnmount, onDeactivated, onMounted, shallowRef } from 'vue'

import {
  createBackup,
  createBackupDownloadTicket,
  deleteBackup,
  getBackupSchedule,
  listBackups,
  setBackupPinned,
  updateBackupSchedule,
} from '../api/backup-api'

import type { BackupRecord, BackupSchedule, UpdateBackupScheduleRequest } from '../model/types'

export function useBackups() {
  const records = shallowRef<BackupRecord[]>([])
  const loading = shallowRef(false)
  const creating = shallowRef(false)
  const deletingId = shallowRef<string | null>(null)
  const schedule = shallowRef<BackupSchedule | null>(null)
  const scheduleLoading = shallowRef(false)
  const scheduleSaving = shallowRef(false)
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
    refresh,
    create,
    remove,
    download,
    loadSchedule,
    saveSchedule,
    setPinned,
  }
}
