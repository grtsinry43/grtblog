<script setup lang="ts">
import * as echarts from 'echarts'
import {
  NCard,
  NDataTable,
  NSelect,
  NSpace,
  NStatistic,
  NTag,
  useMessage,
} from 'naive-ui'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

import { ScrollContainer } from '@/components'
import { getRssAccessStats } from '@/services/rss'

import type { DataTableColumns } from 'naive-ui'
import type { ECharts } from 'echarts'
import type { RssAccessBucket, RssAccessStats } from '@/types/rss'

defineOptions({
  name: 'RssAccessStats',
})

const message = useMessage()
const loading = ref(false)
const stats = ref<RssAccessStats | null>(null)
const days = ref(7)
const topN = ref(12)
const topTab = ref<'clients' | 'ips' | 'platforms' | 'browsers' | 'locations' | 'hints' | 'userAgents'>('clients')

const daysOptions = [
  { label: '最近 7 天', value: 7 },
  { label: '最近 30 天', value: 30 },
  { label: '最近 90 天', value: 90 },
]

const topOptions = [
  { label: 'Top 8', value: 8 },
  { label: 'Top 12', value: 12 },
  { label: 'Top 20', value: 20 },
]

const topData = computed(() => {
  if (!stats.value) return []
  if (topTab.value === 'ips') return stats.value.topIps
  if (topTab.value === 'platforms') return stats.value.topPlatforms
  if (topTab.value === 'browsers') return stats.value.topBrowsers
  if (topTab.value === 'locations') return stats.value.topLocations
  if (topTab.value === 'hints') return stats.value.topHints
  if (topTab.value === 'userAgents') return stats.value.topUserAgents
  return stats.value.topClients
})

const topLabel = computed(() => {
  if (topTab.value === 'ips') return 'IP'
  if (topTab.value === 'platforms') return '操作系统'
  if (topTab.value === 'browsers') return '浏览器'
  if (topTab.value === 'locations') return '地区'
  if (topTab.value === 'hints') return '客户端 Hint'
  if (topTab.value === 'userAgents') return 'User-Agent'
  return '客户端'
})

const trendChartRef = ref<HTMLDivElement | null>(null)
const topChartRef = ref<HTMLDivElement | null>(null)
let trendChart: ECharts | null = null
let topChart: ECharts | null = null

const topTableColumns = computed<DataTableColumns<RssAccessBucket>>(() => [
  {
    title: topLabel.value,
    key: 'name',
    minWidth: 240,
    ellipsis: { tooltip: true },
  },
  {
    title: '请求数',
    key: 'count',
    width: 120,
  },
])

async function loadStats() {
  loading.value = true
  try {
    stats.value = await getRssAccessStats(days.value, topN.value)
    await nextTick()
    renderCharts()
  } catch (error: any) {
    message.error(error?.message || '获取 RSS 统计失败')
  } finally {
    loading.value = false
  }
}

function renderTrendChart() {
  if (!trendChartRef.value || !stats.value) return
  trendChart?.dispose()
  trendChart = echarts.init(trendChartRef.value)
  trendChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { top: 4, data: ['请求量', '去重 IP'] },
    grid: { left: 28, right: 20, top: 32, bottom: 20, containLabel: true },
    xAxis: {
      type: 'category',
      data: stats.value.trend.map((item) => item.hour.slice(5)),
    },
    yAxis: { type: 'value' },
    series: [
      { name: '请求量', type: 'line', smooth: true, data: stats.value.trend.map((item) => item.requests) },
      { name: '去重 IP', type: 'line', smooth: true, data: stats.value.trend.map((item) => item.uniqueIp) },
    ],
  })
}

function renderTopChart() {
  if (!topChartRef.value) return
  topChart?.dispose()
  topChart = echarts.init(topChartRef.value)
  const data = [...topData.value].slice(0, topN.value).reverse()
  topChart.setOption({
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: 120, right: 20, top: 12, bottom: 20, containLabel: false },
    xAxis: { type: 'value' },
    yAxis: {
      type: 'category',
      data: data.map((item) => item.name),
      axisLabel: {
        width: 110,
        overflow: 'truncate',
      },
    },
    series: [
      {
        type: 'bar',
        data: data.map((item) => item.count),
      },
    ],
  })
}

function renderCharts() {
  renderTrendChart()
  renderTopChart()
}

watch([days, topN], async () => {
  await loadStats()
})

watch(topTab, async () => {
  if (!stats.value) return
  await nextTick()
  renderTopChart()
})

onMounted(async () => {
  await loadStats()
  window.addEventListener('resize', renderCharts)
})

onUnmounted(() => {
  window.removeEventListener('resize', renderCharts)
  trendChart?.dispose()
  topChart?.dispose()
})
</script>

<template>
  <ScrollContainer wrapper-class="p-4" :scrollbar-props="{ trigger: 'none' }">
    <NCard title="RSS 访问统计" class="mb-4">
      <template #header-extra>
        <NSpace align="center">
          <NTag size="small" :bordered="false">用户行为埋点聚合</NTag>
          <NSelect v-model:value="days" :options="daysOptions" style="width: 132px" />
          <NSelect v-model:value="topN" :options="topOptions" style="width: 110px" />
        </NSpace>
      </template>

      <div v-if="stats" class="grid grid-cols-1 lg:grid-cols-2 gap-3 mb-4">
        <NCard size="small">
          <NStatistic label="总请求数" :value="stats.total" />
        </NCard>
        <NCard size="small">
          <NStatistic label="去重 IP" :value="stats.uniqueIp" />
        </NCard>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <NCard size="small" title="请求趋势（按小时）" :loading="loading">
          <div ref="trendChartRef" style="height: 300px" />
        </NCard>

        <NCard size="small" :title="`${topLabel} Top`" :loading="loading">
          <template #header-extra>
            <NSelect
              v-model:value="topTab"
              style="width: 140px"
              :options="[
                { label: '客户端', value: 'clients' },
                { label: 'IP', value: 'ips' },
                { label: '操作系统', value: 'platforms' },
                { label: '浏览器', value: 'browsers' },
                { label: '地区', value: 'locations' },
                { label: 'Hint', value: 'hints' },
                { label: 'User-Agent', value: 'userAgents' },
              ]"
            />
          </template>
          <div ref="topChartRef" style="height: 300px" />
        </NCard>
      </div>
    </NCard>

    <NCard size="small" :title="`${topLabel} 明细`" :loading="loading">
      <NDataTable
        :columns="topTableColumns"
        :data="topData"
        :bordered="false"
        :max-height="420"
      />
    </NCard>
  </ScrollContainer>
</template>
