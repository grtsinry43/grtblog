<script setup lang="ts">
import { ref } from 'vue'
import { NButton, NModal, useMessage } from 'naive-ui'

import {
  exportFederationConfigs,
  importFederationConfigs,
  listActivityPubConfigs,
  listFederationConfigs,
  updateActivityPubConfigs,
  updateFederationConfigs,
} from '@/services/sysconfig'
import type { ConfigExportData } from '@/services/sysconfig'

import ConfigPanel from '../ConfigPanel'

const emit = defineEmits<{ 'dirty-change': [dirty: boolean] }>()
const message = useMessage()

const federationDirty = ref(false)
const activityPubDirty = ref(false)

type ConfigPanelExposed = {
  fetch: () => Promise<void>
}

const federationPanelRef = ref<ConfigPanelExposed | null>(null)
const activityPubPanelRef = ref<ConfigPanelExposed | null>(null)

const exporting = ref(false)
const importing = ref(false)
const showImportConfirm = ref(false)
const pendingImportData = ref<ConfigExportData | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)

function updateDirty(source: 'federation' | 'activitypub', dirty: boolean) {
  if (source === 'federation') federationDirty.value = dirty
  else activityPubDirty.value = dirty
  emit('dirty-change', federationDirty.value || activityPubDirty.value)
}

async function handleExport() {
  exporting.value = true
  try {
    const data = await exportFederationConfigs()
    const json = JSON.stringify(data, null, 2)
    const blob = new Blob([json], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const date = new Date().toISOString().slice(0, 10)
    const a = document.createElement('a')
    a.href = url
    a.download = `federation-config-${date}.json`
    a.click()
    URL.revokeObjectURL(url)
    message.success('导出成功')
  } catch (e: any) {
    message.error(e?.message || '导出失败')
  } finally {
    exporting.value = false
  }
}

function triggerImport() {
  fileInputRef.value?.click()
}

function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  input.value = ''

  const reader = new FileReader()
  reader.onload = () => {
    try {
      const data = JSON.parse(reader.result as string) as ConfigExportData
      if (!data.version || !Array.isArray(data.configs) || data.configs.length === 0) {
        message.error('无效的配置文件格式')
        return
      }
      pendingImportData.value = data
      showImportConfirm.value = true
    } catch {
      message.error('JSON 解析失败，请检查文件格式')
    }
  }
  reader.readAsText(file)
}

async function confirmImport() {
  if (!pendingImportData.value) return
  importing.value = true
  try {
    await importFederationConfigs(pendingImportData.value)
    message.success('导入成功')
    showImportConfirm.value = false
    pendingImportData.value = null
    federationPanelRef.value?.fetch()
    activityPubPanelRef.value?.fetch()
  } catch (e: any) {
    message.error(e?.message || '导入失败')
  } finally {
    importing.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center gap-2">
      <NButton size="small" secondary :loading="exporting" @click="handleExport">
        导出配置
      </NButton>
      <NButton size="small" secondary @click="triggerImport">
        导入配置
      </NButton>
      <input
        ref="fileInputRef"
        type="file"
        accept=".json"
        class="hidden"
        @change="handleFileChange"
      />
    </div>

    <ConfigPanel
      ref="federationPanelRef"
      :list-fn="listFederationConfigs"
      :update-fn="updateFederationConfigs"
      title="Federation 联合"
      description="启用后系统会自动生成密钥，仅需填写基础信息即可"
      :on-dirty-change="(dirty: boolean) => updateDirty('federation', dirty)"
    />

    <ConfigPanel
      ref="activityPubPanelRef"
      :list-fn="listActivityPubConfigs"
      :update-fn="updateActivityPubConfigs"
      title="ActivityPub"
      description="兼容功能独立配置，启用后将使用 ActivityPub 专用密钥"
      :on-dirty-change="(dirty: boolean) => updateDirty('activitypub', dirty)"
    />

    <NModal
      v-model:show="showImportConfirm"
      preset="dialog"
      title="确认导入"
      positive-text="确认导入"
      negative-text="取消"
      :positive-button-props="{ loading: importing }"
      @positive-click="confirmImport"
    >
      <template v-if="pendingImportData">
        <p>
          即将导入 <strong>{{ pendingImportData.configs.length }}</strong> 项配置，现有的同名配置将被覆盖。
        </p>
        <p class="mt-1 text-xs text-neutral-500">
          导出时间：{{ pendingImportData.exportedAt }}
        </p>
      </template>
    </NModal>
  </div>
</template>
