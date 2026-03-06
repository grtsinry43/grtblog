<script setup lang="ts">
import { PreviewLink20Regular } from '@vicons/fluent'
import { SaveOutline } from '@vicons/ionicons5'
import {
  NButton,
  NPopover,
  NInput,
  NSwitch,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  useMessage,
} from 'naive-ui'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

// Components
import MarkdownEditor from '@/components/markdown-editor/MarkdownEditor.vue'
import MarkdownPreview from '@/components/markdown-editor/MarkdownPreview.vue'
import { generateSummaryStream } from '@/services/ai'
import { listWebsiteInfo } from '@/services/website-info'

// Composables
import { usePageForm } from './composables/use-page-form'

import type { PageDetail } from '@/services/page'

defineOptions({ name: 'PageEdit' })

const message = useMessage()

// 1. Initialize form logic
const { form, saving, fetch, save } = usePageForm()

// AI 摘要生成
const aiSummaryLoading = ref(false)
const aiSummaryResult = ref('')
const aiSummaryDone = ref(false)

async function handleAISummary() {
  if (!form.content?.trim()) {
    message.warning('请先输入内容')
    return
  }
  aiSummaryLoading.value = true
  aiSummaryResult.value = ''
  aiSummaryDone.value = false
  try {
    await generateSummaryStream(form.content, (chunk) => {
      aiSummaryResult.value += chunk
    })
    aiSummaryDone.value = true
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : 'AI 摘要生成失败')
    aiSummaryResult.value = ''
  } finally {
    aiSummaryLoading.value = false
  }
}

function adoptAISummary() {
  form.aiSummary = aiSummaryResult.value
  aiSummaryResult.value = ''
  aiSummaryDone.value = false
  message.success('已采纳 AI 摘要')
}

function dismissAISummary() {
  aiSummaryResult.value = ''
  aiSummaryDone.value = false
}

// 2. View state management
const showMeta = ref(false)
const showPreview = ref(false)
const previewMode = ref<'markdown' | 'page'>('markdown')
const previewFrameRef = ref<HTMLIFrameElement | null>(null)
const previewReady = ref(false)
const publicUrl = ref('')
const loadedPage = ref<PageDetail | null>(null)

const PREVIEW_READY_TYPE = 'grtblog-preview:ready'
const PREVIEW_PAGE_TYPE = 'grtblog-preview:page'

const previewUrl = computed(() => {
  const base = normalizePublicUrl(publicUrl.value)
  return base ? `${base}/internal/preview/page` : ''
})

const previewOrigin = computed(() => {
  if (!previewUrl.value) return '*'
  try {
    return new URL(previewUrl.value).origin
  } catch {
    return '*'
  }
})

function normalizePublicUrl(value: string) {
  return value.trim().replace(/\/+$/, '')
}

async function fetchWebsiteInfo() {
  try {
    const list = await listWebsiteInfo()
    const item = list?.find((info) => info.key === 'public_url')
    publicUrl.value = item?.value?.trim() ?? ''
  } catch (err) {
    console.error(err)
  }
}

function buildPreviewPayload() {
  const nowIso = new Date().toISOString()
  return {
    id: loadedPage.value?.id ?? 0,
    title: form.title,
    description: form.description || null,
    aiSummary: form.aiSummary || loadedPage.value?.aiSummary || null,
    content: form.content,
    contentHash: loadedPage.value?.contentHash ?? '',
    commentAreaId: loadedPage.value?.commentId ?? null,
    shortUrl: form.shortUrl,
    isEnabled: form.isEnabled,
    isBuiltin: loadedPage.value?.isBuiltin ?? false,
    metrics: loadedPage.value
      ? {
          views: loadedPage.value.views ?? 0,
          likes: loadedPage.value.likes ?? 0,
          comments: loadedPage.value.comments ?? 0,
        }
      : undefined,
    createdAt: loadedPage.value?.createdAt ?? nowIso,
    updatedAt: nowIso,
  }
}

function sendPreviewPayload() {
  if (!showPreview.value || previewMode.value !== 'page') return
  if (!previewUrl.value || !previewReady.value) return
  const frame = previewFrameRef.value
  if (!frame?.contentWindow) return
  frame.contentWindow.postMessage(
    { type: PREVIEW_PAGE_TYPE, payload: buildPreviewPayload() },
    previewOrigin.value,
  )
}

let previewTimer: number | null = null
function schedulePreviewPayload() {
  if (!showPreview.value || previewMode.value !== 'page') return
  if (!previewUrl.value) return
  if (previewTimer) window.clearTimeout(previewTimer)
  previewTimer = window.setTimeout(() => {
    previewTimer = null
    sendPreviewPayload()
  }, 200)
}

function handlePreviewMessage(event: MessageEvent) {
  const frame = previewFrameRef.value
  if (!frame?.contentWindow || event.source !== frame.contentWindow) return
  const data = event.data as { type?: string } | null
  if (!data || data.type !== PREVIEW_READY_TYPE) return
  previewReady.value = true
  sendPreviewPayload()
}

function handlePreviewFrameLoad() {
  previewReady.value = true
  sendPreviewPayload()
}

// 3. Lifecycle
onMounted(() => {
  window.addEventListener('message', handlePreviewMessage)
  Promise.all([fetch(), fetchWebsiteInfo()]).then(([data]) => {
    loadedPage.value = data as PageDetail | null
  })
})

onUnmounted(() => {
  window.removeEventListener('message', handlePreviewMessage)
  if (previewTimer) window.clearTimeout(previewTimer)
})

watch(
  () => [
    form.title,
    form.description,
    form.content,
    form.shortUrl,
    form.isEnabled,
    form.allowComment,
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
  previewReady.value = false
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
          placeholder="页面标题..."
          :bordered="false"
          class="flex-1 text-xl! leading-tight font-bold sm:text-2xl!"
          style="--n-caret-color: var(--primary-color); background-color: transparent"
        />
      </div>

      <div class="flex w-full flex-wrap items-center gap-3 sm:w-auto sm:flex-nowrap sm:gap-4">
        <div class="flex items-baseline gap-1">
          <div class="iconify self-center ph--link-simple" />
          <span class="text-xs leading-none">/pages/</span>
          <input
            v-model="form.shortUrl"
            placeholder="请填写短链接"
            class="w-24 border-b border-current/30 p-0 pb-0.5 text-[11px] leading-none focus:border-primary focus:outline-none sm:w-32"
          />
        </div>

        <div class="flex items-center gap-2">
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
            @click="save"
            class="px-5 font-medium shadow-sm active:scale-95"
          >
            <template #icon><SaveOutline /></template>
            保存
          </NButton>
        </div>
      </div>
    </header>

    <main class="flex min-h-0 flex-1 overflow-hidden">
      <div
        class="editor-container grid h-full min-h-0 w-full"
        :class="showPreview ? 'grid-cols-1 lg:grid-cols-2' : 'grid-cols-1'"
      >
        <div class="pane editor-pane relative h-full overflow-auto">
          <MarkdownEditor
            v-model="form.content"
            class="h-full min-h-full"
          />
        </div>

        <div
          v-if="showPreview"
          class="pane preview-pane relative h-full overflow-auto"
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
              请先在站点设置中配置 public_url
            </div>
          </div>
        </div>
      </div>
    </main>

    <NDrawer
      v-model:show="showMeta"
      placement="right"
      width="400"
    >
      <NDrawerContent
        title="页面设置"
        :native-scrollbar="false"
        closable
        header-style="padding: 24px;"
        body-style="padding: 24px;"
      >
        <div class="flex flex-col gap-6">
          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--article" />
              <span>元信息</span>
            </div>
            <NForm
              label-placement="top"
              label-width="auto"
              class="space-y-4"
            >
              <NFormItem
                label="描述"
                :show-feedback="false"
              >
                <NInput
                  v-model:value="form.description"
                  type="textarea"
                  placeholder="简短的页面描述..."
                  :autosize="{ minRows: 2, maxRows: 4 }"
                />
              </NFormItem>
              <NFormItem
                :show-feedback="false"
              >
                <template #label>
                  <span>AI 摘要</span>
                  <span class="ml-1 text-xs opacity-50">用于正文前的总结导读</span>
                </template>
                <NInput
                  v-model:value="form.aiSummary"
                  type="textarea"
                  placeholder="AI 生成的内容导读，展示在正文之前..."
                  :autosize="{ minRows: 2, maxRows: 4 }"
                />
              </NFormItem>
              <div class="flex flex-col gap-2">
                <NButton
                  size="small"
                  :loading="aiSummaryLoading"
                  :disabled="!form.content?.trim() || aiSummaryLoading"
                  @click="handleAISummary"
                >
                  <template #icon><div class="iconify ph--sparkle" /></template>
                  AI 生成导读摘要
                </NButton>
                <div
                  v-if="aiSummaryLoading || aiSummaryResult"
                  class="rounded-lg border border-current/10 p-3 text-sm leading-relaxed"
                >
                  <span>{{ aiSummaryResult }}</span>
                  <span
                    v-if="aiSummaryLoading"
                    class="inline-block w-1.5 animate-pulse bg-current"
                    >&nbsp;</span
                  >
                </div>
                <div
                  v-if="aiSummaryDone"
                  class="flex justify-end gap-2"
                >
                  <NButton
                    size="small"
                    quaternary
                    @click="dismissAISummary"
                    >放弃</NButton
                  >
                  <NButton
                    size="small"
                    type="primary"
                    @click="adoptAISummary"
                    >采纳</NButton
                  >
                </div>
              </div>
            </NForm>
          </div>

          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--toggle-left" />
              <span>属性</span>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="flex items-center justify-between rounded-lg px-4 py-3">
                <span class="text-sm">是否启用</span>
                <NSwitch
                  v-model:value="form.isEnabled"
                  size="small"
                />
              </div>
              <div class="flex items-center justify-between rounded-lg px-4 py-3">
                <span class="text-sm">允许评论</span>
                <NSwitch
                  v-model:value="form.allowComment"
                  size="small"
                />
              </div>
            </div>
          </div>
        </div>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped>
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
  padding-bottom: 50vh;
  font-family: 'JetBrains Mono', monospace;
  line-height: 1.6;
}
.preview-pane :deep(.markdown-preview) {
  padding-bottom: 50vh;
}
</style>
