<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { NButton, NCard, NDataTable, NForm, NFormItem, NGrid, NGi, NInput, NSelect, NSpace, NTag } from 'naive-ui'
import { h, reactive, ref, computed } from 'vue'

import { ScrollContainer } from '@/components'
import { listEmailSubscriptions } from '@/services/email'
import { useThemeVars } from 'naive-ui'

import type { EmailSubscription } from '@/services/email'
import type { DataTableColumns } from 'naive-ui'

const themeVars = useThemeVars()

const params = reactive({
  page: 1,
  pageSize: 20,
  eventName: undefined as string | undefined,
  status: undefined as string | undefined,
  search: undefined as string | undefined,
})

const statusOptions = [
  { label: '全部状态', value: undefined },
  { label: '待验证', value: 'pending' },
  { label: '已验证', value: 'verified' },
  { label: '已退订', value: 'unsubscribed' },
]

const { data, isLoading, refetch } = useQuery({
  queryKey: ['emailSubscriptions', params],
  queryFn: () => listEmailSubscriptions(params),
})

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
    render: (row) => h(NTag, { type: 'info', size: 'small', bordered: false }, { default: () => row.eventName }),
  },
  {
    title: '状态',
    key: 'status',
    width: 120,
    render: (row) => {
      let type: 'default' | 'success' | 'warning' | 'error' = 'default'
      let label: string = row.status
      if (row.status === 'verified') {
        type = 'success'
        label = '已验证'
      } else if (row.status === 'pending') {
        type = 'warning'
        label = '待验证'
      } else if (row.status === 'unsubscribed') {
        type = 'error'
        label = '已退订'
      }
      return h(NTag, { type, size: 'small', bordered: false }, { default: () => label })
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
        <NButton secondary @click="handleRefresh">刷新</NButton>
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
