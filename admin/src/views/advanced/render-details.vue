<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { ArrowClockwise24Regular, Flash24Regular, Rocket24Regular } from '@vicons/fluent'
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NEmpty,
  NIcon,
  NInput,
  NInputGroup,
  NInputGroupLabel,
  NSpace,
  NSpin,
  NSwitch,
  NTag,
  NTimeline,
  NTimelineItem,
  NTree,
  useMessage,
} from 'naive-ui'
import { computed, h, ref } from 'vue'

import { ScrollContainer } from '@/components'
import {
  bootstrapObservabilityPages,
  getObservabilityPages,
  invalidateObservabilityPages,
} from '@/services/observability'

import type {
  ObservabilityInvalidateReport,
  ObservabilityRenderRecord,
} from '@/types/observability'
import type { DataTableColumns, TreeOption } from 'naive-ui'

defineOptions({
  name: 'AdvancedRenderDetails',
})

const message = useMessage()
const queryClient = useQueryClient()

const trackedLimit = ref(200)
const recentLimit = ref(30)
const routeLimit = ref(500)
const source = ref('admin:render-details')
const syncRender = ref(true)
const depKeysInput = ref('')
const urlsInput = ref('')
const lastInvalidateReport = ref<ObservabilityInvalidateReport | null>(null)

const { data: pageStateData, isPending } = useQuery({
  queryKey: ['obs-pages', trackedLimit, recentLimit, routeLimit],
  queryFn: () =>
    getObservabilityPages({
      tracked_limit: trackedLimit.value,
      recent_limit: recentLimit.value,
      route_limit: routeLimit.value,
    }),
  refetchInterval: 15000,
})

const bootstrapMutation = useMutation({
  mutationFn: bootstrapObservabilityPages,
  onSuccess: (data) => {
    message.success(`冷启动完成：共 ${data.totalRoutes} 路由，成功渲染 ${data.renderedCount} 次`)
    queryClient.invalidateQueries({ queryKey: ['obs-pages'] })
  },
})

const invalidateMutation = useMutation({
  mutationFn: invalidateObservabilityPages,
  onSuccess: (data) => {
    lastInvalidateReport.value = data
    message.success(
      `更新完成：候选 ${ensureArray(data.candidateUrls).length}，已渲染 ${ensureArray(data.rendered).length}`,
    )
    queryClient.invalidateQueries({ queryKey: ['obs-pages'] })
  },
})

const snapshot = computed(() => pageStateData.value?.snapshot)
const routeCatalog = computed(() => pageStateData.value?.routeCatalog)
const bootstrapReport = computed(() => snapshot.value?.lastBootstrap)
const bootstrapPending = computed(() => bootstrapMutation.isPending.value)
const invalidatePending = computed(() => invalidateMutation.isPending.value)

const renderColumns: DataTableColumns<ObservabilityRenderRecord> = [
  {
    title: '页面',
    key: 'urlPath',
    minWidth: 240,
    ellipsis: { tooltip: true },
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) =>
      h(
        NTag,
        {
          size: 'small',
          type:
            row.status === 'error' ? 'error' : row.status === 'not_found' ? 'warning' : 'success',
        },
        { default: () => row.status },
      ),
  },
  {
    title: '触发源',
    key: 'trigger',
    width: 180,
    ellipsis: { tooltip: true },
  },
  {
    title: '依赖数',
    key: 'depsCount',
    width: 90,
    render: (row) => row.deps?.length ?? 0,
  },
  {
    title: '更新文件',
    key: 'updatedFiles',
    width: 90,
    render: (row) => row.updatedFiles?.length ?? 0,
  },
  {
    title: '耗时',
    key: 'durationMs',
    width: 90,
    render: (row) => `${row.durationMs}ms`,
  },
]

const treeOptions = computed<TreeOption[]>(() => {
  const root = pageStateData.value?.tree
  if (!root) {
    return []
  }
  return [toTreeOption(root)]
})

function ensureArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : []
}

function toTreeOption(node: any): TreeOption {
  const details: string[] = []
  if (node.routePath) details.push(node.routePath)
  if (node.tracked) details.push(`deps:${node.deps?.length ?? 0}`)
  if (node.hasHtml) details.push('html')
  if (node.hasData) details.push('data')
  const label = details.length > 0 ? `${node.name} (${details.join(' | ')})` : node.name
  return {
    key: `${node.path}:${node.name}`,
    label,
    children: (node.children ?? []).map((child: any) => toTreeOption(child)),
    isLeaf: node.nodeType === 'file',
  }
}

function splitByLine(input: string) {
  return input
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter((item) => item.length > 0)
}

function refreshPageState() {
  queryClient.invalidateQueries({ queryKey: ['obs-pages'] })
}

function triggerBootstrap() {
  bootstrapMutation.mutate()
}

function triggerInvalidate() {
  const depKeys = splitByLine(depKeysInput.value)
  const urls = splitByLine(urlsInput.value)
  if (depKeys.length === 0 && urls.length === 0) {
    message.warning('至少填写 depKeys 或 urls 之一')
    return
  }
  invalidateMutation.mutate({
    depKeys,
    urls,
    source: source.value.trim() || 'admin:render-details',
    syncRender: syncRender.value,
  })
}
</script>

<template>
  <ScrollContainer wrapper-class="p-4 md:p-6 space-y-4">
    <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
      <div class="flex items-center gap-2">
        <NIcon
          :component="Flash24Regular"
          class="text-xl text-primary"
        />
        <div class="text-lg font-medium">高级信息 / 渲染详情</div>
      </div>
      <div class="flex items-center gap-2">
        <NButton
          size="small"
          secondary
          :loading="isPending"
          @click="refreshPageState"
        >
          <template #icon><NIcon :component="ArrowClockwise24Regular" /></template>
          刷新状态
        </NButton>
        <NButton
          size="small"
          type="warning"
          :loading="bootstrapPending"
          @click="triggerBootstrap"
        >
          <template #icon><NIcon :component="Rocket24Regular" /></template>
          冷启动全量渲染
        </NButton>
      </div>
    </div>

    <NSpin :show="isPending">
      <div class="grid grid-cols-1 gap-4 lg:grid-cols-12">
        <NCard
          title="ISR 状态"
          class="lg:col-span-6"
        >
          <NDescriptions
            :column="2"
            size="small"
          >
            <NDescriptionsItem label="队列深度">{{ snapshot?.queueDepth ?? 0 }}</NDescriptionsItem>
            <NDescriptionsItem label="依赖键">{{ snapshot?.depKeyCount ?? 0 }}</NDescriptionsItem>
            <NDescriptionsItem label="页面键">{{ snapshot?.urlKeyCount ?? 0 }}</NDescriptionsItem>
            <NDescriptionsItem label="已追踪页面">{{
              snapshot?.trackedPages?.length ?? 0
            }}</NDescriptionsItem>
            <NDescriptionsItem label="路由总数">{{ routeCatalog?.total ?? 0 }}</NDescriptionsItem>
            <NDescriptionsItem label="路由截断">{{
              routeCatalog?.truncated ? '是' : '否'
            }}</NDescriptionsItem>
          </NDescriptions>
          <div class="mt-3 text-xs text-neutral-500">
            最近引导：{{
              bootstrapReport?.finishedAt
                ? new Date(bootstrapReport.finishedAt).toLocaleString()
                : '无'
            }}
          </div>
        </NCard>

        <NCard
          title="手动触发更新"
          class="lg:col-span-6"
        >
          <div class="space-y-3">
            <NInputGroup>
              <NInputGroupLabel>source</NInputGroupLabel>
              <NInput
                v-model:value="source"
                placeholder="admin:render-details"
              />
            </NInputGroup>
            <NInput
              v-model:value="depKeysInput"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 6 }"
              placeholder="depKeys，一行一个，例如：&#10;post:list:page:1&#10;post:detail:123"
            />
            <NInput
              v-model:value="urlsInput"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 6 }"
              placeholder="urls，一行一个，例如：&#10;/&#10;/posts/example"
            />
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2 text-xs text-neutral-500">
                <span>同步渲染</span>
                <NSwitch v-model:value="syncRender" />
              </div>
              <NButton
                type="primary"
                :loading="invalidatePending"
                @click="triggerInvalidate"
              >
                触发更新
              </NButton>
            </div>
          </div>
        </NCard>
      </div>

      <div class="grid grid-cols-1 gap-4 lg:grid-cols-12">
        <NCard
          title="页面目录树"
          class="lg:col-span-7"
        >
          <NEmpty
            v-if="!treeOptions.length"
            description="暂无页面目录"
          />
          <NTree
            v-else
            block-line
            expand-on-click
            :default-expand-all="false"
            :data="treeOptions"
          />
        </NCard>

        <NCard
          title="路由目录"
          class="lg:col-span-5"
        >
          <NEmpty
            v-if="!routeCatalog?.items?.length"
            description="暂无路由目录"
          />
          <NSpace
            v-else
            size="small"
          >
            <NTag
              v-for="item in routeCatalog.items"
              :key="item"
              size="small"
              type="default"
            >
              {{ item }}
            </NTag>
          </NSpace>
          <NAlert
            v-if="routeCatalog?.truncated"
            class="mt-3"
            type="info"
            :show-icon="false"
          >
            路由列表已截断，可调大 route_limit 查询更多。
          </NAlert>
        </NCard>
      </div>

      <NCard title="本次更新回执">
        <NEmpty
          v-if="!lastInvalidateReport"
          description="尚未触发手动更新"
        />
        <div
          v-else
          class="space-y-3"
        >
          <NDescriptions
            :column="3"
            size="small"
          >
            <NDescriptionsItem label="来源">{{ lastInvalidateReport.source }}</NDescriptionsItem>
            <NDescriptionsItem label="候选页面">{{
              ensureArray(lastInvalidateReport.candidateUrls).length
            }}</NDescriptionsItem>
            <NDescriptionsItem label="渲染记录">{{
              ensureArray(lastInvalidateReport.rendered).length
            }}</NDescriptionsItem>
            <NDescriptionsItem label="命中依赖">{{
              ensureArray(lastInvalidateReport.matchedUrls).length
            }}</NDescriptionsItem>
            <NDescriptionsItem label="入队页面">{{
              ensureArray(lastInvalidateReport.enqueuedUrls).length
            }}</NDescriptionsItem>
            <NDescriptionsItem label="队列深度">{{
              lastInvalidateReport.queueDepth
            }}</NDescriptionsItem>
          </NDescriptions>
          <NDataTable
            :columns="renderColumns"
            :data="ensureArray(lastInvalidateReport.rendered)"
            :pagination="{ pageSize: 8 }"
            size="small"
          />
        </div>
      </NCard>

      <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <NCard title="最近依赖失效活动">
          <NEmpty
            v-if="!snapshot?.recentInvalidations?.length"
            description="暂无失效记录"
          />
          <NTimeline v-else>
            <NTimelineItem
              v-for="item in snapshot.recentInvalidations"
              :key="`${item.generatedAt}:${item.source}`"
              :title="item.source"
              :time="new Date(item.generatedAt).toLocaleString()"
              type="warning"
            >
              <div class="text-xs text-neutral-500">
                depKeys: {{ ensureArray(item.depKeys).join(', ') || 'none' }}
              </div>
              <div class="text-xs text-neutral-500">
                candidate: {{ ensureArray(item.candidateUrls).length }} / rendered:
                {{ ensureArray(item.renderedUrls).length }}
              </div>
            </NTimelineItem>
          </NTimeline>
        </NCard>

        <NCard title="最近渲染活动">
          <NEmpty
            v-if="!snapshot?.recentRenderActivity?.length"
            description="暂无渲染记录"
          />
          <NTimeline v-else>
            <NTimelineItem
              v-for="item in snapshot.recentRenderActivity"
              :key="`${item.generatedAt}:${item.urlPath}:${item.trigger}`"
              :title="item.urlPath"
              :time="new Date(item.generatedAt).toLocaleString()"
              :type="item.status === 'error' ? 'error' : 'success'"
            >
              <div class="text-xs text-neutral-500">
                {{ item.trigger }} / {{ item.status }} / {{ item.durationMs }}ms
              </div>
              <div class="text-xs text-neutral-500">
                deps: {{ item.deps?.length ?? 0 }}, updated: {{ item.updatedFiles?.length ?? 0 }}
              </div>
            </NTimelineItem>
          </NTimeline>
        </NCard>
      </div>
    </NSpin>
  </ScrollContainer>
</template>
