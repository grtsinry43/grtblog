<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NFormItem,
  NInput,
  NModal,
  NPopconfirm,
  NSpace,
  NTag,
  useMessage,
} from 'naive-ui'
import { computed, h, ref } from 'vue'

import { FormModal } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { createAdminToken, deleteAdminToken, listAdminTokens } from '@/services/admin-tokens'
import { formatDate } from '@/utils/format'

import type { AdminTokenItem } from '@/services/admin-tokens'
import type { DataTableColumns } from 'naive-ui'

const message = useMessage()
const { loading, data: tableData, pagination, refresh } = useTable<AdminTokenItem>(listAdminTokens)

const columns = computed<DataTableColumns<AdminTokenItem>>(() => [
  {
    title: 'ID',
    key: 'id',
    width: 70,
  },
  {
    title: '描述',
    key: 'description',
    render: (row) => row.description || '-',
  },
  {
    title: '创建人',
    key: 'username',
    width: 140,
  },
  {
    title: '到期时间',
    key: 'expireAt',
    width: 180,
    render: (row) => formatDate(row.expireAt),
  },
  {
    title: '状态',
    key: 'isExpired',
    width: 90,
    render: (row) =>
      h(
        NTag,
        { type: row.isExpired ? 'warning' : 'success', size: 'small', bordered: false },
        { default: () => (row.isExpired ? '已过期' : '有效') },
      ),
  },
  {
    title: '操作',
    key: 'actions',
    width: 90,
    render: (row) =>
      h(
        NPopconfirm,
        { onPositiveClick: () => handleDelete(row) },
        {
          trigger: () =>
            h(NButton, { size: 'small', type: 'error', tertiary: true }, { default: () => '删除' }),
          default: () => '确认删除该 token？',
        },
      ),
  },
])

const createVisible = ref(false)
const revealVisible = ref(false)
const saving = ref(false)
const createdToken = ref('')
const formParams = ref({
  description: '',
  expireAt: Date.now() + 1000 * 60 * 60 * 24 * 30,
})

function openCreate() {
  formParams.value = {
    description: '',
    expireAt: Date.now() + 1000 * 60 * 60 * 24 * 30,
  }
  createVisible.value = true
}

async function handleCreate() {
  if (!formParams.value.expireAt) {
    message.error('请选择过期时间')
    return
  }
  saving.value = true
  try {
    const payload = {
      description: formParams.value.description.trim(),
      expireAt: new Date(formParams.value.expireAt).toISOString(),
    }
    const result = await createAdminToken(payload)
    createdToken.value = result.token
    createVisible.value = false
    revealVisible.value = true
    message.success('创建成功')
    refresh()
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: AdminTokenItem) {
  await deleteAdminToken(row.id)
  message.success('删除成功')
  refresh()
}
</script>

<template>
  <NCard title="管理员 API Token">
    <template #header-extra>
      <NButton
        type="primary"
        @click="openCreate"
        >新建 Token</NButton
      >
    </template>

    <NDataTable
      remote
      :loading="loading"
      :columns="columns"
      :data="tableData"
      :pagination="pagination"
      class="mt-4"
    />
  </NCard>

  <FormModal
    v-model:show="createVisible"
    title="新建管理员 Token"
    :loading="saving"
    :label-width="96"
    confirm-text="创建"
    @confirm="handleCreate"
  >
    <NFormItem label="描述">
      <NInput
        v-model:value="formParams.description"
        maxlength="200"
        placeholder="可选，用于区分用途（例如：CI 调用）"
      />
    </NFormItem>
    <NFormItem
      label="过期时间"
      required
    >
      <NDatePicker
        v-model:value="formParams.expireAt"
        type="datetime"
        style="width: 100%"
      />
    </NFormItem>
  </FormModal>

  <NModal
    v-model:show="revealVisible"
    preset="card"
    title="Token 已生成"
    style="width: 560px"
  >
    <div class="mb-3 text-sm text-[var(--text-color-2)]">仅展示一次，请立即复制保存。</div>
    <NInput
      :value="createdToken"
      type="textarea"
      :rows="3"
      readonly
    />
    <template #footer>
      <NSpace justify="end">
        <NButton @click="revealVisible = false">我已保存</NButton>
      </NSpace>
    </template>
  </NModal>
</template>
