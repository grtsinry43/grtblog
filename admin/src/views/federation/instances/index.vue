<script setup lang="ts">
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
  NCard,
  NDataTable,
  NButton,
  NTag,
  NSpace,
  NInput,
  NPagination,
  useMessage,
  NPopconfirm,
} from 'naive-ui'
import { h, ref } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { getFederationInstances, updateFederationInstanceStatus } from '@/services/federation-admin'

import DetailDrawer from './DetailDrawer.vue'

import type { FederationInstanceResp } from '@/types/federation'
import type { DataTableColumns } from 'naive-ui'

const message = useMessage()
const queryClient = useQueryClient()
const router = useRouter()

const page = ref(1)
const pageSize = ref(10)
const searchKeyword = ref('')

// Detail Drawer State
const showDrawer = ref(false)
const currentInstanceId = ref<number | undefined>(undefined)

const queryParams = () => ({
  page: page.value,
  pageSize: pageSize.value,
  keyword: searchKeyword.value || undefined,
})

const { data, isPending } = useQuery({
  queryKey: ['federation-instances', page, pageSize, searchKeyword],
  queryFn: () => getFederationInstances(queryParams()),
})

const columns: DataTableColumns<FederationInstanceResp> = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: '域名',
    key: 'base_url',
    minWidth: 200,
    ellipsis: { tooltip: true },
    render: (row) =>
      h(
        'a',
        { href: row.base_url, target: '_blank', class: 'text-primary hover:underline' },
        row.base_url,
      ),
  },
  { title: '名称', key: 'name', minWidth: 120, render: (row) => row.name || '-' },
  {
    title: '软件版本',
    key: 'protocol_version',
    width: 120,
    render: (row) => row.protocol_version || '-',
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render(row) {
      const typeMap: Record<string, 'default' | 'success' | 'warning' | 'error'> = {
        active: 'success',
        blocked: 'error',
        unknown: 'warning',
      }
      return h(
        NTag,
        { type: typeMap[row.status] || 'default', size: 'small' },
        { default: () => row.status },
      )
    },
  },
  {
    title: '最后可见',
    key: 'last_seen_at',
    width: 180,
    render: (row) => (row.last_seen_at ? new Date(row.last_seen_at).toLocaleString() : '-'),
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render(row) {
      return h(
        NSpace,
        {},
        {
          default: () => [
            h(
              NButton,
              {
                size: 'small',
                onClick: () => handleOpenDetail(row),
              },
              { default: () => '详情' },
            ),
            h(
              NPopconfirm,
              {
                onPositiveClick: () => handleToggleStatus(row),
              },
              {
                trigger: () =>
                  h(
                    NButton,
                    {
                      size: 'small',
                      type: row.status === 'blocked' ? 'success' : 'error',
                      secondary: true,
                    },
                    { default: () => (row.status === 'blocked' ? '解封' : '封禁') },
                  ),
                default: () => `确认${row.status === 'blocked' ? '解封' : '封禁'}该实例吗？`,
              },
            ),
          ],
        },
      )
    },
  },
]

function handleOpenDetail(row: FederationInstanceResp) {
  currentInstanceId.value = row.id
  showDrawer.value = true
}

// Status Toggle Logic
const { mutate: updateStatus } = useMutation({
  mutationFn: ({ id, status }: { id: number; status: string }) =>
    updateFederationInstanceStatus(id, status),
  onSuccess: () => {
    message.success('状态更新成功')
    queryClient.invalidateQueries({ queryKey: ['federation-instances'] })
  },
  onError: (err: any) => {
    message.error('更新失败: ' + (err.message || 'Unknown error'))
  },
})

function handleToggleStatus(row: FederationInstanceResp) {
  const newStatus = row.status === 'blocked' ? 'active' : 'blocked'
  updateStatus({ id: row.id, status: newStatus })
}
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-y-4">
    <NCard :bordered="false">
      <div class="flex items-center justify-between">
        <div class="text-lg font-medium">联合实例</div>
        <div class="flex items-center gap-2">
          <NInput
            v-model:value="searchKeyword"
            placeholder="搜索域名或名称"
            clearable
            class="w-60"
            @keydown.enter="queryClient.invalidateQueries({ queryKey: ['federation-instances'] })"
          />
          <NButton
            secondary
            @click="queryClient.invalidateQueries({ queryKey: ['federation-instances'] })"
          >
            搜索
          </NButton>
          <NButton
            secondary
            type="warning"
            @click="router.push({ name: 'federationDebug' })"
          >
            联合调试
          </NButton>
          <NButton
            type="primary"
            @click="router.push({ name: 'friendLinkList' })"
          >
            去友链管理
          </NButton>
        </div>
      </div>
    </NCard>

    <NCard
      :bordered="false"
      content-style="padding: 0;"
    >
      <NDataTable
        remote
        :columns="columns"
        :data="data?.items || []"
        :loading="isPending"
        :bordered="false"
        :row-key="(row: FederationInstanceResp) => row.id"
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

    <DetailDrawer
      v-model:show="showDrawer"
      :instance-id="currentInstanceId"
    />
  </ScrollContainer>
</template>
