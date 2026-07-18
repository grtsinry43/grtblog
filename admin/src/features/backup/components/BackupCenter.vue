<script setup lang="ts">
import { NAlert, NButton, NCard, useMessage } from 'naive-ui'

import { useBackups } from '../composables/use-backups'

import BackupList from './BackupList.vue'

const message = useMessage()
const { records, loading, creating, deletingId, refresh, create, remove, download } = useBackups()

async function handleCreate() {
  try {
    await create()
    message.success('完整备份已开始，可留在此页查看进度')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '创建备份失败')
  }
}

async function handleDelete(id: string) {
  try {
    await remove(id)
    message.success('备份已删除')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '删除备份失败')
  }
}

async function handleDownload(id: string) {
  try {
    await download(id)
  } catch (error) {
    message.error(error instanceof Error ? error.message : '生成下载链接失败')
  }
}
</script>

<template>
  <div class="space-y-4">
    <NAlert
      type="info"
      :bordered="false"
    >
      完整备份包含站点数据库、系统配置、内容、用户与互动记录，以及上传文件。归档可能含有账号和密钥等敏感信息，请妥善保存。
    </NAlert>

    <NCard :bordered="false">
      <div class="mb-5 flex items-start justify-between gap-4 max-sm:flex-col">
        <div>
          <h2 class="text-base font-semibold">站点完整备份</h2>
          <p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
            创建一致性数据库快照并打包为 tar.gz，任务在后台运行。
          </p>
        </div>
        <div class="flex gap-2">
          <NButton
            secondary
            :loading="loading"
            @click="refresh()"
          >
            <template #icon><span class="iconify ph--arrows-clockwise" /></template>
            刷新
          </NButton>
          <NButton
            type="primary"
            :loading="creating"
            :disabled="
              records.some((item) => item.status === 'queued' || item.status === 'running')
            "
            @click="handleCreate"
          >
            <template #icon><span class="iconify ph--archive" /></template>
            立即备份
          </NButton>
        </div>
      </div>

      <BackupList
        :records="records"
        :loading="loading"
        :deleting-id="deletingId"
        @download="handleDownload"
        @delete="handleDelete"
      />
    </NCard>
  </div>
</template>
