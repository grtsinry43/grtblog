<script setup lang="ts">
import {
  NAlert,
  NButton,
  NCard,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NSelect,
  NSkeleton,
  NSwitch,
  useMessage,
} from 'naive-ui'
import { reactive, shallowRef, watch } from 'vue'

import { formatDate } from '@/utils/format'

import { useBackups } from '../composables/use-backups'

import BackupList from './BackupList.vue'
import LocalRestoreDevHint from './LocalRestoreDevHint.vue'

const message = useMessage()
const {
  records,
  loading,
  creating,
  deletingId,
  schedule,
  scheduleLoading,
  scheduleSaving,
  restoreStatus,
  restoring,
  refresh,
  create,
  remove,
  download,
  saveSchedule,
  setPinned,
  restoreExisting,
  restoreUpload,
} = useBackups()

const restoreModalVisible = shallowRef(false)
const restoreTargetId = shallowRef<string | null>(null)
const selectedArchive = shallowRef<File | null>(null)
const confirmation = shallowRef('')
const archiveInput = shallowRef<HTMLInputElement | null>(null)

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

function openExistingRestore(id: string) {
  restoreTargetId.value = id
  selectedArchive.value = null
  confirmation.value = ''
  restoreModalVisible.value = true
}

function chooseRestoreArchive() {
  archiveInput.value?.click()
}

function handleArchiveSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  selectedArchive.value = file
  restoreTargetId.value = null
  confirmation.value = ''
  restoreModalVisible.value = true
}

async function confirmRestore() {
  try {
    if (selectedArchive.value) {
      await restoreUpload(selectedArchive.value, confirmation.value)
    } else if (restoreTargetId.value) {
      await restoreExisting(restoreTargetId.value, confirmation.value)
    } else {
      return
    }
    message.warning('恢复请求已接受，服务即将重启。恢复完成后此页面会重新可用。', {
      duration: 10000,
    })
    restoreModalVisible.value = false
  } catch (error) {
    message.error(error instanceof Error ? error.message : '创建恢复请求失败')
  }
}

function restoreAlertType() {
  if (restoreStatus.value?.state === 'failed') return 'error'
  if (restoreStatus.value?.state === 'succeeded') return 'success'
  return 'warning'
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

    <NCard
      title="覆盖恢复"
      :bordered="false"
    >
      <LocalRestoreDevHint />
      <NAlert
        v-if="restoreStatus && restoreStatus.state !== 'idle'"
        :type="restoreAlertType()"
        :title="
          restoreStatus.state === 'succeeded'
            ? '最近一次恢复成功'
            : restoreStatus.state === 'failed'
              ? '最近一次恢复失败'
              : '恢复等待执行'
        "
        class="mb-4"
      >
        {{ restoreStatus.message }}
        <span
          v-if="restoreStatus.completedAt"
          class="ml-2 opacity-70"
          >{{ formatDate(restoreStatus.completedAt) }}</span
        >
      </NAlert>
      <div class="flex items-start justify-between gap-4 max-sm:flex-col">
        <div>
          <p class="text-sm text-neutral-600 dark:text-neutral-300">
            可从下方已有备份恢复，也可以上传另一台站点导出的 tar.gz。
          </p>
          <p class="mt-1 text-xs text-neutral-500">
            恢复会重启服务，并覆盖数据库、配置、内容、互动记录与全部上传文件。
          </p>
        </div>
        <NButton
          type="warning"
          secondary
          @click="chooseRestoreArchive"
        >
          <template #icon><span class="iconify ph--upload-simple" /></template>
          上传备份并恢复
        </NButton>
        <input
          ref="archiveInput"
          type="file"
          accept=".tar.gz,application/gzip"
          class="hidden"
          @change="handleArchiveSelected"
        />
      </div>
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
        @restore="openExistingRestore"
      />
    </NCard>

    <NModal
      v-model:show="restoreModalVisible"
      preset="card"
      title="确认覆盖整个网站"
      :mask-closable="!restoring"
      :closable="!restoring"
      style="width: min(560px, calc(100vw - 32px))"
    >
      <NAlert
        type="error"
        title="这是破坏性操作"
        class="mb-4"
      >
        当前网站的数据库和上传文件都会被所选备份替换。恢复期间服务会短暂离线，请勿关闭或手动停止容器。
      </NAlert>
      <div
        v-if="selectedArchive"
        class="mb-4 rounded-md bg-neutral-100 px-3 py-2 text-sm dark:bg-neutral-800"
      >
        将上传：{{ selectedArchive.name }}
      </div>
      <p class="mb-2 text-sm">
        请输入 <code class="font-mono font-semibold">OVERWRITE</code> 继续：
      </p>
      <NInput
        v-model:value="confirmation"
        :disabled="restoring"
        placeholder="OVERWRITE"
        @keyup.enter="confirmation === 'OVERWRITE' && confirmRestore()"
      />
      <template #footer>
        <div class="flex justify-end gap-2">
          <NButton
            :disabled="restoring"
            @click="restoreModalVisible = false"
            >取消</NButton
          >
          <NButton
            type="error"
            :loading="restoring"
            :disabled="confirmation !== 'OVERWRITE'"
            @click="confirmRestore"
            >覆盖并重启恢复</NButton
          >
        </div>
      </template>
    </NModal>
  </div>
</template>
