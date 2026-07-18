<script setup lang="ts">
import {
  NAlert,
  NButton,
  NCard,
  NForm,
  NFormItem,
  NInputNumber,
  NSelect,
  NSkeleton,
  NSwitch,
  useMessage,
} from 'naive-ui'
import { reactive, watch } from 'vue'

import { formatDate } from '@/utils/format'

import { useBackups } from '../composables/use-backups'

import BackupList from './BackupList.vue'

const message = useMessage()
const {
  records,
  loading,
  creating,
  deletingId,
  schedule,
  scheduleLoading,
  scheduleSaving,
  refresh,
  create,
  remove,
  download,
  saveSchedule,
  setPinned,
} = useBackups()

const scheduleForm = reactive({ enabled: false, intervalHours: 24, retentionCount: 7 })
const intervalOptions = [
  { label: '每 6 小时', value: 6 },
  { label: '每 12 小时', value: 12 },
  { label: '每天', value: 24 },
  { label: '每 3 天', value: 72 },
  { label: '每周', value: 168 },
  { label: '每月（30 天）', value: 720 },
]

watch(
  schedule,
  (value) => {
    if (!value) return
    scheduleForm.enabled = value.enabled
    scheduleForm.intervalHours = value.intervalHours
    scheduleForm.retentionCount = value.retentionCount
  },
  { immediate: true },
)

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

async function handleSaveSchedule() {
  try {
    await saveSchedule({ ...scheduleForm })
    message.success(scheduleForm.enabled ? '自动备份计划已启用' : '自动备份计划已停用')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存备份计划失败')
  }
}

async function handlePin(id: string, pinned: boolean) {
  try {
    await setPinned(id, pinned)
    message.success(pinned ? '备份已固定' : '已取消固定')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '更新备份失败')
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

    <NCard
      title="自动备份计划"
      :bordered="false"
    >
      <NSkeleton
        v-if="scheduleLoading && !schedule"
        text
        :repeat="3"
      />
      <NForm
        v-else
        label-placement="left"
        label-width="110"
        class="max-w-2xl"
      >
        <NFormItem label="自动备份">
          <NSwitch v-model:value="scheduleForm.enabled" />
        </NFormItem>
        <NFormItem label="执行频率">
          <NSelect
            v-model:value="scheduleForm.intervalHours"
            :options="intervalOptions"
            :disabled="!scheduleForm.enabled"
          />
        </NFormItem>
        <NFormItem label="保留自动备份">
          <div class="flex items-center gap-2">
            <NInputNumber
              v-model:value="scheduleForm.retentionCount"
              :min="1"
              :max="100"
              :disabled="!scheduleForm.enabled"
            />
            <span class="text-sm text-neutral-500">份；固定备份和手动备份不计入清理</span>
          </div>
        </NFormItem>
        <div
          class="mb-4 grid grid-cols-2 gap-4 text-xs text-neutral-500 max-sm:grid-cols-1 dark:text-neutral-400"
        >
          <div>
            上次执行：{{ schedule?.lastRunAt ? formatDate(schedule.lastRunAt) : '尚未执行' }}
          </div>
          <div>下次执行：{{ schedule?.nextRunAt ? formatDate(schedule.nextRunAt) : '未安排' }}</div>
        </div>
        <NButton
          type="primary"
          :loading="scheduleSaving"
          @click="handleSaveSchedule"
          >保存计划</NButton
        >
      </NForm>
    </NCard>

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
        @pin="handlePin"
      />
    </NCard>
  </div>
</template>
