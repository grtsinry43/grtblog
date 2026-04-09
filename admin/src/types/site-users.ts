export interface SiteUser {
  id: number
  username: string
  nickname: string
  email: string
  avatar: string
  isActive: boolean
  isAdmin: boolean
  createdAt: string
  updatedAt: string
  deletedAt?: string
}

export interface SiteUserListResponse {
  items: SiteUser[]
  total: number
  page: number
  size: number
}

export interface SiteUserListParams {
  keyword?: string
  admin?: boolean
  active?: boolean
  page?: number
  pageSize?: number
}

export interface UpdateSiteUserPayload {
  nickname: string
  email: string
  isActive: boolean
  isAdmin: boolean
}
