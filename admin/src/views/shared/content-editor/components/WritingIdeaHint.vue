<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { ArrowRight16Regular, ArrowSync16Regular, Dismiss16Regular } from '@vicons/fluent'
import { NIcon, useThemeVars } from 'naive-ui'
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { getCachedOpsReportIdeas } from '@/services/ai'
import { setPendingWritingIdea } from '@/views/shared/content-editor/utils/writing-idea-handoff'

const props = defineProps<{
  target: 'article' | 'note'
}>()

const DISMISS_STORAGE_KEY = 'grtblog:writing-idea-hint-dismissed'

const router = useRouter()
const themeVars = useThemeVars()

const currentIndex = ref(0)
const dismissedVersion = ref<string>('')

try {
  dismissedVersion.value = localStorage.getItem(DISMISS_STORAGE_KEY) ?? ''
} catch {
  dismissedVersion.value = ''
}

const { data } = useQuery({
  queryKey: ['ops-report-cached-ideas'],
  queryFn: getCachedOpsReportIdeas,
  staleTime: 1000 * 60 * 10,
  retry: false,
})

const ideas = computed(() => data.value?.ideas ?? [])
const versionKey = computed(() => `${props.target}:${data.value?.generatedAt ?? ''}`)

watch(ideas, (list) => {
  if (currentIndex.value >= list.length) {
    currentIndex.value = 0
  }
})

const currentIdea = computed(() => ideas.value[currentIndex.value] ?? null)

const visible = computed(() => {
  if (!currentIdea.value) return false
  if (dismissedVersion.value && dismissedVersion.value === versionKey.value) return false
  return true
})

function nextIdea() {
  if (ideas.value.length <= 1) return
  currentIndex.value = (currentIndex.value + 1) % ideas.value.length
}

function dismiss() {
  dismissedVersion.value = versionKey.value
  try {
    localStorage.setItem(DISMISS_STORAGE_KEY, versionKey.value)
  } catch {
    // 忽略存储受限
  }
}

function goCreate() {
  if (currentIdea.value) {
    setPendingWritingIdea(props.target, currentIdea.value)
  }
  if (props.target === 'article') {
    void router.push({ name: 'articleCreate' })
  } else {
    void router.push({ name: 'noteCreate' })
  }
}
</script>

<template>
  <Transition name="idea-hint-fade">
    <div
      v-if="visible && currentIdea"
      class="idea-hint-bar"
    >
      <div class="idea-hint-main">
        <span class="idea-hint-label">
          <span class="idea-hint-dot" />
          <span>选题灵感</span>
        </span>

        <Transition
          name="idea-swap"
          mode="out-in"
        >
          <span
            :key="currentIndex"
            class="idea-hint-text"
            :title="currentIdea.why ? `${currentIdea.title} — ${currentIdea.why}` : currentIdea.title"
          >
            <span class="idea-hint-title">{{ currentIdea.title }}</span>
            <span
              v-if="currentIdea.angle"
              class="idea-hint-angle hidden md:inline"
            >
              · {{ currentIdea.angle }}
            </span>
          </span>
        </Transition>
      </div>

      <div class="idea-hint-actions">
        <button
          v-if="ideas.length > 1"
          type="button"
          class="idea-hint-btn"
          title="换一条灵感"
          @click="nextIdea"
        >
          <NIcon
            :component="ArrowSync16Regular"
            :size="12"
          />
          <span>换一条</span>
        </button>

        <button
          type="button"
          class="idea-hint-btn idea-hint-btn--go"
          @click="goCreate"
        >
          <span>{{ target === 'article' ? '去写一篇' : '记一条' }}</span>
          <NIcon
            :component="ArrowRight16Regular"
            :size="12"
          />
        </button>

        <button
          type="button"
          class="idea-hint-close"
          title="暂时收起"
          @click="dismiss"
        >
          <NIcon
            :component="Dismiss16Regular"
            :size="12"
          />
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.idea-hint-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 14px;
  border-radius: v-bind('themeVars.borderRadius');
  border: 1px solid v-bind('themeVars.borderColor');
  background-color: v-bind('themeVars.cardColor');
  font-size: 12.5px;
  line-height: 1.5;
}

.idea-hint-main {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.idea-hint-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
  font-size: 11.5px;
  color: v-bind('themeVars.textColor3');
  user-select: none;
}

.idea-hint-dot {
  width: 5px;
  height: 5px;
  border-radius: 9999px;
  background-color: v-bind('themeVars.primaryColor');
  opacity: 0.7;
}

.idea-hint-text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.idea-hint-title {
  color: v-bind('themeVars.textColor1');
  font-weight: 400;
}

.idea-hint-angle {
  margin-left: 4px;
  color: v-bind('themeVars.textColor3');
  font-size: 12px;
}

.idea-hint-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.idea-hint-btn {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 2px 7px;
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

.idea-hint-btn:hover {
  color: v-bind('themeVars.textColor1');
  background-color: color-mix(in srgb, v-bind('themeVars.textColor1') 5%, transparent);
}

.idea-hint-btn--go {
  color: v-bind('themeVars.textColor2');
  background-color: color-mix(in srgb, v-bind('themeVars.textColor1') 4%, transparent);
}

.idea-hint-btn--go:hover {
  color: v-bind('themeVars.textColor1');
  background-color: color-mix(in srgb, v-bind('themeVars.textColor1') 8%, transparent);
}

.idea-hint-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  border-radius: v-bind('themeVars.borderRadius');
  background: transparent;
  color: v-bind('themeVars.textColor3');
  opacity: 0.65;
  cursor: pointer;
  transition:
    opacity 150ms ease,
    color 150ms ease;
}

.idea-hint-close:hover {
  opacity: 1;
  color: v-bind('themeVars.textColor1');
}

.idea-hint-fade-enter-active,
.idea-hint-fade-leave-active {
  transition:
    opacity 200ms cubic-bezier(0.16, 1, 0.3, 1),
    transform 200ms cubic-bezier(0.16, 1, 0.3, 1);
}

.idea-hint-fade-enter-from,
.idea-hint-fade-leave-to {
  opacity: 0;
  transform: translateY(-3px);
}

.idea-swap-enter-active,
.idea-swap-leave-active {
  transition: opacity 140ms ease;
}

.idea-swap-enter-from,
.idea-swap-leave-to {
  opacity: 0;
}
</style>
