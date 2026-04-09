<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NGrid,
  NGi,
  NInput,
  NPopconfirm,
  NSelect,
  NTag,
  useMessage,
} from 'naive-ui'
import { h, reactive, computed } from 'vue'

import { ScrollContainer } from '@/components'
import { batchUpdateEmailSubscriptionStatus, listEmailSubscriptions } from '@/services/email'

import type { EmailSubscription } from '@/services/email'
import type { DataTableColumns } from 'naive-ui'

const message = useMessage()
const queryClient = useQueryClient()

const params = reactive({
  page: 1,
  pageSize: 20,
  eventName: undefined as string | undefined,
  status: undefined as string | undefined,
  search: undefined as string | undefined,
})

const statusOptions = [
  { label: '全部状态', value: undefined },
  { label: '已订阅', value: 'active' },
  { label: '已退订', value: 'unsubscribed' },
  { label: '已拉黑', value: 'blocked' },
]

const { data, isLoading, refetch } = useQuery({
  queryKey: ['emailSubscriptions', params],
  queryFn: () => listEmailSubscriptions(params),
})

const updateStatusMutation = useMutation({
  mutationFn: ({ id, status }: { id: number; status: string }) =>
    batchUpdateEmailSubscriptionStatus({ ids: [id], status }),
  onSuccess: async (_, vars) => {
    message.success(vars.status === 'blocked' ? '已拉黑订阅用户' : '已解除拉黑')
    await queryClient.invalidateQueries({ queryKey: ['emailSubscriptions'] })
  },
  onError: (err: unknown) => {
    message.error(err instanceof Error ? err.message : '状态更新失败')
  },
})

function toStatusLabel(status: string) {
  switch (status) {
    case 'active':
      return '已订阅'
    case 'blocked':
      return '已拉黑'
    case 'unsubscribed':
      return '已退订'
    default:
      return status
  }
}

function toStatusTagType(status: string): 'default' | 'success' | 'warning' | 'error' {
  switch (status) {
    case 'active':
      return 'success'
    case 'blocked':
      return 'error'
    case 'unsubscribed':
      return 'warning'
    default:
      return 'default'
  }
}

function toggleBlocked(row: EmailSubscription) {
  const nextStatus = row.status === 'blocked' ? 'active' : 'blocked'
  updateStatusMutation.mutate({ id: row.id, status: nextStatus })
}

const columns: DataTableColumns<EmailSubscription> = [
  {
    title: 'ID',
    key: 'id',
    width: 80,
  },
  {
    title: '邮箱',
    key: 'email',
    width: 250,
  },
  {
    title: '订阅事件',
    key: 'eventName',
    width: 200,
    render: (row) =>
      h(NTag, { type: 'info', size: 'small', bordered: false }, { default: () => row.eventName }),
  },
  {
    title: '状态',
    key: 'status',
    width: 120,
    render: (row) => {
      return h(
        NTag,
        { type: toStatusTagType(row.status), size: 'small', bordered: false },
        { default: () => toStatusLabel(row.status) },
      )
    },
  },
  {
    title: '来源 IP',
    key: 'sourceIp',
    width: 150,
    render: (row) => h('div', { class: 'text-xs text-[var(--text-color-3)]' }, row.sourceIp),
  },
  {
    title: '订阅时间',
    key: 'createdAt',
    width: 180,
    render: (row) => new Date(row.createdAt).toLocaleString(),
  },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    render: (row) =>
      h(
        NPopconfirm,
        {
          onPositiveClick: () => toggleBlocked(row),
        },
        {
          trigger: () =>
            h(
              NButton,
              {
                size: 'small',
                type: row.status === 'blocked' ? 'success' : 'error',
                ghost: true,
                loading: updateStatusMutation.isPending.value,
              },
              { default: () => (row.status === 'blocked' ? '解除拉黑' : '拉黑') },
            ),
          default: () =>
            row.status === 'blocked' ? '确认解除拉黑该订阅用户？' : '确认拉黑该订阅用户？',
        },
      ),
  },
]

const pagination = computed(() => ({
  page: params.page,
  pageSize: params.pageSize,
  itemCount: data.value?.total || 0,
  onChange: (page: number) => {
    params.page = page
  },
  onUpdatePageSize: (pageSize: number) => {
    params.pageSize = pageSize
    params.page = 1
  },
}))

function handleRefresh() {
  refetch()
}

function handleReset() {
  params.eventName = undefined
  params.status = undefined
  params.search = undefined
  params.page = 1
}
</script>

<template>
  <ScrollContainer>
    <NCard title="订阅管理">
      <template #header-extra>
        <NButton
          secondary
          @click="handleRefresh"
          >刷新</NButton
        >
      </template>

      <NForm
        label-placement="left"
        label-width="auto"
        class="mb-4"
        :show-feedback="false"
      >
        <NGrid
          cols="1 640:2 900:4"
          :x-gap="16"
          :y-gap="8"
        >
          <NGi>
            <NFormItem label="搜索">
              <NInput
                v-model:value="params.search"
                placeholder="邮箱地址"
                clearable
              />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="状态">
              <NSelect
                v-model:value="params.status"
                :options="statusOptions"
                placeholder="全部"
                clearable
              />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="事件">
              <NInput
                v-model:value="params.eventName"
                placeholder="事件名称"
                clearable
              />
            </NFormItem>
          </NGi>
          <NGi>
            <div class="flex justify-end">
              <NButton @click="handleReset">重置</NButton>
            </div>
          </NGi>
        </NGrid>
      </NForm>

      <NDataTable
        remote
        :columns="columns"
        :data="data?.items || []"
        :loading="isLoading"
        :pagination="pagination"
        :row-key="(row: EmailSubscription) => row.id"
        :scroll-x="1000"
      />
    </NCard>
  </ScrollContainer>
</template>
