import { NButton, NCard, NCollapse, NCollapseItem, NEmpty, NForm, NSpin, NTag } from 'naive-ui'
import { computed, defineComponent, watch, type PropType } from 'vue'

import ConfigItem from '@/components/config/ConfigItem'
import { useConfigCenter, type ConfigListFn, type ConfigUpdateFn } from '@/composables/use-config-center'

import type { SysConfigGroup, SysConfigTreeResponse } from '@/services/sysconfig'

/**
 * Check if a group path matches any of the given prefixes.
 * A group matches if:
 * - It IS one of the prefixes (exact)
 * - It's a descendant of a prefix (group.path starts with prefix/)
 * - It's an ancestor of a prefix (a prefix starts with group.path/)
 */
function groupMatchesPrefixes(groupPath: string, prefixes: string[]): 'exact-or-descendant' | 'ancestor' | false {
  for (const p of prefixes) {
    if (groupPath === p || groupPath.startsWith(`${p}/`)) return 'exact-or-descendant'
    if (p.startsWith(`${groupPath}/`)) return 'ancestor'
  }
  return false
}

function filterGroupsRecursive(
  groups: SysConfigGroup[],
  filterPrefixes?: string[],
  excludePrefixes?: string[],
): SysConfigGroup[] {
  if (!groups || groups.length === 0) return []

  return groups.reduce<SysConfigGroup[]>((acc, group) => {
    // Exclude check first
    if (excludePrefixes && excludePrefixes.length > 0) {
      const excludeMatch = groupMatchesPrefixes(group.path, excludePrefixes)
      if (excludeMatch === 'exact-or-descendant') return acc
      // If ancestor of excluded prefix, keep group but filter children
      if (excludeMatch === 'ancestor') {
        const filteredChildren = filterGroupsRecursive(group.children ?? [], undefined, excludePrefixes)
        if (filteredChildren.length > 0 || (group.items && group.items.length > 0)) {
          acc.push({ ...group, children: filteredChildren })
        }
        return acc
      }
    }

    // Include check
    if (filterPrefixes && filterPrefixes.length > 0) {
      const match = groupMatchesPrefixes(group.path, filterPrefixes)
      if (!match) return acc

      if (match === 'exact-or-descendant') {
        // This group is the target or inside it — include everything
        acc.push(group)
      } else {
        // This group is an ancestor — include it but recurse to filter children
        const filteredChildren = filterGroupsRecursive(group.children ?? [], filterPrefixes, excludePrefixes)
        if (filteredChildren.length > 0) {
          acc.push({ ...group, children: filteredChildren, items: undefined })
        }
      }
      return acc
    }

    // No filter — include as-is (but still apply exclude to children if needed)
    acc.push(group)
    return acc
  }, [])
}

function removeItemsByKey(groups: SysConfigGroup[], keys: string[]): SysConfigGroup[] {
  return groups.map((group) => ({
    ...group,
    items: group.items?.filter((item) => !keys.includes(item.key)),
    children: group.children ? removeItemsByKey(group.children, keys) : undefined,
  }))
}

function filterTree(
  tree: SysConfigTreeResponse,
  filterGroups?: string[],
  filterRootItemKeys?: string[],
  excludeGroups?: string[],
  excludeRootItemKeys?: string[],
  excludeItemKeys?: string[],
): SysConfigTreeResponse {
  let groups = filterGroupsRecursive(tree.groups ?? [], filterGroups, excludeGroups)
  let items = tree.items ?? []

  if (filterRootItemKeys && filterRootItemKeys.length > 0) {
    items = items.filter((item) =>
      filterRootItemKeys.some((p) => item.key === p || item.key.startsWith(`${p}.`)),
    )
  } else if (filterGroups && filterGroups.length > 0) {
    // When filtering by group, don't show root items unless explicitly requested
    items = []
  }

  if (excludeRootItemKeys && excludeRootItemKeys.length > 0) {
    items = items.filter(
      (item) => !excludeRootItemKeys.some((p) => item.key === p || item.key.startsWith(`${p}.`)),
    )
  }

  if (excludeItemKeys && excludeItemKeys.length > 0) {
    items = items.filter((item) => !excludeItemKeys.includes(item.key))
    groups = removeItemsByKey(groups, excludeItemKeys)
  }

  return { groups, items }
}

export default defineComponent({
  name: 'ConfigPanel',
  props: {
    listFn: { type: Function as PropType<ConfigListFn>, required: true },
    updateFn: { type: Function as PropType<ConfigUpdateFn>, required: true },
    title: { type: String, required: true },
    description: { type: String, default: '' },
    filterGroups: { type: Array as PropType<string[]>, default: undefined },
    filterRootItemKeys: { type: Array as PropType<string[]>, default: undefined },
    excludeGroups: { type: Array as PropType<string[]>, default: undefined },
    excludeRootItemKeys: { type: Array as PropType<string[]>, default: undefined },
    excludeItemKeys: { type: Array as PropType<string[]>, default: undefined },
    onDirtyChange: { type: Function as PropType<(dirty: boolean) => void>, default: undefined },
  },
  setup(props, { expose }) {
    const hasFilter = computed(
      () =>
        (props.filterGroups && props.filterGroups.length > 0) ||
        (props.filterRootItemKeys && props.filterRootItemKeys.length > 0) ||
        (props.excludeGroups && props.excludeGroups.length > 0) ||
        (props.excludeRootItemKeys && props.excludeRootItemKeys.length > 0) ||
        (props.excludeItemKeys && props.excludeItemKeys.length > 0),
    )

    const wrappedListFn: ConfigListFn = async (keys?: string[]) => {
      const tree = await props.listFn(keys)
      if (!hasFilter.value) return tree
      return filterTree(
        tree,
        props.filterGroups,
        props.filterRootItemKeys,
        props.excludeGroups,
        props.excludeRootItemKeys,
        props.excludeItemKeys,
      )
    }

    const {
      loading,
      saving,
      tree,
      valueMap,
      jsonBufferMap,
      expandedGroups,
      pendingCount,
      isItemVisible,
      fetch,
      save,
    } = useConfigCenter(wrappedListFn, props.updateFn)

    watch(
      pendingCount,
      (count) => {
        props.onDirtyChange?.(count > 0)
      },
      { immediate: true },
    )

    expose({ save, fetch, pendingCount })

    const renderGroups = (groups?: SysConfigGroup[]) => {
      if (!groups || groups.length === 0) return null

      return groups.map((group) => (
        <NCollapseItem key={group.path} name={group.path} title={group.label || group.key}>
          <div class="space-y-4 pl-2">
            {group.items && group.items.length > 0 && (
              <NForm labelPlacement="left" labelWidth={160}>
                {group.items.map((item) => (
                  <ConfigItem
                    key={item.key}
                    item={item}
                    valueMap={valueMap}
                    jsonBufferMap={jsonBufferMap}
                    visible={isItemVisible}
                  />
                ))}
              </NForm>
            )}

            {group.children && group.children.length > 0 && (
              <NCollapse>{renderGroups(group.children)}</NCollapse>
            )}
          </div>
        </NCollapseItem>
      ))
    }

    return () => (
      <NCard>
        {{
          header: () => (
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <div class="text-base font-semibold">{props.title}</div>
                {props.description && (
                  <div class="text-xs text-neutral-500">{props.description}</div>
                )}
              </div>
              <div class="flex items-center gap-2">
                {pendingCount.value > 0 && (
                  <NTag type="warning">
                    待保存 {pendingCount.value}
                  </NTag>
                )}
                <NButton size="small" secondary loading={loading.value} onClick={fetch}>
                  刷新
                </NButton>
                <NButton size="small" type="primary" loading={saving.value} onClick={save}>
                  保存
                </NButton>
              </div>
            </div>
          ),
          default: () => (
            <NSpin show={loading.value}>
              {!tree.value || (!tree.value.items?.length && !tree.value.groups?.length) ? (
                <div class="py-8">
                  <NEmpty description="暂无配置项" />
                </div>
              ) : (
                <div class="space-y-6">
                  {tree.value.items && tree.value.items.length > 0 && (
                    <NForm labelPlacement="left" labelWidth={160}>
                      {tree.value.items.map((item) => (
                        <ConfigItem
                          key={item.key}
                          item={item}
                          valueMap={valueMap}
                          jsonBufferMap={jsonBufferMap}
                          visible={isItemVisible}
                        />
                      ))}
                    </NForm>
                  )}

                  {tree.value.groups && tree.value.groups.length > 0 && (
                    <NCollapse v-model:expandedNames={expandedGroups.value}>
                      {renderGroups(tree.value.groups)}
                    </NCollapse>
                  )}
                </div>
              )}
            </NSpin>
          ),
        }}
      </NCard>
    )
  },
})
