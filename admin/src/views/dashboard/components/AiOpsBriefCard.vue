<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import {
  ArrowClockwise16Regular,
  ArrowRight16Regular,
  ChevronDown16Regular,
  EyeOff16Regular,
} from '@vicons/fluent'
import { NButton, NIcon, NTooltip, useMessage, useThemeVars } from 'naive-ui'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import {
  getAIFeatureFlags,
  streamDashboardBrief,
  type DashboardBriefAction,
  type DashboardBriefMood,
  type DashboardBriefResp,
} from '@/services/ai'
import { setPendingWritingIdea } from '@/views/shared/content-editor/utils/writing-idea-handoff'

const HIDDEN_STORAGE_KEY = 'grtblog:dashboard-ai-brief-hidden'

const router = useRouter()
const message = useMessage()
const themeVars = useThemeVars()

const expanded = ref(false)
const hidden = ref(false)
const enabled = ref<boolean | null>(null)
const streaming = ref(false)
const stageMessage = ref('')
const streamingHeadline = ref('')
const errorMessage = ref('')
const brief = ref<DashboardBriefResp | null>(null)

const { data: featureFlags } = useQuery({
  queryKey: ['ai-feature-flags'],
  queryFn: getAIFeatureFlags,
  staleTime: 1000 * 60 * 5,
  retry: false,
})

let abortController: AbortController | null = null
let hasTriggeredInitialLoad = false

async function loadBrief(force = false) {
  abortController?.abort()
  abortController = new AbortController()
  const signal = abortController.signal

  streaming.value = true
  errorMessage.value = ''
  if (force) {
    streamingHeadline.value = ''
    stageMessage.value = '正在重新整理近期数据…'
  } else {
    stageMessage.value = '正在整理近期数据…'
  }

  let receivedResult = false

  try {
    await streamDashboardBrief(
      force,
      {
        onDisabled: () => {
          enabled.value = false
          brief.value = null
        },
        onStage: (_stage, msg) => {
          enabled.value = true
          stageMessage.value = msg
        },
        onHeadline: (text) => {
          enabled.value = true
          streamingHeadline.value = text
        },
        onResult: (data) => {
          enabled.value = data.enabled
          brief.value = data
          streamingHeadline.value = ''
          receivedResult = true
          if (force) {
            message.success('站点简报已更新')
          }
        },
        onError: (err) => {
          enabled.value = true
          errorMessage.value = err
        },
      },
      signal,
    )
    if (!signal.aborted && enabled.value === true && !receivedResult && !errorMessage.value) {
      errorMessage.value = '简报生成提前中断'
    }
  } catch (e: unknown) {
    if (signal.aborted) return
    if (enabled.value !== false) {
      errorMessage.value = e instanceof Error ? e.message : '加载简报失败'
    }
  } finally {
    if (!signal.aborted) {
      streaming.value = false
    }
  }
}

onMounted(() => {
  try {
    hidden.value = localStorage.getItem(HIDDEN_STORAGE_KEY) === '1'
  } catch {
    hidden.value = false
  }
})

// 等用户完成首次功能引导确认后再按其选择决定是否拉取简报，避免未确认前偷跑
watch(
  featureFlags,
  (flags) => {
    if (!flags) return
    if (!flags.enabled || !flags.dashboardBrief) {
      abortController?.abort()
      enabled.value = false
      brief.value = null
      streaming.value = false
      hasTriggeredInitialLoad = false
      return
    }
    if (!flags.onboardingSeen) {
      return
    }
    if (!hasTriggeredInitialLoad) {
      hasTriggeredInitialLoad = true
      void loadBrief(false)
    }
  },
  { immediate: true },
)

onUnmounted(() => {
  abortController?.abort()
})

watch(hidden, (val) => {
  try {
    localStorage.setItem(HIDDEN_STORAGE_KEY, val ? '1' : '0')
  } catch {
    // 忽略存储受限
  }
})

const showCard = computed(() => {
  if (hidden.value) return false
  if (enabled.value === false) return false
  return enabled.value === true || !!brief.value?.headline
})

const moodMeta = computed(() => {
  if (streaming.value) {
    return { label: '更新中', dotColor: themeVars.value.textColor3 }
  }
  const mood: DashboardBriefMood = brief.value?.mood ?? 'steady'
  switch (mood) {
    case 'growing':
      return { label: '上升期', dotColor: themeVars.value.successColor }
    case 'attention':
      return { label: '待关注', dotColor: themeVars.value.warningColor }
    case 'quiet':
      return { label: '平稳偏静', dotColor: themeVars.value.textColor3 }
    default:
      return { label: '运行平稳', dotColor: themeVars.value.primaryColor }
  }
})

const displayHeadline = computed(() => {
  if (streaming.value && streamingHeadline.value) {
    return streamingHeadline.value
  }
  return brief.value?.headline ?? ''
})

const updatedTimeText = computed(() => {
  const raw = brief.value?.generatedAt
  if (!raw) return ''
  const ts = Date.parse(raw)
  if (Number.isNaN(ts)) return ''
  const diffMin = Math.max(0, Math.floor((Date.now() - ts) / 60000))
  if (diffMin < 1) return '刚刚'
  if (diffMin < 60) return `${diffMin} 分钟前`
  const diffHour = Math.floor(diffMin / 60)
  if (diffHour < 24) return `${diffHour} 小时前`
  return `${Math.floor(diffHour / 24)} 天前`
})

function onActionClick(action: DashboardBriefAction) {
  if (action.route) {
    if (action.route === '/articles/edit/new') {
      setPendingWritingIdea('article', { title: action.title, angle: action.reason })
    } else if (action.route === '/notes/edit/new') {
      setPendingWritingIdea('note', { title: action.title, angle: action.reason })
    }
    void router.push(action.route)
  } else {
    expanded.value = true
  }
}
</script>

<template>
  <!-- 收起态：极淡文字恢复入口 -->
  <div
    v-if="hidden && enabled"
    class="mb-1 flex items-center justify-end"
  >
    <button
      type="button"
      class="brief-restore-btn"
      @click="hidden = false"
    >
      <span
        class="brief-dot brief-dot--tiny"
        :style="{ backgroundColor: moodMeta.dotColor }"
      />
      <span>显示站点简报</span>
    </button>
  </div>

  <!-- 主简报条：与仪表盘卡片同源，无侧边描边、无冗余图标 -->
  <Transition name="brief-fade">
    <section
      v-if="showCard"
      class="brief-card mb-2 rounded border border-naive-border bg-naive-card transition-[background-color,border-color]"
    >
      <div class="px-5 py-3.5 max-sm:px-4">
        <!-- 顶行：极简状态点 + 摘要 + 低调操作 -->
        <div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
          <div class="flex min-w-0 flex-1 items-baseline gap-2.5">
            <span class="brief-meta-group">
              <span
                class="brief-dot"
                :class="{ 'is-streaming': streaming }"
                :style="{ backgroundColor: moodMeta.dotColor }"
              />
              <span class="brief-label">站点简报</span>
              <span class="brief-sep">·</span>
              <span class="brief-mood">{{ moodMeta.label }}</span>
            </span>

            <!-- 1. 流式/完成摘要 -->
            <p
              v-if="displayHeadline"
              class="m-0 min-w-0 flex-1 text-[13.5px] leading-relaxed text-neutral-700 dark:text-neutral-200"
            >
              {{ displayHeadline }}<span
                v-if="streaming"
                class="brief-caret"
              />
            </p>

            <!-- 2. 静默阶段提示 -->
            <p
              v-else-if="streaming"
              class="brief-stage-text m-0 min-w-0 flex-1 text-[13px] leading-relaxed"
            >
              {{ stageMessage || '正在整理近期数据…' }}
            </p>

            <!-- 3. 异常温和降级 -->
            <p
              v-else-if="errorMessage"
              class="m-0 flex items-center gap-2 text-xs text-neutral-400 dark:text-neutral-500"
            >
              <span>简报暂时未能生成</span>
              <button
                type="button"
                class="brief-retry-link"
                @click="loadBrief(true)"
              >
                重试
              </button>
            </p>
          </div>

          <div class="brief-tools flex shrink-0 items-center gap-0.5">
            <span
              v-if="updatedTimeText && !streaming"
              class="mr-1.5 hidden text-[11px] text-neutral-400 sm:inline dark:text-neutral-500"
            >
              {{ updatedTimeText }}
            </span>

            <NTooltip
              v-if="!streaming && brief?.headline"
              trigger="hover"
            >
              <template #trigger>
                <NButton
                  quaternary
                  size="tiny"
                  @click="router.push('/ai/ops-report')"
                >
                  <span class="text-xs">完整报告</span>
                  <NIcon
                    :component="ArrowRight16Regular"
                    :size="12"
                    class="ml-0.5"
                  />
                </NButton>
              </template>
              进入独立页面查看主题梳理、访客画像与评论分析
            </NTooltip>

            <NTooltip
              v-if="!streaming && (brief?.highlights?.length ?? 0) > 0"
              trigger="hover"
            >
              <template #trigger>
                <NButton
                  quaternary
                  size="tiny"
                  @click="expanded = !expanded"
                >
                  <template #icon>
                    <NIcon
                      :component="ChevronDown16Regular"
                      :size="13"
                      class="brief-chevron"
                      :class="{ 'is-open': expanded }"
                    />
                  </template>
                  <span class="text-xs">{{ expanded ? '收起' : '详情' }}</span>
                </NButton>
              </template>
              {{ expanded ? '收起数据洞察' : '展开近期观察与建议' }}
            </NTooltip>

            <NTooltip trigger="hover">
              <template #trigger>
                <NButton
                  quaternary
                  circle
                  size="tiny"
                  :disabled="streaming"
                  @click="loadBrief(true)"
                >
                  <template #icon>
                    <NIcon
                      :component="ArrowClockwise16Regular"
                      :size="13"
                      :class="{ 'animate-spin': streaming }"
                    />
                  </template>
                </NButton>
              </template>
              刷新简报
            </NTooltip>

            <NTooltip trigger="hover">
              <template #trigger>
                <NButton
                  quaternary
                  circle
                  size="tiny"
                  @click="hidden = true"
                >
                  <template #icon>
                    <NIcon
                      :component="EyeOff16Regular"
                      :size="13"
                    />
                  </template>
                </NButton>
              </template>
              暂时隐藏
            </NTooltip>
          </div>
        </div>

        <!-- 展开态：近期观察三栏（平滑撑开，行动胶囊在其下方自然下移无跳动） -->
        <div
          v-if="brief && brief.highlights && brief.highlights.length > 0"
          class="brief-expand"
          :class="{ 'is-open': expanded && !streaming }"
        >
          <div class="brief-expand-inner">
            <div class="mt-3.5 border-t border-naive-border pt-3.5">
              <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
                <div
                  v-for="(item, idx) in brief.highlights"
                  :key="idx"
                  class="brief-highlight-item"
                >
                  <div class="text-[11px] font-medium tracking-wide text-neutral-400 dark:text-neutral-500">
                    {{ item.label }}
                  </div>
                  <div class="mt-1 text-xs leading-relaxed text-neutral-600 dark:text-neutral-300">
                    {{ item.detail }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 单一持久化行动胶囊：展开时原地无缝变形展示理由，零高度抖动 -->
        <Transition name="brief-fade">
          <div
            v-if="!streaming && brief?.actions && brief.actions.length > 0"
            class="brief-actions-row flex flex-wrap items-center gap-1.5"
            :class="{ 'is-expanded': expanded }"
          >
            <button
              v-for="(act, idx) in brief.actions"
              :key="idx"
              type="button"
              class="brief-action-chip"
              :class="{ 'is-expanded': expanded }"
              :title="expanded ? undefined : act.reason"
              @click="onActionClick(act)"
            >
              <span class="brief-action-title">{{ act.title }}</span>
              <span
                v-if="act.reason"
                class="brief-action-reason-wrap"
                :class="{ 'is-open': expanded }"
              >
                <span class="brief-action-reason text-neutral-400 dark:text-neutral-500">
                  &nbsp;· {{ act.reason }}
                </span>
              </span>
              <NIcon
                v-if="act.route"
                :component="ArrowRight16Regular"
                :size="11"
                class="brief-action-arrow"
              />
            </button>
          </div>
        </Transition>
      </div>
    </section>
  </Transition>
</template>

<style scoped>
.brief-meta-group {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
  font-size: 11.5px;
  line-height: 1;
  color: v-bind('themeVars.textColor3');
  user-select: none;
}

.brief-dot {
  width: 6px;
  height: 6px;
  border-radius: 9999px;
  flex-shrink: 0;
  opacity: 0.8;
}

.brief-dot--tiny {
  width: 5px;
  height: 5px;
}

.brief-dot.is-streaming {
  animation: brief-dot-breathe 1.6s ease-in-out infinite;
}

.brief-label {
  font-weight: 500;
  color: v-bind('themeVars.textColor2');
}

.brief-sep {
  opacity: 0.45;
}

.brief-mood {
  color: v-bind('themeVars.textColor3');
}

.brief-stage-text {
  color: v-bind('themeVars.textColor3');
  background: linear-gradient(
    90deg,
    v-bind('themeVars.textColor3') 0%,
    v-bind('themeVars.textColor1') 50%,
    v-bind('themeVars.textColor3') 100%
  );
  background-size: 200% 100%;
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  animation: brief-shimmer 2.2s linear infinite;
}

.brief-caret {
  display: inline-block;
  width: 1.5px;
  height: 0.92em;
  margin-left: 3px;
  vertical-align: -2px;
  border-radius: 1px;
  background-color: v-bind('themeVars.textColor3');
  animation: brief-caret-blink 0.95s ease-in-out infinite;
}

.brief-retry-link {
  border: none;
  padding: 0;
  background: transparent;
  color: v-bind('themeVars.textColor2');
  text-decoration: underline;
  text-underline-offset: 2px;
  font-size: 12px;
  cursor: pointer;
}

.brief-retry-link:hover {
  color: v-bind('themeVars.textColor1');
}

.brief-tools {
  opacity: 0.72;
  transition: opacity 180ms ease;
}

.brief-card:hover .brief-tools {
  opacity: 1;
}

.brief-chevron {
  transition: transform 220ms cubic-bezier(0.16, 1, 0.3, 1);
}

.brief-chevron.is-open {
  transform: rotate(180deg);
}

.brief-actions-row {
  margin-top: 8px;
  transition: margin-top 240ms cubic-bezier(0.16, 1, 0.3, 1);
}

.brief-actions-row.is-expanded {
  margin-top: 12px;
}

.brief-action-chip {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  padding: 2px 9px;
  border-radius: v-bind('themeVars.borderRadius');
  border: none;
  background-color: color-mix(in srgb, v-bind('themeVars.textColor1') 4%, transparent);
  color: v-bind('themeVars.textColor2');
  font-size: 11.5px;
  line-height: 1.6;
  cursor: pointer;
  transition:
    padding 220ms cubic-bezier(0.16, 1, 0.3, 1),
    color 150ms ease,
    background-color 150ms ease;
}

.brief-action-chip.is-expanded {
  padding: 4px 10px;
}

.brief-action-chip:hover {
  color: v-bind('themeVars.textColor1');
  background-color: color-mix(in srgb, v-bind('themeVars.textColor1') 8%, transparent);
}

.brief-action-title {
  flex-shrink: 0;
  font-weight: 500;
}

.brief-action-reason-wrap {
  display: inline-grid;
  grid-template-columns: 0fr;
  opacity: 0;
  transition:
    grid-template-columns 240ms cubic-bezier(0.16, 1, 0.3, 1),
    opacity 180ms ease;
}

.brief-action-reason-wrap.is-open {
  grid-template-columns: 1fr;
  opacity: 1;
}

.brief-action-reason {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.brief-action-arrow {
  flex-shrink: 0;
  margin-left: 4px;
  opacity: 0.45;
  transition:
    transform 160ms cubic-bezier(0.16, 1, 0.3, 1),
    opacity 160ms ease;
}

.brief-action-chip:hover .brief-action-arrow {
  opacity: 0.9;
  transform: translateX(2px);
}

.brief-highlight-item {
  padding: 2px 0;
}

.brief-expand {
  display: grid;
  grid-template-rows: 0fr;
  opacity: 0;
  transition:
    grid-template-rows 240ms cubic-bezier(0.16, 1, 0.3, 1),
    opacity 200ms ease;
}

.brief-expand.is-open {
  grid-template-rows: 1fr;
  opacity: 1;
}

.brief-expand-inner {
  overflow: hidden;
}

.brief-restore-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 8px;
  border: none;
  border-radius: v-bind('themeVars.borderRadius');
  background: transparent;
  color: v-bind('themeVars.textColor3');
  font-size: 11.5px;
  cursor: pointer;
  transition:
    color 150ms ease,
    background-color 150ms ease;
}

.brief-restore-btn:hover {
  color: v-bind('themeVars.textColor1');
  background-color: color-mix(in srgb, v-bind('themeVars.textColor1') 4%, transparent);
}

@keyframes brief-shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

@keyframes brief-dot-breathe {
  0%,
  100% {
    transform: scale(0.85);
    opacity: 0.35;
  }
  50% {
    transform: scale(1.15);
    opacity: 0.9;
  }
}

@keyframes brief-caret-blink {
  0%,
  100% {
    opacity: 0.85;
  }
  50% {
    opacity: 0.15;
  }
}

.brief-fade-enter-active,
.brief-fade-leave-active {
  transition:
    opacity 200ms cubic-bezier(0.16, 1, 0.3, 1),
    transform 200ms cubic-bezier(0.16, 1, 0.3, 1);
}

.brief-fade-enter-from,
.brief-fade-leave-to {
  opacity: 0;
  transform: translateY(-3px);
}
</style>
