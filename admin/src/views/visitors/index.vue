<script setup lang="ts">
import * as echarts from 'echarts'
import {
  NButton,
  NCard,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NDrawer,
  NDrawerContent,
  NInput,
  NSelect,
  NSpace,
  NStatistic,
  NTag,
  NText,
  useMessage,
} from 'naive-ui'
import { computed, h, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

import { useTable } from '@/composables/table/use-table'
import { getVisitorInsights, getVisitorProfile, listVisitors } from '@/services/visitors'
import { ScrollContainer } from '@/components'

import type { DataTableColumns } from 'naive-ui'
import type { ECharts } from 'echarts'
import type { VisitorInsights, VisitorProfile, VisitorRecentComment } from '@/types/visitors'

defineOptions({
  name: 'VisitorProfileList',
})

const message = useMessage()
const keyword = ref('')
const queryState = ref({ keyword: '' })

const {
  loading,
  data: tableData,
  pagination,
  refresh,
} = useTable<VisitorProfile>(listVisitors, queryState.value)

const detailVisible = ref(false)
const detailLoading = ref(false)
const currentProfile = ref<VisitorProfile | null>(null)
const recentComments = ref<VisitorRecentComment[]>([])

const insightDays = ref<number>(30)
const sourceTab = ref<'platform' | 'browser' | 'location'>('platform')
const insightsLoading = ref(false)
const insights = ref<VisitorInsights | null>(null)

const sourceChartRef = ref<HTMLDivElement | null>(null)
const trendChartRef = ref<HTMLDivElement | null>(null)
const funnelChartRef = ref<HTMLDivElement | null>(null)
let sourceChart: ECharts | null = null
let trendChart: ECharts | null = null
let funnelChart: ECharts | null = null

const daysOptions = [
  { label: '最近 7 天', value: 7 },
  { label: '最近 30 天', value: 30 },
  { label: '最近 90 天', value: 90 },
]

const statusTagTypeMap: Record<string, 'default' | 'info' | 'warning' | 'success' | 'error'> = {
  pending: 'warning',
  approved: 'success',
  rejected: 'error',
  blocked: 'default',
}

const sourceSeries = computed(() => {
  if (!insights.value) return []
  if (sourceTab.value === 'platform') return insights.value.platformTop
  if (sourceTab.value === 'browser') return insights.value.browserTop
  return insights.value.locationTop
})

const dataSourceLabel = computed(() => {
  if (!insights.value) return '-'
  return insights.value.dataSource === 'api' ? '用户行为埋点聚合' : '浏览埋点聚合'
})

const columns = computed<DataTableColumns<VisitorProfile>>(() => [
  {
    title: '访客 ID',
    key: 'visitorId',
    minWidth: 220,
    ellipsis: { tooltip: true },
    render: (row) => h('code', {}, row.visitorId),
  },
  { title: '昵称', key: 'nickName', width: 120, render: (row) => row.nickName || '-' },
  {
    title: '邮箱',
    key: 'email',
    minWidth: 180,
    ellipsis: { tooltip: true },
    render: (row) => row.email || '-',
  },
  { title: '地区', key: 'location', width: 140, render: (row) => row.location || '-' },
  {
    title: '设备',
    key: 'device',
    minWidth: 180,
    render: (row) => [row.browser, row.platform].filter(Boolean).join(' / ') || '-',
  },
  { title: '浏览', key: 'totalViews', width: 90 },
  { title: '点赞', key: 'totalLikes', width: 90 },
  { title: '评论', key: 'totalComments', width: 90 },
  {
    title: '最近活跃',
    key: 'lastSeenAt',
    width: 180,
    render: (row) => formatDate(row.lastSeenAt),
  },
  {
    title: '操作',
    key: 'actions',
    width: 96,
    render: (row) =>
      h(
        NButton,
        { size: 'small', tertiary: true, onClick: () => openProfile(row.visitorId) },
        { default: () => '详情' },
      ),
  },
])

function formatDate(value?: string) {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

function toPercent(value: number) {
  return `${(value * 100).toFixed(1)}%`
}

async function loadInsights() {
  insightsLoading.value = true
  try {
    insights.value = await getVisitorInsights(insightDays.value)
    await nextTick()
    renderCharts()
  } catch (error: any) {
    message.error(error?.message || '获取访客统计失败')
  } finally {
    insightsLoading.value = false
  }
}

function renderSourceChart() {
  if (!sourceChartRef.value) return
  sourceChart?.dispose()
  sourceChart = echarts.init(sourceChartRef.value)
  sourceChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { top: 4 },
    series: [
      {
        type: 'pie',
        radius: ['38%', '66%'],
        center: ['50%', '58%'],
        data: sourceSeries.value.map((item) => ({ name: item.name, value: item.count })),
      },
    ],
  })
}

function renderTrendChart() {
  if (!trendChartRef.value || !insights.value) return
  trendChart?.dispose()
  trendChart = echarts.init(trendChartRef.value)
  trendChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { top: 4 },
    grid: { left: 28, right: 20, top: 30, bottom: 20, containLabel: true },
    xAxis: {
      type: 'category',
      data: insights.value.trend.map((item) => item.date.slice(5)),
    },
    yAxis: { type: 'value' },
    series: [
      { name: '活跃访客', type: 'line', smooth: true, data: insights.value.trend.map((item) => item.activeVisitors) },
      { name: '浏览', type: 'line', smooth: true, data: insights.value.trend.map((item) => item.views) },
      { name: '点赞', type: 'line', smooth: true, data: insights.value.trend.map((item) => item.likes) },
      { name: '评论', type: 'line', smooth: true, data: insights.value.trend.map((item) => item.comments) },
    ],
  })
}

function renderFunnelChart() {
  if (!funnelChartRef.value || !insights.value) return
  funnelChart?.dispose()
  funnelChart = echarts.init(funnelChartRef.value)
  funnelChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: ['浏览访客', '点赞访客', '评论访客'],
    },
    yAxis: { type: 'value' },
    series: [
      {
        type: 'bar',
        data: [
          insights.value.funnel.viewVisitors,
          insights.value.funnel.likeVisitors,
          insights.value.funnel.commentVisitors,
        ],
      },
    ],
  })
}

function renderCharts() {
  renderSourceChart()
  renderTrendChart()
  renderFunnelChart()
}

function doSearch() {
  queryState.value.keyword = keyword.value.trim()
  pagination.page = 1
  refresh()
}

function resetSearch() {
  keyword.value = ''
  queryState.value.keyword = ''
  pagination.page = 1
  refresh()
}

async function openProfile(visitorId: string) {
  detailVisible.value = true
  detailLoading.value = true
  currentProfile.value = null
  recentComments.value = []
  try {
    const detail = await getVisitorProfile(visitorId, 20)
    currentProfile.value = detail.profile
    recentComments.value = detail.recentComments || []
  } catch (error: any) {
    message.error(error?.message || '获取访客详情失败')
    detailVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

watch(insightDays, async () => {
  await loadInsights()
})

watch(sourceTab, async () => {
  if (!insights.value) return
  await nextTick()
  renderSourceChart()
})

onMounted(async () => {
  await loadInsights()
  window.addEventListener('resize', renderCharts)
})

onUnmounted(() => {
  window.removeEventListener('resize', renderCharts)
  sourceChart?.dispose()
  trendChart?.dispose()
  funnelChart?.dispose()
})
</script>

<template>
  <ScrollContainer wrapper-class="p-4" :scrollbar-props="{ trigger: 'none' }">
    <NCard title="访客画像管理" class="mb-4">
      <template #header-extra>
        <NSpace align="center">
          <NTag size="small" :bordered="false">
            数据来源：{{ dataSourceLabel }}
          </NTag>
          <NSelect v-model:value="insightDays" :options="daysOptions" style="width: 132px" />
        </NSpace>
      </template>

      <div v-if="insights" class="grid grid-cols-1 lg:grid-cols-4 gap-3 mb-4">
        <NCard size="small">
          <NStatistic label="1天活跃访客" :value="insights.segments.active1d" />
        </NCard>
        <NCard size="small">
          <NStatistic label="7天活跃访客" :value="insights.segments.active7d" />
        </NCard>
        <NCard size="small">
          <NStatistic label="30天活跃访客" :value="insights.segments.active30d" />
        </NCard>
        <NCard size="small">
          <NStatistic label="高活跃访客" :value="insights.segments.highlyEngaged" />
        </NCard>
      </div>

      <div v-if="insights" class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <NCard size="small" title="来源分布">
          <template #header-extra>
            <NSpace align="center">
              <NTag size="tiny" :bordered="false">用户行为埋点聚合</NTag>
              <NSelect
                v-model:value="sourceTab"
                style="width: 120px"
                :options="[
                  { label: '操作系统', value: 'platform' },
                  { label: '浏览器', value: 'browser' },
                  { label: '地区', value: 'location' },
                ]"
              />
            </NSpace>
          </template>
          <div ref="sourceChartRef" style="height: 280px" />
        </NCard>

        <NCard size="small" title="行为漏斗">
          <template #header-extra>
            <NTag size="tiny" :bordered="false">用户行为埋点聚合</NTag>
          </template>
          <div ref="funnelChartRef" style="height: 280px" />
          <NSpace class="mt-2">
            <NTag type="info">点赞率 {{ toPercent(insights.funnel.likeRate) }}</NTag>
            <NTag type="warning">评论率(按浏览) {{ toPercent(insights.funnel.commentRateByView) }}</NTag>
            <NTag type="success">评论率(按点赞) {{ toPercent(insights.funnel.commentRateByLike) }}</NTag>
          </NSpace>
        </NCard>
      </div>

      <NCard v-if="insights" size="small" title="活跃趋势" class="mt-4">
        <template #header-extra>
          <NTag size="tiny" :bordered="false">用户行为埋点聚合</NTag>
        </template>
        <div ref="trendChartRef" style="height: 320px" />
      </NCard>
    </NCard>

    <NCard title="访客列表">
      <NSpace class="mb-4" align="center">
        <NInput
          v-model:value="keyword"
          placeholder="搜索 visitorId / 昵称 / 邮箱 / IP / 地区 / 设备"
          clearable
          style="width: 380px"
          @keyup.enter="doSearch"
        />
        <NButton type="primary" @click="doSearch">查询</NButton>
        <NButton @click="resetSearch">重置</NButton>
      </NSpace>

      <NDataTable
        remote
        :loading="loading || insightsLoading"
        :columns="columns"
        :data="tableData"
        :pagination="pagination"
      />
    </NCard>

    <NDrawer v-model:show="detailVisible" width="760">
      <NDrawerContent title="访客画像详情" :native-scrollbar="false">
        <div v-if="detailLoading" class="py-8 text-center">
          <NText depth="3">加载中...</NText>
        </div>

        <template v-else-if="currentProfile">
          <NDescriptions bordered label-placement="left" :column="2" class="mb-4">
            <NDescriptionsItem label="访客 ID">
              <code>{{ currentProfile.visitorId }}</code>
            </NDescriptionsItem>
            <NDescriptionsItem label="昵称">{{ currentProfile.nickName || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="邮箱">{{ currentProfile.email || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="网站">
              <a
                v-if="currentProfile.website"
                :href="currentProfile.website"
                target="_blank"
                class="text-primary hover:underline"
              >
                {{ currentProfile.website }}
              </a>
              <span v-else>-</span>
            </NDescriptionsItem>
            <NDescriptionsItem label="IP">{{ currentProfile.ip || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="地区">{{ currentProfile.location || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="浏览器 / 平台">
              {{ [currentProfile.browser, currentProfile.platform].filter(Boolean).join(' / ') || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem label="首次出现">{{ formatDate(currentProfile.firstSeenAt) }}</NDescriptionsItem>
            <NDescriptionsItem label="最近活跃">{{ formatDate(currentProfile.lastSeenAt) }}</NDescriptionsItem>
            <NDescriptionsItem label="最近浏览">{{ formatDate(currentProfile.lastViewedAt) }}</NDescriptionsItem>
            <NDescriptionsItem label="最近点赞">{{ formatDate(currentProfile.lastLikedAt) }}</NDescriptionsItem>
          </NDescriptions>

          <NSpace class="mb-4">
            <NTag type="info">浏览 {{ currentProfile.totalViews }}</NTag>
            <NTag type="info">浏览内容数 {{ currentProfile.uniqueViewItems }}</NTag>
            <NTag type="success">点赞 {{ currentProfile.totalLikes }}</NTag>
            <NTag type="success">点赞内容数 {{ currentProfile.uniqueLikedItems }}</NTag>
            <NTag type="warning">评论 {{ currentProfile.totalComments }}</NTag>
          </NSpace>

          <NCard title="最近评论" size="small">
            <div v-if="recentComments.length === 0" class="text-center py-4 text-[var(--text-color-3)]">
              暂无评论记录
            </div>
            <NSpace v-else vertical :size="12">
              <div v-for="item in recentComments" :key="item.id" class="rounded border border-gray-200 p-3">
                <NSpace justify="space-between" align="center" class="mb-2">
                  <NSpace align="center">
                    <NTag size="small" :type="statusTagTypeMap[item.status] || 'default'">
                      {{ item.status }}
                    </NTag>
                    <NTag v-if="item.isDeleted" size="small" type="error">已删除</NTag>
                  </NSpace>
                  <NText depth="3" style="font-size: 12px">
                    {{ formatDate(item.createdAt) }}
                  </NText>
                </NSpace>
                <div class="text-sm whitespace-pre-wrap break-all">{{ item.content }}</div>
              </div>
            </NSpace>
          </NCard>
        </template>
      </NDrawerContent>
    </NDrawer>
  </ScrollContainer>
</template>
