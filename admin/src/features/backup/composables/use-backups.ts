import { onActivated, onBeforeUnmount, onDeactivated, onMounted, shallowRef } from 'vue'

import {
  createBackup,
  createBackupDownloadTicket,
  deleteBackup,
  listBackups,
} from '../api/backup-api'

import type { BackupRecord } from '../model/types'

export function useBackups() {
  const records = shallowRef<BackupRecord[]>([])
  const loading = shallowRef(false)
  const creating = shallowRef(false)
  const deletingId = shallowRef<string | null>(null)
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
    startPolling()
  })
  onActivated(startPolling)
  onDeactivated(stopPolling)
  onBeforeUnmount(stopPolling)

  return { records, loading, creating, deletingId, refresh, create, remove, download }
}
