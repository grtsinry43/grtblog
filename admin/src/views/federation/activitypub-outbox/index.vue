<script setup lang="ts">
import { h, ref } from 'vue'
import { NDataTable, NButton, NTag, NInput, NCard, NSelect, NPagination } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useQuery } from '@tanstack/vue-query'
import { ScrollContainer } from '@/components'
import { listActivityPubOutbox } from '@/services/federation-admin'
import type { ActivityPubOutboxItemResp } from '@/types/federation'
import DetailDrawer from './DetailDrawer.vue'

const page = ref(1)
const pageSize = ref(20)
const filterStatus = ref<string | null>(null)
const filterSourceType = ref<string | null>(null)
const searchKeyword = ref('')

const queryParams = () => ({
  page: page.value,
  pageSize: pageSize.value,
  status: filterStatus.value || undefined,
  sourceType: filterSourceType.value || undefined,
  search: searchKeyword.value || undefined,
})

const { data, isPending, refetch } = useQuery({
  queryKey: ['activitypub-outbox', page, pageSize, filterStatus, filterSourceType, searchKeyword],
  queryFn: () => listActivityPubOutbox(queryParams()),
})

const showDrawer = ref(false)
const currentItem = ref<ActivityPubOutboxItemResp | undefined>(undefined)

function openDetail(row: ActivityPubOutboxItemResp) {
  currentItem.value = row
  showDrawer.value = true
}

function statusTagType(status: string) {
  switch (status) {
    case 'completed': return 'success'
    case 'partial': return 'warning'
    case 'failed': return 'error'
    case 'sending': return 'info'
    default: return 'default'
  }
}

function sourceTypeLabel(type: string) {
  switch (type) {
    case 'article': return '文章'
    case 'moment': return '手记'
    case 'thinking': return '思考'
    default: return type
  }
}

const columns: DataTableColumns<ActivityPubOutboxItemResp> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: '摘要',
    key: 'summary',
    minWidth: 220,
    ellipsis: { tooltip: true },
  },
  {
    title: '类型',
    key: 'source_type',
    width: 100,
    render(row) {
      return h(NTag, { size: 'small', bordered: false }, { default: () => sourceTypeLabel(row.source_type) })
    },
  },
  {
    title: '状态',
    key: 'status',
    width: 110,
    render(row) {
      return h(NTag, { type: statusTagType(row.status), size: 'small', bordered: false }, { default: () => row.status })
    },
  },
  {
    title: '投递统计',
    key: 'stats',
    width: 110,
    render(row) {
      return `${row.success_count}/${row.total_targets}`
    },
  },
  {
    title: '触发来源',
    key: 'trigger_source',
    width: 110,
  },
  {
    title: '发布时间',
    key: 'published_at',
    width: 180,
    render(row) {
      return new Date(row.published_at).toLocaleString()
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 90,
    render(row) {
      return h(NButton, { size: 'small', onClick: () => openDetail(row) }, { default: () => '详情' })
    },
  },
]

const statusOptions = [
  { label: '全部状态', value: null },
  { label: 'queued', value: 'queued' },
  { label: 'sending', value: 'sending' },
  { label: 'completed', value: 'completed' },
  { label: 'partial', value: 'partial' },
  { label: 'failed', value: 'failed' },
] as any

const sourceTypeOptions = [
  { label: '全部类型', value: null },
  { label: '文章', value: 'article' },
  { label: '手记', value: 'moment' },
  { label: '思考', value: 'thinking' },
] as any
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-y-4">
    <NCard :bordered="false">
      <div class="flex items-center justify-between">
        <div class="text-lg font-medium">ActivityPub 出站</div>
        <div class="flex items-center gap-2">
          <NInput v-model:value="searchKeyword" placeholder="搜索摘要 / activity_id / object_id" clearable class="w-72" />
          <NSelect v-model:value="filterSourceType" :options="sourceTypeOptions" class="w-32" placeholder="内容类型" clearable />
          <NSelect v-model:value="filterStatus" :options="statusOptions" class="w-32" placeholder="状态" clearable />
        </div>
      </div>
    </NCard>

    <NCard :bordered="false" content-style="padding: 0;">
      <NDataTable
        remote
        :columns="columns"
        :data="data?.items || []"
        :loading="isPending"
        :bordered="false"
        :row-key="(row: ActivityPubOutboxItemResp) => row.id"
        :scroll-x="1200"
      />
      <div class="flex justify-end p-4">
        <NPagination
          v-model:page="page"
          v-model:page-size="pageSize"
          :item-count="data?.total || 0"
          show-size-picker
          :page-sizes="[10, 20, 50]"
          @update:page="(p: number) => (page = p)"
          @update:page-size="(s: number) => { pageSize = s; page = 1 }"
        />
      </div>
    </NCard>

    <DetailDrawer v-model:show="showDrawer" :item="currentItem" @refresh="refetch()" />
  </ScrollContainer>
</template>
