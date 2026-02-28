<script setup lang="ts">
import { NButton, NCard, NDataTable, NForm, NFormItem, NInput, NModal, NPopconfirm, NSpace, useMessage } from 'naive-ui'
import { h, onMounted, reactive, ref } from 'vue'

import { ScrollContainer } from '@/components'
import { createColumn, deleteColumn, listColumns, updateColumn } from '@/services/taxonomy'

import type { ColumnItem } from '@/services/taxonomy'
import type { DataTableColumns } from 'naive-ui'

defineOptions({
  name: 'MomentColumnManagement',
})

const message = useMessage()
const loading = ref(false)
const saving = ref(false)
const items = ref<ColumnItem[]>([])
const editVisible = ref(false)
const editingId = ref<number | null>(null)
const formModel = reactive({
  name: '',
  shortUrl: '',
})

const columns: DataTableColumns<ColumnItem> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '专栏名称', key: 'name', minWidth: 200 },
  { title: '短链接', key: 'shortUrl', minWidth: 180 },
  {
    title: '更新时间',
    key: 'updatedAt',
    width: 180,
    render: (row) => formatDate(row.updatedAt),
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render: (row) =>
      h(NSpace, { size: 'small' }, () => [
        h(
          NButton,
          { size: 'small', tertiary: true, onClick: () => openEdit(row) },
          { default: () => '编辑' },
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => handleDelete(row) },
          {
            trigger: () =>
              h(
                NButton,
                { size: 'small', type: 'error', secondary: true },
                { default: () => '删除' },
              ),
            default: () => '确认删除该专栏？',
          },
        ),
      ]),
  },
]

const modalTitle = ref('新建专栏')

function formatDate(value?: string) {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

async function fetchData() {
  loading.value = true
  try {
    items.value = await listColumns()
  } catch (error: any) {
    message.error(error?.message || '获取专栏列表失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  modalTitle.value = '新建专栏'
  editingId.value = null
  formModel.name = ''
  formModel.shortUrl = ''
  editVisible.value = true
}

function openEdit(row: ColumnItem) {
  modalTitle.value = '编辑专栏'
  editingId.value = row.id
  formModel.name = row.name
  formModel.shortUrl = row.shortUrl
  editVisible.value = true
}

async function handleSubmit() {
  const name = formModel.name.trim()
  const shortUrl = formModel.shortUrl.trim()
  if (!name) {
    message.warning('请输入专栏名称')
    return
  }
  if (!shortUrl) {
    message.warning('请输入专栏短链接')
    return
  }

  saving.value = true
  try {
    if (editingId.value) {
      await updateColumn(editingId.value, { name, shortUrl })
      message.success('专栏已更新')
    } else {
      await createColumn({ name, shortUrl })
      message.success('专栏已创建')
    }
    editVisible.value = false
    await fetchData()
  } catch (error: any) {
    message.error(error?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: ColumnItem) {
  try {
    await deleteColumn(row.id)
    message.success('删除成功')
    await fetchData()
  } catch (error: any) {
    message.error(error?.message || '删除失败')
  }
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <ScrollContainer wrapper-class="p-4" :scrollbar-props="{ trigger: 'none' }">
    <NCard title="手记专栏管理">
      <template #header-extra>
        <NButton type="primary" @click="openCreate">新建专栏</NButton>
      </template>

      <NDataTable
        :columns="columns"
        :data="items"
        :loading="loading"
        :row-key="(row: ColumnItem) => row.id"
      />
    </NCard>

    <NModal v-model:show="editVisible" preset="card" :title="modalTitle" style="width: 560px">
      <NForm label-placement="left" label-width="84">
        <NFormItem label="专栏名称">
          <NInput v-model:value="formModel.name" placeholder="请输入专栏名称" />
        </NFormItem>
        <NFormItem label="专栏短链">
          <NInput v-model:value="formModel.shortUrl" placeholder="例如 life-notes" />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="editVisible = false">取消</NButton>
          <NButton type="primary" :loading="saving" @click="handleSubmit">保存</NButton>
        </NSpace>
      </template>
    </NModal>
  </ScrollContainer>
</template>
