export interface NavMenuIconOption {
  label: string
  value: string
  iconClass: string
}

export const navMenuIconOptions: NavMenuIconOption[] = [
  { label: 'House / 首页', value: 'house', iconClass: 'iconify lucide--house' },
  { label: 'BookOpen / 文章', value: 'book-open', iconClass: 'iconify lucide--book-open' },
  { label: 'PenTool / 手记', value: 'pen-tool', iconClass: 'iconify lucide--pen-tool' },
  { label: 'Archive / 归档', value: 'archive', iconClass: 'iconify lucide--archive' },
  { label: 'Image / 相册', value: 'image', iconClass: 'iconify lucide--image' },
  { label: 'User / 关于', value: 'user', iconClass: 'iconify lucide--user' },
  { label: 'Terminal / 技术', value: 'terminal', iconClass: 'iconify lucide--terminal' },
  { label: 'Coffee / 生活', value: 'coffee', iconClass: 'iconify lucide--coffee' },
  { label: 'Sparkles / 随想', value: 'sparkles', iconClass: 'iconify lucide--sparkles' },
  { label: 'Code / 代码', value: 'code', iconClass: 'iconify lucide--code' },
  { label: 'List / 列表', value: 'list', iconClass: 'iconify lucide--list' },
]

export const navMenuIconValueSet = new Set(navMenuIconOptions.map((item) => item.value))

export function normalizeNavMenuIconValue(value?: string | null): string | null {
  if (!value) return null
  const trimmed = value.trim()
  if (!trimmed) return null
  return navMenuIconValueSet.has(trimmed) ? trimmed : null
}
