import { API_BASE_URL, ApiError, getAuthToken, request } from './http'

export type FileType = 'picture' | 'file'

export interface UploadImageMeta {
  width?: number
  height?: number
  dominantColor?: string
}

export interface UploadFileResponse {
  id: number
  name: string
  path: string
  publicUrl: string
  thumbnailUrl?: string
  imageMeta?: UploadImageMeta | null
  type: FileType
  size: number
  createdAt: string
  duplicated: boolean
}

export interface UploadFileListResponse {
  items: UploadFileResponse[]
  total: number
  page: number
  size: number
}

export interface UploadSyncResponse {
  scanned: number
  indexed: number
  created: number
  updated: number
  deleted: number
  skippedDuplicates: number
}

export interface ListUploadsParams {
  page?: number
  pageSize?: number
}

export interface RenameFilePayload {
  name: string
}

export interface UploadProgressEvent {
  percent: number
  loaded: number
  total: number
}

/**
 * Upload a file to the server
 * @param file - The file to upload
 * @param type - File type: 'picture' or 'file'
 */
export async function uploadFile(
  file: File,
  type: FileType,
  onProgress?: (event: UploadProgressEvent) => void,
): Promise<UploadFileResponse> {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('type', type)

  if (!onProgress) {
    return request<UploadFileResponse>('/upload', {
      method: 'POST',
      body: formData,
    })
  }

  return new Promise<UploadFileResponse>((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open('POST', `${API_BASE_URL}/upload`)

    const token = getAuthToken()
    if (token) {
      xhr.setRequestHeader('Authorization', `Bearer ${token}`)
    }

    xhr.upload.onprogress = (event) => {
      if (!event.lengthComputable) return
      onProgress({
        percent: Math.round((event.loaded / event.total) * 100),
        loaded: event.loaded,
        total: event.total,
      })
    }

    xhr.onerror = () => {
      reject(new ApiError('网络异常，请稍后重试'))
    }

    xhr.onload = () => {
      const status = xhr.status
      let payload: { code: number; bizErr: string; msg: string; data: UploadFileResponse } | null =
        null

      if (xhr.responseText) {
        try {
          payload = JSON.parse(xhr.responseText)
        } catch {
          reject(new ApiError('无法解析服务端响应', { status }))
          return
        }
      }

      if (status < 200 || status >= 300) {
        reject(
          new ApiError(payload?.msg || `请求失败（${status}）`, {
            code: payload?.code,
            bizErr: payload?.bizErr,
            status,
          }),
        )
        return
      }

      if (!payload) {
        reject(new ApiError('无法解析服务端响应', { status }))
        return
      }

      if (payload.code !== 0) {
        reject(
          new ApiError(payload.msg || payload.bizErr || '请求失败', {
            code: payload.code,
            bizErr: payload.bizErr,
            status,
          }),
        )
        return
      }

      onProgress({ percent: 100, loaded: file.size, total: file.size })
      resolve(payload.data)
    }

    xhr.send(formData)
  })
}

/**
 * List uploaded files with pagination
 * @param params - Query parameters for listing files
 */
export function listUploads(params: ListUploadsParams = {}): Promise<UploadFileListResponse> {
  const query: Record<string, string> = {}

  if (params.page !== undefined) {
    query.page = String(params.page)
  }
  if (params.pageSize !== undefined) {
    query.pageSize = String(params.pageSize)
  }

  return request<UploadFileListResponse>('/uploads', {
    method: 'GET',
    query,
  })
}

/**
 * Rename a file (display name only)
 * @param id - File ID
 * @param payload - New name
 */
export function renameFile(id: number, payload: RenameFilePayload): Promise<UploadFileResponse> {
  return request<UploadFileResponse>(`/upload/${id}`, {
    method: 'PUT',
    body: payload,
  })
}

/**
 * Delete a file
 * @param id - File ID
 */
export function deleteFile(id: number): Promise<null> {
  return request<null>(`/upload/${id}`, {
    method: 'DELETE',
  })
}

/**
 * Download a file
 * @param id - File ID
 * @returns Blob URL for download
 */
export async function downloadFile(id: number, fileName: string): Promise<void> {
  // For download, we need to handle this differently since it returns binary data
  // We'll need to use a direct fetch approach
  const token = localStorage.getItem('token')
  const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '/api/v2').replace(/\/$/, '')

  const response = await fetch(`${API_BASE_URL}/upload/${id}/download`, {
    headers: {
      Authorization: token ? `Bearer ${token}` : '',
    },
  })

  if (!response.ok) {
    throw new Error('Download failed')
  }

  const blob = await response.blob()
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = fileName
  document.body.appendChild(a)
  a.click()
  window.URL.revokeObjectURL(url)
  document.body.removeChild(a)
}

/**
 * Get the full public URL for a file
 * @param path - Virtual path from the upload response
 */
export function getPublicUrl(path: string): string {
  const baseUrl = window.location.origin
  return `${baseUrl}${path}`
}

/**
 * Sync files on disk into the upload index
 */
export function syncUploads(): Promise<UploadSyncResponse> {
  return request<UploadSyncResponse>('/uploads/sync', {
    method: 'POST',
  })
}
