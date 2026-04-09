import { computed, ref } from 'vue'

import { listWebsiteInfo } from '@/services/website-info'

import type { MessageApi } from 'naive-ui'

interface UsePreviewFrameOptions<TPayload> {
  previewPath: string
  readyType: string
  postType: string
  buildPayload: () => TPayload
  message?: MessageApi
}

function normalizePublicUrl(value: string) {
  return value.trim().replace(/\/+$/, '')
}

export function usePreviewFrame<TPayload>(options: UsePreviewFrameOptions<TPayload>) {
  const showPreview = ref(false)
  const previewMode = ref<'markdown' | 'page'>('markdown')
  const previewFrameRef = ref<HTMLIFrameElement | null>(null)
  const previewReady = ref(false)
  const publicUrl = ref('')

  const previewUrl = computed(() => {
    const base = normalizePublicUrl(publicUrl.value)
    return base ? `${base}${options.previewPath}` : ''
  })

  const previewOrigin = computed(() => {
    if (!previewUrl.value) return '*'
    try {
      return new URL(previewUrl.value).origin
    } catch {
      return '*'
    }
  })

  async function fetchWebsiteInfo() {
    try {
      const list = await listWebsiteInfo()
      const item = list?.find((info) => info.key === 'public_url')
      publicUrl.value = item?.value?.trim() ?? ''
    } catch (error) {
      if (options.message) {
        options.message.error(error instanceof Error ? error.message : '加载站点地址失败')
      } else {
        console.error(error)
      }
    }
  }

  function sendPreviewPayload() {
    if (!showPreview.value || previewMode.value !== 'page') return
    if (!previewUrl.value || !previewReady.value) return

    const frame = previewFrameRef.value
    if (!frame?.contentWindow) return

    frame.contentWindow.postMessage(
      { type: options.postType, payload: options.buildPayload() },
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
    if (!data || data.type !== options.readyType) return

    previewReady.value = true
    sendPreviewPayload()
  }

  function handlePreviewFrameLoad() {
    previewReady.value = true
    sendPreviewPayload()
  }

  function resetPreviewReady() {
    previewReady.value = false
  }

  function cleanup() {
    if (previewTimer) {
      window.clearTimeout(previewTimer)
      previewTimer = null
    }
  }

  return {
    showPreview,
    previewMode,
    previewFrameRef,
    previewReady,
    previewUrl,
    fetchWebsiteInfo,
    sendPreviewPayload,
    schedulePreviewPayload,
    handlePreviewMessage,
    handlePreviewFrameLoad,
    resetPreviewReady,
    cleanup,
  }
}
