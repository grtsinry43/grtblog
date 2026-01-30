<script setup lang="ts">
import {
  NAlert,
  NButton,
  NCard,
  NDivider,
  NForm,
  NFormItem,
  NInput,
  NSpin,
  NTag,
  useMessage,
} from 'naive-ui'
import { computed, onMounted, reactive, ref, watch } from 'vue'

import { ScrollContainer } from '@/components'
import TemplateEditor from '@/components/template-editor/TemplateEditor.vue'
import { listWebsiteInfo, updateWebsiteInfo } from '@/services/website-info'

import type { WebsiteInfoItem } from '@/services/website-info'

const message = useMessage()
const loading = ref(false)
const saving = ref(false)
const items = ref<WebsiteInfoItem[]>([])

const themeKey = 'theme_extend_info'

const valueMap = reactive<Record<string, string>>({})
const originalValues = ref<Record<string, string>>({})
const themeJsonText = ref('')
const originalThemeJson = ref('')
const themeJsonError = ref<string | null>(null)

const normalItems = computed(() => items.value.filter((item) => item.key !== themeKey))
const themeItem = computed(() => items.value.find((item) => item.key === themeKey))
const themeJsonValid = computed(() => !themeJsonError.value)

function normalizeJsonSource(value: unknown) {
  return JSON.stringify(value ?? {}, null, 2)
}

async function fetchInfo() {
  loading.value = true
  try {
    const list = await listWebsiteInfo()
    items.value = list || []
    const nextValues: Record<string, string> = {}
    items.value.forEach((item) => {
      nextValues[item.key] = item.value ?? ''
    })
    Object.keys(valueMap).forEach((key) => delete valueMap[key])
    Object.assign(valueMap, nextValues)
    originalValues.value = { ...nextValues }

    const themeSource = normalizeJsonSource(themeItem.value?.infoJson)
    themeJsonText.value = themeSource
    originalThemeJson.value = themeSource
  } catch (err) {
    message.error(err instanceof Error ? err.message : '加载站点信息失败')
  } finally {
    loading.value = false
  }
}

function collectUpdates() {
  if (!themeJsonValid.value) {
    throw new Error(themeJsonError.value || '主题扩展 JSON 格式不正确')
  }
  const updates: Array<{ key: string; value?: string; infoJson?: unknown }> = []
  normalItems.value.forEach((item) => {
    const nextValue = valueMap[item.key] ?? ''
    const prevValue = originalValues.value[item.key] ?? ''
    if (nextValue !== prevValue) {
      updates.push({ key: item.key, value: nextValue })
    }
  })

  if (themeItem.value) {
    const nextText = themeJsonText.value.trim()
    const prevText = originalThemeJson.value.trim()
    if (nextText !== prevText) {
      let parsed: unknown
      try {
        parsed = JSON.parse(nextText || '{}')
      } catch (err) {
        throw new Error(err instanceof Error ? err.message : '主题扩展 JSON 格式不正确')
      }
      updates.push({ key: themeKey, infoJson: parsed })
    }
  }

  return updates
}

function formatThemeJson() {
  const source = themeJsonText.value.trim() || '{}'
  try {
    const parsed = JSON.parse(source)
    themeJsonText.value = JSON.stringify(parsed, null, 2)
    themeJsonError.value = null
    message.success('已格式化')
  } catch (err) {
    themeJsonError.value = err instanceof Error ? err.message : 'JSON 格式不正确'
    message.error('JSON 格式不正确')
  }
}

async function handleSave() {
  if (saving.value) return
  let updates: Array<{ key: string; value?: string; infoJson?: unknown }>
  try {
    updates = collectUpdates()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '保存失败')
    return
  }
  if (updates.length === 0) {
    message.warning('没有检测到更改')
    return
  }

  saving.value = true
  try {
    await Promise.all(
      updates.map((item) =>
        updateWebsiteInfo(item.key, {
          value: item.value,
          infoJson: item.infoJson,
        }),
      ),
    )
    message.success('保存成功')
    await fetchInfo()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(fetchInfo)

watch(
  themeJsonText,
  (value) => {
    const source = value.trim()
    if (!source) {
      themeJsonError.value = null
      return
    }
    try {
      JSON.parse(source)
      themeJsonError.value = null
    } catch (err) {
      themeJsonError.value = err instanceof Error ? err.message : 'JSON 格式不正确'
    }
  },
  { immediate: true },
)
</script>

<template>
  <ScrollContainer wrapper-class="p-4">
    <NCard>
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <div class="text-base font-semibold">站点信息</div>
            <div class="text-xs text-neutral-500">用于前端渲染的基础信息与 OG 默认值</div>
          </div>
          <div class="flex items-center gap-2">
            <NButton
              size="small"
              secondary
              :loading="loading"
              @click="fetchInfo"
            >
              刷新
            </NButton>
            <NButton
              size="small"
              type="primary"
              :loading="saving"
              :disabled="!themeJsonValid"
              @click="handleSave"
            >
              保存
            </NButton>
          </div>
        </div>
      </template>

      <NSpin :show="loading">
        <NForm
          label-placement="left"
          label-width="140"
          class="space-y-2"
        >
          <template v-for="item in normalItems" :key="item.key">
            <NFormItem :label="item.name || item.key">
              <div class="w-full">
                <NInput
                  v-model:value="valueMap[item.key]"
                  :placeholder="item.key"
                />
              </div>
            </NFormItem>
          </template>
        </NForm>

        <NDivider>
          <div class="flex items-center gap-2">
            <span>主题扩展</span>
            <NTag
              size="small"
              :type="themeJsonValid ? 'success' : 'error'"
              :bordered="false"
            >
              {{ themeJsonValid ? 'JSON 有效' : 'JSON 无效' }}
            </NTag>
          </div>
        </NDivider>
        <div class="space-y-3">
          <NAlert type="info" :show-icon="false">仅支持 JSON，对应 theme_extend_info。</NAlert>
          <div class="flex justify-end">
            <NButton
              size="small"
              tertiary
              @click="formatThemeJson"
            >
              格式化
            </NButton>
          </div>
          <TemplateEditor v-model="themeJsonText" />
          <NAlert
            v-if="themeJsonError"
            type="error"
            :show-icon="false"
          >
            {{ themeJsonError }}
          </NAlert>
        </div>
      </NSpin>
    </NCard>
  </ScrollContainer>
</template>
