<script setup lang="ts">
import { NAlert, NButton, NEmpty, NInput, NScrollbar, NTree } from 'naive-ui'
import { computed, h, shallowRef } from 'vue'

import {
  buildRouteCatalogTree,
  collectBranchKeys,
  collectTopLevelBranchKeys,
} from './route-catalog-tree'

import type { RouteCatalogTreeOption } from './route-catalog-tree'
import type { TreeOption } from 'naive-ui'

type TreeKey = string | number
type TreeRenderContext = { option: TreeOption }

interface Props {
  routes: string[]
  total?: number
  truncated?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  total: 0,
  truncated: false,
})

const searchPattern = shallowRef('')
const expandedKeys = shallowRef<TreeKey[]>([])
const treeOptions = computed(() => buildRouteCatalogTree(props.routes))
const groupCount = computed(() => treeOptions.value.filter((node) => node.routePath !== '/').length)

function expandFirstLevel() {
  expandedKeys.value = collectTopLevelBranchKeys(treeOptions.value)
}

function expandAll() {
  expandedKeys.value = collectBranchKeys(treeOptions.value)
}

function collapseAll() {
  searchPattern.value = ''
  expandedKeys.value = []
}

function updateExpandedKeys(keys: TreeKey[]) {
  expandedKeys.value = keys
}

function filterRoute(pattern: string, option: TreeOption): boolean {
  return String(option.routePath).toLocaleLowerCase().includes(pattern.toLocaleLowerCase())
}

function renderLabel({ option }: TreeRenderContext) {
  const route = option as RouteCatalogTreeOption
  return h(
    'span',
    {
      class: 'font-mono text-[13px] text-neutral-700 dark:text-neutral-200',
      title: route.routePath,
    },
    route.label,
  )
}

function renderSuffix({ option }: TreeRenderContext) {
  const route = option as RouteCatalogTreeOption
  if (route.isLeaf) return null

  return h(
    'span',
    {
      class:
        'mr-2 rounded-full bg-neutral-100 px-2 py-0.5 text-[11px] tabular-nums text-neutral-500 dark:bg-neutral-800 dark:text-neutral-400',
    },
    route.routeCount,
  )
}
</script>

<template>
  <NEmpty
    v-if="!routes.length"
    description="暂无路由目录"
  />
  <div
    v-else
    class="space-y-3"
  >
    <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
      <NInput
        v-model:value="searchPattern"
        clearable
        size="small"
        placeholder="搜索完整路由"
      />
      <div class="flex shrink-0 gap-1.5">
        <NButton
          size="small"
          quaternary
          @click="expandFirstLevel"
        >
          展开一级
        </NButton>
        <NButton
          size="small"
          quaternary
          @click="expandAll"
        >
          全部展开
        </NButton>
        <NButton
          size="small"
          quaternary
          @click="collapseAll"
        >
          全部收起
        </NButton>
      </div>
    </div>

    <div class="text-xs text-neutral-500">
      当前载入 {{ routes.length }} 条，共 {{ total }} 条 · {{ groupCount }} 个顶层分组
    </div>

    <NScrollbar style="max-height: 30rem">
      <NTree
        class="pr-2"
        block-line
        expand-on-click
        :data="treeOptions"
        :expanded-keys="expandedKeys"
        :filter="filterRoute"
        :pattern="searchPattern"
        :render-label="renderLabel"
        :render-suffix="renderSuffix"
        :selectable="false"
        :show-irrelevant-nodes="false"
        @update:expanded-keys="updateExpandedKeys"
      />
    </NScrollbar>

    <NAlert
      v-if="truncated"
      type="info"
      :show-icon="false"
    >
      当前目录已按 route_limit 截断；如需查看完整路由，可在页面上方调大查询上限。
    </NAlert>
  </div>
</template>
