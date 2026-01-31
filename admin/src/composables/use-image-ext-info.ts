import { computed, ref, watch, type Ref } from 'vue'

import type { ContentExtInfo, ImageExtInfoItem } from '@/types/ext-info'

const MARKDOWN_IMAGE_REGEX = /!\[[^\]]*]\(([^)\s]+)(?:\s+"[^"]*")?\)/g
const HTML_IMAGE_REGEX = /<img\s+[^>]*src=["']([^"']+)["'][^>]*>/gi

function normalizeUrl(value: string) {
  return value.trim().replace(/^<|>$/g, '')
}

function extractImageUrlsFromMarkdown(markdown: string) {
  const results: string[] = []
  let match: RegExpExecArray | null

  while ((match = MARKDOWN_IMAGE_REGEX.exec(markdown)) !== null) {
    if (match[1]) results.push(normalizeUrl(match[1]))
  }
  while ((match = HTML_IMAGE_REGEX.exec(markdown)) !== null) {
    if (match[1]) results.push(normalizeUrl(match[1]))
  }

  return results.filter(Boolean)
}

function splitImagesInput(value?: string) {
  if (!value) return []
  return value
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean)
}

async function loadImageElement(url: string) {
  return new Promise<HTMLImageElement>((resolve, reject) => {
    const image = new Image()
    image.crossOrigin = 'anonymous'
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error('image load failed'))
    image.src = url
  })
}

function calcDominantColor(ctx: CanvasRenderingContext2D, width: number, height: number) {
  const imageData = ctx.getImageData(0, 0, width, height)
  const data = imageData.data
  let r = 0
  let g = 0
  let b = 0
  let count = 0

  for (let i = 0; i < data.length; i += 4) {
    const alpha = data[i + 3]
    if (alpha < 16) continue
    r += data[i]
    g += data[i + 1]
    b += data[i + 2]
    count += 1
  }

  if (count === 0) return undefined
  const avgR = Math.round(r / count)
  const avgG = Math.round(g / count)
  const avgB = Math.round(b / count)
  return `#${avgR.toString(16).padStart(2, '0')}${avgG.toString(16).padStart(2, '0')}${avgB.toString(16).padStart(2, '0')}`
}

async function resolveImageInfo(url: string): Promise<ImageExtInfoItem | null> {
  try {
    const image = await loadImageElement(url)
    const width = image.naturalWidth || image.width
    const height = image.naturalHeight || image.height
    const sampleSize = 32
    const canvas = document.createElement('canvas')
    canvas.width = sampleSize
    canvas.height = sampleSize
    const ctx = canvas.getContext('2d')
    if (!ctx) return { id: url, width, height }
    ctx.drawImage(image, 0, 0, sampleSize, sampleSize)
    let color: string | undefined
    try {
      color = calcDominantColor(ctx, sampleSize, sampleSize)
    } catch {
      color = undefined
    }
    return { id: url, width, height, color }
  } catch {
    return { id: url }
  }
}

async function resolveImageInfos(urls: string[]) {
  const results = await Promise.all(urls.map((url) => resolveImageInfo(url)))
  return results.filter((item): item is ImageExtInfoItem => Boolean(item))
}

function buildExtInfo(base: ContentExtInfo | null | undefined, images: ImageExtInfoItem[]) {
  const next: ContentExtInfo = { ...(base ?? {}) }
  if (images.length) {
    next.images = images
  } else {
    delete next.images
  }
  return Object.keys(next).length ? next : null
}

export function useImageExtInfo(params: {
  content: Ref<string>
  extraImages?: Ref<string>
  baseExtInfo?: Ref<ContentExtInfo | null | undefined>
}) {
  const processing = ref(false)
  const extInfo = ref<ContentExtInfo | null>(null)
  const taskId = ref(0)
  const cache = new Map<string, ImageExtInfoItem>()

  const urls = computed(() => {
    const contentUrls = extractImageUrlsFromMarkdown(params.content.value || '')
    const extraUrls = params.extraImages ? splitImagesInput(params.extraImages.value) : []
    return Array.from(new Set([...contentUrls, ...extraUrls])).filter(Boolean)
  })

  let debounceTimer: number | undefined
  watch(
    [urls, params.baseExtInfo ?? ref<ContentExtInfo | null>(null)],
    (next) => {
      if (debounceTimer) window.clearTimeout(debounceTimer)
      const nextUrls = Array.isArray(next) ? (next[0] as string[]) : next
      debounceTimer = window.setTimeout(async () => {
        const currentTask = taskId.value + 1
        taskId.value = currentTask
        if (!nextUrls.length) {
          extInfo.value = buildExtInfo(params.baseExtInfo?.value ?? null, [])
          processing.value = false
          return
        }
        const missing = nextUrls.filter((url) => !cache.has(url))
        if (!missing.length) {
          const existing = nextUrls.map((url) => cache.get(url) || { id: url })
          extInfo.value = buildExtInfo(params.baseExtInfo?.value ?? null, existing)
          processing.value = false
          return
        }
        processing.value = true
        const result = await resolveImageInfos(missing)
        if (taskId.value !== currentTask) return
        for (const item of result) {
          cache.set(item.id, item)
        }
        const merged = nextUrls.map((url) => cache.get(url) || { id: url })
        extInfo.value = buildExtInfo(params.baseExtInfo?.value ?? null, merged)
        processing.value = false
      }, 400)
    },
    { immediate: true },
  )

  return {
    extInfo,
    processing,
    imageUrls: urls,
  }
}
