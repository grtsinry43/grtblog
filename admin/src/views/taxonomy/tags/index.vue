<script setup lang="ts">
import { NButton, NCard, NDataTable, NForm, NFormItem, NInput, NModal, NPopconfirm, NSpace, useMessage } from 'naive-ui'
import { h, onMounted, reactive, ref } from 'vue'

import { ScrollContainer } from '@/components'
import { createTag, deleteTag, listTags, updateTag } from '@/services/taxonomy'

import type { TagItem } from '@/services/taxonomy'
import type { DataTableColumns } from 'naive-ui'

defineOptions({
  name: 'TagManagement',
})

const message = useMessage()
const loading = ref(false)
const saving = ref(false)
const items = ref<TagItem[]>([])
const editVisible = ref(false)
const editingId = ref<number | null>(null)
const formModel = reactive({
  name: '',
})

const columns: DataTableColumns<TagItem> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '标签名称', key: 'name', minWidth: 220 },
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
            default: () => '确认删除该标签？',
          },
        ),
      ]),
  },
]

const modalTitle = ref('新建标签')

function formatDate(value?: string) {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

async function fetchData() {
  loading.value = true
  try {
    items.value = await listTags()
  } catch (error: any) {
    message.error(error?.message || '获取标签列表失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  modalTitle.value = '新建标签'
  editingId.value = null
  formModel.name = ''
  editVisible.value = true
}

function openEdit(row: TagItem) {
  modalTitle.value = '编辑标签'
  editingId.value = row.id
  formModel.name = row.name
  editVisible.value = true
}

async function handleSubmit() {
  const name = formModel.name.trim()
  if (!name) {
    message.warning('请输入标签名称')
    return
  }

  saving.value = true
  try {
    if (editingId.value) {
      await updateTag(editingId.value, { name })
      message.success('标签已更新')
    } else {
      await createTag(name)
      message.success('标签已创建')
    }
    editVisible.value = false
    await fetchData()
  } catch (error: any) {
    message.error(error?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: TagItem) {
  try {
    await deleteTag(row.id)
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
    <NCard title="标签管理">
      <template #header-extra>
        <NButton type="primary" @click="openCreate">新建标签</NButton>
      </template>

      <NDataTable
        :columns="columns"
        :data="items"
        :loading="loading"
        :row-key="(row: TagItem) => row.id"
      />
    </NCard>

    <NModal v-model:show="editVisible" preset="card" :title="modalTitle" style="width: 520px">
      <NForm label-placement="left" label-width="84">
        <NFormItem label="标签名称">
          <NInput v-model:value="formModel.name" placeholder="请输入标签名称" />
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
