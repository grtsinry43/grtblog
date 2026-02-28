<script setup lang="ts">
import { NAlert, NButton, NTag, useMessage } from 'naive-ui'
import { computed, onMounted, ref, watch } from 'vue'

import TemplateEditor from '@/components/template-editor/TemplateEditor.vue'
import { listWebsiteInfo, updateWebsiteInfo } from '@/services/website-info'

const emit = defineEmits<{ 'dirty-change': [dirty: boolean] }>()

const message = useMessage()
const loading = ref(false)
const saving = ref(false)

const jsonText = ref('')
const originalJson = ref('')
const jsonError = ref<string | null>(null)

const jsonValid = computed(() => !jsonError.value)
const isDirty = computed(() => jsonText.value.trim() !== originalJson.value.trim())

watch(isDirty, (dirty) => emit('dirty-change', dirty), { immediate: true })

watch(
  jsonText,
  (value) => {
    const source = value.trim()
    if (!source) {
      jsonError.value = null
      return
    }
    try {
      JSON.parse(source)
      jsonError.value = null
    } catch (err) {
      jsonError.value = err instanceof Error ? err.message : 'JSON 格式不正确'
    }
  },
  { immediate: true },
)

async function fetchData() {
  loading.value = true
  try {
    const list = await listWebsiteInfo()
    const item = (list || []).find((i) => i.key === 'theme_extend_info')
    const source = JSON.stringify(item?.infoJson ?? {}, null, 2)
    jsonText.value = source
    originalJson.value = source
  } catch (err) {
    message.error(err instanceof Error ? err.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function formatJson() {
  const source = jsonText.value.trim() || '{}'
  try {
    const parsed = JSON.parse(source)
    jsonText.value = JSON.stringify(parsed, null, 2)
    jsonError.value = null
    message.success('已格式化')
  } catch (err) {
    jsonError.value = err instanceof Error ? err.message : 'JSON 格式不正确'
    message.error('JSON 格式不正确')
  }
}

async function handleSave() {
  if (saving.value) return
  if (!isDirty.value) {
    message.warning('没有检测到更改')
    return
  }
  if (!jsonValid.value) {
    message.error(jsonError.value || 'JSON 格式不正确')
    return
  }

  saving.value = true
  try {
    const parsed = JSON.parse(jsonText.value.trim() || '{}')
    await updateWebsiteInfo('theme_extend_info', { infoJson: parsed })
    message.success('保存成功')
    originalJson.value = jsonText.value.trim()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="flex h-full flex-col gap-2">
    <!-- Header bar -->
    <div
      class="flex shrink-0 flex-wrap items-center justify-between gap-3 rounded-lg border border-neutral-200 bg-white px-4 py-2.5 dark:border-neutral-700 dark:bg-neutral-900"
    >
      <div>
        <div class="text-sm font-semibold">主题扩展信息</div>
        <div class="text-xs text-neutral-500">
          对应 theme_extend_info，主题可读取此 JSON 进行自定义扩展
        </div>
      </div>
      <div class="flex items-center gap-2">
        <NTag v-if="isDirty" type="warning" size="small"> 未保存 </NTag>
        <NTag size="small" :type="jsonValid ? 'success' : 'error'" :bordered="false">
          {{ jsonValid ? 'JSON 有效' : 'JSON 无效' }}
        </NTag>
        <NButton size="small" tertiary @click="formatJson"> 格式化 </NButton>
        <NButton size="small" secondary :loading="loading" @click="fetchData"> 刷新 </NButton>
        <NButton
          size="small"
          type="primary"
          :loading="saving"
          :disabled="!jsonValid || !isDirty"
          @click="handleSave"
        >
          保存
        </NButton>
      </div>
    </div>

    <!-- Editor fills remaining space -->
    <TemplateEditor v-model="jsonText" class="theme-editor min-h-0 flex-1" />

    <!-- Error bar -->
    <NAlert v-if="jsonError" type="error" :show-icon="false" class="shrink-0">
      {{ jsonError }}
    </NAlert>
  </div>
</template>

<style scoped>
.theme-editor :deep(.cm-editor) {
  height: 100%;
  min-height: unset;
}
</style>
