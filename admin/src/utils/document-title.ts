import type { RouteLocationNormalizedLoaded } from 'vue-router'

const ADMIN_PANEL_TITLE = '管理后台'
const DEFAULT_SITE_NAME = 'grtblog'
const FALLBACK_SITE_NAME = (import.meta.env.VITE_APP_NAME || DEFAULT_SITE_NAME).trim() || DEFAULT_SITE_NAME
const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '/api/v2').replace(/\/$/, '')

interface WebsiteInfoItem {
  key: string
  value?: string | null
}

interface ApiEnvelope<T> {
  code: number
  data: T
}

let cachedSiteName: string | null = null
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

function extractSiteName(items: unknown): string {
  if (!Array.isArray(items)) return FALLBACK_SITE_NAME
  const websiteNameItem = items.find(
    (item): item is WebsiteInfoItem =>
      !!item && typeof item === 'object' && 'key' in item && (item as WebsiteInfoItem).key === 'website_name',
  )
  return normalizeSiteName(websiteNameItem?.value)
}

async function fetchSiteNameFromBackend() {
  const response = await fetch(`${API_BASE_URL}/website-info`, {
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

  return extractSiteName(payload.data)
}

export function getCachedSiteName() {
  return normalizeSiteName(cachedSiteName)
}

export function resolveDocumentTitle(route: RouteLocationNormalizedLoaded, siteName: string) {
  return [resolveMetaTitle(route), ADMIN_PANEL_TITLE, normalizeSiteName(siteName)].filter(Boolean).join(' - ')
}

export function applyDocumentTitle(route: RouteLocationNormalizedLoaded, siteName: string) {
  document.title = resolveDocumentTitle(route, siteName)
}

export async function ensureBackendSiteName() {
  if (cachedSiteName) return cachedSiteName

  if (!pendingSiteNameRequest) {
    pendingSiteNameRequest = fetchSiteNameFromBackend()
      .then((siteName) => {
        cachedSiteName = normalizeSiteName(siteName)
        return cachedSiteName
      })
      .catch(() => normalizeSiteName(cachedSiteName))
      .finally(() => {
        pendingSiteNameRequest = null
      })
  }

  return pendingSiteNameRequest
}
