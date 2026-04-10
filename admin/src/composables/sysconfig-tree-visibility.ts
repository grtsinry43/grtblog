import type { SysConfigGroup, SysConfigItem } from '@/services/sysconfig'

/** 子树内是否存在任一当前可见的配置项（与 ConfigItem 的 visibleWhen 一致，由调用方传入 isItemVisible） */
export function subtreeHasAnyVisibleField(
  group: SysConfigGroup,
  isItemVisible: (item: SysConfigItem) => boolean,
): boolean {
  if (group.items?.some(isItemVisible)) return true
  if (!group.children?.length) return false
  return group.children.some((ch) => subtreeHasAnyVisibleField(ch, isItemVisible))
}

/**
 * 前序遍历第一个将渲染为 NCollapseItem 的分组 path（与 ConfigPanel 跳过纯容器、隐藏级联空项一致）。
 */
export function firstVisibleCollapsiblePath(
  groups: SysConfigGroup[] | undefined,
  isItemVisible: (item: SysConfigItem) => boolean,
): string | null {
  if (!groups?.length) return null
  for (const g of groups) {
    const visibleDirect = g.items?.some(isItemVisible) ?? false
    const hasCh = (g.children?.length ?? 0) > 0

    if (!visibleDirect && hasCh) {
      if (!g.children!.some((ch) => subtreeHasAnyVisibleField(ch, isItemVisible))) continue
      const sub = firstVisibleCollapsiblePath(g.children, isItemVisible)
      if (sub) return sub
      continue
    }
    if (visibleDirect) return g.path
  }
  return null
}
