<script setup lang="ts">
import { ChevronDown16Regular, Sparkle20Filled, Sparkle20Regular } from '@vicons/fluent'
import { NButton, NButtonGroup, NIcon, NPopover, NSwitch, NTooltip, useThemeVars } from 'naive-ui'
import { onMounted, onUnmounted, watch } from 'vue'

import type { RealtimeAIStatus } from '@/composables/markdown-editor/use-realtime-ai'

export interface RealtimeAIPrefs {
  enabled: boolean
  completion: boolean
  lint: boolean
}

const STORAGE_KEY = 'grtblog:realtime-ai-prefs'

defineProps<{
  status?: RealtimeAIStatus
}>()

const prefs = defineModel<RealtimeAIPrefs>({
  default: () => ({
    enabled: false,
    completion: true,
    lint: true,
  }),
})

const themeVars = useThemeVars()

function loadPrefs() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw) as Partial<RealtimeAIPrefs>
    prefs.value = {
      enabled: typeof parsed.enabled === 'boolean' ? parsed.enabled : prefs.value.enabled,
      completion: typeof parsed.completion === 'boolean' ? parsed.completion : prefs.value.completion,
      lint: typeof parsed.lint === 'boolean' ? parsed.lint : prefs.value.lint,
    }
  } catch {
    // 忽略无效缓存
  }
}

watch(
  prefs,
  (val) => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(val))
    } catch {
      // 忽略存储受限
    }
  },
  { deep: true },
)

function toggleEnabled() {
  prefs.value = { ...prefs.value, enabled: !prefs.value.enabled }
}

function toggleCompletion(val: boolean) {
  prefs.value = { ...prefs.value, completion: val }
}

function toggleLint(val: boolean) {
  prefs.value = { ...prefs.value, lint: val }
}

function onKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key.toLowerCase() === 'a') {
    e.preventDefault()
    toggleEnabled()
  }
}

onMounted(() => {
  loadPrefs()
  window.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <NButtonGroup size="small">
    <NTooltip trigger="hover">
      <template #trigger>
        <NButton
          :type="prefs.enabled ? 'primary' : 'default'"
          :secondary="prefs.enabled"
          :quaternary="!prefs.enabled"
          @click="toggleEnabled"
        >
          <template #icon>
            <span
              class="ai-icon-wrap"
              :class="{ 'is-thinking': prefs.enabled && status === 'thinking' }"
            >
              <NIcon :component="prefs.enabled ? Sparkle20Filled : Sparkle20Regular" />
              <span
                v-if="prefs.enabled"
                class="ai-status-dot"
              />
            </span>
          </template>
          实时帮写
        </NButton>
      </template>
      <span>{{ prefs.enabled ? '关闭' : '开启' }}实时 AI 帮写 · ⌘/Ctrl+Shift+A</span>
    </NTooltip>

    <NPopover
      trigger="click"
      placement="bottom-end"
    >
      <template #trigger>
        <NButton
          :type="prefs.enabled ? 'primary' : 'default'"
          :secondary="prefs.enabled"
          :quaternary="!prefs.enabled"
          style="padding: 0 6px"
          title="实时帮写子项设置"
        >
          <template #icon>
            <NIcon
              :component="ChevronDown16Regular"
              :size="14"
            />
          </template>
        </NButton>
      </template>

      <div class="ai-prefs-panel">
        <div class="ai-prefs-row">
          <div class="ai-prefs-info">
            <span class="ai-prefs-title">幽灵续写</span>
            <span class="ai-prefs-desc">停顿后在光标后浮现续写建议</span>
          </div>
          <NSwitch
            size="small"
            :value="prefs.completion"
            @update:value="toggleCompletion"
          />
        </div>

        <div class="ai-prefs-row">
          <div class="ai-prefs-info">
            <span class="ai-prefs-title">行文纠错与优化</span>
            <span class="ai-prefs-desc">标记错别字与可润色的表达</span>
          </div>
          <NSwitch
            size="small"
            :value="prefs.lint"
            @update:value="toggleLint"
          />
        </div>
      </div>
    </NPopover>
  </NButtonGroup>
</template>

<style scoped>
.ai-icon-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.ai-status-dot {
  position: absolute;
  top: -1px;
  right: -2px;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background-color: v-bind('themeVars.primaryColor');
  animation: ai-dot-breathe 2.4s ease-in-out infinite;
}

.ai-icon-wrap.is-thinking .ai-status-dot {
  animation: ai-dot-pulse 0.75s ease-in-out infinite;
}

.ai-prefs-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 224px;
  padding: 2px 0;
}

.ai-prefs-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.ai-prefs-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.ai-prefs-title {
  font-size: 13px;
  font-weight: 500;
  color: v-bind('themeVars.textColor1');
  line-height: 1.3;
}

.ai-prefs-desc {
  font-size: 11px;
  color: v-bind('themeVars.textColor3');
  line-height: 1.3;
}

@keyframes ai-dot-breathe {
  0%,
  100% {
    opacity: 0.45;
    transform: scale(0.9);
  }
  50% {
    opacity: 1;
    transform: scale(1.1);
  }
}

@keyframes ai-dot-pulse {
  0%,
  100% {
    opacity: 0.35;
    transform: scale(0.8);
  }
  50% {
    opacity: 1;
    transform: scale(1.35);
  }
}
</style>
