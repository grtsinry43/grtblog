<script setup lang="ts">
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { ArrowClockwise24Regular, Desktop24Regular } from '@vicons/fluent'
import { useIntervalFn } from '@vueuse/core'
import chroma from 'chroma-js'
import * as echarts from 'echarts'
import {
  NButton,
  NCard,
  NDescriptions,
  NDescriptionsItem,
  NEmpty,
  NGrid,
  NGi,
  NIcon,
  NNumberAnimation,
  NProgress,
  NSelect,
  NSpin,
  NTag,
  NTimeline,
  NTimelineItem,
  NSkeleton,
} from 'naive-ui'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

import { ScrollContainer } from '@/components'
import {
  getObservabilityAlerts,
  getObservabilityControlPlane,
  getObservabilityFederation,
  getObservabilityOverview,
  getObservabilityRealtime,
  getObservabilityStorage,
  getObservabilityTimeline,
} from '@/services/observability'
import { getSystemStatus, getSystemUpdateCheck } from '@/services/system'
import { toRefsPreferencesStore } from '@/stores'
import twc from '@/utils/tailwindColor'

import type { ECharts } from 'echarts'

defineOptions({
  name: 'AdvancedInfo',
})

const { isDark, themeColor } = toRefsPreferencesStore()
const queryClient = useQueryClient()
const lastRefreshAt = ref(new Date())
const timelineWindow = ref<'24h' | '7d'>('24h')

const windowOptions = [
  { label: '最近 24 小时', value: '24h' },
  { label: '最近 7 天', value: '7d' },
]

const nowISO = () => new Date().toISOString()
const sinceISO = computed(() => {
  const now = new Date()
  if (timelineWindow.value === '7d') {
    return new Date(now.getTime() - 7 * 24 * 3600 * 1000).toISOString()
  }
  return new Date(now.getTime() - 24 * 3600 * 1000).toISOString()
})

const { data: overviewData, isPending: overviewPending } = useQuery({
  queryKey: ['obs-overview'],
  queryFn: getObservabilityOverview,
  refetchInterval: 15000,
})
const { data: controlData, isPending: controlPending } = useQuery({
  queryKey: ['obs-control'],
  queryFn: () => getObservabilityControlPlane('5m'),
  refetchInterval: 15000,
})
const { data: realtimeData, isPending: realtimePending } = useQuery({
  queryKey: ['obs-realtime'],
  queryFn: getObservabilityRealtime,
  refetchInterval: 8000,
})
const { data: federationData, isPending: federationPending } = useQuery({
  queryKey: ['obs-federation'],
  queryFn: () => getObservabilityFederation('24h'),
  refetchInterval: 15000,
})
const { data: storageData, isPending: storagePending } = useQuery({
  queryKey: ['obs-storage'],
  queryFn: getObservabilityStorage,
  refetchInterval: 20000,
})
const { data: systemData, isPending: systemPending } = useQuery({
  queryKey: ['system-status-advanced'],
  queryFn: getSystemStatus,
  refetchInterval: 15000,
})
const { data: updateData, isPending: updatePending } = useQuery({
  queryKey: ['system-update-check'],
  queryFn: () => getSystemUpdateCheck(false),
  staleTime: 30 * 60 * 1000,
  refetchOnWindowFocus: false,
})
const { data: alertsData } = useQuery({
  queryKey: ['obs-alerts'],
  queryFn: () => getObservabilityAlerts(12),
  refetchInterval: 20000,
})
const { data: timelineData } = useQuery({
  queryKey: ['obs-timeline', timelineWindow],
  queryFn: () =>
    getObservabilityTimeline({
      since: sinceISO.value,
      until: nowISO(),
      group_by: timelineWindow.value === '7d' ? 'day' : 'hour',
    }),
  refetchInterval: 30000,
})

const loading = computed(
  () =>
    overviewPending.value ||
    controlPending.value ||
    realtimePending.value ||
    federationPending.value ||
    storagePending.value ||
    systemPending.value ||
    updatePending.value,
)

const componentHealths = computed(() => systemData.value?.components ?? [])
const updateInfo = computed(() => updateData.value)

function componentTagType(item: { healthy: boolean; status: string }) {
  if (item.healthy) return 'success'
  if (item.status === 'not_configured') return 'warning'
  return 'error'
}

function updateTagType() {
  const info = updateInfo.value
  if (!info) return 'default'
  if (info.status === 'error') return 'error'
  if (info.status === 'disabled') return 'warning'
  if (info.hasUpdate) return 'info'
  return 'success'
}

function formatBytes(bytes?: number) {
  if (!bytes || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = bytes
  let unitIdx = 0
  while (value >= 1024 && unitIdx < units.length - 1) {
    value /= 1024
    unitIdx++
  }
  return `${value.toFixed(value >= 100 ? 0 : 1)} ${units[unitIdx]}`
}

function formatPercent(value?: number) {
  if (value == null || Number.isNaN(value)) return '0%'
  return `${(value * 100).toFixed(2)}%`
}

const cardStats = computed(() => {
  const ov = overviewData.value
  return [
    {
      title: 'API 请求(5m)',
      value: ov?.api.requests ?? 0,
      suffix: 'req',
      iconClass: 'iconify ph--arrows-left-right-bold text-indigo-50 dark:text-indigo-150',
      iconBgClass:
        'text-indigo-500/5 bg-indigo-400 ring-4 ring-indigo-200 dark:bg-indigo-650 dark:ring-indigo-500/30 transition-all',
      description: '最近5分钟请求',
    },
    {
      title: 'API 错误率',
      value: (ov?.api.errorRate ?? 0) * 100,
      suffix: '%',
      precision: 2,
      iconClass: 'iconify ph--warning-circle-bold text-rose-50 dark:text-rose-150',
      iconBgClass:
        'text-rose-500/5 bg-rose-400 ring-4 ring-rose-200 dark:bg-rose-650 dark:ring-rose-500/30 transition-all',
      description: '接口调用异常比例',
    },
    {
      title: '在线连接',
      value: ov?.realtime.currentOnline ?? 0,
      suffix: 'ws',
      iconClass: 'iconify ph--users-three-bold text-blue-50 dark:text-blue-150',
      iconBgClass:
        'text-blue-500/5 bg-blue-400 ring-4 ring-blue-200 dark:bg-blue-650 dark:ring-blue-500/30 transition-all',
      description: '实时 WebSocket 连接',
    },
    {
      title: '联合成功率(24h)',
      value: (ov?.federation.deliverySuccessRate ?? 0) * 100,
      suffix: '%',
      precision: 2,
      iconClass: 'iconify ph--planet-bold text-emerald-50 dark:text-emerald-150',
      iconBgClass:
        'text-emerald-500/5 bg-emerald-400 ring-4 ring-emerald-200 dark:bg-emerald-650 dark:ring-emerald-500/30 transition-all',
      description: '联合投递成功率',
    },
  ]
})

function refreshAll() {
  lastRefreshAt.value = new Date()
  queryClient.invalidateQueries({ queryKey: ['obs-overview'] })
  queryClient.invalidateQueries({ queryKey: ['obs-control'] })
  queryClient.invalidateQueries({ queryKey: ['obs-realtime'] })
  queryClient.invalidateQueries({ queryKey: ['obs-federation'] })
  queryClient.invalidateQueries({ queryKey: ['obs-storage'] })
  queryClient.invalidateQueries({ queryKey: ['system-status-advanced'] })
  queryClient.invalidateQueries({ queryKey: ['obs-alerts'] })
  queryClient.invalidateQueries({ queryKey: ['obs-timeline'] })
}

useIntervalFn(refreshAll, 30000)

const trafficChartEl = ref<HTMLDivElement | null>(null)
const federationChartEl = ref<HTMLDivElement | null>(null)
let trafficChart: ECharts | null = null
let federationChart: ECharts | null = null

const createTooltipConfig = (formatter?: any) => ({
  trigger: 'axis',
  backgroundColor: isDark.value ? twc.neutral[750] : '#fff',
  borderWidth: 1,
  borderColor: isDark.value ? twc.neutral[700] : twc.neutral[150],
  padding: 8,
  extraCssText: 'box-shadow: none;',
  textStyle: {
    color: isDark.value ? twc.neutral[400] : twc.neutral[600],
    fontSize: 12,
  },
  axisPointer: {
    type: 'none',
  },
  ...(formatter && { formatter }),
})

const trafficSeries = computed(() => {
  const items = timelineData.value?.series ?? []
  const xSet = new Set<string>()
  const pvMap = new Map<string, number>()
  const onlineMap = new Map<string, number>()
  const outboundMap = new Map<string, number>()
  for (const item of items) {
    const x = new Date(item.timestamp).toLocaleString()
    xSet.add(x)
    if (item.metric === 'pv') pvMap.set(x, item.value)
    if (item.metric === 'online_peak_avg') onlineMap.set(x, item.value)
    if (item.metric === 'federation_outbound_total') outboundMap.set(x, item.value)
  }
  const xAxis = Array.from(xSet).sort((a, b) => new Date(a).getTime() - new Date(b).getTime())
  return {
    xAxis,
    pv: xAxis.map((x) => pvMap.get(x) ?? 0),
    online: xAxis.map((x) => onlineMap.get(x) ?? 0),
    outbound: xAxis.map((x) => outboundMap.get(x) ?? 0),
  }
})

function renderTrafficChart() {
  if (!trafficChartEl.value) return
  if (!trafficChart) trafficChart = echarts.init(trafficChartEl.value)
  const data = trafficSeries.value
  const color = themeColor.value

  trafficChart.setOption({
    tooltip: createTooltipConfig(),
    legend: {
      data: ['PV', '在线峰值', '联合出站'],
      right: 0,
      top: 0,
      textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] },
    },
    grid: { left: 12, right: 16, top: 36, bottom: 8, containLabel: true },
    xAxis: {
      type: 'category',
      data: data.xAxis,
      axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] },
      axisLine: { show: false },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] },
      splitLine: {
        lineStyle: { color: isDark.value ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.08)' },
      },
    },
    series: [
      {
        name: 'PV',
        type: 'line',
        smooth: true,
        data: data.pv,
        lineStyle: { width: 3, color: color },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: chroma(color).alpha(0.2).hex() },
            { offset: 1, color: chroma(color).alpha(0.02).hex() },
          ]),
        },
        itemStyle: { color },
      },
      {
        name: '在线峰值',
        type: 'line',
        smooth: true,
        data: data.online,
        lineStyle: { width: 2, color: twc.amber[500] },
        itemStyle: { color: twc.amber[500] },
      },
      {
        name: '联邦出站',
        type: 'bar',
        data: data.outbound,
        itemStyle: { color: twc.emerald[500] },
      },
    ],
  })
}

function renderFederationChart() {
  if (!federationChartEl.value) return
  if (!federationChart) federationChart = echarts.init(federationChartEl.value)
  const statusMap = federationData.value?.outboundByStatus ?? {}
  const pieData = Object.entries(statusMap).map(([name, value]) => ({ name, value }))
  federationChart.setOption({
    tooltip: { trigger: 'item' },
    legend: {
      bottom: 0,
      textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] },
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '72%'],
        center: ['50%', '45%'],
        itemStyle: {
          borderRadius: 5,
          borderColor: isDark.value ? twc.neutral[800] : '#fff',
          borderWidth: 2,
        },
        label: {
          formatter: '{b}: {d}%',
          color: isDark.value ? twc.neutral[400] : twc.neutral[600],
        },
        data: pieData,
      },
    ],
  })
}

watch([trafficSeries, isDark, themeColor], () => nextTick(renderTrafficChart), { deep: true })
watch([federationData, isDark], () => nextTick(renderFederationChart), { deep: true })
watch(timelineWindow, () => queryClient.invalidateQueries({ queryKey: ['obs-timeline'] }))

onMounted(() => {
  nextTick(() => {
    renderTrafficChart()
    renderFederationChart()
  })
  window.addEventListener('resize', renderTrafficChart)
  window.addEventListener('resize', renderFederationChart)
})

onUnmounted(() => {
  window.removeEventListener('resize', renderTrafficChart)
  window.removeEventListener('resize', renderFederationChart)
  trafficChart?.dispose()
  federationChart?.dispose()
})
</script>

<template>
  <ScrollContainer wrapper-class="p-4 md:p-6 space-y-4">
    <div class="mb-4 flex flex-col gap-y-3 md:flex-row md:items-center md:justify-between">
      <div class="flex items-center gap-2">
        <NIcon
          :component="Desktop24Regular"
          class="text-xl text-primary"
        />
        <div class="text-lg font-medium">高级信息 / Observability</div>
        <NTag
          size="small"
          type="info"
          round
          >实时</NTag
        >
      </div>
      <div class="flex items-center gap-2 self-end md:self-auto">
        <span class="mr-2 text-xs whitespace-nowrap text-neutral-400"
          >最后刷新：{{ lastRefreshAt.toLocaleTimeString() }}</span
        >
        <NSelect
          v-model:value="timelineWindow"
          :options="windowOptions"
          size="small"
          class="w-32"
        />
        <NButton
          size="small"
          secondary
          :loading="loading"
          @click="refreshAll"
        >
          <template #icon><NIcon :component="ArrowClockwise24Regular" /></template>
          刷新
        </NButton>
      </div>
    </div>

    <!-- Top Cards -->
    <div class="grid grid-cols-1 gap-4 max-sm:gap-2 md:grid-cols-2 lg:grid-cols-4">
      <div
        v-for="(item, index) in cardStats"
        :key="index"
        class="flex items-center justify-between gap-x-4 overflow-hidden rounded border border-naive-border bg-naive-card p-6 transition-[background-color,border-color]"
      >
        <template v-if="!loading || item.value">
          <div class="flex-1">
            <span class="text-sm font-medium text-neutral-450">{{ item.title }}</span>
            <div class="mt-1 mb-1.5 flex gap-x-1 text-2xl text-neutral-700 dark:text-neutral-400">
              <NNumberAnimation
                :to="item.value"
                show-separator
                :precision="item.precision || 0"
              />
              <span class="mb-1 self-end text-xs text-neutral-400">{{ item.suffix }}</span>
            </div>
            <div class="flex items-center">
              <span class="text-xs text-neutral-500 dark:text-neutral-400">{{
                item.description
              }}</span>
            </div>
          </div>
          <div>
            <div
              class="grid place-items-center rounded-full p-3"
              :class="item.iconBgClass"
            >
              <span
                class="size-7"
                :class="item.iconClass"
              />
            </div>
          </div>
        </template>
        <template v-else>
          <div class="flex w-full gap-4">
            <div class="flex-1 space-y-2">
              <NSkeleton
                text
                style="width: 40%"
              />
              <NSkeleton
                text
                style="width: 80%; height: 28px"
              />
              <NSkeleton
                text
                style="width: 60%"
              />
            </div>
            <NSkeleton
              circle
              size="medium"
              style="width: 48px; height: 48px"
            />
          </div>
        </template>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="grid grid-cols-1 gap-4 overflow-hidden max-sm:gap-2 lg:grid-cols-12">
      <!-- Traffic Chart -->
      <div class="col-span-1 lg:col-span-8">
        <div
          class="flex flex-col rounded border border-naive-border bg-naive-card transition-[background-color,border-color]"
          style="height: 420px"
        >
          <div class="flex items-center justify-between px-5 pt-4">
            <span class="text-base font-medium text-neutral-600 dark:text-neutral-300"
              >全链路趋势</span
            >
          </div>
          <div class="flex-1 px-4 pt-2 pb-4">
            <div
              ref="trafficChartEl"
              class="h-full w-full"
            />
          </div>
        </div>
      </div>

      <!-- Federation Chart -->
      <div class="col-span-1 lg:col-span-4">
        <div
          class="flex flex-col rounded border border-naive-border bg-naive-card transition-[background-color,border-color]"
          style="height: 420px"
        >
          <div class="flex items-center justify-between px-5 pt-4">
            <span class="text-base font-medium text-neutral-600 dark:text-neutral-300"
              >联邦出站状态分布</span
            >
          </div>
          <div class="flex-1 px-4 pt-2 pb-4">
            <div
              ref="federationChartEl"
              class="h-full w-full"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- Info Row -->
    <div class="grid grid-cols-1 gap-4 overflow-hidden max-sm:gap-2 lg:grid-cols-12">
      <!-- Control Plane -->
      <div class="col-span-1 lg:col-span-6">
        <div
          class="flex flex-col rounded border border-naive-border bg-naive-card p-5 transition-[background-color,border-color]"
        >
          <div class="mb-4 text-base font-medium text-neutral-600 dark:text-neutral-300">
            控制平面概况
          </div>
          <NDescriptions
            :column="2"
            size="small"
          >
            <NDescriptionsItem label="RPS">{{
              controlData?.api.rps?.toFixed(2)
            }}</NDescriptionsItem>
            <NDescriptionsItem label="P95 延迟"
              >{{ controlData?.api.p95LatencyMs?.toFixed(1) }} ms</NDescriptionsItem
            >
            <NDescriptionsItem label="API 错误率">{{
              formatPercent(controlData?.api.errorRate)
            }}</NDescriptionsItem>
            <NDescriptionsItem label="Go Goroutines">{{
              controlData?.goRuntime.numGoroutine
            }}</NDescriptionsItem>
            <NDescriptionsItem label="DB 连接状态">
              <NTag
                :type="controlData?.database.status === 'connected' ? 'success' : 'error'"
                size="small"
                round
              >
                {{ controlData?.database.status || 'unknown' }}
              </NTag>
            </NDescriptionsItem>
            <NDescriptionsItem label="DB 等待">{{
              controlData?.database.waitCount
            }}</NDescriptionsItem>
          </NDescriptions>
        </div>
      </div>

      <!-- Realtime & Storage -->
      <div class="col-span-1 lg:col-span-6">
        <div
          class="flex flex-col rounded border border-naive-border bg-naive-card p-5 transition-[background-color,border-color]"
        >
          <div class="mb-4 text-base font-medium text-neutral-600 dark:text-neutral-300">
            实时与存储
          </div>
          <NDescriptions
            :column="2"
            size="small"
          >
            <NDescriptionsItem label="WS 在线">{{
              realtimeData?.snapshot.currentOnline
            }}</NDescriptionsItem>
            <NDescriptionsItem label="WS 广播错误率">{{
              formatPercent(realtimeData?.snapshot.broadcastErrorRate)
            }}</NDescriptionsItem>
            <NDescriptionsItem label="WS Fanout P95">
              {{ realtimeData?.snapshot.broadcastP95Ms?.toFixed(1) }} ms
            </NDescriptionsItem>
            <NDescriptionsItem label="平均接收人数">{{
              realtimeData?.snapshot.avgRecipients?.toFixed(2)
            }}</NDescriptionsItem>
            <NDescriptionsItem label="HTML 存储">{{
              formatBytes(storageData?.storageHtml.size)
            }}</NDescriptionsItem>
            <NDescriptionsItem label="日志存储">{{
              formatBytes(storageData?.storageLogs.size)
            }}</NDescriptionsItem>
          </NDescriptions>
          <div class="mt-4 border-t border-neutral-100 pt-4 dark:border-neutral-800">
            <div class="mb-1 flex items-center justify-between">
              <span class="text-xs text-neutral-500">Redis 队列深度</span>
              <span class="text-xs text-neutral-400">{{
                storageData?.redis.analyticsQueueDepth || 0
              }}</span>
            </div>
            <NProgress
              type="line"
              status="info"
              :percentage="
                Math.min(((storageData?.redis.analyticsQueueDepth || 0) / 1000) * 100, 100)
              "
              :show-indicator="false"
              processing
            />
          </div>
        </div>
      </div>
    </div>

    <!-- Component Health -->
    <div
      class="rounded border border-naive-border bg-naive-card p-5 transition-[background-color,border-color]"
    >
      <div class="mb-4 text-base font-medium text-neutral-600 dark:text-neutral-300">组件健康状态</div>
      <NEmpty
        v-if="!componentHealths.length"
        description="暂无组件健康数据"
      />
      <div
        v-else
        class="space-y-3"
      >
        <div class="flex flex-wrap gap-2">
          <NTag
            v-for="item in componentHealths"
            :key="item.name"
            :type="componentTagType(item)"
            size="small"
            round
          >
            {{ item.name }} · {{ item.status }} · v{{ item.version || 'n/a' }}
          </NTag>
        </div>
        <div class="text-xs text-neutral-500">
          检查时间：{{ componentHealths[0]?.checkedAt ? new Date(componentHealths[0].checkedAt).toLocaleString() : '-' }}
        </div>
      </div>
    </div>

    <!-- Update Check -->
    <div
      class="rounded border border-naive-border bg-naive-card p-5 transition-[background-color,border-color]"
    >
      <div class="mb-4 flex items-center justify-between">
        <div class="text-base font-medium text-neutral-600 dark:text-neutral-300">更新检查</div>
        <NTag
          size="small"
          :type="updateTagType()"
          round
        >
          {{
            updateInfo?.status === 'error'
              ? '检查失败'
              : updateInfo?.status === 'disabled'
                ? '已关闭'
                : updateInfo?.hasUpdate
                  ? '可更新'
                  : '已最新'
          }}
        </NTag>
      </div>
      <NEmpty
        v-if="!updateInfo"
        description="暂无更新信息"
      />
      <NDescriptions
        v-else
        :column="2"
        size="small"
      >
        <NDescriptionsItem label="当前版本">{{ updateInfo.currentVersion }}</NDescriptionsItem>
        <NDescriptionsItem label="更新通道">{{ updateInfo.channel }}</NDescriptionsItem>
        <NDescriptionsItem label="目标版本">
          {{ updateInfo.targetRelease?.tag || '-' }}
          <NTag
            v-if="updateInfo.targetRelease?.prerelease"
            class="ml-2"
            size="tiny"
            type="warning"
            round
            >测试版</NTag
          >
        </NDescriptionsItem>
        <NDescriptionsItem label="检查时间">{{
          updateInfo.checkedAt ? new Date(updateInfo.checkedAt).toLocaleString() : '-'
        }}</NDescriptionsItem>
      </NDescriptions>
      <div class="mt-3 flex items-center justify-between gap-2">
        <span class="text-xs text-neutral-500">{{ updateInfo?.message || '版本来源：GitHub Releases' }}</span>
        <NButton
          v-if="updateInfo?.upgradeUrl"
          size="small"
          secondary
          tag="a"
          :href="updateInfo.upgradeUrl"
          target="_blank"
        >
          查看 Release
        </NButton>
      </div>
    </div>

    <!-- Alerts -->
    <div
      class="rounded border border-naive-border bg-naive-card p-5 transition-[background-color,border-color]"
    >
      <div class="mb-4 text-base font-medium text-neutral-600 dark:text-neutral-300">
        系统告警流
      </div>
      <NEmpty
        v-if="!alertsData?.items?.length"
        description="暂无告警"
      />
      <NTimeline v-else>
        <NTimelineItem
          v-for="item in alertsData.items"
          :key="item.id"
          :type="item.isRead ? 'default' : 'warning'"
          :title="item.title"
          :time="new Date(item.createdAt).toLocaleString()"
        >
          <div class="mb-1 text-xs text-neutral-500">{{ item.type }}</div>
          <div class="text-sm text-neutral-700 dark:text-neutral-300">{{ item.content }}</div>
        </NTimelineItem>
      </NTimeline>
    </div>
  </ScrollContainer>
</template>
