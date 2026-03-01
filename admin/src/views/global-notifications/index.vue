<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NFormItem,
  NInput,
  NPopconfirm,
  NSpace,
  NSwitch,
  NDatePicker,
  NTag,
  useMessage,
} from 'naive-ui'
import { h, ref, computed } from 'vue'
import { FormModal, ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import {
  listGlobalNotifications,
  createGlobalNotification,
  updateGlobalNotification,
  deleteGlobalNotification,
} from '@/services/global-notifications'
import type { DataTableColumns } from 'naive-ui'
import type { GlobalNotificationItem } from '@/services/global-notifications'
import { formatDate } from '@/utils/format'

defineOptions({
  name: 'GlobalNotificationList',
})

const message = useMessage()

// Table Logic
const {
  loading,
  data: tableData,
  pagination,
  refresh,
} = useTable(listGlobalNotifications)

const columns = computed<DataTableColumns<GlobalNotificationItem>>(() => [
  {
    title: 'ID',
    key: 'id',
    width: 80,
  },
  {
    title: '内容',
    key: 'content',
    minWidth: 200,
    ellipsis: { tooltip: true },
    render: (row) => h('div', { class: 'truncate max-w-md' }, row.content),
  },
  {
    title: '开始时间',
    key: 'publishAt',
    width: 180,
    render: (row) => formatDate(row.publishAt),
  },
  {
    title: '结束时间',
    key: 'expireAt',
    width: 180,
    render: (row) => formatDate(row.expireAt),
  },
  {
    title: '允许"不再提示"',
    key: 'allowClose',
    width: 180,
    render: (row) =>
      h(
        NTag,
        { size: 'small', type: row.allowClose ? 'success' : 'warning', bordered: false },
        { default: () => (row.allowClose ? '是' : '否') },
      ),
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render: (row) =>
      h(
        NSpace,
        { size: 'small' },
        {
          default: () => [
            h(
              NButton,
              {
                size: 'small',
                secondary: true,
                type: 'primary',
                onClick: () => openEdit(row),
              },
              { default: () => '编辑' },
            ),
            h(
              NPopconfirm,
              {
                onPositiveClick: () => handleDelete(row),
              },
              {
                trigger: () =>
                  h(
                    NButton,
                    {
                      size: 'small',
                      secondary: true,
                      type: 'error',
                    },
                    { default: () => '删除' },
                  ),
                default: () => '确认删除该通知？',
              },
            ),
          ],
        },
      ),
  },
])

// Form Logic
const formVisible = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const formParams = ref({
  content: '',
  publishAt: Date.now(),
  expireAt: Date.now() + 86400000 * 7, // Default 7 days
  allowClose: true,
})

const formTitle = computed(() => (editingId.value ? '编辑通知' : '新建通知'))

function openCreate() {
  editingId.value = null
  formParams.value = {
    content: '',
    publishAt: Date.now(),
    expireAt: Date.now() + 86400000 * 7,
    allowClose: true,
  }
  formVisible.value = true
}

function openEdit(row: GlobalNotificationItem) {
  editingId.value = row.id
  formParams.value = {
    content: row.content,
    publishAt: new Date(row.publishAt).getTime(),
    expireAt: new Date(row.expireAt).getTime(),
    allowClose: row.allowClose,
  }
  formVisible.value = true
}

async function handleSave() {
  if (!formParams.value.content) {
    message.error('请输入内容')
    return
  }
  saving.value = true
  try {
    const payload = {
      content: formParams.value.content,
      publishAt: new Date(formParams.value.publishAt).toISOString(),
      expireAt: new Date(formParams.value.expireAt).toISOString(),
      allowClose: formParams.value.allowClose,
    }
    if (editingId.value) {
      await updateGlobalNotification(editingId.value, payload)
      message.success('更新成功')
    } else {
      await createGlobalNotification(payload)
      message.success('创建成功')
    }
    formVisible.value = false
    refresh()
  } catch (e) {
    // Error handling is generic in request
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: GlobalNotificationItem) {
  try {
    await deleteGlobalNotification(row.id)
    message.success('删除成功')
    refresh()
  } catch (e) {
    // Error handling
  }
}

</script>

<template>
  <ScrollContainer wrapper-class="p-4">
    <NCard title="全站通知管理">
      <template #header-extra>
        <NButton type="primary" @click="openCreate">新建通知</NButton>
      </template>

      <NDataTable
        remote
        :loading="loading"
        :columns="columns"
        :data="tableData"
        :pagination="pagination"
        class="mt-4"
        :scroll-x="900"
      />
    </NCard>

    <FormModal
      v-model:show="formVisible"
      :title="formTitle"
      :loading="saving"
      :label-width="100"
      @confirm="handleSave"
    >
      <NFormItem label="内容" required>
        <NInput
          v-model:value="formParams.content"
          type="textarea"
          placeholder="请输入通知内容"
        />
      </NFormItem>
      <NFormItem label="开始时间" required>
        <NDatePicker v-model:value="formParams.publishAt" type="datetime" style="width: 100%" />
      </NFormItem>
      <NFormItem label="结束时间" required>
        <NDatePicker v-model:value="formParams.expireAt" type="datetime" style="width: 100%" />
      </NFormItem>
      <NFormItem>
        <template #label>允许&ldquo;不再提示&rdquo;</template>
        <NSwitch v-model:value="formParams.allowClose" />
      </NFormItem>
    </FormModal>
  </ScrollContainer>
</template>
