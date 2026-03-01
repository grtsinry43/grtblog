<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NFormItem,
  NInput,
  NSelect,
  NSpace,
  NSwitch,
  NTag,
  useMessage,
} from 'naive-ui'
import { computed, h, reactive, ref } from 'vue'

import { FormModal, ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { listSiteUsers, updateSiteUser } from '@/services/site-users'
import { toRefsUserStore } from '@/stores'
import { formatDate } from '@/utils/format'

import type { DataTableColumns } from 'naive-ui'
import type { SiteUser } from '@/types/site-users'

defineOptions({
  name: 'SiteUserManagement',
})

const message = useMessage()
const { user } = toRefsUserStore()

const keyword = ref('')
const adminFilter = ref<string>('all')
const activeFilter = ref<string>('all')
const queryState = ref({
  keyword: '',
  admin: undefined as boolean | undefined,
  active: undefined as boolean | undefined,
})

const {
  loading,
  data: tableData,
  pagination,
  refresh,
} = useTable<SiteUser>(listSiteUsers, queryState.value)

const editVisible = ref(false)
const saving = ref(false)
const editingUserId = ref<number>(0)
const formModel = reactive({
  username: '',
  nickname: '',
  email: '',
  isActive: true,
  isAdmin: false,
})

const isEditingSelf = computed(() => {
  if (!editingUserId.value || !user.value?.id) return false
  return editingUserId.value === user.value.id
})

const columns = computed<DataTableColumns<SiteUser>>(() => [
  {
    title: 'ID',
    key: 'id',
    width: 72,
  },
  {
    title: '用户名',
    key: 'username',
    width: 160,
  },
  {
    title: '昵称',
    key: 'nickname',
    width: 140,
    render: (row) => row.nickname || '-',
  },
  {
    title: '邮箱',
    key: 'email',
    minWidth: 220,
    ellipsis: {
      tooltip: true,
    },
    render: (row) => row.email || '-',
  },
  {
    title: '角色',
    key: 'isAdmin',
    width: 90,
    render: (row) =>
      h(
        NTag,
        { type: row.isAdmin ? 'success' : 'default', size: 'small', bordered: false },
        { default: () => (row.isAdmin ? '管理员' : '用户') },
      ),
  },
  {
    title: '状态',
    key: 'isActive',
    width: 90,
    render: (row) =>
      h(
        NTag,
        { type: row.isActive ? 'success' : 'warning', size: 'small', bordered: false },
        { default: () => (row.isActive ? '启用' : '停用') },
      ),
  },
  {
    title: '创建时间',
    key: 'createdAt',
    width: 180,
    render: (row) => formatDate(row.createdAt),
  },
  {
    title: '操作',
    key: 'actions',
    width: 96,
    render: (row) =>
      h(
        NButton,
        { size: 'small', tertiary: true, onClick: () => openEdit(row) },
        { default: () => '编辑' },
      ),
  },
])


function resolveBoolFilter(raw: string): boolean | undefined {
  if (raw === 'true') return true
  if (raw === 'false') return false
  return undefined
}

function doSearch() {
  queryState.value.keyword = keyword.value.trim()
  queryState.value.admin = resolveBoolFilter(adminFilter.value)
  queryState.value.active = resolveBoolFilter(activeFilter.value)
  pagination.page = 1
  refresh()
}

function resetSearch() {
  keyword.value = ''
  adminFilter.value = 'all'
  activeFilter.value = 'all'
  queryState.value.keyword = ''
  queryState.value.admin = undefined
  queryState.value.active = undefined
  pagination.page = 1
  refresh()
}

function openEdit(row: SiteUser) {
  editingUserId.value = row.id
  formModel.username = row.username
  formModel.nickname = row.nickname
  formModel.email = row.email
  formModel.isActive = row.isActive
  formModel.isAdmin = row.isAdmin
  editVisible.value = true
}

async function saveEdit() {
  if (!editingUserId.value) return
  saving.value = true
  try {
    await updateSiteUser(editingUserId.value, {
      nickname: formModel.nickname.trim(),
      email: formModel.email.trim(),
      isActive: formModel.isActive,
      isAdmin: formModel.isAdmin,
    })
    message.success('用户信息已更新')
    editVisible.value = false
    refresh()
  } catch (error: any) {
    message.error(error?.message || '更新失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <ScrollContainer wrapper-class="p-4" :scrollbar-props="{ trigger: 'none' }">
    <NCard title="本站用户管理">
      <NSpace class="mb-4" align="center">
        <NInput
          v-model:value="keyword"
          placeholder="搜索用户名 / 昵称 / 邮箱"
          clearable
          style="width: 280px"
          @keyup.enter="doSearch"
        />
        <NSelect
          v-model:value="adminFilter"
          style="width: 140px"
          :options="[
            { label: '全部角色', value: 'all' },
            { label: '管理员', value: 'true' },
            { label: '普通用户', value: 'false' },
          ]"
        />
        <NSelect
          v-model:value="activeFilter"
          style="width: 140px"
          :options="[
            { label: '全部状态', value: 'all' },
            { label: '已启用', value: 'true' },
            { label: '已停用', value: 'false' },
          ]"
        />
        <NButton type="primary" @click="doSearch">查询</NButton>
        <NButton @click="resetSearch">重置</NButton>
      </NSpace>

      <NDataTable
        remote
        :loading="loading"
        :columns="columns"
        :data="tableData"
        :pagination="pagination"
        :scroll-x="1100"
      />
    </NCard>

    <FormModal v-model:show="editVisible" title="编辑用户" :loading="saving" @confirm="saveEdit">
      <NFormItem label="用户名">
        <NInput :value="formModel.username" disabled />
      </NFormItem>
      <NFormItem label="昵称">
        <NInput v-model:value="formModel.nickname" placeholder="请输入昵称" />
      </NFormItem>
      <NFormItem label="邮箱">
        <NInput v-model:value="formModel.email" placeholder="请输入邮箱（可留空）" />
      </NFormItem>
      <NFormItem label="启用">
        <NSwitch v-model:value="formModel.isActive" :disabled="isEditingSelf" />
      </NFormItem>
      <NFormItem label="管理员">
        <NSwitch v-model:value="formModel.isAdmin" :disabled="isEditingSelf" />
      </NFormItem>
    </FormModal>
  </ScrollContainer>
</template>
