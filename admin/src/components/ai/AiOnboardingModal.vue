<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { ArrowRight16Regular, Checkmark16Regular, Dismiss16Regular } from '@vicons/fluent'
import { NButton, NIcon, NModal, NSwitch, useMessage, useThemeVars } from 'naive-ui'
import { computed, onUnmounted, reactive, ref, watch } from 'vue'

import { useAiOnboardingModal } from '@/composables/ai/use-ai-onboarding'
import {
  getAIFeatureFlags,
  updateAIFeatureFlags,
  type AIFeatureFlags,
} from '@/services/ai'
import { getSystemStatus } from '@/services/system'

const themeVars = useThemeVars()
const message = useMessage()
const queryClient = useQueryClient()
const { showModal, initialStep, close } = useAiOnboardingModal()

// step 0 = 作者前言页, step 1~4 = 四幕功能演示
const step = ref(0)
const autoPrompted = ref(false)

const draft = reactive({
  marginNotes: true,
  inlineLint: true,
  ghostCompletion: false,
  dashboardBrief: true,
  opsReport: true,
})

const { data: flags } = useQuery({
  queryKey: ['ai-feature-flags'],
  queryFn: getAIFeatureFlags,
  staleTime: 1000 * 60 * 5,
  retry: false,
})

const { data: systemStatus } = useQuery({
  queryKey: ['system-status-version'],
  queryFn: getSystemStatus,
  enabled: computed(() => showModal.value),
  staleTime: 1000 * 60 * 10,
  retry: false,
})

/** 动态读取当前后端版本号；非 semver（如 dev / commit hash）或加载前优雅降级 */
const welcomeLine = computed(() => {
  const raw = systemStatus.value?.app?.version?.trim()
  if (raw && /^v?\d+\.\d+(?:\.\d+)?(?:-[\w.]+)?$/i.test(raw)) {
    const normalized = raw.startsWith('v') || raw.startsWith('V') ? raw : `v${raw}`
    return `欢迎来到 grtblog ${normalized}`
  }
  return '欢迎来到新版 grtblog'
})

function syncDraftFromFlags(f?: AIFeatureFlags | null) {
  if (!f) return
  draft.marginNotes = f.marginNotes
  draft.inlineLint = f.inlineLint
  draft.ghostCompletion = f.ghostCompletion
  draft.dashboardBrief = f.dashboardBrief
  draft.opsReport = f.opsReport
}

watch(
  flags,
  (f) => {
    if (!f) return
    syncDraftFromFlags(f)
    if (f.enabled && !f.onboardingSeen && !autoPrompted.value) {
      autoPrompted.value = true
      step.value = 0
      showModal.value = true
    }
  },
  { immediate: true },
)

watch(showModal, (open) => {
  if (open) {
    step.value = initialStep.value
    syncDraftFromFlags(flags.value)
    startDemoLoop()
  } else {
    stopDemoLoop()
  }
})

watch(step, () => {
  if (showModal.value) {
    startDemoLoop()
  }
})

// ── 拟真 UI 微动画编排 ──
const demoPhase = ref(0)
const ghostTextFull = '但顺手的环境能少很多打断。'
const ghostTypedCount = ref(0)
let demoTimers: ReturnType<typeof setTimeout>[] = []

function clearDemoTimers() {
  demoTimers.forEach((t) => clearTimeout(t))
  demoTimers = []
}

function schedule(fn: () => void, ms: number) {
  demoTimers.push(setTimeout(fn, ms))
}

function stopDemoLoop() {
  clearDemoTimers()
  demoPhase.value = 0
  ghostTypedCount.value = 0
}

function startDemoLoop() {
  clearDemoTimers()
  demoPhase.value = 0
  ghostTypedCount.value = 0

  if (step.value === 1) {
    schedule(() => (demoPhase.value = 1), 650)
    schedule(() => (demoPhase.value = 2), 1450)
    schedule(() => (demoPhase.value = 3), 2100)
    schedule(() => startDemoLoop(), 6400)
  } else if (step.value === 2) {
    schedule(() => (demoPhase.value = 1), 650)
    schedule(() => (demoPhase.value = 2), 1500)
    schedule(() => (demoPhase.value = 3), 3400)
    schedule(() => startDemoLoop(), 6000)
  } else if (step.value === 3) {
    schedule(() => {
      demoPhase.value = 1
      let idx = 0
      const typeNext = () => {
        idx++
        ghostTypedCount.value = idx
        if (idx < ghostTextFull.length) {
          schedule(typeNext, 78)
        } else {
          schedule(() => (demoPhase.value = 2), 280)
          schedule(() => (demoPhase.value = 3), 1850)
          schedule(() => startDemoLoop(), 4400)
        }
      }
      schedule(typeNext, 140)
    }, 650)
  } else if (step.value === 4) {
    schedule(() => (demoPhase.value = 1), 950)
    schedule(() => (demoPhase.value = 2), 2550)
    schedule(() => startDemoLoop(), 6600)
  }
}

onUnmounted(() => {
  stopDemoLoop()
})

const featureSteps = [
  {
    eyebrow: '写作辅助 · 01 / 04',
    title: '段落批注',
    desc: '写完一段后，在段尾标记可精简的表达、结构问题或补充思路。点击圆点查看，不修改正文。',
  },
  {
    eyebrow: '写作辅助 · 02 / 04',
    title: '错字校对',
    desc: '自动标记错别字、「的 / 地 / 得」误用与常见标点错误，点击即可替换，不干涉个人文风。',
  },
  {
    eyebrow: '写作辅助 · 03 / 04',
    title: '行内续写',
    desc: '停顿时在光标后给出续写建议，按 Tab 采纳，Ctrl/Cmd + → 逐句采纳。默认关闭，可随时按 Alt+\\ 手动触发。',
  },
  {
    eyebrow: '站点运营 · 04 / 04',
    title: '站点简报与复盘',
    desc: '在仪表盘顶部汇总近期数据变化；独立复盘页提供主题、访客与评论分析，选题建议可直接带入编辑器生成大纲。',
  },
]

const totalSteps = 5
const currentFeatureMeta = computed(() => featureSteps[step.value - 1] ?? featureSteps[0]!)
const isCoverStep = computed(() => step.value === 0)
const isLastStep = computed(() => step.value === totalSteps - 1)

// NModal 会 Teleport 到 body，显式在弹窗根容器挂载主题 CSS 变量
const shellStyle = computed(() => ({
  '--ob-card': themeVars.value.cardColor,
  '--ob-popover': themeVars.value.popoverColor,
  '--ob-border': themeVars.value.borderColor,
  '--ob-text-1': themeVars.value.textColor1,
  '--ob-text-2': themeVars.value.textColor2,
  '--ob-text-3': themeVars.value.textColor3,
  '--ob-primary': themeVars.value.primaryColor,
  '--ob-warning': themeVars.value.warningColor,
  '--ob-success': themeVars.value.successColor,
  '--ob-shadow-2': themeVars.value.boxShadow2,
  '--ob-shadow-3': themeVars.value.boxShadow3,
  backgroundColor: themeVars.value.cardColor,
  borderColor: themeVars.value.borderColor,
  boxShadow: themeVars.value.boxShadow3,
}))

const { mutate: savePreferences, isPending: saving } = useMutation({
  mutationFn: (payload: { dismissOnly?: boolean }) =>
    updateAIFeatureFlags(
      payload.dismissOnly
        ? { onboardingSeen: true }
        : {
            marginNotes: draft.marginNotes,
            inlineLint: draft.inlineLint,
            ghostCompletion: draft.ghostCompletion,
            dashboardBrief: draft.dashboardBrief,
            opsReport: draft.opsReport,
            onboardingSeen: true,
          },
    ),
  onSuccess: (updated, vars) => {
    queryClient.setQueryData(['ai-feature-flags'], updated)
    queryClient.invalidateQueries({ queryKey: ['dashboard-ai-brief'] })
    queryClient.invalidateQueries({ queryKey: ['ops-report-cached-ideas'] })
    close()
    if (!vars.dismissOnly) {
      message.success('AI 功能偏好已保存')
    }
  },
  onError: () => {
    close()
  },
})

/** 仅临时关闭，不写入已读标记，防止手滑误关 */
function handleTempClose() {
  close()
}

/** 明确确认不再自动弹出 */
function handleNeverShowAgain() {
  savePreferences({ dismissOnly: true })
}

function handleNext() {
  if (!isLastStep.value) {
    step.value++
  } else {
    savePreferences({ dismissOnly: false })
  }
}
</script>

<template>
  <NModal
    v-model:show="showModal"
    :mask-closable="true"
    @after-leave="stopDemoLoop"
  >
    <div
      class="onboard-shell"
      :style="shellStyle"
    >
      <!-- 右上角仅临时关闭，不标记“不再展示” -->
      <button
        type="button"
        class="onboard-close"
        title="暂且关闭（下次仍会提醒）"
        @click="handleTempClose"
      >
        <NIcon
          :component="Dismiss16Regular"
          :size="15"
        />
      </button>

      <!-- ── Step 0：前置作者前言页 ── -->
      <template v-if="isCoverStep">
        <div class="cover-wrap">
          <div class="cover-eyebrow">AI is now available in grtblog Admin</div>
          <h2 class="cover-title">
            {{ welcomeLine }}<br />
            现在你可以在 Grtblog Admin 中体验 AI 辅助，<br />
            不过我还是希望它克制一点
          </h2>

          <div class="cover-letter">
            <p>
              对我而言，还是希望这套博客系统能够跟上 AI 发展的脚步。但写作始终是一件很私人、需要心流、不该被轻易打断的事，所以这里的 AI 做得比较克制。
            </p>
            <p>
              这些能力都可以由你按需开关。即便如此，我依然不太建议打开「行内续写」——它很容易像代码补全一样打断思路；不过如果你确实需要，我们也准备了三档节奏可供切换。
            </p>
            <p>
              希望 AI 在这里只是你的辅助：帮你处理文字里那些细致的问题，把站点运营得更好——不会打扰你的写作状态，更不是代替你写作。
            </p>
          </div>

          <div class="cover-sign">—— grtsinry43</div>

          <div class="onboard-footer">
            <NButton
              size="small"
              quaternary
              :disabled="saving"
              @click="handleNeverShowAgain"
            >
              不再自动展示
            </NButton>

            <NButton
              size="small"
              type="primary"
              @click="step = 1"
            >
              <span>了解这些功能</span>
              <template #icon>
                <NIcon :component="ArrowRight16Regular" />
              </template>
            </NButton>
          </div>
        </div>
      </template>

      <!-- ── Step 1~4：拟真迷你界面演示 + 真实开关 ── -->
      <template v-else>
        <div class="onboard-stage">
          <div class="mock-window">
            <div class="mock-chrome">
              <div class="mock-traffic">
                <span />
                <span />
                <span />
              </div>
              <div class="mock-crumb">
                {{ step < 4 ? '文章编辑 · 随笔草稿' : '仪表板 · 站点总览' }}
              </div>
              <div class="mock-chrome-right">
                <span
                  v-if="step < 4"
                  class="mock-status-pill"
                >
                  <span
                    class="mock-status-dot"
                    :class="{ 'is-pulse': step === 3 && demoPhase === 0 }"
                  />
                  <span>{{ step === 3 && demoPhase === 0 ? '思考中' : '实时帮写' }}</span>
                </span>
                <span
                  v-else
                  class="mock-status-pill"
                >
                  <span>近 7 天</span>
                </span>
              </div>
            </div>

            <div class="mock-viewport">
              <Transition
                name="stage-fade"
                mode="out-in"
              >
                <!-- 第 1 幕：段落批注 -->
                <div
                  v-if="step === 1"
                  key="step-1"
                  class="mock-editor"
                >
                  <div class="mock-editor-title">Go 并发排查的一点笔记</div>
                  <div class="mock-editor-body">
                    <div class="mock-gutter">
                      <span>1</span>
                      <span>2</span>
                      <span>3</span>
                    </div>
                    <div class="mock-lines">
                      <p class="mock-text">
                        排查完这次内存泄漏后，其实一开始只是怀疑缓存没设过期时间，其实真正的问题出在后台协程没有正常退出。<span
                          class="mock-margin-dot"
                          :class="{ 'is-visible': demoPhase >= 1, 'is-active': demoPhase >= 2 }"
                        />
                      </p>
                      <p class="mock-text mock-text--muted">
                        补上 context 取消信号之后，内存曲线就恢复平稳了。
                      </p>
                    </div>
                  </div>

                  <span
                    class="mock-pointer mock-pointer--margin"
                    :class="{ 'is-hover': demoPhase >= 2 }"
                  >
                    <svg
                      viewBox="0 0 16 16"
                      width="14"
                      height="14"
                      fill="currentColor"
                    >
                      <path d="M3 1.5v11.2l3.1-2.8 1.9 4.1 1.7-.8-1.9-4.1h4.1L3 1.5z" />
                    </svg>
                  </span>

                  <div
                    class="mock-float-card mock-float-card--margin"
                    :class="{ 'is-visible': demoPhase >= 3 }"
                  >
                    <div class="mock-float-head">
                      <span class="mock-badge mock-badge--amber">
                        <span class="mock-dot mock-dot--amber" />
                        表达推敲
                      </span>
                      <span class="mock-float-action">基于此改写</span>
                    </div>
                    <div class="mock-quote">“其实一开始…其实真正的问题”</div>
                    <p class="mock-float-desc">
                      这句连用了两次“其实”，删去后一处会更简洁。
                    </p>
                  </div>
                </div>

                <!-- 第 2 幕：错字校对 -->
                <div
                  v-else-if="step === 2"
                  key="step-2"
                  class="mock-editor"
                >
                  <div class="mock-editor-title">迁移到 Vite 8 的踩坑记录</div>
                  <div class="mock-editor-body">
                    <div class="mock-gutter">
                      <span>1</span>
                      <span>2</span>
                      <span>3</span>
                    </div>
                    <div class="mock-lines">
                      <p class="mock-text">
                        把构建配置拆清楚之后，整个流程<span
                          class="mock-lint-word"
                          :class="{
                            'has-wave': demoPhase === 1 || demoPhase === 2,
                            'is-fixed': demoPhase === 3,
                          }"
                        >{{ demoPhase === 3 ? '实现' : '实先' }}</span>起来反而简单很多。
                      </p>
                      <p class="mock-text mock-text--muted">
                        构建耗时从原来的 8 秒降到了 2 秒出头。
                      </p>
                    </div>
                  </div>

                  <div
                    class="mock-float-card mock-float-card--lint"
                    :class="{ 'is-visible': demoPhase === 2 }"
                  >
                    <div class="mock-lint-top">
                      <span class="mock-dot mock-dot--amber" />
                      <span class="mock-lint-reason">错别字</span>
                    </div>
                    <div class="mock-lint-row">
                      <span class="mock-lint-old">实先</span>
                      <NIcon
                        :component="ArrowRight16Regular"
                        :size="11"
                        class="opacity-50"
                      />
                      <span class="mock-lint-new">实现</span>
                      <span class="mock-lint-btn">替换</span>
                    </div>
                  </div>
                </div>

                <!-- 第 3 幕：行内续写 -->
                <div
                  v-else-if="step === 3"
                  key="step-3"
                  class="mock-editor"
                >
                  <div class="mock-editor-title">今年换了新的写作工具</div>
                  <div class="mock-editor-body">
                    <div class="mock-gutter">
                      <span>1</span>
                      <span>2</span>
                      <span>3</span>
                    </div>
                    <div class="mock-lines">
                      <p class="mock-text">
                        工具本身并不会让人写得更快，<span
                          v-if="demoPhase === 0"
                          class="mock-caret"
                        /><span
                          v-if="demoPhase >= 1 && demoPhase < 3"
                          class="mock-ghost-text"
                        >{{ ghostTextFull.slice(0, ghostTypedCount) }}</span><span
                          v-else-if="demoPhase === 3"
                          class="mock-ghost-settled"
                        >{{ ghostTextFull }}</span><span
                          v-if="demoPhase === 2"
                          class="mock-tab-pill"
                        >Tab</span>
                      </p>
                    </div>
                  </div>

                  <div class="mock-editor-statusbar">
                    <span>行 1, 列 {{ demoPhase === 3 ? 27 : 15 }}</span>
                    <div class="mock-keycaps">
                      <kbd :class="{ 'is-pressed': demoPhase === 3 }">Tab</kbd>
                      <span>采纳</span>
                      <span class="opacity-35">·</span>
                      <kbd>Alt</kbd>+<kbd>\</kbd>
                      <span>手动触发</span>
                    </div>
                  </div>
                </div>

                <!-- 第 4 幕：拟真仪表盘 -->
                <div
                  v-else
                  key="step-4"
                  class="mock-dashboard"
                >
                  <div class="mock-brief">
                    <div class="mock-brief-head">
                      <span class="mock-dot mock-dot--green" />
                      <span class="mock-brief-tag">站点简报 · 上升期</span>
                      <span class="mock-brief-summary">近 7 天访问量增长 18%，当前有 2 篇草稿未发布。</span>
                    </div>

                    <div
                      class="mock-brief-expand"
                      :class="{ 'is-open': demoPhase >= 1 }"
                    >
                      <div class="mock-brief-expand-inner">
                        <div class="mock-brief-cols">
                          <div><span>流量走势</span><p>周末阅读量高于工作日</p></div>
                          <div><span>更新频率</span><p>近 30 天发布 4 篇</p></div>
                          <div><span>互动概况</span><p>评论回复率保持高位</p></div>
                        </div>
                      </div>
                    </div>

                    <div class="mock-brief-pills">
                      <span class="mock-mini-pill">
                        查看草稿
                        <span
                          class="mock-pill-reason"
                          :class="{ 'is-open': demoPhase >= 1 }"
                        >
                          <span>&nbsp;· 2 篇待发布</span>
                        </span>
                      </span>
                      <span
                        class="mock-mini-pill"
                        :class="{ 'is-highlight': demoPhase >= 2 }"
                      >
                        新建文章 → &lt;!-- 自动带入大纲 --&gt;
                      </span>
                    </div>
                  </div>

                  <div class="mock-stat-grid">
                    <div class="mock-stat-card">
                      <span>用户总数</span>
                      <strong>128</strong>
                    </div>
                    <div class="mock-stat-card">
                      <span>总访问量</span>
                      <strong>24,610</strong>
                    </div>
                    <div class="mock-stat-card">
                      <span>在线峰值</span>
                      <strong>19</strong>
                    </div>
                    <div class="mock-stat-card">
                      <span>待办事项</span>
                      <strong>2</strong>
                    </div>
                  </div>
                </div>
              </Transition>
            </div>
          </div>
        </div>

        <div class="onboard-body">
          <div class="onboard-eyebrow">{{ currentFeatureMeta.eyebrow }}</div>
          <h3 class="onboard-title">{{ currentFeatureMeta.title }}</h3>
          <p class="onboard-desc">{{ currentFeatureMeta.desc }}</p>

          <div class="onboard-toggles">
            <div
              v-if="step === 1"
              class="onboard-toggle-row"
            >
              <div>
                <div class="onboard-toggle-label">段落批注</div>
                <div class="onboard-toggle-sub">默认开启 · 段落完成后在段尾提示</div>
              </div>
              <NSwitch
                v-model:value="draft.marginNotes"
                size="small"
              />
            </div>

            <div
              v-else-if="step === 2"
              class="onboard-toggle-row"
            >
              <div>
                <div class="onboard-toggle-label">错字校对</div>
                <div class="onboard-toggle-sub">默认开启 · 仅检查错别字与标点错误</div>
              </div>
              <NSwitch
                v-model:value="draft.inlineLint"
                size="small"
              />
            </div>

            <div
              v-else-if="step === 3"
              class="onboard-toggle-row"
            >
              <div>
                <div class="onboard-toggle-label">行内续写</div>
                <div class="onboard-toggle-sub">默认关闭 · 按需开启自动续写</div>
              </div>
              <NSwitch
                v-model:value="draft.ghostCompletion"
                size="small"
              />
            </div>

            <div
              v-else
              class="flex flex-col gap-2.5"
            >
              <div class="onboard-toggle-row">
                <div>
                  <div class="onboard-toggle-label">仪表盘站点简报</div>
                  <div class="onboard-toggle-sub">在首页顶部展示近期数据摘要（缓存 6 小时）</div>
                </div>
                <NSwitch
                  v-model:value="draft.dashboardBrief"
                  size="small"
                />
              </div>
              <div class="onboard-toggle-row">
                <div>
                  <div class="onboard-toggle-label">运营复盘与选题提示</div>
                  <div class="onboard-toggle-sub">启用独立复盘页，并在列表页显示选题提示</div>
                </div>
                <NSwitch
                  v-model:value="draft.opsReport"
                  size="small"
                />
              </div>
            </div>
          </div>

          <div class="onboard-footer">
            <div class="flex items-center gap-3">
              <div class="flex items-center gap-1.5">
                <button
                  v-for="idx in totalSteps"
                  :key="idx"
                  type="button"
                  class="onboard-dot"
                  :class="{ 'is-active': step === idx - 1 }"
                  :aria-label="`第 ${idx} 页`"
                  @click="step = idx - 1"
                />
              </div>
              <NButton
                size="small"
                quaternary
                @click="step--"
              >
                上一步
              </NButton>
            </div>

            <div class="flex items-center gap-2">
              <NButton
                size="small"
                quaternary
                :disabled="saving"
                @click="handleNeverShowAgain"
              >
                不再展示
              </NButton>
              <NButton
                size="small"
                type="primary"
                :loading="saving"
                @click="handleNext"
              >
                <template
                  v-if="isLastStep"
                  #icon
                >
                  <NIcon :component="Checkmark16Regular" />
                </template>
                {{ isLastStep ? '保存偏好并开始' : '下一步' }}
              </NButton>
            </div>
          </div>
        </div>
      </template>
    </div>
  </NModal>
</template>

<style scoped>
.onboard-shell {
  position: relative;
  width: 640px;
  max-width: calc(100vw - 32px);
  border-radius: 12px;
  border: 1px solid;
  overflow: hidden;
}

.onboard-close {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 5;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--ob-text-3);
  cursor: pointer;
  transition:
    color 150ms ease,
    background-color 150ms ease;
}

.onboard-close:hover {
  color: var(--ob-text-1);
  background-color: color-mix(in srgb, var(--ob-text-1) 6%, transparent);
}

/* ── Step 0：前置作者前言页 ── */
.cover-wrap {
  padding: 36px 36px 24px;
}

.cover-eyebrow {
  font-size: 11.5px;
  font-weight: 500;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--ob-text-3);
}

.cover-title {
  margin: 10px 0 0;
  font-size: 20px;
  font-weight: 500;
  line-height: 1.45;
  letter-spacing: -0.01em;
  color: var(--ob-text-1);
}

.cover-letter {
  margin-top: 22px;
  padding-top: 20px;
  border-top: 1px solid var(--ob-border);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.cover-letter p {
  margin: 0;
  font-size: 14px;
  line-height: 1.9;
  color: var(--ob-text-2);
}

.cover-sign {
  margin-top: 18px;
  text-align: right;
  font-size: 12.5px;
  letter-spacing: 0.02em;
  color: var(--ob-text-3);
}

/* ── 顶部工作台画布 ── */
.onboard-stage {
  height: 276px;
  padding: 24px 28px 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: color-mix(in srgb, var(--ob-text-1) 3%, var(--ob-card));
  background-image: radial-gradient(
    color-mix(in srgb, var(--ob-text-1) 10%, transparent) 1px,
    transparent 1px
  );
  background-size: 16px 16px;
  border-bottom: 1px solid var(--ob-border);
  overflow: hidden;
}

.mock-window {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  border-radius: 10px;
  border: 1px solid var(--ob-border);
  background-color: var(--ob-card);
  box-shadow: 0 10px 28px -12px rgba(0, 0, 0, 0.14);
  overflow: hidden;
}

.mock-chrome {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 32px;
  padding: 0 12px;
  border-bottom: 1px solid var(--ob-border);
  background-color: color-mix(in srgb, var(--ob-text-1) 2.5%, var(--ob-card));
  flex-shrink: 0;
}

.mock-traffic {
  display: flex;
  align-items: center;
  gap: 5px;
}

.mock-traffic span {
  width: 7px;
  height: 7px;
  border-radius: 9999px;
  background-color: color-mix(in srgb, var(--ob-text-1) 14%, transparent);
}

.mock-crumb {
  font-size: 11px;
  color: var(--ob-text-3);
}

.mock-chrome-right {
  display: flex;
  align-items: center;
}

.mock-status-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 1px 7px;
  border-radius: 9999px;
  font-size: 10.5px;
  color: var(--ob-text-2);
  background-color: color-mix(in srgb, var(--ob-text-1) 5%, transparent);
}

.mock-status-dot {
  width: 5px;
  height: 5px;
  border-radius: 9999px;
  background-color: var(--ob-primary);
}

.mock-status-dot.is-pulse {
  animation: mock-breathe 1.1s ease-in-out infinite;
}

.mock-viewport {
  position: relative;
  flex: 1;
  min-height: 0;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
}

/* ── 拟真编辑器 ── */
.mock-editor {
  position: relative;
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.mock-editor-title {
  font-size: 13.5px;
  font-weight: 500;
  color: var(--ob-text-1);
  padding-bottom: 8px;
  margin-bottom: 10px;
  border-bottom: 1px dashed var(--ob-border);
}

.mock-editor-body {
  display: flex;
  gap: 12px;
  flex: 1;
}

.mock-gutter {
  display: flex;
  flex-direction: column;
  gap: 5px;
  width: 16px;
  text-align: right;
  font-size: 10.5px;
  line-height: 1.85;
  color: var(--ob-text-3);
  opacity: 0.45;
  user-select: none;
  font-family: 'Fira Code', monospace;
}

.mock-lines {
  flex: 1;
  min-width: 0;
}

.mock-text {
  margin: 0 0 6px;
  font-size: 12.5px;
  line-height: 1.8;
  color: var(--ob-text-1);
}

.mock-text--muted {
  color: var(--ob-text-3);
  font-size: 12px;
}

.mock-pointer {
  position: absolute;
  pointer-events: none;
  color: var(--ob-text-1);
  opacity: 0;
  transition:
    transform 650ms cubic-bezier(0.22, 1, 0.36, 1),
    opacity 260ms ease;
}

.mock-pointer--margin {
  top: 96px;
  left: 300px;
  transform: translate(36px, 24px);
}

.mock-pointer--margin.is-hover {
  opacity: 0.85;
  transform: translate(-18px, -22px);
}

/* Step 1: 段落批注 */
.mock-margin-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  margin-left: 6px;
  border-radius: 9999px;
  vertical-align: middle;
  background-color: var(--ob-warning);
  opacity: 0;
  transform: scale(0.4);
  transition:
    opacity 320ms cubic-bezier(0.16, 1, 0.3, 1),
    transform 320ms cubic-bezier(0.16, 1, 0.3, 1),
    box-shadow 300ms ease;
}

.mock-margin-dot.is-visible {
  opacity: 0.88;
  transform: scale(1);
}

.mock-margin-dot.is-active {
  opacity: 1;
  box-shadow: 0 0 0 5px color-mix(in srgb, var(--ob-warning) 20%, transparent);
}

.mock-float-card {
  position: absolute;
  border-radius: 8px;
  border: 1px solid var(--ob-border);
  background-color: var(--ob-popover);
  box-shadow: var(--ob-shadow-2);
  opacity: 0;
  transform: translateY(8px) scale(0.98);
  transition:
    opacity 280ms cubic-bezier(0.16, 1, 0.3, 1),
    transform 280ms cubic-bezier(0.16, 1, 0.3, 1);
}

.mock-float-card.is-visible {
  opacity: 1;
  transform: translateY(0) scale(1);
}

.mock-float-card--margin {
  right: 8px;
  bottom: 4px;
  width: 264px;
  padding: 9px 11px;
}

.mock-float-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.mock-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 10.5px;
  font-weight: 500;
  color: var(--ob-text-2);
}

.mock-dot {
  width: 5px;
  height: 5px;
  border-radius: 9999px;
  flex-shrink: 0;
}

.mock-dot--amber {
  background-color: var(--ob-warning);
}

.mock-dot--green {
  background-color: var(--ob-success);
}

.mock-float-action {
  font-size: 10.5px;
  color: var(--ob-primary);
}

.mock-quote {
  margin-top: 5px;
  padding-left: 7px;
  border-left: 2px solid color-mix(in srgb, var(--ob-warning) 45%, transparent);
  font-size: 10.5px;
  color: var(--ob-text-3);
}

.mock-float-desc {
  margin: 5px 0 0;
  font-size: 11.5px;
  line-height: 1.5;
  color: var(--ob-text-2);
}

/* Step 2: 错字校对 */
.mock-lint-word {
  border-radius: 2px;
  transition: background-color 320ms ease;
}

.mock-lint-word.has-wave {
  text-decoration: underline wavy var(--ob-warning);
  text-underline-offset: 3px;
}

.mock-lint-word.is-fixed {
  padding: 0 2px;
  background-color: color-mix(in srgb, var(--ob-primary) 16%, transparent);
}

.mock-float-card--lint {
  left: 164px;
  bottom: 14px;
  padding: 8px 11px;
}

.mock-lint-top {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 10.5px;
  color: var(--ob-text-3);
}

.mock-lint-row {
  margin-top: 5px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}

.mock-lint-old {
  text-decoration: line-through;
  color: var(--ob-text-3);
}

.mock-lint-new {
  font-weight: 500;
  color: var(--ob-primary);
}

.mock-lint-btn {
  margin-left: 6px;
  padding: 1px 7px;
  border-radius: 4px;
  font-size: 10.5px;
  color: var(--ob-primary);
  background-color: color-mix(in srgb, var(--ob-primary) 12%, transparent);
}

/* Step 3: 行内续写 */
.mock-caret {
  display: inline-block;
  width: 1.5px;
  height: 1em;
  vertical-align: -2px;
  background-color: var(--ob-primary);
  animation: mock-caret-blink 0.9s ease-in-out infinite;
}

.mock-ghost-text {
  color: var(--ob-text-3);
  opacity: 0.54;
}

.mock-ghost-settled {
  border-radius: 3px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    color-mix(in srgb, var(--ob-primary) 18%, transparent) 45%,
    transparent 100%
  );
  background-size: 220% 100%;
  animation: mock-settle-sweep 520ms cubic-bezier(0.22, 1, 0.36, 1) forwards;
}

.mock-tab-pill {
  display: inline-flex;
  align-items: center;
  margin-left: 6px;
  padding: 0 5px;
  border-radius: 4px;
  font-size: 10px;
  line-height: 1.5;
  color: var(--ob-text-3);
  border: 1px solid var(--ob-border);
  background-color: var(--ob-popover);
  animation: mock-fade-in 220ms ease forwards;
}

.mock-editor-statusbar {
  margin-top: auto;
  padding-top: 8px;
  border-top: 1px solid var(--ob-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 10.5px;
  color: var(--ob-text-3);
}

.mock-keycaps {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.mock-keycaps kbd {
  padding: 1px 5px;
  border-radius: 4px;
  border: 1px solid var(--ob-border);
  background-color: var(--ob-popover);
  font-size: 10px;
  font-family: inherit;
  transition:
    transform 150ms ease,
    border-color 150ms ease,
    color 150ms ease;
}

.mock-keycaps kbd.is-pressed {
  transform: translateY(1px) scale(0.95);
  border-color: var(--ob-primary);
  color: var(--ob-primary);
}

/* Step 4: 拟真仪表盘 */
.mock-dashboard {
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1;
}

.mock-brief {
  padding: 10px 12px;
  border-radius: 6px;
  border: 1px solid var(--ob-border);
  background-color: color-mix(in srgb, var(--ob-text-1) 1.5%, var(--ob-card));
}

.mock-brief-head {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11.5px;
}

.mock-brief-tag {
  font-size: 10.5px;
  color: var(--ob-text-3);
  flex-shrink: 0;
}

.mock-brief-summary {
  color: var(--ob-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mock-brief-expand {
  display: grid;
  grid-template-rows: 0fr;
  opacity: 0;
  transition:
    grid-template-rows 280ms cubic-bezier(0.16, 1, 0.3, 1),
    opacity 220ms ease;
}

.mock-brief-expand.is-open {
  grid-template-rows: 1fr;
  opacity: 1;
}

.mock-brief-expand-inner {
  overflow: hidden;
}

.mock-brief-cols {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--ob-border);
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.mock-brief-cols span {
  font-size: 10px;
  color: var(--ob-text-3);
}

.mock-brief-cols p {
  margin: 2px 0 0;
  font-size: 10.5px;
  color: var(--ob-text-2);
}

.mock-brief-pills {
  margin-top: 7px;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.mock-mini-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 7px;
  border-radius: 4px;
  font-size: 10.5px;
  color: var(--ob-text-2);
  background-color: color-mix(in srgb, var(--ob-text-1) 5%, transparent);
  transition:
    color 200ms ease,
    background-color 200ms ease;
}

.mock-mini-pill.is-highlight {
  color: var(--ob-primary);
  background-color: color-mix(in srgb, var(--ob-primary) 10%, transparent);
}

.mock-pill-reason {
  display: inline-grid;
  grid-template-columns: 0fr;
  opacity: 0;
  transition:
    grid-template-columns 240ms cubic-bezier(0.16, 1, 0.3, 1),
    opacity 180ms ease;
}

.mock-pill-reason.is-open {
  grid-template-columns: 1fr;
  opacity: 1;
}

.mock-pill-reason > span {
  overflow: hidden;
  white-space: nowrap;
  color: var(--ob-text-3);
}

.mock-stat-grid {
  margin-top: auto;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.mock-stat-card {
  padding: 8px 10px;
  border-radius: 6px;
  border: 1px solid var(--ob-border);
  background-color: var(--ob-card);
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.mock-stat-card span {
  font-size: 10px;
  color: var(--ob-text-3);
}

.mock-stat-card strong {
  font-size: 13px;
  font-weight: 500;
  color: var(--ob-text-1);
}

/* ── 下半区 ── */
.onboard-body {
  padding: 20px 28px 20px;
}

.onboard-eyebrow {
  font-size: 11px;
  letter-spacing: 0.04em;
  color: var(--ob-text-3);
}

.onboard-title {
  margin: 4px 0 0;
  font-size: 16px;
  font-weight: 500;
  color: var(--ob-text-1);
}

.onboard-desc {
  margin: 7px 0 0;
  font-size: 13px;
  line-height: 1.65;
  color: var(--ob-text-2);
}

.onboard-toggles {
  margin-top: 15px;
  padding: 12px 14px;
  border-radius: 8px;
  background-color: color-mix(in srgb, var(--ob-text-1) 3.5%, transparent);
}

.onboard-toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.onboard-toggle-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--ob-text-1);
}

.onboard-toggle-sub {
  font-size: 11.5px;
  color: var(--ob-text-3);
  margin-top: 1px;
}

.onboard-footer {
  margin-top: 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.onboard-dot {
  width: 6px;
  height: 6px;
  padding: 0;
  border: none;
  border-radius: 9999px;
  background-color: color-mix(in srgb, var(--ob-text-1) 18%, transparent);
  cursor: pointer;
  transition:
    width 200ms cubic-bezier(0.16, 1, 0.3, 1),
    background-color 200ms ease;
}

.onboard-dot.is-active {
  width: 18px;
  background-color: var(--ob-primary);
}

@keyframes mock-settle-sweep {
  0% {
    background-position: 100% 0;
  }
  100% {
    background-position: -100% 0;
  }
}

@keyframes mock-fade-in {
  from {
    opacity: 0;
    transform: translateX(-2px);
  }
  to {
    opacity: 0.8;
    transform: translateX(0);
  }
}

@keyframes mock-breathe {
  0%,
  100% {
    opacity: 0.35;
    transform: scale(0.85);
  }
  50% {
    opacity: 1;
    transform: scale(1.2);
  }
}

@keyframes mock-caret-blink {
  0%,
  100% {
    opacity: 0.9;
  }
  50% {
    opacity: 0.15;
  }
}

.stage-fade-enter-active,
.stage-fade-leave-active {
  transition:
    opacity 200ms cubic-bezier(0.16, 1, 0.3, 1),
    transform 200ms cubic-bezier(0.16, 1, 0.3, 1);
}

.stage-fade-enter-from,
.stage-fade-leave-to {
  opacity: 0;
  transform: translateY(4px);
}
</style>
