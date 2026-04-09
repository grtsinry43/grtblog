import { request } from './http'

import type {
  FriendLink,
  FriendLinkApplication,
  FriendLinkCreateReq,
  FriendLinkFederationRequestReq,
  FriendLinkFederationRequestResp,
  FriendLinkListAppsParams,
  FriendLinkListSyncJobsParams,
  FriendLinkListParams,
  FriendLinkSyncJob,
  FriendLinkUpdateReq,
} from '@/types/friend-link'

export interface ListResponse<T> {
  items: T[]
  total: number
  page: number
  size: number
}

// 友链管理相关 API
export const friendLinkService = {
  // 获取友链列表
  getFriendLinks: (params: FriendLinkListParams) => {
    return request<ListResponse<FriendLink>>('/admin/friend-links', {
      method: 'GET',
      query: params,
    })
  },

  // 创建友链
  createFriendLink: (data: FriendLinkCreateReq) => {
    return request<FriendLink>('/admin/friend-links', {
      method: 'POST',
      body: data,
    })
  },

  // 以联合协议发起对外友链申请
  requestFederationFriendLink: (payload: FriendLinkFederationRequestReq) => {
    return request<FriendLinkFederationRequestResp>('/admin/friend-links/federation/request', {
      method: 'POST',
      body: payload,
    })
  },

  // 更新友链
  updateFriendLink: (id: number, data: FriendLinkUpdateReq) => {
    return request<FriendLink>(`/admin/friend-links/${id}`, {
      method: 'PUT',
      body: data,
    })
  },

  // 删除友链
  deleteFriendLink: (id: number) => {
    return request<any>(`/admin/friend-links/${id}`, {
      method: 'DELETE',
    })
  },

  // 获取申请列表
  getApplications: (params: FriendLinkListAppsParams) => {
    return request<ListResponse<FriendLinkApplication>>('/admin/friend-links/applications', {
      method: 'GET',
      query: params,
    })
  },

  // 获取同步作业列表
  getSyncJobs: (params: FriendLinkListSyncJobsParams) => {
    return request<ListResponse<FriendLinkSyncJob>>('/admin/friend-links/sync-jobs', {
      method: 'GET',
      query: params,
    })
  },

  // 审核通过申请
  approveApplication: (id: number) => {
    return request<FriendLinkApplication>(`/admin/friend-links/applications/${id}/approve`, {
      method: 'PUT',
    })
  },

  // 拒绝申请
  rejectApplication: (id: number) => {
    return request<FriendLinkApplication>(`/admin/friend-links/applications/${id}/reject`, {
      method: 'PUT',
    })
  },

  // 封禁申请
  blockApplication: (id: number) => {
    return request<FriendLinkApplication>(`/admin/friend-links/applications/${id}/block`, {
      method: 'PUT',
    })
  },

  // 变更申请状态
  updateApplicationStatus: (id: number, status: string) => {
    return request<FriendLinkApplication>(`/admin/friend-links/applications/${id}/status`, {
      method: 'PUT',
      body: { status },
    })
  },

  // 封禁友链
  blockFriendLink: (id: number) => {
    return request<FriendLink>(`/admin/friend-links/${id}/block`, {
      method: 'PUT',
    })
  },
}
