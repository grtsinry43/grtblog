<script setup lang="ts">
import { h, ref } from 'vue'
import { NDataTable, NButton, NTag, NInput, NCard, NSelect, NPagination } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { listAITaskLogs } from '@/services/ai'
import type { AITaskLog } from '@/services/ai'
import { useQuery } from '@tanstack/vue-query'
import DetailDrawer from './DetailDrawer.vue'
import { ScrollContainer } from '@/components'

const page = ref(1)
const pageSize = ref(20)
const filterTaskType = ref<string | null>(null)
const filterStatus = ref<string | null>(null)
const searchKeyword = ref('')

const queryParams = () => ({
  page: page.value,
  pageSize: pageSize.value,
  taskType: filterTaskType.value || undefined,
  status: filterStatus.value || undefined,
  search: searchKeyword.value || undefined,
})

const { data, isPending } = useQuery({
  queryKey: ['ai-task-logs', page, pageSize, filterTaskType, filterStatus, searchKeyword],
  queryFn: () => listAITaskLogs(queryParams()),
})

function statusTagType(status: string) {
  switch (status) {
    case 'completed': return 'success'
    case 'failed': return 'error'
    case 'interrupted': return 'warning'
    case 'running': return 'warning'
    case 'pending': return 'info'
    default: return 'default'
  }
}

function statusLabel(status: string) {
  switch (status) {
    case 'pending': return '等待中'
    case 'running': return '运行中'
    case 'completed': return '已完成'
    case 'failed': return '失败'
    case 'interrupted': return '已中断'
    default: return status
  }
}

function taskTypeLabel(type_: string) {
  switch (type_) {
    case 'comment_moderation': return '评论审核'
    case 'title_generation': return '标题生成'
    case 'content_rewrite': return '内容改写'
    case 'summary_generation': return '摘要生成'
    default: return type_
  }
}

function triggerLabel(trigger: string) {
  switch (trigger) {
    case 'manual': return '手动'
    case 'auto': return '自动'
    default: return trigger
  }
}

function formatDuration(ms: number) {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(1)}s`
}

const columns: DataTableColumns<AITaskLog> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: '任务类型',
    key: 'taskType',
    width: 120,
    render(row) {
      return h(NTag, { type: 'info', bordered: false, size: 'small' }, { default: () => taskTypeLabel(row.taskType) })
    },
  },
  {
    title: '模型',
    key: 'modelName',
    width: 150,
    ellipsis: { tooltip: true },
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
    title: '耗时',
    key: 'durationMs',
    width: 90,
    render: (row) => formatDuration(row.durationMs),
  },
  {
    title: '触发来源',
    key: 'triggerSource',
    width: 90,
    render(row) {
      return h(NTag, {
        type: row.triggerSource === 'auto' ? 'warning' : 'default',
        bordered: false,
        size: 'small',
      }, { default: () => triggerLabel(row.triggerSource) })
    },
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
const currentLog = ref<AITaskLog | undefined>(undefined)

function openDetail(row: AITaskLog) {
  currentLog.value = row
  showDrawer.value = true
}

const taskTypeOptions = [
  { label: '全部类型', value: null },
  { label: '评论审核', value: 'comment_moderation' },
  { label: '标题生成', value: 'title_generation' },
  { label: '内容改写', value: 'content_rewrite' },
  { label: '摘要生成', value: 'summary_generation' },
] as any

const statusOptions = [
  { label: '全部状态', value: null },
  { label: '等待中', value: 'pending' },
  { label: '运行中', value: 'running' },
  { label: '已完成', value: 'completed' },
  { label: '失败', value: 'failed' },
  { label: '已中断', value: 'interrupted' },
] as any
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-y-4">
    <NCard :bordered="false">
      <div class="flex items-center justify-between">
        <div class="text-lg font-medium">AI 任务日志</div>
        <div class="flex items-center gap-2">
          <NInput
            v-model:value="searchKeyword"
            placeholder="搜索内容 / 模型名"
            clearable
            class="w-52"
          />
          <NSelect
            v-model:value="filterTaskType"
            :options="taskTypeOptions"
            class="w-32"
            placeholder="类型"
            clearable
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
        :row-key="(row: AITaskLog) => row.id"
        :scroll-x="1000"
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

    <DetailDrawer v-model:show="showDrawer" :task-log="currentLog" />
  </ScrollContainer>
</template>
