/**
 * Shared formatting utilities.
 */

/** Format an ISO date string to a locale string. Returns '-' for falsy/invalid input. */
export function formatDate(value?: string): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

/** Format an ISO date string using zh-CN locale with year/month/day/hour/minute. */
export function formatDateZhCN(value?: string): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

/** Format byte count to a human-readable string (e.g. "1.5 MB"). */
export function formatBytes(bytes?: number | string, decimals = 2): string {
  const b = typeof bytes === 'string' ? parseFloat(bytes) : (bytes ?? 0)
  if (isNaN(b) || b <= 0) return '0 B'
  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return `${parseFloat((b / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
}

/** Alias for formatBytes for backward compatibility. */
export const formatFileSize = formatBytes
