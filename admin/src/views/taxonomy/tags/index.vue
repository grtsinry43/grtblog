<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NDrawer,
  NDrawerContent,
  NEmpty,
  NFormItem,
  NInput,
  NPopconfirm,
  NSpace,
  NSpin,
  NTag,
  useMessage,
} from 'naive-ui'
import { h, onMounted, reactive, ref } from 'vue'

import { FormModal, ScrollContainer } from '@/components'
import { createTag, deleteTag, getTagContents, listTags, updateTag } from '@/services/taxonomy'
import { formatDate } from '@/utils/format'

import type { TagItem, TagRelatedContents } from '@/services/taxonomy'
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

// 关联内容 Drawer
const contentsDrawer = ref(false)
const contentsTag = ref<TagItem | null>(null)
const contentsLoading = ref(false)
const contentsData = ref<TagRelatedContents | null>(null)

async function openContents(row: TagItem) {
  contentsTag.value = row
  contentsDrawer.value = true
  contentsLoading.value = true
  contentsData.value = null
  try {
    contentsData.value = await getTagContents(row.id)
  } catch (e: any) {
    message.error(e?.message || '获取关联内容失败')
  } finally {
    contentsLoading.value = false
  }
}

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
    width: 260,
    render: (row) =>
      h(NSpace, { size: 'small' }, () => [
        h(
          NButton,
          { size: 'small', tertiary: true, onClick: () => openContents(row) },
          {
            icon: () => h('div', { class: 'iconify ph--article' }),
            default: () => '关联内容',
          },
        ),
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
  <ScrollContainer
    wrapper-class="p-4"
    :scrollbar-props="{ trigger: 'none' }"
  >
    <NCard title="标签管理">
      <template #header-extra>
        <NButton
          type="primary"
          @click="openCreate"
          >新建标签</NButton
        >
      </template>

      <NDataTable
        :columns="columns"
        :data="items"
        :loading="loading"
        :row-key="(row: TagItem) => row.id"
      />
    </NCard>

    <FormModal
      v-model:show="editVisible"
      :title="modalTitle"
      :loading="saving"
      @confirm="handleSubmit"
    >
      <NFormItem label="标签名称">
        <NInput
          v-model:value="formModel.name"
          placeholder="请输入标签名称"
        />
      </NFormItem>
    </FormModal>

    <NDrawer
      v-model:show="contentsDrawer"
      placement="right"
      width="380"
    >
      <NDrawerContent
        :title="`「${contentsTag?.name ?? ''}」关联内容`"
        :native-scrollbar="false"
        closable
        header-style="padding: 20px 24px;"
        body-style="padding: 0 24px 24px;"
      >
        <NSpin :show="contentsLoading">
          <template v-if="contentsData">
            <div class="space-y-5">
              <!-- 文章 -->
              <div>
                <div class="mb-2 flex items-center gap-2 text-sm font-medium">
                  <div class="iconify ph--article" />
                  <span>文章</span>
                  <NTag
                    size="small"
                    :bordered="false"
                    round
                    >{{ contentsData.articles?.length ?? 0 }}</NTag
                  >
                </div>
                <div
                  v-if="contentsData.articles?.length"
                  class="space-y-1"
                >
                  <router-link
                    v-for="article in contentsData.articles"
                    :key="article.id"
                    :to="{ name: 'articleEdit', params: { id: article.id } }"
                    class="flex items-start gap-2 rounded-lg px-3 py-2.5 no-underline transition-colors hover:bg-gray-100 dark:hover:bg-gray-800"
                    @click="contentsDrawer = false"
                  >
                    <div class="min-w-0 flex-1">
                      <div class="truncate text-sm font-medium">{{ article.title }}</div>
                      <div class="mt-0.5 truncate text-xs opacity-50">
                        {{ formatDate(article.createdAt) }}
                      </div>
                    </div>
                  </router-link>
                </div>
                <NEmpty
                  v-else
                  description="无关联文章"
                  size="small"
                  class="py-3"
                />
              </div>

              <!-- 手记 -->
              <div>
                <div class="mb-2 flex items-center gap-2 text-sm font-medium">
                  <div class="iconify ph--notebook" />
                  <span>手记</span>
                  <NTag
                    size="small"
                    :bordered="false"
                    round
                    >{{ contentsData.moments?.length ?? 0 }}</NTag
                  >
                </div>
                <div
                  v-if="contentsData.moments?.length"
                  class="space-y-1"
                >
                  <router-link
                    v-for="moment in contentsData.moments"
                    :key="moment.id"
                    :to="{ name: 'noteEdit', params: { id: moment.id } }"
                    class="flex items-start gap-2 rounded-lg px-3 py-2.5 no-underline transition-colors hover:bg-gray-100 dark:hover:bg-gray-800"
                    @click="contentsDrawer = false"
                  >
                    <div class="min-w-0 flex-1">
                      <div class="truncate text-sm font-medium">{{ moment.title }}</div>
                      <div class="mt-0.5 truncate text-xs opacity-50">
                        {{ formatDate(moment.createdAt) }}
                      </div>
                    </div>
                  </router-link>
                </div>
                <NEmpty
                  v-else
                  description="无关联手记"
                  size="small"
                  class="py-3"
                />
              </div>
            </div>
          </template>
          <div
            v-else-if="!contentsLoading"
            class="py-8"
          >
            <NEmpty description="加载失败" />
          </div>
        </NSpin>
      </NDrawerContent>
    </NDrawer>
  </ScrollContainer>
</template>
