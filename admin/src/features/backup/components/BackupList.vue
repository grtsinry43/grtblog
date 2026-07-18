<script setup lang="ts">
import { NButton, NEmpty, NPopconfirm, NSpin, NTag } from 'naive-ui'

import { formatDate } from '@/utils/format'

import type { BackupRecord, BackupStatus } from '../model/types'

defineProps<{
  records: BackupRecord[]
  loading: boolean
  deletingId: string | null
}>()

const emit = defineEmits<{
  download: [id: string]
  delete: [id: string]
}>()

const statusMeta: Record<
  BackupStatus,
  { label: string; type: 'default' | 'info' | 'success' | 'error' }
> = {
  queued: { label: '等待中', type: 'default' },
  running: { label: '备份中', type: 'info' },
  completed: { label: '已完成', type: 'success' },
  failed: { label: '失败', type: 'error' },
}

const stageLabels: Record<string, string> = {
  queued: '等待执行',
  preparing: '准备快照',
  snapshotting_uploads: '复制上传文件',
  dumping_database: '导出数据库',
  packing_archive: '打包归档',
  completed: '完成',
  failed: '失败',
  interrupted: '被服务重启中断',
}

function formatSize(size: number) {
  if (!size) return '—'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const index = Math.min(Math.floor(Math.log(size) / Math.log(1024)), units.length - 1)
  return `${(size / 1024 ** index).toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}
</script>

<template>
  <NSpin :show="loading">
    <NEmpty
      v-if="!records.length && !loading"
      description="还没有备份，创建第一份完整站点快照吧"
      class="py-16"
    />
    <div
      v-else
      class="space-y-3"
    >
      <article
        v-for="item in records"
        :key="item.id"
        class="rounded-xl border border-neutral-200 bg-white p-4 transition-shadow hover:shadow-sm dark:border-neutral-700 dark:bg-neutral-900"
      >
        <div class="flex gap-4 max-md:flex-col">
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <span class="iconify text-xl text-[var(--primary-color)] ph--archive" />
              <h3 class="truncate font-medium">{{ item.filename }}</h3>
              <NTag
                size="small"
                :type="statusMeta[item.status].type"
                :bordered="false"
              >
                {{ statusMeta[item.status].label }}
              </NTag>
              <NTag
                v-if="item.triggerType === 'manual'"
                size="small"
                :bordered="false"
                >手动</NTag
              >
            </div>

            <div
              class="mt-3 grid grid-cols-2 gap-x-8 gap-y-2 text-xs text-neutral-500 md:grid-cols-4 dark:text-neutral-400"
            >
              <div>
                <span class="block text-neutral-400">创建时间</span>{{ formatDate(item.createdAt) }}
              </div>
              <div>
                <span class="block text-neutral-400">归档大小</span>{{ formatSize(item.sizeBytes) }}
              </div>
              <div>
                <span class="block text-neutral-400">上传文件</span>{{ item.uploadFileCount }} 个
              </div>
              <div>
                <span class="block text-neutral-400">应用版本</span>{{ item.appVersion || '—' }}
              </div>
            </div>

            <div
              v-if="item.status === 'queued' || item.status === 'running'"
              class="mt-3 flex items-center gap-2 text-xs text-blue-600 dark:text-blue-400"
            >
              <span class="iconify animate-spin ph--spinner" />
              {{ stageLabels[item.stage] || item.stage }}
            </div>
            <div
              v-if="item.errorMessage"
              class="mt-3 rounded-md bg-red-50 px-3 py-2 text-xs text-red-600 dark:bg-red-950/30 dark:text-red-400"
            >
              {{ item.errorMessage }}
            </div>
            <div
              v-if="item.sha256"
              class="mt-3 truncate font-mono text-[11px] text-neutral-400"
              :title="item.sha256"
            >
              SHA-256 · {{ item.sha256 }}
            </div>
          </div>

          <div class="flex shrink-0 items-start gap-2">
            <NButton
              size="small"
              type="primary"
              secondary
              :disabled="item.status !== 'completed'"
              @click="emit('download', item.id)"
            >
              <template #icon><span class="iconify ph--download-simple" /></template>
              下载
            </NButton>
            <NPopconfirm
              :disabled="item.status === 'queued' || item.status === 'running'"
              @positive-click="emit('delete', item.id)"
            >
              <template #trigger>
                <NButton
                  size="small"
                  tertiary
                  type="error"
                  :loading="deletingId === item.id"
                  :disabled="item.status === 'queued' || item.status === 'running'"
                >
                  <template #icon><span class="iconify ph--trash" /></template>
                </NButton>
              </template>
              删除后归档文件也会被永久移除，确认继续？
            </NPopconfirm>
          </div>
        </div>
      </article>
    </div>
  </NSpin>
</template>
