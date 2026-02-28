<script setup lang="ts">
import {
  ArrowUp20Filled,
  Checkmark20Filled,
  Dismiss20Filled,
  Sparkle20Filled,
} from '@vicons/fluent'
import { NIcon, NScrollbar, useThemeVars } from 'naive-ui'
import { computed, nextTick, ref, watch } from 'vue'

import { computeLineDiff } from '@/composables/markdown-editor/use-ai-toolbar'

import type { DiffLine } from '@/composables/markdown-editor/use-ai-toolbar'

const props = defineProps<{
  visible: boolean
  instruction: string
  loading: boolean
  resultContent: string
  showResult: boolean
  originalContent: string
}>()

const emit = defineEmits<{
  'update:instruction': [value: string]
  execute: []
  accept: []
  reject: []
  close: []
}>()

const themeVars = useThemeVars()
const inputRef = ref<HTMLInputElement>()
const scrollRef = ref<InstanceType<typeof NScrollbar>>()

const phase = computed<'input' | 'streaming' | 'diff'>(() => {
  if (!props.showResult) return 'input'
  if (props.loading) return 'streaming'
  return 'diff'
})

const diffLines = computed<DiffLine[]>(() => {
  if (phase.value !== 'diff') return []
  return computeLineDiff(props.originalContent, props.resultContent)
})

const diffStats = computed(() => {
  let added = 0
  let removed = 0
  for (const line of diffLines.value) {
    if (line.type === 'added') added++
    else if (line.type === 'removed') removed++
  }
  return { added, removed }
})

// Auto-scroll streaming content to bottom
watch(
  () => props.resultContent,
  () => {
    if (phase.value === 'streaming') {
      nextTick(() => {
        scrollRef.value?.scrollTo({ top: 999999 })
      })
    }
  },
)

// Auto-focus input when toolbar becomes visible
watch(
  () => props.visible,
  (v) => {
    if (v) {
      nextTick(() => inputRef.value?.focus())
    }
  },
)

function onInputKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    emit('execute')
  } else if (e.key === 'Escape') {
    e.preventDefault()
    emit('close')
  }
}
</script>

<template>
  <Transition name="ai-toolbar">
    <div v-if="visible" class="ai-toolbar-container">
      <!-- Upper panel: streaming or diff -->
      <div v-if="phase === 'streaming'" class="ai-panel-upper">
        <NScrollbar ref="scrollRef" style="max-height: 260px">
          <pre class="ai-stream-pre">{{ resultContent }}<span class="ai-cursor" /></pre>
        </NScrollbar>
      </div>

      <div v-else-if="phase === 'diff'" class="ai-panel-upper">
        <div class="ai-diff-header">
          <span class="ai-diff-title">变更预览</span>
          <span class="ai-diff-stats">
            <span class="ai-diff-stat-removed">-{{ diffStats.removed }}</span>
            <span class="ai-diff-stat-added">+{{ diffStats.added }}</span>
          </span>
        </div>
        <NScrollbar style="max-height: 260px">
          <div class="ai-diff-body">
            <div
              v-for="(line, idx) in diffLines"
              :key="idx"
              class="ai-diff-line"
              :class="`ai-diff-line--${line.type}`"
            >
              <span class="ai-diff-gutter">{{
                line.type === 'added' ? '+' : line.type === 'removed' ? '-' : ' '
              }}</span>
              <span class="ai-diff-text">{{ line.text }}</span>
            </div>
          </div>
        </NScrollbar>
      </div>

      <!-- Bottom input bar (always present) -->
      <div class="ai-input-bar">
        <NIcon :component="Sparkle20Filled" :size="16" class="ai-input-icon" />
        <input
          ref="inputRef"
          class="ai-input"
          :value="instruction"
          :disabled="phase === 'streaming'"
          placeholder="输入改写指令，如：扩写、改为正式语气、翻译为英文..."
          @input="emit('update:instruction', ($event.target as HTMLInputElement).value)"
          @keydown="onInputKeydown"
        />
        <div class="ai-input-actions">
          <template v-if="phase === 'diff'">
            <button class="ai-btn ai-btn--reject" title="放弃" @click="emit('reject')">
              <NIcon :component="Dismiss20Filled" :size="14" />
            </button>
            <button class="ai-btn ai-btn--accept" title="采纳" @click="emit('accept')">
              <NIcon :component="Checkmark20Filled" :size="14" />
              <span>采纳</span>
            </button>
          </template>
          <template v-else-if="phase === 'streaming'">
            <span class="ai-spinner" />
          </template>
          <template v-else>
            <button
              class="ai-btn ai-btn--send"
              :disabled="!instruction.trim()"
              title="发送"
              @click="emit('execute')"
            >
              <NIcon :component="ArrowUp20Filled" :size="14" />
            </button>
          </template>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.ai-toolbar-container {
  position: absolute;
  left: 50%;
  bottom: 16px;
  transform: translateX(-50%);
  z-index: 20;
  width: min(560px, calc(100% - 32px));
  border-radius: 8px;
  background-color: v-bind('themeVars.popoverColor');
  border: 1px solid v-bind('themeVars.dividerColor');
  box-shadow: v-bind('themeVars.boxShadow2');
  overflow: hidden;
}

/* ── Upper panels ── */
.ai-panel-upper {
  border-bottom: 1px solid v-bind('themeVars.dividerColor');
}

/* ── Streaming preview ── */
.ai-stream-pre {
  margin: 0;
  padding: 12px 14px;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Fira Code', 'SFMono-Regular', Menlo, Monaco, Consolas, monospace;
  color: v-bind('themeVars.textColor1');
}

.ai-cursor {
  display: inline-block;
  width: 2px;
  height: 1em;
  vertical-align: text-bottom;
  background-color: v-bind('themeVars.primaryColor');
  animation: ai-blink 1s step-end infinite;
}

@keyframes ai-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

/* ── Diff view ── */
.ai-diff-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 14px;
  border-bottom: 1px solid v-bind('themeVars.dividerColor');
}

.ai-diff-title {
  font-size: 12px;
  font-weight: 500;
  color: v-bind('themeVars.textColor2');
}

.ai-diff-stats {
  display: flex;
  gap: 8px;
  font-size: 12px;
  font-family: 'Fira Code', 'SFMono-Regular', Menlo, Monaco, Consolas, monospace;
}

.ai-diff-stat-removed {
  color: v-bind('themeVars.errorColor');
}

.ai-diff-stat-added {
  color: v-bind('themeVars.successColor');
}

.ai-diff-body {
  padding: 4px 0;
}

.ai-diff-line {
  display: flex;
  font-size: 13px;
  line-height: 1.6;
  font-family: 'Fira Code', 'SFMono-Regular', Menlo, Monaco, Consolas, monospace;
}

.ai-diff-gutter {
  flex-shrink: 0;
  width: 28px;
  text-align: center;
  user-select: none;
  color: v-bind('themeVars.textColor3');
}

.ai-diff-text {
  flex: 1;
  white-space: pre-wrap;
  word-break: break-word;
  padding-right: 14px;
}

.ai-diff-line--added {
  background-color: color-mix(
    in srgb,
    v-bind('themeVars.successColor') 8%,
    v-bind('themeVars.popoverColor')
  );
}

.ai-diff-line--added .ai-diff-gutter {
  color: v-bind('themeVars.successColor');
}

.ai-diff-line--removed {
  background-color: color-mix(
    in srgb,
    v-bind('themeVars.errorColor') 8%,
    v-bind('themeVars.popoverColor')
  );
}

.ai-diff-line--removed .ai-diff-gutter {
  color: v-bind('themeVars.errorColor');
}

/* ── Bottom input bar ── */
.ai-input-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
}

.ai-input-icon {
  flex-shrink: 0;
  color: v-bind('themeVars.primaryColor');
}

.ai-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  line-height: 1.4;
  color: v-bind('themeVars.textColor1');
  font-family: inherit;
}

.ai-input::placeholder {
  color: v-bind('themeVars.textColor3');
}

.ai-input:disabled {
  opacity: 0.5;
}

.ai-input-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

/* ── Buttons ── */
.ai-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  border: none;
  cursor: pointer;
  border-radius: 6px;
  font-size: 12px;
  line-height: 1;
  transition: all 0.15s ease;
  font-family: inherit;
}

.ai-btn--send {
  width: 24px;
  height: 24px;
  padding: 0;
  background-color: v-bind('themeVars.primaryColor');
  color: #fff;
  border-radius: 50%;
}

.ai-btn--send:hover:not(:disabled) {
  opacity: 0.85;
}

.ai-btn--send:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.ai-btn--accept {
  height: 26px;
  padding: 0 10px;
  background: transparent;
  color: v-bind('themeVars.successColor');
  border: 1px solid v-bind('themeVars.successColor');
}

.ai-btn--accept:hover {
  background-color: color-mix(
    in srgb,
    v-bind('themeVars.successColor') 10%,
    v-bind('themeVars.popoverColor')
  );
}

.ai-btn--reject {
  width: 26px;
  height: 26px;
  padding: 0;
  background: transparent;
  color: v-bind('themeVars.textColor3');
  border: 1px solid v-bind('themeVars.dividerColor');
}

.ai-btn--reject:hover {
  background-color: v-bind('themeVars.hoverColor');
  color: v-bind('themeVars.textColor1');
}

/* ── Spinner ── */
.ai-spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid v-bind('themeVars.dividerColor');
  border-top-color: v-bind('themeVars.primaryColor');
  border-radius: 50%;
  animation: ai-spin 0.7s linear infinite;
}

@keyframes ai-spin {
  to { transform: rotate(360deg); }
}

/* ── Transition ── */
.ai-toolbar-enter-active,
.ai-toolbar-leave-active {
  transition: all 0.2s ease;
}

.ai-toolbar-enter-from,
.ai-toolbar-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(8px);
}
</style>
