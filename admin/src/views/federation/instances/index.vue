<script setup lang="ts">
import { h, ref } from 'vue'
import { NCard, NDataTable, NButton, NTag, NSpace, NInput, NPagination, useMessage, NPopconfirm, NModal, NForm, NFormItem } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { getFederationInstances, updateFederationInstanceStatus, requestFederationFriendlink } from '@/services/federation-admin'
import type { FederationInstanceResp, FederationAdminFriendLinkRequestReq } from '@/types/federation'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { ScrollContainer } from '@/components'
import DetailDrawer from './DetailDrawer.vue'
import { useRouter } from 'vue-router'

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
  { title: '域名', key: 'base_url', render: (row) => h('a', { href: row.base_url, target: '_blank', class: 'text-primary hover:underline' }, row.base_url) },
  { title: '名称', key: 'name', render: (row) => row.name || '-' },
  { title: '软件版本', key: 'protocol_version', render: (row) => row.protocol_version || '-' },
  { title: '状态', key: 'status', width: 100, render(row) {
    const typeMap: Record<string, 'default' | 'success' | 'warning' | 'error'> = {
        active: 'success',
        blocked: 'error',
        unknown: 'warning'
    }
    return h(NTag, { type: typeMap[row.status] || 'default', size: 'small' }, { default: () => row.status })
  }},
  { title: '最后可见', key: 'last_seen_at', width: 180, render: (row) => row.last_seen_at ? new Date(row.last_seen_at).toLocaleString() : '-' },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render(row) {
      return h(NSpace, {}, {
        default: () => [
          h(NButton, {
            size: 'small',
            onClick: () => handleOpenDetail(row)
          }, { default: () => '详情' }),
           h(NPopconfirm, {
            onPositiveClick: () => handleToggleStatus(row)
          }, {
            trigger: () => h(NButton, {
              size: 'small',
              type: row.status === 'blocked' ? 'success' : 'error',
              secondary: true
            }, { default: () => row.status === 'blocked' ? '解封' : '封禁' }),
            default: () => `确认${row.status === 'blocked' ? '解封' : '封禁'}该实例吗？`
          })
        ]
      })
    }
  }
]

function handleOpenDetail(row: FederationInstanceResp) {
    currentInstanceId.value = row.id
    showDrawer.value = true
}

// Status Toggle Logic
const { mutate: updateStatus } = useMutation({
  mutationFn: ({ id, status }: { id: number, status: string }) => updateFederationInstanceStatus(id, status),
  onSuccess: () => {
    message.success('状态更新成功')
    queryClient.invalidateQueries({ queryKey: ['federation-instances'] })
  },
  onError: (err: any) => {
    message.error('更新失败: ' + (err.message || 'Unknown error'))
  }
})

function handleToggleStatus(row: FederationInstanceResp) {
  const newStatus = row.status === 'blocked' ? 'active' : 'blocked'
  updateStatus({ id: row.id, status: newStatus })
}

// Initiate United Modal (Friend Link Request)
const showCreateModal = ref(false)
const createForm = ref<FederationAdminFriendLinkRequestReq>({
  target_url: '',
  message: '',
  rss_url: ''
})

const { mutate: requestFriendLink, isPending: isRequesting } = useMutation({
  mutationFn: requestFederationFriendlink,
  onSuccess: () => {
    message.success('联合申请已发送')
    showCreateModal.value = false
    createForm.value = { target_url: '', message: '', rss_url: '' }
    // Optionally refresh list if the new instance appears immediately
    queryClient.invalidateQueries({ queryKey: ['federation-instances'] })
  },
  onError: (err: any) => {
    message.error('申请发送失败: ' + (err.message || 'Unknown error'))
  }
})

function handleInitiateUnited() {
  showCreateModal.value = true
}

function submitCreate() {
  if (!createForm.value.target_url) {
    message.warning('请输入目标地址')
    return
  }
  requestFriendLink(createForm.value)
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
           <NButton secondary @click="queryClient.invalidateQueries({ queryKey: ['federation-instances'] })">
            搜索
          </NButton>
          <NButton secondary type="warning" @click="router.push({ name: 'federationDebug' })">
            联合调试
          </NButton>
          <NButton type="primary" @click="handleInitiateUnited">
            发起联合
          </NButton>
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
        :row-key="(row: FederationInstanceResp) => row.id"
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

    <NModal
      v-model:show="showCreateModal"
      preset="dialog"
      title="发起联合 (申请友链)"
      positive-text="发送申请"
      negative-text="取消"
      @positive-click="submitCreate"
      @negative-click="showCreateModal = false"
      :loading="isRequesting"
    >
      <div class="py-4">
        <NForm label-placement="left" label-width="100">
          <NFormItem label="目标地址" required>
            <NInput v-model:value="createForm.target_url" placeholder="https://target.com/friend" />
          </NFormItem>
          <NFormItem label="你的 RSS 地址">
            <NInput v-model:value="createForm.rss_url" placeholder="https://mysite.com/feed (可选)" />
          </NFormItem>
          <NFormItem label="留言信息">
            <NInput v-model:value="createForm.message" type="textarea" placeholder="你好，我想与贵站建立联合..." />
          </NFormItem>
        </NForm>
      </div>
    </NModal>

    <DetailDrawer 
      v-model:show="showDrawer" 
      :instance-id="currentInstanceId" 
    />
  </ScrollContainer>
</template>
