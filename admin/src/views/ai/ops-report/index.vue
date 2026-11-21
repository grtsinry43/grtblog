<script setup lang="ts">
import { ArrowClockwise16Regular, ArrowRight16Regular } from '@vicons/fluent'
import { NButton, NIcon, useMessage, useThemeVars } from 'naive-ui'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import {
  streamOpsReport,
  type DashboardBriefMood,
  type OpsReportActionItem,
  type OpsReportResp,
  type OpsReportTopicIdea,
} from '@/services/ai'
import { setPendingWritingIdea } from '@/views/shared/content-editor/utils/writing-idea-handoff'

defineOptions({
  name: 'AIOpsReport',
})

const router = useRouter()
const message = useMessage()
const themeVars = useThemeVars()

const days = ref<7 | 30>(30)
const enabled = ref<boolean | null>(null)
const streaming = ref(false)
const stageMessage = ref('')
const streamingSummary = ref('')
const errorMessage = ref('')
const report = ref<OpsReportResp | null>(null)

let abortController: AbortController | null = null

async function loadReport(targetDays: 7 | 30, force = false) {
  abortController?.abort()
  abortController = new AbortController()
  const signal = abortController.signal

  streaming.value = true
  errorMessage.value = ''
  if (force || report.value?.days !== targetDays) {
    streamingSummary.value = ''
    if (report.value?.days !== targetDays) {
      report.value = null
    }
    stageMessage.value = `正在汇总近 ${targetDays} 天站点趋势与内容结构…`
  }

  let receivedResult = false

  try {
    await streamOpsReport(
      { days: targetDays, force },
      {
        onDisabled: () => {
          enabled.value = false
          report.value = null
        },
        onStage: (_stage, msg) => {
          enabled.value = true
          stageMessage.value = msg
        },
        onSummary: (text) => {
          enabled.value = true
          streamingSummary.value = text
        },
        onResult: (data) => {
          enabled.value = data.enabled
          report.value = data
          streamingSummary.value = ''
          receivedResult = true
          if (force) {
            message.success('运营复盘报告已更新')
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
      errorMessage.value = '报告生成提前中断，请重试'
    }
  } catch (e: unknown) {
    if (signal.aborted) return
    if (enabled.value !== false) {
      errorMessage.value = e instanceof Error ? e.message : '加载报告失败'
    }
  } finally {
    if (!signal.aborted) {
      streaming.value = false
    }
  }
}

function switchDays(nextDays: 7 | 30) {
  if (days.value === nextDays && !errorMessage.value) return
  days.value = nextDays
  void loadReport(nextDays, false)
}

onMounted(() => {
  void loadReport(days.value, false)
})

onUnmounted(() => {
  abortController?.abort()
})

const moodMeta = computed(() => {
  if (streaming.value) {
    return { label: '分析中', dotColor: themeVars.value.textColor3 }
  }
  const mood: DashboardBriefMood = report.value?.mood ?? 'steady'
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

const displaySummary = computed(() => {
  if (streaming.value && streamingSummary.value) {
    return streamingSummary.value
  }
  return report.value?.summary ?? ''
})

const updatedTimeText = computed(() => {
  const raw = report.value?.generatedAt
  if (!raw) return ''
  const ts = Date.parse(raw)
  if (Number.isNaN(ts)) return ''
  return new Date(ts).toLocaleString('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
})

function strengthLabel(strength: 'core' | 'growing' | 'quiet') {
  if (strength === 'core') return '核心支柱'
  if (strength === 'growing') return '成长中'
  return '偏静'
}

function priorityMeta(priority: 'high' | 'medium' | 'low') {
  if (priority === 'high') {
    return { label: '优先', color: themeVars.value.warningColor }
  }
  if (priority === 'medium') {
    return { label: '建议', color: themeVars.value.primaryColor }
  }
  return { label: '可选', color: themeVars.value.textColor3 }
}

function goWriteIdea(idea: OpsReportTopicIdea) {
  setPendingWritingIdea('article', idea)
  void router.push('/articles/edit/new')
}

function goActionRoute(act: OpsReportActionItem) {
  if (!act.route) return
  if (act.route === '/articles/edit/new') {
    setPendingWritingIdea('article', { title: act.title, angle: act.detail })
  } else if (act.route === '/notes/edit/new') {
    setPendingWritingIdea('note', { title: act.title, angle: act.detail })
  }
  void router.push(act.route)
}
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-y-4">
    <!-- 顶部页眉 -->
    <div class="mt-2 flex flex-wrap items-end justify-between gap-4">
      <div class="space-y-1">
        <div class="flex items-center gap-2 text-xs text-neutral-400 dark:text-neutral-500">
          <span
            class="report-dot"
            :class="{ 'is-streaming': streaming }"
            :style="{ backgroundColor: moodMeta.dotColor }"
          />
          <span>{{ moodMeta.label }}</span>
          <template v-if="updatedTimeText && !streaming">
            <span>·</span>
            <span>更新于 {{ updatedTimeText }}</span>
          </template>
        </div>
        <h1 class="m-0 text-xl font-normal text-neutral-800 dark:text-neutral-100">
          站点运营复盘
        </h1>
      </div>

      <div class="flex items-center gap-2">
        <div class="report-segment">
          <button
            type="button"
            class="report-segment-btn"
            :class="{ 'is-active': days === 7 }"
            :disabled="streaming"
            @click="switchDays(7)"
          >
            近 7 天
          </button>
          <button
            type="button"
            class="report-segment-btn"
            :class="{ 'is-active': days === 30 }"
            :disabled="streaming"
            @click="switchDays(30)"
          >
            近 30 天
          </button>
        </div>

        <NButton
          size="small"
          quaternary
          :disabled="streaming || enabled === false"
          @click="loadReport(days, true)"
        >
          <template #icon>
            <NIcon
              :component="ArrowClockwise16Regular"
              :class="{ 'animate-spin': streaming }"
            />
          </template>
          重新分析
        </NButton>
      </div>
    </div>

    <!-- 未启用 AI 时的温和空态 -->
    <section
      v-if="enabled === false"
      class="rounded border border-naive-border bg-naive-card p-8 text-center"
    >
      <p class="m-0 text-sm text-neutral-600 dark:text-neutral-300">
        尚未启用 AI 服务或未配置可用模型
      </p>
      <p class="mt-1.5 mb-4 text-xs text-neutral-400 dark:text-neutral-500">
        请在系统设置中开启 AI 并配置摘要或改写模型后，再查看深度运营复盘。
      </p>
      <NButton
        size="small"
        secondary
        @click="router.push('/settings')"
      >
        前往系统设置
      </NButton>
    </section>

    <template v-else>
      <!-- 执行摘要卡片（流式直出） -->
      <section class="rounded border border-naive-border bg-naive-card p-6">
        <div class="mb-2 text-xs font-medium tracking-wide text-neutral-400 dark:text-neutral-500">
          执行摘要 · 近 {{ days }} 天
        </div>

        <p
          v-if="displaySummary"
          class="m-0 max-w-4xl text-[14.5px] leading-7 text-neutral-700 dark:text-neutral-200"
        >
          {{ displaySummary }}<span
            v-if="streaming"
            class="report-caret"
          />
        </p>

        <p
          v-else-if="streaming"
          class="report-stage-text m-0 text-sm leading-7"
        >
          {{ stageMessage || '正在准备复盘分析…' }}
        </p>

        <div
          v-else-if="errorMessage"
          class="flex items-center gap-3 text-sm text-neutral-500 dark:text-neutral-400"
        >
          <span>{{ errorMessage }}</span>
          <NButton
            size="tiny"
            secondary
            @click="loadReport(days, true)"
          >
            重试
          </NButton>
        </div>
      </section>

      <!-- 主体四维分析模块 -->
      <Transition name="report-reveal">
        <div
          v-if="report && !streaming"
          class="flex flex-col gap-4"
        >
          <!-- 1. 主题与内容梳理 -->
          <section
            v-if="report.topics"
            class="rounded border border-naive-border bg-naive-card p-6"
          >
            <div class="flex flex-wrap items-center justify-between gap-2">
              <h2 class="m-0 text-sm font-medium text-neutral-800 dark:text-neutral-100">
                主题与内容梳理
              </h2>
              <button
                type="button"
                class="report-nav-link"
                @click="router.push('/articles/list')"
              >
                <span>管理文章</span>
                <NIcon
                  :component="ArrowRight16Regular"
                  :size="12"
                />
              </button>
            </div>
            <p class="mt-2 mb-0 text-[13.5px] leading-relaxed text-neutral-600 dark:text-neutral-300">
              {{ report.topics.overview }}
            </p>

            <!-- 内容支柱 -->
            <div
              v-if="report.topics.pillars?.length"
              class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-4"
            >
              <div
                v-for="(pillar, idx) in report.topics.pillars"
                :key="idx"
                class="report-sub-card"
              >
                <div class="flex items-center justify-between gap-2">
                  <span class="text-xs font-medium text-neutral-800 dark:text-neutral-100">
                    {{ pillar.name }}
                  </span>
                  <span class="report-strength-tag">
                    {{ strengthLabel(pillar.strength) }}
                  </span>
                </div>
                <p class="mt-1.5 mb-0 text-xs leading-relaxed text-neutral-500 dark:text-neutral-400">
                  {{ pillar.observation }}
                </p>
              </div>
            </div>

            <!-- 后续选题方向 -->
            <div
              v-if="report.topics.ideas?.length"
              class="mt-5 border-t border-naive-border pt-4"
            >
              <div class="mb-3 text-xs font-medium text-neutral-400 dark:text-neutral-500">
                推荐后续选题方向
              </div>
              <div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
                <div
                  v-for="(idea, idx) in report.topics.ideas"
                  :key="idx"
                  class="report-idea-card flex flex-col justify-between"
                >
                  <div>
                    <div class="text-[13px] font-medium text-neutral-800 dark:text-neutral-100">
                      {{ idea.title }}
                    </div>
                    <p
                      v-if="idea.angle"
                      class="mt-1.5 mb-0 text-xs leading-relaxed text-neutral-600 dark:text-neutral-300"
                    >
                      {{ idea.angle }}
                    </p>
                    <p
                      v-if="idea.why"
                      class="mt-1 mb-0 text-[11.5px] leading-relaxed text-neutral-400 dark:text-neutral-500"
                    >
                      {{ idea.why }}
                    </p>
                  </div>
                  <div class="mt-3 flex justify-end">
                    <button
                      type="button"
                      class="report-nav-link"
                      @click="goWriteIdea(idea)"
                    >
                      <span>去写一篇</span>
                      <NIcon
                        :component="ArrowRight16Regular"
                        :size="12"
                      />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <!-- 2 & 3. 访客画像 + 评论区声音（双栏） -->
          <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <!-- 访客画像与留存 -->
            <section
              v-if="report.audience"
              class="flex flex-col justify-between rounded border border-naive-border bg-naive-card p-6"
            >
              <div>
                <div class="flex items-center justify-between gap-2">
                  <h2 class="m-0 text-sm font-medium text-neutral-800 dark:text-neutral-100">
                    访客画像与留存
                  </h2>
                  <button
                    type="button"
                    class="report-nav-link"
                    @click="router.push('/audience/visitors')"
                  >
                    <span>访客明细</span>
                    <NIcon
                      :component="ArrowRight16Regular"
                      :size="12"
                    />
                  </button>
                </div>
                <p class="mt-2 mb-0 text-[13.5px] leading-relaxed text-neutral-600 dark:text-neutral-300">
                  {{ report.audience.overview }}
                </p>

                <ul
                  v-if="report.audience.highlights?.length"
                  class="mt-3.5 mb-0 space-y-2 pl-4 text-xs leading-relaxed text-neutral-600 dark:text-neutral-300"
                >
                  <li
                    v-for="(item, idx) in report.audience.highlights"
                    :key="idx"
                  >
                    {{ item }}
                  </li>
                </ul>
              </div>

              <div
                v-if="report.audience.retentionNote"
                class="report-note-box mt-4"
              >
                <div class="text-[11px] font-medium text-neutral-400 dark:text-neutral-500">
                  转化与留存观察
                </div>
                <p class="mt-1 mb-0 text-xs leading-relaxed text-neutral-700 dark:text-neutral-200">
                  {{ report.audience.retentionNote }}
                </p>
              </div>
            </section>

            <!-- 评论区声音与互动 -->
            <section
              v-if="report.comments"
              class="flex flex-col justify-between rounded border border-naive-border bg-naive-card p-6"
            >
              <div>
                <div class="flex items-center justify-between gap-2">
                  <h2 class="m-0 text-sm font-medium text-neutral-800 dark:text-neutral-100">
                    评论区声音与互动
                  </h2>
                  <button
                    type="button"
                    class="report-nav-link"
                    @click="router.push('/comments')"
                  >
                    <span>评论管理</span>
                    <NIcon
                      :component="ArrowRight16Regular"
                      :size="12"
                    />
                  </button>
                </div>
                <p class="mt-2 mb-0 text-[13.5px] leading-relaxed text-neutral-600 dark:text-neutral-300">
                  {{ report.comments.overview }}
                </p>

                <ul
                  v-if="report.comments.themes?.length"
                  class="mt-3.5 mb-0 space-y-2 pl-4 text-xs leading-relaxed text-neutral-600 dark:text-neutral-300"
                >
                  <li
                    v-for="(theme, idx) in report.comments.themes"
                    :key="idx"
                  >
                    {{ theme }}
                  </li>
                </ul>
              </div>

              <div
                v-if="report.comments.suggestion"
                class="report-note-box mt-4"
              >
                <div class="text-[11px] font-medium text-neutral-400 dark:text-neutral-500">
                  互动建议
                </div>
                <p class="mt-1 mb-0 text-xs leading-relaxed text-neutral-700 dark:text-neutral-200">
                  {{ report.comments.suggestion }}
                </p>
              </div>
            </section>
          </div>

          <!-- 4. 整体运营行动清单 -->
          <section
            v-if="report.operations"
            class="rounded border border-naive-border bg-naive-card p-6"
          >
            <h2 class="m-0 text-sm font-medium text-neutral-800 dark:text-neutral-100">
              整体运营建议
            </h2>
            <p
              v-if="report.operations.overview"
              class="mt-2 mb-0 text-[13.5px] leading-relaxed text-neutral-600 dark:text-neutral-300"
            >
              {{ report.operations.overview }}
            </p>

            <div
              v-if="report.operations.actions?.length"
              class="mt-4 divide-y divide-naive-border border-t border-naive-border"
            >
              <div
                v-for="(act, idx) in report.operations.actions"
                :key="idx"
                class="flex flex-wrap items-center justify-between gap-3 py-3.5"
              >
                <div class="min-w-0 flex-1 space-y-1">
                  <div class="flex items-center gap-2">
                    <span
                      class="report-dot report-dot--tiny"
                      :style="{ backgroundColor: priorityMeta(act.priority).color }"
                    />
                    <span class="text-[11px] text-neutral-400 dark:text-neutral-500">
                      {{ priorityMeta(act.priority).label }}
                    </span>
                    <span class="text-[13.5px] font-medium text-neutral-800 dark:text-neutral-100">
                      {{ act.title }}
                    </span>
                  </div>
                  <p class="m-0 pl-4 text-xs leading-relaxed text-neutral-500 dark:text-neutral-400">
                    {{ act.detail }}
                  </p>
                </div>

                <button
                  v-if="act.route"
                  type="button"
                  class="report-action-btn"
                  @click="goActionRoute(act)"
                >
                  <span>前往处理</span>
                  <NIcon
                    :component="ArrowRight16Regular"
                    :size="12"
                  />
                </button>
              </div>
            </div>
          </section>
        </div>
      </Transition>
    </template>
  </ScrollContainer>
</template>

<style scoped>
.report-dot {
  width: 6px;
  height: 6px;
  border-radius: 9999px;
  flex-shrink: 0;
  opacity: 0.85;
}

.report-dot--tiny {
  width: 5px;
  height: 5px;
}

.report-dot.is-streaming {
  animation: report-dot-breathe 1.6s ease-in-out infinite;
}

.report-segment {
  display: inline-flex;
  gap: 2px;
  padding: 2px;
  border-radius: v-bind('themeVars.borderRadius');
  border: 1px solid v-bind('themeVars.borderColor');
  background-color: v-bind('themeVars.cardColor');
}

.report-segment-btn {
  border: none;
  border-radius: calc(v-bind('themeVars.borderRadius') - 1px);
  padding: 3px 10px;
  font-size: 12px;
  line-height: 1.4;
  background: transparent;
  color: v-bind('themeVars.textColor3');
  cursor: pointer;
  transition:
    color 150ms ease,
    background-color 150ms ease;
}

.report-segment-btn:hover:not(:disabled) {
  color: v-bind('themeVars.textColor1');
}

.report-segment-btn.is-active {
  background-color: color-mix(in srgb, v-bind('themeVars.textColor1') 7%, transparent);
  color: v-bind('themeVars.textColor1');
  font-weight: 500;
}

.report-segment-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.report-stage-text {
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
  animation: report-shimmer 2.2s linear infinite;
}

.report-caret {
  display: inline-block;
  width: 1.5px;
  height: 0.92em;
  margin-left: 3px;
  vertical-align: -2px;
  border-radius: 1px;
  background-color: v-bind('themeVars.textColor3');
  animation: report-caret-blink 0.95s ease-in-out infinite;
}

.report-sub-card {
  padding: 12px 14px;
  border-radius: v-bind('themeVars.borderRadius');
  background-color: color-mix(in srgb, v-bind('themeVars.textColor1') 3%, transparent);
}

.report-idea-card {
  padding: 14px 16px;
  border-radius: v-bind('themeVars.borderRadius');
  border: 1px solid v-bind('themeVars.borderColor');
}

.report-strength-tag {
  font-size: 10.5px;
  color: v-bind('themeVars.textColor3');
  flex-shrink: 0;
}

.report-note-box {
  padding: 10px 12px;
  border-radius: v-bind('themeVars.borderRadius');
  background-color: color-mix(in srgb, v-bind('themeVars.textColor1') 3%, transparent);
}

.report-nav-link {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 0;
  border: none;
  background: transparent;
  font-size: 12px;
  color: v-bind('themeVars.textColor3');
  cursor: pointer;
  transition: color 150ms ease;
}

.report-nav-link:hover {
  color: v-bind('themeVars.textColor1');
}

.report-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border: none;
  border-radius: v-bind('themeVars.borderRadius');
  background-color: color-mix(in srgb, v-bind('themeVars.textColor1') 4%, transparent);
  color: v-bind('themeVars.textColor2');
  font-size: 12px;
  cursor: pointer;
  flex-shrink: 0;
  transition:
    color 150ms ease,
    background-color 150ms ease;
}

.report-action-btn:hover {
  color: v-bind('themeVars.textColor1');
  background-color: color-mix(in srgb, v-bind('themeVars.textColor1') 8%, transparent);
}

@keyframes report-shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

@keyframes report-dot-breathe {
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

@keyframes report-caret-blink {
  0%,
  100% {
    opacity: 0.85;
  }
  50% {
    opacity: 0.15;
  }
}

.report-reveal-enter-active {
  transition:
    opacity 280ms cubic-bezier(0.16, 1, 0.3, 1),
    transform 280ms cubic-bezier(0.16, 1, 0.3, 1);
}

.report-reveal-enter-from {
  opacity: 0;
  transform: translateY(6px);
}
</style>
