<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { ChevronDown16Regular, Sparkle20Filled, Sparkle20Regular } from '@vicons/fluent'
import { NButton, NButtonGroup, NIcon, NPopover, NSwitch, NTooltip, useThemeVars } from 'naive-ui'
import { computed, onMounted, onUnmounted, watch } from 'vue'

import { useAiOnboardingModal } from '@/composables/ai/use-ai-onboarding'
import {
  getAIFeatureFlags,
  updateAIFeatureFlags,
  type UpdateAIFeatureFlagsReq,
} from '@/services/ai'

import type {
  CompletionPace,
  RealtimeAIStatus,
} from '@/composables/markdown-editor/use-realtime-ai'

export interface RealtimeAIPrefs {
  enabled: boolean
  completion: boolean
  completionPace: CompletionPace
  marginNotes: boolean
  lint: boolean
}

const STORAGE_KEY = 'grtblog:realtime-ai-prefs'

const paceOptions: { value: CompletionPace; label: string; tip: string }[] = [
  { value: 'restrained', label: '克制', tip: '停顿约 3 秒触发' },
  { value: 'manual', label: '手动', tip: '仅按 Alt+\\ 手动触发' },
  { value: 'active', label: '积极', tip: '停顿约 1.6 秒触发' },
]

defineProps<{
  status?: RealtimeAIStatus
}>()

const prefs = defineModel<RealtimeAIPrefs>({
  default: () => ({
    enabled: false,
    completion: false,
    completionPace: 'restrained',
    marginNotes: true,
    lint: true,
  }),
})

const themeVars = useThemeVars()
const queryClient = useQueryClient()
const { open: openOnboarding } = useAiOnboardingModal()

const { data: featureFlags } = useQuery({
  queryKey: ['ai-feature-flags'],
  queryFn: getAIFeatureFlags,
  staleTime: 1000 * 60 * 5,
  retry: false,
})

const { mutate: syncFeatureFlags } = useMutation({
  mutationFn: (patch: UpdateAIFeatureFlagsReq) => updateAIFeatureFlags(patch),
  onSuccess: (updated) => {
    queryClient.setQueryData(['ai-feature-flags'], updated)
  },
})

// 如果全局关闭了 AI 或关闭了编辑器实时帮写，隐藏按钮
const featureAvailable = computed(() => {
  if (!featureFlags.value) return true
  return featureFlags.value.enabled && featureFlags.value.realtimeAssist
})

function isValidPace(v: unknown): v is CompletionPace {
  return v === 'restrained' || v === 'manual' || v === 'active'
}

function loadLocalPrefs() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw) as Partial<RealtimeAIPrefs>
    prefs.value = {
      ...prefs.value,
      enabled: typeof parsed.enabled === 'boolean' ? parsed.enabled : prefs.value.enabled,
      completionPace: isValidPace(parsed.completionPace)
        ? parsed.completionPace
        : prefs.value.completionPace,
    }
  } catch {
    // 忽略无效缓存
  }
}

// 服务端数据库配置为子项开关的权威来源（幽灵续写默认关闭，眉批与校对默认开启）
watch(
  featureFlags,
  (flags) => {
    if (!flags) return
    const disabledGlobally = !flags.enabled || !flags.realtimeAssist
    prefs.value = {
      ...prefs.value,
      enabled: disabledGlobally ? false : prefs.value.enabled,
      completion: flags.ghostCompletion,
      marginNotes: flags.marginNotes,
      lint: flags.inlineLint,
    }
  },
  { immediate: true },
)

watch(
  prefs,
  (val) => {
    try {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({ enabled: val.enabled, completionPace: val.completionPace }),
      )
    } catch {
      // 忽略存储受限
    }
  },
  { deep: true },
)

function toggleEnabled() {
  const next = !prefs.value.enabled
  prefs.value = { ...prefs.value, enabled: next }
  if (next && featureFlags.value && !featureFlags.value.onboardingSeen) {
    openOnboarding(0)
  }
}

function toggleCompletion(val: boolean) {
  prefs.value = { ...prefs.value, completion: val }
  syncFeatureFlags({ ghostCompletion: val })
}

function setPace(val: CompletionPace) {
  prefs.value = { ...prefs.value, completionPace: val }
}

function toggleMarginNotes(val: boolean) {
  prefs.value = { ...prefs.value, marginNotes: val }
  syncFeatureFlags({ marginNotes: val })
}

function toggleLint(val: boolean) {
  prefs.value = { ...prefs.value, lint: val }
  syncFeatureFlags({ inlineLint: val })
}

function onKeydown(e: KeyboardEvent) {
  if (!featureAvailable.value) return
  if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key.toLowerCase() === 'a') {
    e.preventDefault()
    toggleEnabled()
  }
}

onMounted(() => {
  loadLocalPrefs()
  window.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <NButtonGroup
    v-if="featureAvailable"
    size="small"
  >
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
            <span class="ai-prefs-title">段落批注</span>
            <span class="ai-prefs-desc">写完段落后在段尾提示修改与思路</span>
          </div>
          <NSwitch
            size="small"
            :value="prefs.marginNotes"
            @update:value="toggleMarginNotes"
          />
        </div>

        <div class="ai-prefs-row">
          <div class="ai-prefs-info">
            <span class="ai-prefs-title">错字校对</span>
            <span class="ai-prefs-desc">标记错别字与常见标点错误</span>
          </div>
          <NSwitch
            size="small"
            :value="prefs.lint"
            @update:value="toggleLint"
          />
        </div>

        <div class="ai-prefs-section">
          <div class="ai-prefs-row">
            <div class="ai-prefs-info">
              <span class="ai-prefs-title">行内续写</span>
              <span class="ai-prefs-desc">停顿时给出续写建议 · Alt+\ 手动触发</span>
            </div>
            <NSwitch
              size="small"
              :value="prefs.completion"
              @update:value="toggleCompletion"
            />
          </div>

          <div
            v-if="prefs.completion"
            class="ai-pace-segment"
          >
            <button
              v-for="opt in paceOptions"
              :key="opt.value"
              type="button"
              class="ai-pace-btn"
              :class="{ 'is-active': prefs.completionPace === opt.value }"
              :title="opt.tip"
              @click="setPace(opt.value)"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>

        <div class="ai-prefs-footer">
          <button
            type="button"
            class="ai-demo-link"
            @click="openOnboarding(0)"
          >
            查看功能演示
          </button>
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
  gap: 12px;
  width: 248px;
  padding: 2px 0;
}

.ai-prefs-section {
  display: flex;
  flex-direction: column;
  gap: 7px;
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

.ai-pace-segment {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 4px;
  padding: 2px;
  border-radius: 6px;
  background-color: color-mix(in srgb, v-bind('themeVars.textColor1') 5%, transparent);
}

.ai-pace-btn {
  border: none;
  border-radius: 4px;
  padding: 3px 6px;
  font-size: 11.5px;
  line-height: 1.4;
  background: transparent;
  color: v-bind('themeVars.textColor3');
  cursor: pointer;
  transition:
    background-color 150ms ease,
    color 150ms ease;
}

.ai-pace-btn:hover {
  color: v-bind('themeVars.textColor1');
}

.ai-pace-btn.is-active {
  background-color: v-bind('themeVars.popoverColor');
  color: v-bind('themeVars.primaryColor');
  font-weight: 500;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
}

.ai-prefs-footer {
  padding-top: 8px;
  border-top: 1px solid v-bind('themeVars.dividerColor');
  display: flex;
  justify-content: flex-end;
}

.ai-demo-link {
  border: none;
  padding: 0;
  background: transparent;
  font-size: 11.5px;
  color: v-bind('themeVars.textColor3');
  cursor: pointer;
  transition: color 150ms ease;
}

.ai-demo-link:hover {
  color: v-bind('themeVars.primaryColor');
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
