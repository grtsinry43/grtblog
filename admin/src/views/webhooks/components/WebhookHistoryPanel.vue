<script setup lang="ts">
import {
  NButton,
  NCard,
  NCheckbox,
  NDataTable,
  NDivider,
  NEmpty,
  NForm,
  NFormItem,
  NGi,
  NGrid,
  NPagination,
  NSelect,
  NSpace,
  NTag,
} from 'naive-ui'
import { computed, h } from 'vue'

import { formatDate } from '@/utils/format'

import type { WebhookHistoryItem } from '@/services/webhooks'
import type { DataTableColumns, SelectOption } from 'naive-ui'

const historyFilters = defineModel<{
  webhookId: number | null
  eventName: string | null
  isTest: boolean | null
}>('historyFilters', { required: true })

const isTestOnly = defineModel<boolean>('isTestOnly', { required: true })

const props = defineProps<{
  history: WebhookHistoryItem[]
  historyLoading: boolean
  historyPage: number
  historyPageSize: number
  historyTotal: number
  historyFailureCount: number
  webhookMap: Map<number, string>
  webhookOptions: SelectOption[]
  eventOptions: SelectOption[]
}>()

const emit = defineEmits<{
  'update:historyPage': [value: number]
  'update:historyPageSize': [value: number]
  applyFilters: []
  resetFilters: []
  refresh: []
  viewDetail: [item: WebhookHistoryItem]
  replay: [item: WebhookHistoryItem]
}>()

function renderHistoryStatus(row: WebhookHistoryItem) {
  const success = row.responseStatus >= 200 && row.responseStatus < 300
  const label = success ? '成功' : '失败'
  return h(
    NTag,
    { size: 'small', type: success ? 'success' : 'error', bordered: false },
    { default: () => (row.responseStatus ? `${label} ${row.responseStatus}` : label) },
  )
}

const historyColumns = computed<DataTableColumns<WebhookHistoryItem>>(() => [
  { title: '事件', key: 'eventName', width: 220 },
  {
    title: 'Webhook',
    key: 'webhookId',
    width: 180,
    render: (row) => props.webhookMap.get(row.webhookId) || `#${row.webhookId}`,
  },
  {
    title: '状态',
    key: 'responseStatus',
    width: 120,
    render: (row) => renderHistoryStatus(row),
  },
  {
    title: '测试',
    key: 'isTest',
    width: 90,
    render: (row) =>
      h(
        NTag,
        { size: 'small', type: row.isTest ? 'success' : 'default', bordered: false },
        { default: () => (row.isTest ? '是' : '否') },
      ),
  },
  {
    title: '时间',
    key: 'createdAt',
    width: 180,
    render: (row) => formatDate(row.createdAt),
  },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    render: (row) =>
      h('div', { class: 'flex gap-2' }, [
        h(
          NButton,
          { size: 'small', secondary: true, onClick: () => emit('viewDetail', row) },
          { default: () => '详情' },
        ),
        h(
          NButton,
          { size: 'small', type: 'primary', secondary: true, onClick: () => emit('replay', row) },
          { default: () => '重放' },
        ),
      ]),
  },
])
</script>

<template>
  <NCard title="投递历史">
    <template #header-extra>
      <NSpace size="small">
        <NTag
          size="small"
          type="warning"
          :bordered="false"
        >
          本页失败 {{ historyFailureCount }}
        </NTag>
        <NTag
          size="small"
          type="info"
          :bordered="false"
        >
          共 {{ historyTotal }} 条
        </NTag>
      </NSpace>
    </template>
    <NForm
      label-placement="left"
      label-width="60"
      :show-feedback="false"
    >
      <NGrid
        cols="1 640:2 900:4"
        x-gap="16"
        y-gap="8"
      >
        <NGi>
          <NFormItem label="Webhook">
            <NSelect
              :value="historyFilters.webhookId"
              :options="webhookOptions"
              clearable
              placeholder="全部"
              @update:value="historyFilters.webhookId = $event"
            />
          </NFormItem>
        </NGi>
        <NGi>
          <NFormItem label="事件">
            <NSelect
              :value="historyFilters.eventName"
              :options="eventOptions"
              clearable
              placeholder="全部"
              @update:value="historyFilters.eventName = $event"
            />
          </NFormItem>
        </NGi>
        <NGi>
          <NFormItem label="测试">
            <NCheckbox
              :checked="isTestOnly"
              @update:checked="isTestOnly = $event"
            >
              仅测试
            </NCheckbox>
          </NFormItem>
        </NGi>
        <NGi>
          <NFormItem label="操作">
            <NSpace>
              <NButton
                type="primary"
                secondary
                @click="emit('applyFilters')"
                >筛选</NButton
              >
              <NButton @click="emit('resetFilters')">重置</NButton>
              <NButton
                tertiary
                @click="emit('refresh')"
                >刷新</NButton
              >
            </NSpace>
          </NFormItem>
        </NGi>
      </NGrid>
    </NForm>
    <NDivider class="my-3" />
    <div
      v-if="history.length === 0 && !historyLoading"
      class="py-10"
    >
      <NEmpty description="暂无投递记录" />
    </div>
    <NDataTable
      v-else
      :columns="historyColumns"
      :data="history"
      :loading="historyLoading"
      :row-key="(row: WebhookHistoryItem) => row.id"
      striped
      :scroll-x="960"
    />
    <div class="flex justify-end pt-4">
      <NPagination
        :page="historyPage"
        :page-size="historyPageSize"
        :item-count="historyTotal"
        show-size-picker
        :page-sizes="[10, 20, 50]"
        @update:page="emit('update:historyPage', $event)"
        @update:page-size="emit('update:historyPageSize', $event)"
      />
    </div>
  </NCard>
</template>
