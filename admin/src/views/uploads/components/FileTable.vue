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
  NSpace,
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

const columns = computed<DataTableColumns<UploadFileResponse>>(() => [
  {
    title: '预览',
    key: 'preview',
    width: 80,
    render: (row) => {
      if (row.type === 'picture') {
        return h(
          'div',
          {
            style: { cursor: 'pointer', width: '50px', height: '50px', display: 'flex', alignItems: 'center', justifyContent: 'center' },
            onClick: () => emit('preview', row.publicUrl),
          },
          h(NImage, { src: row.publicUrl, width: 50, height: 50, objectFit: 'cover', style: 'border-radius: 4px', previewDisabled: true }),
        )
      }
      return h(NIcon, { size: 32, color: '#18a058' }, { default: () => h(Document24Regular) })
    },
  },
  {
    title: '文件名',
    key: 'name',
    minWidth: 200,
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
    width: 120,
    render: (row) => formatFileSize(row.size),
  },
  {
    title: '上传时间',
    key: 'createdAt',
    width: 180,
    render: (row) => formatDate(row.createdAt),
  },
  {
    title: '操作',
    key: 'actions',
    width: 240,
    render: (row) =>
      h(NSpace, { size: 'small' }, {
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
    :columns="columns"
    :data="files"
    :loading="loading"
    :bordered="false"
    :single-line="false"
    :scroll-x="900"
  />
</template>
