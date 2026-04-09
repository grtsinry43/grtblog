<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NFormItem,
  NInput,
  NPopconfirm,
  NSpace,
  useMessage,
} from 'naive-ui'
import { h, onMounted, reactive, ref } from 'vue'

import { FormModal, ScrollContainer } from '@/components'
import { createCategory, deleteCategory, listCategories, updateCategory } from '@/services/taxonomy'
import { formatDate } from '@/utils/format'

import type { CategoryItem } from '@/services/taxonomy'
import type { DataTableColumns } from 'naive-ui'

defineOptions({
  name: 'ArticleCategoryManagement',
})

const message = useMessage()
const loading = ref(false)
const saving = ref(false)
const items = ref<CategoryItem[]>([])
const editVisible = ref(false)
const editingId = ref<number | null>(null)
const formModel = reactive({
  name: '',
  shortUrl: '',
})

const columns: DataTableColumns<CategoryItem> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '分类名称', key: 'name', minWidth: 200 },
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
            default: () => '确认删除该分类？',
          },
        ),
      ]),
  },
]

const modalTitle = ref('新建分类')

async function fetchData() {
  loading.value = true
  try {
    items.value = await listCategories()
  } catch (error: any) {
    message.error(error?.message || '获取分类列表失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  modalTitle.value = '新建分类'
  editingId.value = null
  formModel.name = ''
  formModel.shortUrl = ''
  editVisible.value = true
}

function openEdit(row: CategoryItem) {
  modalTitle.value = '编辑分类'
  editingId.value = row.id
  formModel.name = row.name
  formModel.shortUrl = row.shortUrl
  editVisible.value = true
}

async function handleSubmit() {
  const name = formModel.name.trim()
  const shortUrl = formModel.shortUrl.trim()
  if (!name) {
    message.warning('请输入分类名称')
    return
  }
  if (!shortUrl) {
    message.warning('请输入分类短链接')
    return
  }

  saving.value = true
  try {
    if (editingId.value) {
      await updateCategory(editingId.value, { name, shortUrl })
      message.success('分类已更新')
    } else {
      await createCategory({ name, shortUrl })
      message.success('分类已创建')
    }
    editVisible.value = false
    await fetchData()
  } catch (error: any) {
    message.error(error?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: CategoryItem) {
  try {
    await deleteCategory(row.id)
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
  <ScrollContainer
    wrapper-class="p-4"
    :scrollbar-props="{ trigger: 'none' }"
  >
    <NCard title="文章分类管理">
      <template #header-extra>
        <NButton
          type="primary"
          @click="openCreate"
          >新建分类</NButton
        >
      </template>

      <NDataTable
        :columns="columns"
        :data="items"
        :loading="loading"
        :row-key="(row: CategoryItem) => row.id"
      />
    </NCard>

    <FormModal
      v-model:show="editVisible"
      :title="modalTitle"
      :loading="saving"
      @confirm="handleSubmit"
    >
      <NFormItem label="分类名称">
        <NInput
          v-model:value="formModel.name"
          placeholder="请输入分类名称"
        />
      </NFormItem>
      <NFormItem label="分类短链">
        <NInput
          v-model:value="formModel.shortUrl"
          placeholder="例如 frontend"
        />
      </NFormItem>
    </FormModal>
  </ScrollContainer>
</template>
