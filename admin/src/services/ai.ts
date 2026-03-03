import { request, getAuthToken, API_BASE_URL } from './http'

// ── Types ──

export interface AIProvider {
  id: number
  name: string
  type: 'openai' | 'openrouter' | 'gemini'
  apiUrl: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface AIModel {
  id: number
  providerId: number
  providerName?: string
  providerType?: string
  name: string
  modelId: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateAIProviderReq {
  name: string
  type: string
  apiUrl: string
  apiKey: string
  isActive?: boolean
}

export interface UpdateAIProviderReq {
  name?: string
  type?: string
  apiUrl?: string
  apiKey?: string
  isActive?: boolean
}

export interface CreateAIModelReq {
  providerId: number
  name: string
  modelId: string
  isActive?: boolean
}

export interface UpdateAIModelReq {
  providerId?: number
  name?: string
  modelId?: string
  isActive?: boolean
}

export interface ModerationResult {
  approved: boolean
  reason: string
  score: number
}

export interface TitleResult {
  title: string
  shortUrl: string
}

export interface RewriteResult {
  content: string
}

// ── TaskLog ──

export interface AITaskLog {
  id: number
  taskType: string
  modelName: string
  providerName: string
  status: string
  inputText?: string
  outputText?: string
  errorMessage?: string
  durationMs: number
  triggerSource: string
  createdAt: string
  updatedAt: string
}

export interface AITaskLogListResp {
  items: AITaskLog[]
  total: number
  page: number
  size: number
}

export interface AITaskLogListParams {
  page?: number
  pageSize?: number
  taskType?: string
  status?: string
  search?: string
}

// ── Provider CRUD ──

export function listAIProviders() {
  return request<AIProvider[]>('/admin/ai/providers')
}

export function createAIProvider(data: CreateAIProviderReq) {
  return request<AIProvider>('/admin/ai/providers', {
    method: 'POST',
    body: data,
  })
}

export function updateAIProvider(id: number, data: UpdateAIProviderReq) {
  return request<AIProvider>(`/admin/ai/providers/${id}`, {
    method: 'PUT',
    body: data,
  })
}

export function deleteAIProvider(id: number) {
  return request<null>(`/admin/ai/providers/${id}`, {
    method: 'DELETE',
  })
}

// ── Model CRUD ──

export function listAIModels() {
  return request<AIModel[]>('/admin/ai/models')
}

export function createAIModel(data: CreateAIModelReq) {
  return request<AIModel>('/admin/ai/models', {
    method: 'POST',
    body: data,
  })
}

export function updateAIModel(id: number, data: UpdateAIModelReq) {
  return request<AIModel>(`/admin/ai/models/${id}`, {
    method: 'PUT',
    body: data,
  })
}

export function deleteAIModel(id: number) {
  return request<null>(`/admin/ai/models/${id}`, {
    method: 'DELETE',
  })
}

// ── AI 功能 ──

export function moderateComment(content: string) {
  return request<ModerationResult>('/admin/ai/moderate-comment', {
    method: 'POST',
    body: { content },
  })
}

export function generateTitle(content: string) {
  return request<TitleResult>('/admin/ai/generate-title', {
    method: 'POST',
    body: { content },
  })
}

export function rewriteContent(content: string, instruction: string) {
  return request<RewriteResult>('/admin/ai/rewrite-content', {
    method: 'POST',
    body: { content, instruction },
  })
}

export async function generateSummaryStream(
  content: string,
  onChunk: (text: string) => void,
): Promise<void> {
  const token = getAuthToken()
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const resp = await fetch(`${API_BASE_URL}/admin/ai/generate-summary/stream`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ content }),
  })

  if (!resp.ok) {
    throw new Error(`请求失败（${resp.status}）`)
  }

  const reader = resp.body?.getReader()
  if (!reader) throw new Error('无法读取响应流')

  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''

    for (const line of lines) {
      if (!line.startsWith('data: ')) continue
      const data = line.slice(6)
      if (data === '[DONE]') return
      try {
        const parsed = JSON.parse(data) as { content?: string; error?: string }
        if (parsed.error) throw new Error(parsed.error)
        if (parsed.content) onChunk(parsed.content)
      } catch (e) {
        if (e instanceof Error && e.message !== data) throw e
      }
    }
  }
}

export async function rewriteContentStream(
  content: string,
  instruction: string,
  onChunk: (text: string) => void,
): Promise<void> {
  const token = getAuthToken()
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const resp = await fetch(`${API_BASE_URL}/admin/ai/rewrite-content/stream`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ content, instruction }),
  })

  if (!resp.ok) {
    throw new Error(`请求失败（${resp.status}）`)
  }

  const reader = resp.body?.getReader()
  if (!reader) throw new Error('无法读取响应流')

  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''

    for (const line of lines) {
      if (!line.startsWith('data: ')) continue
      const data = line.slice(6)
      if (data === '[DONE]') return
      try {
        const parsed = JSON.parse(data) as { content?: string; error?: string }
        if (parsed.error) throw new Error(parsed.error)
        if (parsed.content) onChunk(parsed.content)
      } catch (e) {
        if (e instanceof Error && e.message !== data) throw e
      }
    }
  }
}

// ── TaskLog API ──

export function listAITaskLogs(params: AITaskLogListParams) {
  const query = new URLSearchParams()
  if (params.page) query.set('page', String(params.page))
  if (params.pageSize) query.set('pageSize', String(params.pageSize))
  if (params.taskType) query.set('taskType', params.taskType)
  if (params.status) query.set('status', params.status)
  if (params.search) query.set('search', params.search)
  return request<AITaskLogListResp>(`/admin/ai/task-logs?${query.toString()}`)
}

export function getAITaskLog(id: number) {
  return request<AITaskLog>(`/admin/ai/task-logs/${id}`)
}
