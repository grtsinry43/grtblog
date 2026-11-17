<script setup lang="ts">
import { ArrowRight16Regular, Checkmark16Filled, Dismiss16Regular } from '@vicons/fluent'
import { NButton, NCard, NIcon, useThemeVars } from 'naive-ui'
import { computed } from 'vue'

import type { LintMatch } from '@/composables/markdown-editor/extensions/ai-inline-lint'

const props = defineProps<{
  match: LintMatch | null
  pos: { top: number; left: number } | null
}>()

const emit = defineEmits<{
  accept: [match: LintMatch]
  dismiss: [match: LintMatch]
}>()

const themeVars = useThemeVars()

const style = computed(() => {
  if (!props.pos) return {}
  const clampedLeft = Math.max(16, Math.min(props.pos.left, window.innerWidth - 316))
  const flipBelow = props.pos.top < 140
  return {
    position: 'fixed' as const,
    top: `${flipBelow ? props.pos.top + 32 : props.pos.top}px`,
    left: `${clampedLeft}px`,
    transform: flipBelow ? 'translateY(0)' : 'translateY(-100%)',
    zIndex: 10000,
  }
})
</script>

<template>
  <Teleport to="body">
    <Transition name="lint-pop">
      <div
        v-if="match && pos"
        class="lint-popover"
        :style="style"
        @mousedown.prevent
      >
        <NCard
          size="small"
          class="lint-card"
          :style="{ boxShadow: themeVars.boxShadow2, backgroundColor: themeVars.popoverColor }"
          content-style="padding: 10px 12px;"
        >
          <div class="lint-body">
            <div class="lint-top">
              <span
                class="lint-dot"
                :style="{
                  backgroundColor:
                    match.severity === 'typo' ? themeVars.warningColor : themeVars.primaryColor,
                }"
              />
              <span
                class="lint-reason"
                :style="{ color: themeVars.textColor3 }"
              >
                {{ match.reason }}
              </span>
            </div>

            <div class="lint-diff">
              <span
                class="lint-old"
                :style="{ color: themeVars.textColor3 }"
              >
                {{ match.original }}
              </span>
              <NIcon
                :component="ArrowRight16Regular"
                size="13"
                :style="{ color: themeVars.textColor3, flexShrink: 0 }"
              />
              <span
                class="lint-new"
                :style="{ color: themeVars.primaryColor }"
              >
                {{ match.suggestion }}
              </span>
            </div>

            <div class="lint-actions">
              <NButton
                size="tiny"
                type="primary"
                secondary
                @click="emit('accept', match)"
              >
                <template #icon><NIcon :component="Checkmark16Filled" /></template>
                替换
              </NButton>
              <NButton
                size="tiny"
                quaternary
                @click="emit('dismiss', match)"
              >
                <template #icon><NIcon :component="Dismiss16Regular" /></template>
                忽略
              </NButton>
            </div>
          </div>
        </NCard>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.lint-card {
  min-width: 200px;
  max-width: 300px;
}

.lint-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.lint-top {
  display: flex;
  align-items: center;
  gap: 6px;
}

.lint-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.lint-reason {
  font-size: 11px;
  line-height: 1.3;
}

.lint-diff {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 5px;
  font-size: 12.5px;
  line-height: 1.45;
}

.lint-old {
  text-decoration: line-through;
}

.lint-new {
  font-weight: 500;
}

.lint-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
  margin-top: 2px;
}

.lint-pop-enter-active,
.lint-pop-leave-active {
  transition:
    opacity 150ms cubic-bezier(0.22, 1, 0.36, 1),
    transform 150ms cubic-bezier(0.22, 1, 0.36, 1);
}

.lint-pop-enter-from,
.lint-pop-leave-to {
  opacity: 0;
  transform: translateY(-92%) scale(0.96);
}
</style>
