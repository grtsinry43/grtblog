export interface ContentTagItem {
  id: number
  name: string
}

export interface ContentTagOption {
  label: string
  value: number
  [key: string]: unknown
}

export function normalizeContentTagName(value: string) {
  return value.trim().replace(/\s+/g, ' ')
}

export function contentTagNameKey(value: string) {
  return normalizeContentTagName(value).toLocaleLowerCase()
}

export function findContentTagByName(options: ContentTagOption[], name: string) {
  const key = contentTagNameKey(name)
  if (!key) return undefined
  return options.find((option) => contentTagNameKey(option.label) === key)
}

export function mergeContentTagOptions(
  current: ContentTagOption[],
  items: ContentTagItem[],
): ContentTagOption[] {
  const byId = new Map<number, ContentTagOption>()

  for (const option of current) {
    const label = normalizeContentTagName(option.label)
    if (label) byId.set(option.value, { label, value: option.value })
  }
  for (const item of items) {
    const label = normalizeContentTagName(item.name)
    if (label) byId.set(item.id, { label, value: item.id })
  }

  return Array.from(byId.values()).sort((a, b) => a.label.localeCompare(b.label, 'zh-CN'))
}

export function appendUniqueId(ids: number[], id: number) {
  return ids.includes(id) ? ids : [...ids, id]
}
