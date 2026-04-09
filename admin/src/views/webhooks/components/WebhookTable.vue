<script setup lang="ts">
import {
  NButton,
  NButtonGroup,
  NCard,
  NDataTable,
  NDivider,
  NEmpty,
  NForm,
  NFormItem,
  NGi,
  NGrid,
  NInput,
  NPopconfirm,
  NSelect,
  NSpace,
  NTag,
} from 'naive-ui'
import { computed, h } from 'vue'

import { formatDate } from '@/utils/format'

import type { WebhookItem } from '@/services/webhooks'
import type { DataTableColumns, SelectOption } from 'naive-ui'

const listFilters = defineModel<{
  keyword: string
  status: 'all' | 'enabled' | 'disabled'
  event: string | null
}>('listFilters', { required: true })

const props = defineProps<{
  webhooks: WebhookItem[]
  loading: boolean
  eventOptions: SelectOption[]
  statusOptions: SelectOption[]
}>()

const emit = defineEmits<{
  edit: [item: WebhookItem]
  test: [item: WebhookItem]
  delete: [item: WebhookItem]
  resetFilters: []
}>()

const filteredWebhooks = computed(() => {
  const keyword = listFilters.value.keyword.trim().toLowerCase()
  return props.webhooks.filter((item) => {
    if (listFilters.value.status === 'enabled' && !item.isEnabled) return false
    if (listFilters.value.status === 'disabled' && item.isEnabled) return false
    if (listFilters.value.event && !item.events?.includes(listFilters.value.event)) return false
    if (!keyword) return true
    return item.name.toLowerCase().includes(keyword) || item.url.toLowerCase().includes(keyword)
  })
})

const columns = computed<DataTableColumns<WebhookItem>>(() => [
  {
    title: '名称',
    key: 'name',
    width: 160,
    render: (row) => h('div', { class: 'font-medium' }, row.name),
  },
  {
    title: '地址',
    key: 'url',
    minWidth: 200,
    ellipsis: { tooltip: true },
    render: (row) => h('div', { class: 'text-xs text-[var(--text-color-3)]' }, row.url),
  },
  {
    title: '事件',
    key: 'events',
    minWidth: 200,
    render: (row) => {
      if (!row.events || row.events.length === 0) return '-'
      return h(
        'div',
        { class: 'flex flex-wrap gap-1' },
        row.events.map((item) =>
          h(NTag, { size: 'small', type: 'info', bordered: false }, { default: () => item }),
        ),
      )
    },
  },
  {
    title: '状态',
    key: 'isEnabled',
    width: 90,
    render: (row) =>
      h(
        NTag,
        { size: 'small', type: row.isEnabled ? 'success' : 'warning', bordered: false },
        { default: () => (row.isEnabled ? '启用' : '停用') },
      ),
  },
  {
    title: '更新时间',
    key: 'updatedAt',
    width: 180,
    render: (row) => formatDate(row.updatedAt),
  },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    render: (row) =>
      h(
        NButtonGroup,
        { size: 'small' },
        {
          default: () => [
            h(
              NButton,
              { type: 'primary', secondary: true, onClick: () => emit('edit', row) },
              { default: () => '编辑' },
            ),
            h(
              NButton,
              { tertiary: true, onClick: () => emit('test', row) },
              { default: () => '测试' },
            ),
            h(
              NPopconfirm,
              {
                positiveText: '删除',
                negativeText: '取消',
                onPositiveClick: () => emit('delete', row),
              },
              {
                trigger: () =>
                  h(NButton, { type: 'error', secondary: true }, { default: () => '删除' }),
                default: () => '确认删除该 Webhook？',
              },
            ),
          ],
        },
      ),
  },
])
</script>

<template>
  <NCard title="Webhook 列表">
    <template #header-extra>
      <NTag
        size="small"
        type="info"
        :bordered="false"
      >
        共 {{ filteredWebhooks.length }} 条
      </NTag>
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
          <NFormItem label="关键词">
            <NInput
              :value="listFilters.keyword"
              clearable
              placeholder="名称 / URL"
              @update:value="listFilters.keyword = $event"
            />
          </NFormItem>
        </NGi>
        <NGi>
          <NFormItem label="状态">
            <NSelect
              :value="listFilters.status"
              :options="statusOptions"
              @update:value="listFilters.status = $event"
            />
          </NFormItem>
        </NGi>
        <NGi>
          <NFormItem label="事件">
            <NSelect
              :value="listFilters.event"
              :options="eventOptions"
              clearable
              placeholder="全部"
              @update:value="listFilters.event = $event"
            />
          </NFormItem>
        </NGi>
        <NGi>
          <NFormItem label="操作">
            <NSpace>
              <NButton
                secondary
                @click="emit('resetFilters')"
                >重置</NButton
              >
            </NSpace>
          </NFormItem>
        </NGi>
      </NGrid>
    </NForm>
    <NDivider class="my-3" />
    <div
      v-if="filteredWebhooks.length === 0 && !loading"
      class="py-10"
    >
      <NEmpty description="暂无 Webhook" />
    </div>
    <NDataTable
      v-else
      :columns="columns"
      :data="filteredWebhooks"
      :loading="loading"
      :row-key="(row: WebhookItem) => row.id"
      striped
      :scroll-x="1000"
    />
  </NCard>
</template>
