<script setup lang="ts">
import { h, ref } from 'vue'
import { NDataTable, NButton, NTag, NInput, NCard, NSelect, NPagination } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { listEmailOutbox } from '@/services/email'
import type { EmailOutbox } from '@/services/email'
import { useQuery } from '@tanstack/vue-query'
import DetailDrawer from './DetailDrawer.vue'
import { ScrollContainer } from '@/components'

const page = ref(1)
const pageSize = ref(20)
const filterStatus = ref<string | null>(null)
const filterEventName = ref<string | null>(null)
const searchKeyword = ref('')

const queryParams = () => ({
  page: page.value,
  pageSize: pageSize.value,
  status: filterStatus.value || undefined,
  eventName: filterEventName.value || undefined,
  search: searchKeyword.value || undefined,
})

const { data, isPending } = useQuery({
  queryKey: ['email-outbox', page, pageSize, filterStatus, filterEventName, searchKeyword],
  queryFn: () => listEmailOutbox(queryParams()),
})

function statusTagType(status: string) {
  switch (status) {
    case 'sent': return 'success'
    case 'failed': return 'error'
    case 'sending': return 'warning'
    case 'pending': return 'info'
    default: return 'default'
  }
}

function statusLabel(status: string) {
  switch (status) {
    case 'pending': return '等待中'
    case 'sending': return '发送中'
    case 'sent': return '已发送'
    case 'failed': return '失败'
    default: return status
  }
}

const columns: DataTableColumns<EmailOutbox> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: '主题',
    key: 'subject',
    minWidth: 180,
    ellipsis: { tooltip: true },
  },
  {
    title: '模板',
    key: 'templateCode',
    width: 160,
    ellipsis: { tooltip: true },
  },
  {
    title: '事件',
    key: 'eventName',
    width: 160,
    render(row) {
      return h(NTag, { type: 'info', bordered: false, size: 'small' }, { default: () => row.eventName })
    },
  },
  {
    title: '收件人',
    key: 'toEmails',
    width: 180,
    ellipsis: { tooltip: true },
    render(row) {
      return row.toEmails?.join(', ') || '-'
    },
  },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render(row) {
      return h(NTag, { type: statusTagType(row.status), bordered: false, size: 'small' }, { default: () => statusLabel(row.status) })
    },
  },
  {
    title: '重试',
    key: 'retryCount',
    width: 60,
    render: (row) => `${row.retryCount}`,
  },
  {
    title: '创建时间',
    key: 'createdAt',
    width: 170,
    render: (row) => new Date(row.createdAt).toLocaleString(),
  },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render(row) {
      return h(
        NButton,
        { size: 'small', onClick: () => openDetail(row) },
        { default: () => '详情' },
      )
    },
  },
]

const showDrawer = ref(false)
const currentOutbox = ref<EmailOutbox | undefined>(undefined)

function openDetail(row: EmailOutbox) {
  currentOutbox.value = row
  showDrawer.value = true
}

const statusOptions = [
  { label: '全部状态', value: null },
  { label: '等待中', value: 'pending' },
  { label: '发送中', value: 'sending' },
  { label: '已发送', value: 'sent' },
  { label: '失败', value: 'failed' },
] as any
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-y-4">
    <NCard :bordered="false">
      <div class="flex items-center justify-between">
        <div class="text-lg font-medium">出站队列</div>
        <div class="flex items-center gap-2">
          <NInput
            v-model:value="searchKeyword"
            placeholder="搜索主题 / 模板编码"
            clearable
            class="w-60"
          />
          <NInput
            v-model:value="filterEventName"
            placeholder="事件名"
            clearable
            class="w-40"
          />
          <NSelect
            v-model:value="filterStatus"
            :options="statusOptions"
            class="w-32"
            placeholder="状态"
            clearable
          />
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
        :row-key="(row: EmailOutbox) => row.id"
        :scroll-x="1100"
      />
      <div class="flex justify-end p-4">
        <NPagination
          v-model:page="page"
          v-model:page-size="pageSize"
          :item-count="data?.total || 0"
          show-size-picker
          :page-sizes="[10, 20, 50]"
          @update:page="(p: number) => (page = p)"
          @update:page-size="
            (s: number) => {
              pageSize = s
              page = 1
            }
          "
        />
      </div>
    </NCard>

    <DetailDrawer v-model:show="showDrawer" :outbox="currentOutbox" />
  </ScrollContainer>
</template>
