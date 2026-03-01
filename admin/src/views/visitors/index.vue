<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NInput,
  NSelect,
  NSpace,
  NStatistic,
  NTag,
  useMessage,
} from 'naive-ui'
import { computed, h, ref } from 'vue'
import { formatDate } from '@/utils/format'

import { useTable } from '@/composables/table/use-table'
import { getVisitorProfile, listVisitors } from '@/services/visitors'
import { ScrollContainer } from '@/components'
import { useVisitorInsights } from './composables/use-visitor-insights'
import VisitorDetailDrawer from './components/VisitorDetailDrawer.vue'

import type { DataTableColumns } from 'naive-ui'
import type { VisitorProfile, VisitorRecentComment } from '@/types/visitors'

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

const {
  insightDays,
  sourceTab,
  insightsLoading,
  insights,
  sourceChartRef,
  trendChartRef,
  funnelChartRef,
  daysOptions,
  dataSourceLabel,
  toPercent,
} = useVisitorInsights(message)

const columns = computed<DataTableColumns<VisitorProfile>>(() => [
  {
    title: '访客 ID', key: 'visitorId', minWidth: 220,
    ellipsis: { tooltip: true },
    render: (row) => h('code', {}, row.visitorId),
  },
  { title: '昵称', key: 'nickName', width: 120, render: (row) => row.nickName || '-' },
  { title: '邮箱', key: 'email', minWidth: 180, ellipsis: { tooltip: true }, render: (row) => row.email || '-' },
  { title: '地区', key: 'location', width: 140, render: (row) => row.location || '-' },
  { title: '设备', key: 'device', minWidth: 180, render: (row) => [row.browser, row.platform].filter(Boolean).join(' / ') || '-' },
  { title: '浏览', key: 'totalViews', width: 90 },
  { title: '点赞', key: 'totalLikes', width: 90 },
  { title: '评论', key: 'totalComments', width: 90 },
  { title: '最近活跃', key: 'lastSeenAt', width: 180, render: (row) => formatDate(row.lastSeenAt) },
  {
    title: '操作', key: 'actions', width: 96,
    render: (row) => h(NButton, { size: 'small', tertiary: true, onClick: () => openProfile(row.visitorId) }, { default: () => '详情' }),
  },
])

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
</script>

<template>
  <ScrollContainer wrapper-class="p-4" :scrollbar-props="{ trigger: 'none' }">
    <NCard title="访客画像管理" class="mb-4">
      <template #header-extra>
        <NSpace align="center">
          <NTag size="small" :bordered="false">数据来源：{{ dataSourceLabel }}</NTag>
          <NSelect v-model:value="insightDays" :options="daysOptions" style="width: 132px" />
        </NSpace>
      </template>

      <div v-if="insights" class="grid grid-cols-1 lg:grid-cols-4 gap-3 mb-4">
        <NCard size="small"><NStatistic label="1天活跃访客" :value="insights.segments.active1d" /></NCard>
        <NCard size="small"><NStatistic label="7天活跃访客" :value="insights.segments.active7d" /></NCard>
        <NCard size="small"><NStatistic label="30天活跃访客" :value="insights.segments.active30d" /></NCard>
        <NCard size="small"><NStatistic label="高活跃访客" :value="insights.segments.highlyEngaged" /></NCard>
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
        :scroll-x="1400"
      />
    </NCard>

    <VisitorDetailDrawer
      :visible="detailVisible"
      :loading="detailLoading"
      :profile="currentProfile"
      :recent-comments="recentComments"
      @update:visible="detailVisible = $event"
    />
  </ScrollContainer>
</template>
