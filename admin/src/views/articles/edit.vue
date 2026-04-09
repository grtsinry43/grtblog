<script setup lang="ts">
import { PreviewLink20Regular } from '@vicons/fluent'
import { PaperPlaneOutline, SaveOutline } from '@vicons/ionicons5'
import { NButton, NButtonGroup, NInput, NPopover, useMessage } from 'naive-ui'
import { computed, onMounted, onUnmounted, ref, toRaw, toRef, watch } from 'vue'

// 组件
import MarkdownEditor from '@/components/markdown-editor/MarkdownEditor.vue'
import MarkdownPreview from '@/components/markdown-editor/MarkdownPreview.vue'
import { parseFederationSignals } from '@/composables/markdown-editor/utils/federation-signals'
import {
  getArticleFederationInteractions,
  resetArticleFederationSignals,
  type ArticleDetail,
} from '@/services/articles'
import { publishFederationActivityPub } from '@/services/federation-admin'
import EditorStatsOverlay from '@/views/shared/content-editor/components/EditorStatsOverlay.vue'
import {
  useAiSummaryGeneration,
  useAiTitleGeneration,
} from '@/views/shared/content-editor/composables/use-ai-tools'
import { usePreviewFrame } from '@/views/shared/content-editor/composables/use-preview-frame'

import ArticleCategoryModal from './components/ArticleCategoryModal.vue'
import ArticleMetaDrawer from './components/ArticleMetaDrawer.vue'
// 逻辑 Hooks
import { useArticleForm } from './composables/use-article-form'
import { useEditorStats } from './composables/use-editor-stats'
import { useTaxonomySelect } from './composables/use-taxonomy-select'

import type { FederationSignalRow } from './components/ArticleMetaDrawer.vue'
import type { FederationOutboundInteractionResp } from '@/types/federation'

defineOptions({ name: 'ArticleEdit' })

const message = useMessage()

// 1. 初始化表单核心逻辑
const { form, saving, imageProcessing, isCreating, fetch, save, extInfo, baseExtInfo } =
  useArticleForm()

// 2. 初始化分类与标签逻辑
// 将表单中的响应式属性传给 Hook，实现双向绑定
const {
  categoryOptions,
  dynamicTags,
  tagSearchValue,
  autoCompleteOptions,
  newCatModal,
  setInitialTags,
  handleTagsChange,
  addTagFromSearch,
  createNewCategory,
} = useTaxonomySelect(toRef(form, 'tagIds'), toRef(form, 'categoryId'), message)

// 3. 初始化编辑器统计逻辑
const { cursorPos, selectionStats, statsIdle, markActivity, handleCursorChange, getStats } =
  useEditorStats()

// 4. 视图状态管理
const showMeta = ref(false)
const loadedArticle = ref<ArticleDetail | null>(null)
const apPublishing = ref(false)
const isYearSummary = ref(false)
const yearSummaryYear = ref(new Date().getFullYear())
const yearSummaryReady = ref(false)
const federationInteractionsLoading = ref(false)
const federationInteractionsError = ref('')
const federationOutbounds = ref<FederationOutboundInteractionResp[]>([])
const resetAllFederationLoading = ref(false)
const resetSignalLoadingKeys = ref<Record<string, boolean>>({})

type FederationSignalType = FederationSignalRow['type']

const { loading: aiGenerating, generate: handleAIGenerate } = useAiTitleGeneration({
  getContent: () => form.content,
  applyResult: (result) => {
    form.title = result.title
    form.shortUrl = result.shortUrl
  },
  message,
})

const {
  loading: aiSummaryLoading,
  result: aiSummaryResult,
  done: aiSummaryDone,
  generate: handleAISummary,
  adopt: adoptAISummary,
  dismiss: dismissAISummary,
} = useAiSummaryGeneration({
  getContent: () => form.content,
  adoptSummary: (summary) => {
    form.aiSummary = summary
  },
  message,
})

// 6. 计算属性
const stats = computed(() => getStats(form.content))
const actionLabel = computed(() => {
  if (!form.isPublished) return '保存'
  return isCreating.value ? '发布' : '发布新版本'
})
const actionIcon = computed(() => (form.isPublished ? PaperPlaneOutline : SaveOutline))
const {
  showPreview,
  previewMode,
  previewFrameRef,
  previewUrl,
  fetchWebsiteInfo,
  schedulePreviewPayload,
  handlePreviewMessage,
  handlePreviewFrameLoad,
  resetPreviewReady,
  cleanup,
} = usePreviewFrame({
  previewPath: '/internal/preview/post',
  readyType: 'grtblog-preview:ready',
  postType: 'grtblog-preview:post',
  buildPayload: buildPreviewPayload,
  message,
})

function formatDateTime(value?: string | null) {
  if (!value) return '-'
  const timestamp = Date.parse(value)
  if (Number.isNaN(timestamp)) return '-'
  return new Date(timestamp).toLocaleString()
}

const apStatusText = computed(() => {
  if (isCreating.value) return '未创建'
  return loadedArticle.value?.activityPubObjectId ? '已发布' : '未发布'
})

const apLastPublishedAtText = computed(() =>
  formatDateTime(loadedArticle.value?.activityPubLastPublishedAt ?? null),
)

const canRepublishToActivityPub = computed(
  () => !isCreating.value && !!loadedArticle.value?.id && form.isPublished,
)

async function handleRepublishActivityPub() {
  if (!loadedArticle.value?.id) return
  if (!form.isPublished) {
    message.warning('请先设为发布并保存，再手动补发')
    return
  }
  apPublishing.value = true
  try {
    const resp = await publishFederationActivityPub({
      source_type: 'article',
      source_id: loadedArticle.value.id,
    })
    loadedArticle.value.activityPubObjectId =
      resp.object_id || loadedArticle.value.activityPubObjectId
    loadedArticle.value.activityPubLastPublishedAt = resp.published_at
    message.success(`补发完成：成功 ${resp.success_count}，失败 ${resp.failure_count}`)
  } catch (err) {
    message.error(err instanceof Error ? err.message : '补发失败')
  } finally {
    apPublishing.value = false
  }
}

function readFederationRegistry(value: unknown) {
  const root = value && typeof value === 'object' ? (value as Record<string, unknown>) : {}
  const rawRegistry = root._federation_delivery_registry
  const registry =
    rawRegistry && typeof rawRegistry === 'object' ? (rawRegistry as Record<string, unknown>) : {}
  const mentionRegistry =
    registry.mentions && typeof registry.mentions === 'object'
      ? (registry.mentions as Record<string, unknown>)
      : {}
  const citationRegistry =
    registry.citations && typeof registry.citations === 'object'
      ? (registry.citations as Record<string, unknown>)
      : {}

  const mentions: Record<string, string> = {}
  const citations: Record<string, string> = {}

  for (const [key, rawValue] of Object.entries(mentionRegistry)) {
    const normalizedKey = key.trim()
    if (!normalizedKey) continue
    mentions[normalizedKey] = typeof rawValue === 'string' ? rawValue : ''
  }
  for (const [key, rawValue] of Object.entries(citationRegistry)) {
    const normalizedKey = key.trim()
    if (!normalizedKey) continue
    citations[normalizedKey] = typeof rawValue === 'string' ? rawValue : ''
  }

  return { mentions, citations }
}

const federationOutboundBySignalKey = computed(() => {
  const map = new Map<string, FederationOutboundInteractionResp>()
  for (const item of federationOutbounds.value) {
    const key = item.signal_key?.trim()
    if (!key) continue
    const prev = map.get(key)
    if (!prev) {
      map.set(key, item)
      continue
    }
    const prevTs = Date.parse(prev.updated_at || prev.created_at || '')
    const curTs = Date.parse(item.updated_at || item.created_at || '')
    if ((Number.isFinite(curTs) ? curTs : 0) >= (Number.isFinite(prevTs) ? prevTs : 0)) {
      map.set(key, item)
    }
  }
  return map
})

const federationSignalRows = computed<FederationSignalRow[]>(() => {
  const rows = new Map<string, FederationSignalRow>()
  const parsedSignals = parseFederationSignals(form.content || '')
  const registry = readFederationRegistry(extInfo.value ?? baseExtInfo.value ?? null)

  const ensureMentionRow = (key: string) => {
    const splitAt = key.indexOf('@')
    if (splitAt <= 0 || splitAt >= key.length - 1) return
    const target = key.slice(0, splitAt)
    const instance = key.slice(splitAt + 1)
    if (!target || !instance) return
    if (!rows.has(key)) {
      rows.set(key, {
        key,
        type: 'mention',
        instance,
        target,
        marker: `<@${target}@${instance}>`,
        inContent: false,
        deliveredAt: null,
        outbound: null,
      })
    }
  }

  const ensureCitationRow = (key: string) => {
    const splitAt = key.indexOf('|')
    if (splitAt <= 0 || splitAt >= key.length - 1) return
    const instance = key.slice(0, splitAt)
    const target = key.slice(splitAt + 1)
    if (!instance || !target) return
    if (!rows.has(key)) {
      rows.set(key, {
        key,
        type: 'citation',
        instance,
        target,
        marker: `<cite:${instance}|${target}>`,
        inContent: false,
        deliveredAt: null,
        outbound: null,
      })
    }
  }

  parsedSignals.mentions.forEach((item) => {
    ensureMentionRow(item.key)
    const row = rows.get(item.key)
    if (row) row.inContent = true
  })
  parsedSignals.citations.forEach((item) => {
    ensureCitationRow(item.key)
    const row = rows.get(item.key)
    if (row) row.inContent = true
  })

  Object.entries(registry.mentions).forEach(([key, deliveredAt]) => {
    ensureMentionRow(key)
    const row = rows.get(key)
    if (row) row.deliveredAt = deliveredAt || null
  })
  Object.entries(registry.citations).forEach(([key, deliveredAt]) => {
    ensureCitationRow(key)
    const row = rows.get(key)
    if (row) row.deliveredAt = deliveredAt || null
  })

  federationOutboundBySignalKey.value.forEach((outbound, key) => {
    if (outbound.type === 'mention') {
      ensureMentionRow(key)
    } else if (outbound.type === 'citation') {
      ensureCitationRow(key)
    }
    const row = rows.get(key)
    if (row) row.outbound = outbound
  })

  return Array.from(rows.values()).sort((a, b) => {
    if (a.type !== b.type) return a.type === 'mention' ? -1 : 1
    return a.key.localeCompare(b.key)
  })
})

function outboundStatusTagType(status?: string | null) {
  const normalized = (status || '').trim().toLowerCase()
  if (normalized === 'approved' || normalized === 'accepted') return 'success'
  if (normalized === 'queued' || normalized === 'sending') return 'warning'
  if (
    normalized === 'rejected' ||
    normalized === 'failed' ||
    normalized === 'timeout' ||
    normalized === 'dead'
  )
    return 'error'
  return 'default'
}

function outboundStatusText(status?: string | null) {
  const normalized = (status || '').trim().toLowerCase()
  switch (normalized) {
    case 'queued':
      return '排队中'
    case 'sending':
      return '投递中'
    case 'accepted':
      return '已接收'
    case 'approved':
      return '已通过'
    case 'rejected':
      return '已拒绝'
    case 'failed':
      return '失败'
    case 'timeout':
      return '超时'
    case 'dead':
      return '终止'
    default:
      return normalized || '未知'
  }
}

function signalStatusText(row: FederationSignalRow) {
  if (row.outbound) return `队列：${outboundStatusText(row.outbound.status)}`
  if (row.deliveredAt) return '已记录（未发现出站记录）'
  if (row.inContent) return '正文存在，待触发'
  return '已脱离正文'
}

async function fetchFederationInteractions(articleID?: number) {
  const id = articleID ?? loadedArticle.value?.id
  if (!id) {
    federationOutbounds.value = []
    federationInteractionsError.value = ''
    return
  }
  federationInteractionsLoading.value = true
  federationInteractionsError.value = ''
  try {
    const data = await getArticleFederationInteractions(id)
    federationOutbounds.value = data.outbound ?? []
  } catch (err) {
    federationOutbounds.value = []
    federationInteractionsError.value = err instanceof Error ? err.message : '加载联合状态失败'
  } finally {
    federationInteractionsLoading.value = false
  }
}

function setSignalResetLoading(key: string, loading: boolean) {
  const next = { ...resetSignalLoadingKeys.value }
  if (loading) next[key] = true
  else delete next[key]
  resetSignalLoadingKeys.value = next
}

function applyFederationResetResult(extInfoValue: ArticleDetail['extInfo']) {
  const cloned = extInfoValue ? JSON.parse(JSON.stringify(extInfoValue)) : null
  baseExtInfo.value = cloned
  extInfo.value = cloned
  if (loadedArticle.value) {
    loadedArticle.value.extInfo = cloned
  }
}

async function handleResetFederationSignal(row: FederationSignalRow) {
  if (!loadedArticle.value?.id) return
  setSignalResetLoading(row.key, true)
  try {
    const payload = row.type === 'mention' ? { mentions: [row.key] } : { citations: [row.key] }
    const result = await resetArticleFederationSignals(loadedArticle.value.id, payload)
    applyFederationResetResult(result.extInfo ?? null)
    await fetchFederationInteractions(loadedArticle.value.id)
    if (result.retriggered) {
      message.success('已重置并重新触发该联合条目')
    } else {
      message.success('已重置该联合条目（当前未触发出站）')
    }
  } catch (err) {
    message.error(err instanceof Error ? err.message : '重置失败')
  } finally {
    setSignalResetLoading(row.key, false)
  }
}

async function handleResetAllFederationSignals() {
  if (!loadedArticle.value?.id) return
  resetAllFederationLoading.value = true
  try {
    const result = await resetArticleFederationSignals(loadedArticle.value.id)
    applyFederationResetResult(result.extInfo ?? null)
    await fetchFederationInteractions(loadedArticle.value.id)
    if (result.retriggered) {
      message.success('已重置全部联合条目并重新触发')
    } else {
      message.success('已重置全部联合条目（当前未触发出站）')
    }
  } catch (err) {
    message.error(err instanceof Error ? err.message : '重置失败')
  } finally {
    resetAllFederationLoading.value = false
  }
}

function buildPreviewPayload() {
  const nowIso = new Date().toISOString()
  const safeExtInfo = extInfo.value ? JSON.parse(JSON.stringify(toRaw(extInfo.value))) : null
  const safeTags = loadedArticle.value?.tags
    ? JSON.parse(JSON.stringify(toRaw(loadedArticle.value.tags)))
    : []
  return {
    id: loadedArticle.value?.id ?? 0,
    title: form.title,
    summary: form.summary,
    leadIn: form.leadIn || null,
    content: form.content,
    contentHash: loadedArticle.value?.contentHash ?? '',
    shortUrl: form.shortUrl,
    cover: form.cover || null,
    categoryId: form.categoryId,
    commentAreaId: null,
    extInfo: safeExtInfo,
    toc: undefined,
    tags: safeTags,
    metrics: loadedArticle.value ? { views: 0, likes: 0, comments: 0 } : undefined,
    isPublished: form.isPublished,
    isTop: form.isTop,
    isHot: false, // isHot removed from form, default false for preview or use loaded value if needed, but preview implies 'draft' context often.
    allowComment: form.allowComment,
    isOriginal: form.isOriginal,
    createdAt: loadedArticle.value?.createdAt ?? nowIso,
    updatedAt: nowIso,
    authorId: loadedArticle.value?.authorId ?? 0,
  }
}
function normalizeYearSummaryValue(value: unknown): number | null {
  if (typeof value === 'number' && Number.isFinite(value)) {
    const year = Math.floor(value)
    return year >= 1900 && year <= 3000 ? year : null
  }
  if (typeof value === 'string') {
    const parsed = Number.parseInt(value.trim(), 10)
    return Number.isFinite(parsed) && parsed >= 1900 && parsed <= 3000 ? parsed : null
  }
  return null
}

function readYearSummaryFromExtInfo(value: unknown): number | null {
  if (!value || typeof value !== 'object') return null
  return normalizeYearSummaryValue((value as Record<string, unknown>).is_year_summary)
}

function applyYearSummaryToExtInfo(target: Record<string, unknown>) {
  if (isYearSummary.value) {
    target.is_year_summary = yearSummaryYear.value
  } else {
    delete target.is_year_summary
  }
}

function syncYearSummaryToExtInfo() {
  const nextBase = baseExtInfo.value ? { ...baseExtInfo.value } : {}
  applyYearSummaryToExtInfo(nextBase)
  baseExtInfo.value = Object.keys(nextBase).length > 0 ? nextBase : null

  const nextExtInfo = extInfo.value ? { ...extInfo.value } : {}
  applyYearSummaryToExtInfo(nextExtInfo)
  extInfo.value = Object.keys(nextExtInfo).length > 0 ? nextExtInfo : null
}

async function handleSave() {
  syncYearSummaryToExtInfo()
  await save()
}

// 6. 生命周期
onMounted(async () => {
  window.addEventListener('message', handlePreviewMessage)

  const [data] = await Promise.all([fetch(), fetchWebsiteInfo()])
  loadedArticle.value = data as ArticleDetail | null
  await fetchFederationInteractions(data?.id)
  const summaryYear = readYearSummaryFromExtInfo(data?.extInfo ?? null)
  if (summaryYear) {
    isYearSummary.value = true
    yearSummaryYear.value = summaryYear
  } else {
    isYearSummary.value = false
    yearSummaryYear.value = new Date().getFullYear()
  }
  yearSummaryReady.value = true
  syncYearSummaryToExtInfo()
  if (data?.tags) {
    setInitialTags(data.tags)
  }
})

onUnmounted(() => {
  window.removeEventListener('message', handlePreviewMessage)
  cleanup()
})

watch(
  () => [
    form.title,
    form.summary,
    form.leadIn,
    form.content,
    form.cover,
    form.shortUrl,
    form.isPublished,
    form.isTop,
    form.allowComment,
    form.isOriginal,
    extInfo.value,
  ],
  () => {
    schedulePreviewPayload()
  },
  { deep: true },
)

watch([showPreview, previewMode, previewUrl], () => {
  schedulePreviewPayload()
})

watch(previewUrl, () => {
  resetPreviewReady()
})

watch([isYearSummary, yearSummaryYear], () => {
  if (!yearSummaryReady.value) return
  syncYearSummaryToExtInfo()
})
</script>

<template>
  <div class="flex h-full min-h-0 flex-col">
    <header
      class="z-10 flex shrink-0 flex-col gap-3 px-10 py-8 backdrop-blur sm:h-24 sm:flex-row sm:items-center sm:justify-between sm:py-0"
    >
      <div class="flex w-full items-center gap-4 sm:flex-1">
        <NInput
          v-model:value="form.title"
          placeholder="在这里开始你的写作吧..."
          :bordered="false"
          class="flex-1 text-xl! leading-tight font-bold sm:text-2xl!"
          style="--n-caret-color: var(--primary-color); background-color: transparent"
        />
      </div>

      <div class="flex w-full flex-wrap items-center gap-3 sm:w-auto sm:flex-nowrap sm:gap-4">
        <div class="flex items-baseline gap-1">
          <div class="iconify self-center ph--link-simple" />
          <span class="text-xs leading-none">/posts/</span>
          <input
            v-model="form.shortUrl"
            placeholder="请填写短链接"
            class="w-24 border-b border-current/30 p-0 pb-0.5 text-[11px] leading-none focus:border-primary focus:outline-none sm:w-32"
          />
        </div>

        <NButton
          quaternary
          size="small"
          :loading="aiGenerating"
          :disabled="!form.content?.trim()"
          @click="handleAIGenerate"
        >
          <template #icon><div class="iconify ph--robot" /></template>
          AI
        </NButton>

        <NButtonGroup>
          <NButton
            :type="!form.isPublished ? 'primary' : 'default'"
            :ghost="form.isPublished"
            @click="form.isPublished = false"
          >
            草稿
          </NButton>
          <NButton
            :type="form.isPublished ? 'primary' : 'default'"
            :ghost="!form.isPublished"
            @click="form.isPublished = true"
          >
            发布
          </NButton>
        </NButtonGroup>

        <div class="flex items-center gap-2">
          <span
            v-if="imageProcessing"
            class="text-xs text-amber-600"
          >
            正在处理图片…
          </span>
          <NButton
            quaternary
            circle
            size="small"
            @click="showMeta = true"
          >
            <template #icon><div class="iconify text-xl ph--sliders-horizontal" /></template>
          </NButton>

          <NButton
            quaternary
            circle
            size="small"
            :type="showPreview ? 'primary' : 'default'"
            @click="showPreview = !showPreview"
          >
            <template #icon><PreviewLink20Regular /></template>
          </NButton>

          <NButton
            type="primary"
            size="medium"
            :loading="saving"
            :disabled="saving || imageProcessing"
            @click="handleSave"
            class="px-5 font-medium shadow-sm active:scale-95"
          >
            <template #icon><component :is="actionIcon" /></template>
            {{ actionLabel }}
          </NButton>
        </div>
      </div>
    </header>

    <main class="flex min-h-0 flex-1 overflow-hidden">
      <div
        class="editor-container grid h-full min-h-0 w-full"
        :class="showPreview ? 'grid-cols-1 lg:grid-cols-2' : 'grid-cols-1'"
      >
        <div
          class="pane editor-pane relative h-full overflow-auto"
          @scroll="markActivity"
          @wheel="markActivity"
        >
          <MarkdownEditor
            v-model="form.content"
            class="h-full min-h-full"
            @cursor-change="handleCursorChange"
          />

          <EditorStatsOverlay
            :idle="statsIdle"
            :cursor-line="cursorPos.line"
            :cursor-column="cursorPos.column"
            :reading-minutes="stats.readingMinutes"
            :char-count="stats.charCount"
            :chinese-char-count="stats.chineseCharCount"
            :word-count="stats.wordCount"
            :total-char-count="stats.totalCharCount"
            :paragraph-count="stats.paragraphCount"
            :selection-total="selectionStats.total"
            :selection-chars="selectionStats.chars"
          />
        </div>

        <div
          v-if="showPreview"
          class="pane preview-pane relative h-full overflow-auto"
          @scroll="markActivity"
        >
          <div class="absolute top-3 right-3 z-10">
            <NPopover
              trigger="click"
              placement="bottom-end"
            >
              <template #trigger>
                <NButton
                  tertiary
                  type="primary"
                  circle
                  size="small"
                  class="shadow-sm"
                >
                  <template #icon><div class="iconify text-lg ph--dots-three-vertical" /></template>
                </NButton>
              </template>
              <div class="flex flex-col gap-1 p-1">
                <NButton
                  :type="previewMode === 'markdown' ? 'primary' : 'default'"
                  quaternary
                  size="small"
                  class="w-full justify-start px-2"
                  @click="previewMode = 'markdown'"
                  >Markdown 预览</NButton
                >
                <NButton
                  :type="previewMode === 'page' ? 'primary' : 'default'"
                  quaternary
                  size="small"
                  class="w-full justify-start px-2"
                  @click="previewMode = 'page'"
                  >网页预览</NButton
                >
              </div>
            </NPopover>
          </div>

          <MarkdownPreview
            v-if="previewMode === 'markdown'"
            :source="form.content"
            class="p-4 sm:p-8"
          />
          <div
            v-else
            class="h-full w-full"
          >
            <iframe
              v-if="previewUrl"
              :src="previewUrl"
              ref="previewFrameRef"
              class="h-full w-full border-0"
              @load="handlePreviewFrameLoad"
            />
            <div
              v-else
              class="flex h-full items-center justify-center text-sm opacity-60"
            >
              请先在站点信息中设置 public_url
            </div>
          </div>
        </div>
      </div>
    </main>

    <ArticleMetaDrawer
      v-model:show="showMeta"
      v-model:form="form"
      :category-options="categoryOptions"
      :dynamic-tags="dynamicTags"
      :tag-search-value="tagSearchValue"
      :auto-complete-options="autoCompleteOptions"
      :new-category-modal="newCatModal"
      :ai-summary-loading="aiSummaryLoading"
      :ai-summary-result="aiSummaryResult"
      :ai-summary-done="aiSummaryDone"
      :is-year-summary="isYearSummary"
      :year-summary-year="yearSummaryYear"
      :ap-status-text="apStatusText"
      :ap-last-published-at-text="apLastPublishedAtText"
      :loaded-article="loadedArticle"
      :can-republish-to-activity-pub="canRepublishToActivityPub"
      :ap-publishing="apPublishing"
      :is-creating="isCreating"
      :federation-interactions-error="federationInteractionsError"
      :federation-interactions-loading="federationInteractionsLoading"
      :federation-signal-rows="federationSignalRows"
      :reset-all-federation-loading="resetAllFederationLoading"
      :reset-signal-loading-keys="resetSignalLoadingKeys"
      :format-date-time="formatDateTime"
      :signal-status-text="signalStatusText"
      @update:tag-search-value="tagSearchValue = $event"
      @update:is-year-summary="isYearSummary = $event"
      @update:year-summary-year="yearSummaryYear = $event ?? new Date().getFullYear()"
      @open-category-modal="newCatModal.show = true"
      @tags-change="handleTagsChange"
      @add-tag="addTagFromSearch"
      @generate-ai-summary="handleAISummary"
      @adopt-ai-summary="adoptAISummary"
      @dismiss-ai-summary="dismissAISummary"
      @republish-activity-pub="handleRepublishActivityPub"
      @reset-all-federation-signals="handleResetAllFederationSignals"
      @reset-federation-signal="handleResetFederationSignal"
    />

    <ArticleCategoryModal
      v-model:modal="newCatModal"
      @create="createNewCategory"
    />
  </div>
</template>

<style scoped>
.editor-container {
  /* Clean grid layout handled by Tailwind classes */
}

/* Custom scrollbar refinements for a cleaner look */
.pane::-webkit-scrollbar,
.editor-pane :deep(.cm-scroller::-webkit-scrollbar),
.preview-pane :deep(.markdown-preview::-webkit-scrollbar) {
  width: 5px;
  height: 5px;
}

.pane::-webkit-scrollbar-track,
.editor-pane :deep(.cm-scroller::-webkit-scrollbar-track),
.preview-pane :deep(.markdown-preview::-webkit-scrollbar-track) {
  background: transparent;
}

:global(.dark) .pane::-webkit-scrollbar-thumb,
:global(.dark) .editor-pane :deep(.cm-scroller::-webkit-scrollbar-thumb),
:global(.dark) .preview-pane :deep(.markdown-preview::-webkit-scrollbar-thumb) {
  background-color: #374151;
}

.pane::-webkit-scrollbar-thumb:hover,
.editor-pane :deep(.cm-scroller::-webkit-scrollbar-thumb:hover),
.preview-pane :deep(.markdown-preview::-webkit-scrollbar-thumb:hover) {
  background-color: #d1d5db;
}

:global(.dark) .pane::-webkit-scrollbar-thumb:hover,
:global(.dark) .editor-pane :deep(.cm-scroller::-webkit-scrollbar-thumb:hover),
:global(.dark) .preview-pane :deep(.markdown-preview::-webkit-scrollbar-thumb:hover) {
  background-color: #4b5563;
}

.editor-pane :deep(.cm-editor) {
  height: 100% !important;
  font-family: inherit;
}

.editor-pane :deep(.cm-scroller) {
  padding-bottom: 50vh; /* Allow scrolling past end */
  font-family: 'JetBrains Mono', monospace; /* Optional: technical font for code */
  line-height: 1.6;
}

.preview-pane :deep(.markdown-preview) {
  padding-bottom: 50vh;
}
</style>
