import { request } from './http'

export interface PhotoExif {
  make?: string
  model?: string
  lensModel?: string
  focalLength?: string
  fNumber?: string
  exposureTime?: string
  iso?: number
  gpsLatitude?: number
  gpsLongitude?: number
  gpsAltitude?: number
  dateTimeOriginal?: string
  imageWidth?: number
  imageHeight?: number
  orientation?: number
  dominantColor?: string
  [key: string]: unknown
}

export interface PhotoItem {
  id: number
  albumId?: number
  url: string
  thumbnailUrl?: string
  description?: string | null
  caption?: string | null
  exif?: PhotoExif | null
  sortOrder: number
  createdAt: string
}

export interface AlbumListItem {
  id: number
  title: string
  description?: string | null
  cover?: string | null
  shortUrl: string
  isPublished: boolean
  photoCount: number
  views: number
  likes: number
  comments: number
  createdAt: string
  updatedAt: string
}

export interface AlbumListResponse {
  items: AlbumListItem[]
  total: number
  page: number
  size: number
}

export interface AlbumDetail {
  id: number
  title: string
  description?: string | null
  cover?: string | null
  shortUrl: string
  authorId: number
  commentAreaId?: number | null
  isPublished: boolean
  allowComment: boolean
  photoCount: number
  metrics?: { views: number; likes: number; comments: number } | null
  photos: PhotoItem[]
  createdAt: string
  updatedAt: string
}

export interface ListAlbumsParams {
  page?: number
  pageSize?: number
  published?: boolean
  search?: string
}

export interface CreateAlbumPayload {
  title: string
  description?: string | null
  cover?: string | null
  shortUrl?: string | null
  isPublished: boolean
  allowComment?: boolean
  createdAt?: string | null
}

export interface UpdateAlbumPayload {
  title: string
  description?: string | null
  cover?: string | null
  shortUrl: string
  isPublished: boolean
  allowComment?: boolean
}

export interface CreatePhotoPayload {
  url: string
  description?: string | null
  caption?: string | null
  exif?: PhotoExif | null
  sortOrder: number
}

export interface UpdatePhotoPayload {
  url: string
  description?: string | null
  caption?: string | null
  exif?: PhotoExif | null
  sortOrder: number
}

function stripEmpty<T extends object>(value: T): Record<string, unknown> {
  return Object.fromEntries(
    Object.entries(value).filter(
      ([, entry]) => entry !== undefined && entry !== null && entry !== '',
    ),
  )
}

export function listAlbums(params: ListAlbumsParams) {
  return request<AlbumListResponse>('/admin/albums', {
    method: 'GET',
    query: stripEmpty(params),
  })
}

export function getAlbum(id: number) {
  return request<AlbumDetail>(`/admin/albums/${id}`, {
    method: 'GET',
  })
}

export function createAlbum(payload: CreateAlbumPayload) {
  return request<AlbumDetail>('/albums', {
    method: 'POST',
    body: payload,
  })
}

export function updateAlbum(id: number, payload: UpdateAlbumPayload) {
  return request<AlbumDetail>(`/albums/${id}`, {
    method: 'PUT',
    body: payload,
  })
}

export function deleteAlbum(id: number) {
  return request<void>(`/albums/${id}`, {
    method: 'DELETE',
  })
}

export function batchSetAlbumPublished(payload: { ids: number[]; isPublished: boolean }) {
  return request<void>('/admin/albums/published', {
    method: 'PUT',
    body: payload,
  })
}

export function batchDeleteAlbums(payload: { ids: number[] }) {
  return request<void>('/admin/albums/batch-delete', {
    method: 'POST',
    body: payload,
  })
}

export function addPhotos(albumId: number, payload: { photos: CreatePhotoPayload[] }) {
  return request<PhotoItem[]>(`/albums/${albumId}/photos`, {
    method: 'POST',
    body: payload,
  })
}

export function updatePhoto(albumId: number, photoId: number, payload: UpdatePhotoPayload) {
  return request<PhotoItem>(`/albums/${albumId}/photos/${photoId}`, {
    method: 'PUT',
    body: payload,
  })
}

export function deletePhoto(albumId: number, photoId: number) {
  return request<void>(`/albums/${albumId}/photos/${photoId}`, {
    method: 'DELETE',
  })
}

export function reorderPhotos(albumId: number, payload: { photoIds: number[] }) {
  return request<void>(`/albums/${albumId}/photos/reorder`, {
    method: 'PUT',
    body: payload,
  })
}
