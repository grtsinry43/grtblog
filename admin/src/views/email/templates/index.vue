<script setup lang="ts">
import { NButton, NCard, NDataTable, NPopconfirm, NSpace, NTag, useMessage } from 'naive-ui'
import { h, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { deleteEmailTemplate, listEmailTemplates } from '@/services/email'

import type { EmailTemplate } from '@/services/email'
import type { DataTableColumns } from 'naive-ui'

const message = useMessage()
const router = useRouter()

const loading = ref(false)
const items = ref<EmailTemplate[]>([])

const columns: DataTableColumns<EmailTemplate> = [
  {
    title: '名称',
    key: 'name',
    width: 200,
  },
  {
    title: '编码',
    key: 'code',
    width: 150,
    render: (row) => h(NTag, { type: 'info', size: 'small', bordered: false }, { default: () => row.code }),
  },
  {
    title: '事件',
    key: 'eventName',
    width: 150,
  },
  {
    title: '状态',
    key: 'isEnabled',
    width: 100,
    render: (row) =>
      h(
        NTag,
        { type: row.isEnabled ? 'success' : 'warning', size: 'small', bordered: false },
        { default: () => (row.isEnabled ? '启用' : '禁用') },
      ),
  },
  {
    title: '更新时间',
    key: 'updatedAt',
    width: 180,
    render: (row) => new Date(row.updatedAt).toLocaleString(),
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render: (row) =>
      h(NSpace, { size: 'small' }, () => [
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            secondary: true,
            onClick: () => router.push(`/email/templates/${row.code}`),
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
                  type: 'error',
                  secondary: true,
                },
                { default: () => '删除' },
              ),
            default: () => '确认删除该模版？',
          },
        ),
      ]),
  },
]

async function fetchData() {
  loading.value = true
  try {
    items.value = await listEmailTemplates()
  } finally {
    loading.value = false
  }
}

async function handleDelete(row: EmailTemplate) {
  try {
    await deleteEmailTemplate(row.code)
    message.success('删除成功')
    fetchData()
  } catch (err) {
    //
  }
}

function handleCreate() {
  router.push('/email/templates/new')
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <ScrollContainer>
    <NCard title="邮件模版">
      <template #header-extra>
        <NButton
          type="primary"
          @click="handleCreate"
        >
          新建模版
        </NButton>
      </template>

      <NDataTable
        :columns="columns"
        :data="items"
        :loading="loading"
        :row-key="(row: EmailTemplate) => row.id"
      />
    </NCard>
  </ScrollContainer>
</template>
