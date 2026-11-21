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

export interface WritingIssue {
  original: string
  suggestion: string
  reason: string
  severity: 'typo' | 'style'
}

export interface CheckWritingResp {
  issues: WritingIssue[]
}

export type MarginAnnotationKind = 'refine' | 'spark' | 'structure'

export interface MarginAnnotation {
  kind: MarginAnnotationKind
  quote?: string
  comment: string
  suggestion?: string
}

export interface AnnotateParagraphResp {
  annotations: MarginAnnotation[]
}

export type DashboardBriefMood = 'growing' | 'steady' | 'quiet' | 'attention'

export interface DashboardBriefHighlight {
  label: string
  detail: string
}

export interface DashboardBriefAction {
  title: string
  reason: string
  route?: string
}

export interface DashboardBriefResp {
  enabled: boolean
  headline?: string
  mood?: DashboardBriefMood
  highlights?: DashboardBriefHighlight[]
  actions?: DashboardBriefAction[]
  generatedAt?: string
  cached?: boolean
}

export interface OpsReportTopicPillar {
  name: string
  strength: 'core' | 'growing' | 'quiet'
  observation: string
}

export interface OpsReportTopicIdea {
  title: string
  angle: string
  why: string
}

export interface OpsReportTopicsSection {
  overview: string
  pillars: OpsReportTopicPillar[]
  ideas: OpsReportTopicIdea[]
}

export interface OpsReportAudienceSection {
  overview: string
  highlights: string[]
  retentionNote: string
}

export interface OpsReportCommentsSection {
  overview: string
  themes: string[]
  suggestion: string
}

export interface OpsReportActionItem {
  priority: 'high' | 'medium' | 'low'
  title: string
  detail: string
  route?: string
}

export interface OpsReportOperationsSection {
  overview: string
  actions: OpsReportActionItem[]
}

export interface OpsReportResp {
  enabled: boolean
  summary?: string
  mood?: DashboardBriefMood
  days?: number
  topics?: OpsReportTopicsSection
  audience?: OpsReportAudienceSection
  comments?: OpsReportCommentsSection
  operations?: OpsReportOperationsSection
  generatedAt?: string
  cached?: boolean
}

export interface OpsReportIdeasResp {
  ideas: OpsReportTopicIdea[]
  generatedAt?: string
}

export interface AIFeatureFlags {
  enabled: boolean
  realtimeAssist: boolean
  ghostCompletion: boolean
  marginNotes: boolean
  inlineLint: boolean
  dashboardBrief: boolean
  opsReport: boolean
  onboardingSeen: boolean
}

export interface UpdateAIFeatureFlagsReq {
  realtimeAssist?: boolean
  ghostCompletion?: boolean
  marginNotes?: boolean
  inlineLint?: boolean
  dashboardBrief?: boolean
  opsReport?: boolean
  onboardingSeen?: boolean
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

async function consumeAIStream(
  path: string,
  body: Record<string, unknown>,
  onChunk: (text: string) => void,
  signal?: AbortSignal,
): Promise<void> {
  const token = getAuthToken()
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const resp = await fetch(`${API_BASE_URL}${path}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
    signal,
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

export async function generateSummaryStream(
  content: string,
  onChunk: (text: string) => void,
  signal?: AbortSignal,
): Promise<void> {
  return consumeAIStream('/admin/ai/generate-summary/stream', { content }, onChunk, signal)
}

export async function rewriteContentStream(
  content: string,
  instruction: string,
  onChunk: (text: string) => void,
  signal?: AbortSignal,
): Promise<void> {
  return consumeAIStream('/admin/ai/rewrite-content/stream', { content, instruction }, onChunk, signal)
}

export async function completeTextStream(
  params: { prefix: string; suffix?: string; title?: string },
  onChunk: (text: string) => void,
  signal?: AbortSignal,
): Promise<void> {
  return consumeAIStream('/admin/ai/complete/stream', params, onChunk, signal)
}

export function checkWriting(content: string, signal?: AbortSignal) {
  return request<CheckWritingResp>('/admin/ai/check-writing', {
    method: 'POST',
    body: { content },
    signal,
  })
}

export function annotateParagraph(
  params: { paragraph: string; prevContext?: string; title?: string },
  signal?: AbortSignal,
) {
  return request<AnnotateParagraphResp>('/admin/ai/annotate-paragraph', {
    method: 'POST',
    body: params,
    signal,
  })
}

export function getDashboardBrief() {
  return request<DashboardBriefResp>('/admin/ai/dashboard-brief')
}

export function refreshDashboardBrief() {
  return request<DashboardBriefResp>('/admin/ai/dashboard-brief/refresh', {
    method: 'POST',
  })
}

export function getCachedOpsReportIdeas() {
  return request<OpsReportIdeasResp>('/admin/ai/ops-report/ideas')
}

export function getAIFeatureFlags() {
  return request<AIFeatureFlags>('/admin/ai/features')
}

export function updateAIFeatureFlags(data: UpdateAIFeatureFlagsReq) {
  return request<AIFeatureFlags>('/admin/ai/features', {
    method: 'PUT',
    body: data,
  })
}

export interface DashboardBriefStreamHandlers {
  onDisabled?: () => void
  onStage?: (stage: string, message: string) => void
  onHeadline?: (text: string) => void
  onResult?: (data: DashboardBriefResp) => void
  onError?: (err: string) => void
}

export async function streamDashboardBrief(
  force: boolean,
  handlers: DashboardBriefStreamHandlers,
  signal?: AbortSignal,
): Promise<void> {
  const token = getAuthToken()
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const resp = await fetch(`${API_BASE_URL}/admin/ai/dashboard-brief/stream`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ force }),
    signal,
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
      const raw = line.slice(6)
      if (raw === '[DONE]') return
      try {
        const parsed = JSON.parse(raw) as {
          type?: string
          stage?: string
          message?: string
          text?: string
          data?: DashboardBriefResp
          error?: string
        }
        if (parsed.type === 'disabled') {
          handlers.onDisabled?.()
        } else if (parsed.type === 'stage' && parsed.message) {
          handlers.onStage?.(parsed.stage ?? '', parsed.message)
        } else if (parsed.type === 'headline' && typeof parsed.text === 'string') {
          handlers.onHeadline?.(parsed.text)
        } else if (parsed.type === 'result' && parsed.data) {
          handlers.onResult?.(parsed.data)
        } else if (parsed.type === 'error' && parsed.error) {
          handlers.onError?.(parsed.error)
        }
      } catch {
        // 忽略单行解析异常
      }
    }
  }
}

export interface OpsReportStreamHandlers {
  onDisabled?: () => void
  onStage?: (stage: string, message: string) => void
  onSummary?: (text: string) => void
  onResult?: (data: OpsReportResp) => void
  onError?: (err: string) => void
}

export async function streamOpsReport(
  params: { days: 7 | 30; force?: boolean },
  handlers: OpsReportStreamHandlers,
  signal?: AbortSignal,
): Promise<void> {
  const token = getAuthToken()
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const resp = await fetch(`${API_BASE_URL}/admin/ai/ops-report/stream`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ days: params.days, force: !!params.force }),
    signal,
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
      const raw = line.slice(6)
      if (raw === '[DONE]') return
      try {
        const parsed = JSON.parse(raw) as {
          type?: string
          stage?: string
          message?: string
          text?: string
          data?: OpsReportResp
          error?: string
        }
        if (parsed.type === 'disabled') {
          handlers.onDisabled?.()
        } else if (parsed.type === 'stage' && parsed.message) {
          handlers.onStage?.(parsed.stage ?? '', parsed.message)
        } else if (parsed.type === 'summary' && typeof parsed.text === 'string') {
          handlers.onSummary?.(parsed.text)
        } else if (parsed.type === 'result' && parsed.data) {
          handlers.onResult?.(parsed.data)
        } else if (parsed.type === 'error' && parsed.error) {
          handlers.onError?.(parsed.error)
        }
      } catch {
        // 忽略单行解析异常
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
