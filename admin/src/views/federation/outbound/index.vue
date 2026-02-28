<script setup lang="ts">
import { h, ref } from 'vue'
import { NDataTable, NButton, NTag, NInput, NSpace, NCard, NSelect } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { getFederationOutboundLog } from '@/services/federation-admin'
import type { FederationOutboundDeliveryResp, FederationOutboundDeliveryListResp } from '@/types/federation'
import { useQuery } from '@tanstack/vue-query'
import DetailDrawer from './DetailDrawer.vue'
import { ScrollContainer } from '@/components'

const page = ref(1)
const pageSize = ref(10)
const filterType = ref<string | null>(null)
const filterStatus = ref<string | null>(null)
const searchRequestId = ref('')

const queryParams = () => ({
  page: page.value,
  pageSize: pageSize.value,
  type: filterType.value || undefined,
  status: filterStatus.value || undefined,
  request_id: searchRequestId.value || undefined,
})

const { data, isPending } = useQuery({
  queryKey: ['federation-outbound-logs', page, pageSize, filterType, filterStatus, searchRequestId],
  queryFn: () => getFederationOutboundLog(queryParams()),
})

const columns: DataTableColumns<FederationOutboundDeliveryResp> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '类型', key: 'type', width: 120, render(row) {
    return h(NTag, { type: 'info', bordered: false }, { default: () => row.type })
  }},
  { title: '目标实例', key: 'target_instance_url', minWidth: 200, ellipsis: { tooltip: true } },
  { title: '状态', key: 'status', width: 100, render(row) {
    return h(NTag, { type: row.status === 'success' ? 'success' : row.status === 'failed' ? 'error' : 'warning' }, { default: () => row.status })
  }},
  { title: '尝试次数', key: 'attempt_count', width: 100, render: (row) => `${row.attempt_count}/${row.max_attempts}` },
  { title: '创建时间', key: 'created_at', width: 180, render: (row) => new Date(row.created_at).toLocaleString() },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render(row) {
      return h(NButton, {
        size: 'small',
        onClick: () => openDetail(row)
      }, { default: () => '详情' })
    }
  }
]

// Detail Drawer Logic
const showDrawer = ref(false)
const currentDelivery = ref<FederationOutboundDeliveryResp | undefined>(undefined)

function openDetail(row: FederationOutboundDeliveryResp) {
  currentDelivery.value = row
  showDrawer.value = true
}

const typeOptions = [
  { label: '全部类型', value: null },
  { label: '友链申请', value: 'friend_link' },
  { label: '引用通知', value: 'citation' },
  { label: '提及通知', value: 'mention' },
] as any

const statusOptions = [
  { label: '全部状态', value: null },
  { label: '等待中', value: 'pending' },
  { label: '成功', value: 'success' },
  { label: '失败', value: 'failed' },
  { label: '重试中', value: 'retrying' },
] as any
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-y-4">
    <NCard :bordered="false">
      <div class="flex items-center justify-between">
        <div class="text-lg font-medium">出站记录</div>
        <div class="flex items-center gap-2">
          <NInput 
            v-model:value="searchRequestId" 
            placeholder="搜索 Request ID" 
            clearable 
            class="w-60"
          />
          <NSelect 
            v-model:value="filterType" 
            :options="typeOptions" 
            class="w-40" 
            placeholder="类型" 
            clearable
          />
          <NSelect 
            v-model:value="filterStatus" 
            :options="statusOptions" 
            class="w-40" 
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
        :row-key="(row: FederationOutboundDeliveryResp) => row.id"
        :scroll-x="900"
      />
      <div class="p-4 flex justify-end">
        <NPagination
          v-model:page="page"
          v-model:page-size="pageSize"
          :item-count="data?.total || 0"
          show-size-picker
          :page-sizes="[10, 20, 50]"
          @update:page="(p: number) => page = p"
          @update:page-size="(s: number) => { pageSize = s; page = 1 }"
        />
      </div>
    </NCard>

    <DetailDrawer 
      v-model:show="showDrawer"
      :delivery="currentDelivery"
    />
  </ScrollContainer>
</template>
