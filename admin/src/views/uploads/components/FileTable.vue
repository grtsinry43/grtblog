<script setup lang="ts">
import {
  Copy24Regular,
  Delete24Regular,
  Document24Regular,
  Edit24Regular,
} from '@vicons/fluent'
import {
  NButton,
  NDataTable,
  NIcon,
  NImage,
  NTag,
} from 'naive-ui'
import { computed, h } from 'vue'
import { formatDateZhCN as formatDate, formatFileSize } from '@/utils/format'

import type { UploadFileResponse } from '@/services/uploads'
import type { DataTableColumns } from 'naive-ui'

const props = defineProps<{
  files: UploadFileResponse[]
  loading: boolean
}>()

const emit = defineEmits<{
  copyUrl: [file: UploadFileResponse]
  rename: [file: UploadFileResponse]
  download: [file: UploadFileResponse]
  delete: [file: UploadFileResponse]
  preview: [url: string]
}>()

const actionCellStyle = {
  display: 'flex',
  alignItems: 'center',
  gap: '4px',
  flexWrap: 'nowrap',
  whiteSpace: 'nowrap',
} as const

const previewCellStyle = {
  width: '100%',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
} as const

const columns = computed<DataTableColumns<UploadFileResponse>>(() => [
  {
    title: '预览',
    key: 'preview',
    width: 68,
    render: (row) => {
      if (row.type === 'picture') {
        return h(
          'div',
          {
            style: { ...previewCellStyle, cursor: 'pointer', height: '40px' },
            onClick: () => emit('preview', row.publicUrl),
          },
          h(NImage, { src: row.publicUrl, width: 40, height: 40, objectFit: 'cover', style: 'border-radius: 4px', previewDisabled: true }),
        )
      }
      return h(
        'div',
        { style: previewCellStyle },
        h(NIcon, { size: 24, color: '#18a058' }, { default: () => h(Document24Regular) }),
      )
    },
  },
  {
    title: '文件名',
    key: 'name',
    minWidth: 180,
    ellipsis: { tooltip: true },
  },
  {
    title: '类型',
    key: 'type',
    width: 100,
    render: (row) =>
      h(NTag, { type: row.type === 'picture' ? 'success' : 'info', size: 'small' }, { default: () => (row.type === 'picture' ? '图片' : '文件') }),
  },
  {
    title: '大小',
    key: 'size',
    width: 96,
    render: (row) => formatFileSize(row.size),
  },
  {
    title: '上传时间',
    key: 'createdAt',
    width: 168,
    render: (row) => formatDate(row.createdAt),
  },
  {
    title: '操作',
    key: 'actions',
    width: 320,
    render: (row) =>
      h('div', { style: actionCellStyle }, {
        default: () => [
          h(NButton, { size: 'small', quaternary: true, onClick: () => emit('copyUrl', row) }, {
            icon: () => h(NIcon, null, { default: () => h(Copy24Regular) }),
            default: () => '复制链接',
          }),
          h(NButton, { size: 'small', quaternary: true, onClick: () => emit('rename', row) }, {
            icon: () => h(NIcon, null, { default: () => h(Edit24Regular) }),
            default: () => '重命名',
          }),
          h(NButton, { size: 'small', quaternary: true, onClick: () => emit('download', row) }, { default: () => '下载' }),
          h(NButton, { size: 'small', quaternary: true, type: 'error', onClick: () => emit('delete', row) }, {
            icon: () => h(NIcon, null, { default: () => h(Delete24Regular) }),
            default: () => '删除',
          }),
        ],
      }),
  },
])
</script>

<template>
  <NDataTable
    class="file-table"
    :columns="columns"
    :data="files"
    :loading="loading"
    :bordered="false"
    :single-line="false"
    size="small"
    :scroll-x="900"
  />
</template>

<style scoped>
.file-table :deep(.n-data-table-th),
.file-table :deep(.n-data-table-td) {
  padding-top: 8px;
  padding-bottom: 8px;
}

.file-table :deep(.n-button) {
  padding-left: 6px;
  padding-right: 6px;
}
</style>
