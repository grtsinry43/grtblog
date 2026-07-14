import type { SysConfigGroup, SysConfigItem, SysConfigTreeResponse } from '@/services/sysconfig'

export interface HotArticleThresholds {
  views: number
  likes: number
  comments: number
}

export const DEFAULT_HOT_ARTICLE_THRESHOLDS: HotArticleThresholds = {
  views: 100,
  likes: 10,
  comments: 5,
}

function collectGroupItems(groups: SysConfigGroup[], result: SysConfigItem[]) {
  for (const group of groups) {
    if (group.items) result.push(...group.items)
    if (group.children) collectGroupItems(group.children, result)
  }
}

function getAllItems(tree: SysConfigTreeResponse) {
  const result = [...(tree.items ?? [])]
  collectGroupItems(tree.groups ?? [], result)
  return result
}

function parseThreshold(value: unknown, fallback: number) {
  if (value === null || value === undefined || String(value).trim() === '') return fallback
  const parsed = typeof value === 'number' ? value : Number(String(value ?? '').trim())
  return Number.isSafeInteger(parsed) && parsed >= 0 ? parsed : fallback
}

export function resolveHotArticleThresholds(tree: SysConfigTreeResponse): HotArticleThresholds {
  const items = new Map(getAllItems(tree).map((item) => [item.key, item]))
  const read = (key: string, fallback: number) => {
    const item = items.get(key)
    return parseThreshold(item?.value ?? item?.defaultValue, fallback)
  }

  return {
    views: read('article.hot.views', DEFAULT_HOT_ARTICLE_THRESHOLDS.views),
    likes: read('article.hot.likes', DEFAULT_HOT_ARTICLE_THRESHOLDS.likes),
    comments: read('article.hot.comments', DEFAULT_HOT_ARTICLE_THRESHOLDS.comments),
  }
}

export function formatHotArticleThresholds(thresholds: HotArticleThresholds) {
  return `热门标准：浏览量 ≥ ${thresholds.views}、点赞数 ≥ ${thresholds.likes} 或评论数 ≥ ${thresholds.comments}`
}
