import { ref } from 'vue'

import type { RouteLocationNormalizedLoaded } from 'vue-router'

const ADMIN_PANEL_TITLE = '管理后台'
const DEFAULT_SITE_NAME = 'Grtblog Admin'
const FALLBACK_SITE_NAME =
  (import.meta.env.VITE_APP_NAME || DEFAULT_SITE_NAME).trim() || DEFAULT_SITE_NAME
const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '/api/v2').replace(/\/$/, '')

interface WebsiteInfoItem {
  key: string
  value?: string | null
}

interface ApiEnvelope<T> {
  code: number
  data: T
}

const cachedSiteName = ref<string | null>(null)
const cachedFavicon = ref<string | null>(null)
let pendingSiteNameRequest: Promise<string> | null = null

function toText(value: unknown): string {
  return typeof value === 'string' ? value.trim() : ''
}

function resolveMetaTitle(route: RouteLocationNormalizedLoaded): string {
  const renderedTitle = route.meta.renderTabTitle?.(route.params)
  const renderedText = toText(renderedTitle)
  if (renderedText) return renderedText

  const currentMetaTitle = route.meta.title
  if (typeof currentMetaTitle === 'function') {
    const title = currentMetaTitle()
    const text = toText(title)
    if (text) return text
  } else {
    const text = toText(currentMetaTitle)
    if (text) return text
  }

  for (let i = route.matched.length - 1; i >= 0; i -= 1) {
    const matchedTitle = route.matched[i]?.meta?.title
    if (typeof matchedTitle === 'function') {
      const text = toText(matchedTitle())
      if (text) return text
    } else {
      const text = toText(matchedTitle)
      if (text) return text
    }
  }

  if (typeof route.name === 'string' && route.name.trim()) {
    return route.name.trim()
  }

  return route.path
}

function normalizeSiteName(siteName: string | null | undefined) {
  const text = toText(siteName)
  return text || FALLBACK_SITE_NAME
}

function extractField(data: unknown, key: string): string | null {
  if (data && typeof data === 'object' && !Array.isArray(data) && key in data) {
    const val = (data as Record<string, unknown>)[key]
    return typeof val === 'string' ? val.trim() || null : null
  }
  if (Array.isArray(data)) {
    const item = data.find(
      (it): it is WebsiteInfoItem =>
        !!it && typeof it === 'object' && 'key' in it && (it as WebsiteInfoItem).key === key,
    )
    return item?.value?.trim() || null
  }
  return null
}

function extractSiteName(data: unknown): string {
  return normalizeSiteName(extractField(data, 'website_name'))
}

async function fetchSiteInfoFromBackend() {
  const response = await fetch(`${API_BASE_URL}/public/website-info`, {
    method: 'GET',
    headers: {
      Accept: 'application/json',
    },
  })

  if (!response.ok) {
    throw new Error(`request failed (${response.status})`)
  }

  const payload = (await response.json()) as ApiEnvelope<unknown>
  if (!payload || payload.code !== 0) {
    throw new Error('invalid website info payload')
  }

  cachedFavicon.value = extractField(payload.data, 'favicon')
  return extractSiteName(payload.data)
}

export function getCachedSiteName() {
  return normalizeSiteName(cachedSiteName.value)
}

export function getCachedFavicon() {
  return cachedFavicon.value
}

export function resolveDocumentTitle(route: RouteLocationNormalizedLoaded, siteName: string) {
  return [resolveMetaTitle(route), ADMIN_PANEL_TITLE, normalizeSiteName(siteName)]
    .filter(Boolean)
    .join(' - ')
}

export function applyDocumentTitle(route: RouteLocationNormalizedLoaded, siteName: string) {
  document.title = resolveDocumentTitle(route, siteName)
}

export async function ensureBackendSiteName() {
  if (cachedSiteName.value) return cachedSiteName.value

  if (!pendingSiteNameRequest) {
    pendingSiteNameRequest = fetchSiteInfoFromBackend()
      .then((siteName) => {
        cachedSiteName.value = normalizeSiteName(siteName)
        return cachedSiteName.value
      })
      .catch(() => normalizeSiteName(cachedSiteName.value))
      .finally(() => {
        pendingSiteNameRequest = null
      })
  }

  return pendingSiteNameRequest
}
