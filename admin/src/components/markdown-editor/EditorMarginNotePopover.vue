<script setup lang="ts">
import { Dismiss16Regular, Sparkle16Regular } from '@vicons/fluent'
import { NButton, NCard, NIcon, useThemeVars } from 'naive-ui'
import { computed } from 'vue'

import type { MarginNote } from '@/composables/markdown-editor/extensions/ai-margin-notes'
import type { MarginAnnotationKind } from '@/services/ai'

const props = defineProps<{
  note: MarginNote | null
  pos: { top: number; left: number } | null
}>()

const emit = defineEmits<{
  rewrite: [note: MarginNote]
  dismiss: [note: MarginNote]
  close: []
}>()

const themeVars = useThemeVars()

const kindMeta = computed(() => {
  const kind: MarginAnnotationKind = props.note?.kind ?? 'refine'
  if (kind === 'spark') {
    return { label: '思路延伸', color: themeVars.value.primaryColor }
  }
  if (kind === 'structure') {
    return { label: '逻辑结构', color: themeVars.value.successColor }
  }
  return { label: '表达推敲', color: themeVars.value.warningColor }
})

const style = computed(() => {
  if (!props.pos) return {}
  const cardWidth = 310
  const clampedLeft = Math.max(16, Math.min(props.pos.left - cardWidth + 24, window.innerWidth - cardWidth - 16))
  const flipBelow = props.pos.top < 190
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
    <Transition name="margin-note-pop">
      <div
        v-if="note && pos"
        class="margin-note-popover"
        :style="style"
        @mousedown.prevent
      >
        <NCard
          size="small"
          class="margin-note-card"
          :style="{ boxShadow: themeVars.boxShadow2, backgroundColor: themeVars.popoverColor }"
          content-style="padding: 12px 14px;"
        >
          <div class="margin-note-body">
            <div class="margin-note-header">
              <span
                class="margin-note-badge"
                :style="{
                  color: kindMeta.color,
                  backgroundColor: `color-mix(in srgb, ${kindMeta.color} 12%, transparent)`,
                }"
              >
                <span
                  class="margin-note-dot"
                  :style="{ backgroundColor: kindMeta.color }"
                />
                {{ kindMeta.label }}
              </span>
              <button
                type="button"
                class="margin-note-close"
                :style="{ color: themeVars.textColor3 }"
                title="收起"
                @click="emit('close')"
              >
                <NIcon
                  :component="Dismiss16Regular"
                  size="13"
                />
              </button>
            </div>

            <div
              v-if="note.quote"
              class="margin-note-quote"
              :style="{
                color: themeVars.textColor3,
                borderLeftColor: `color-mix(in srgb, ${kindMeta.color} 45%, transparent)`,
              }"
            >
              “{{ note.quote }}”
            </div>

            <div
              class="margin-note-comment"
              :style="{ color: themeVars.textColor1 }"
            >
              {{ note.comment }}
            </div>

            <div
              v-if="note.suggestion"
              class="margin-note-suggestion"
              :style="{
                color: themeVars.textColor2,
                backgroundColor: `color-mix(in srgb, ${themeVars.textColor1} 4%, transparent)`,
              }"
            >
              {{ note.suggestion }}
            </div>

            <div class="margin-note-actions">
              <NButton
                size="tiny"
                type="primary"
                secondary
                @click="emit('rewrite', note)"
              >
                <template #icon><NIcon :component="Sparkle16Regular" /></template>
                基于此改写
              </NButton>
              <NButton
                size="tiny"
                quaternary
                @click="emit('dismiss', note)"
              >
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
.margin-note-card {
  width: 310px;
  max-width: calc(100vw - 32px);
}

.margin-note-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.margin-note-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.margin-note-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 1px 7px;
  border-radius: 9999px;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.5;
}

.margin-note-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
}

.margin-note-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  border-radius: 4px;
  background: transparent;
  cursor: pointer;
  opacity: 0.7;
  transition: opacity 150ms ease;
}

.margin-note-close:hover {
  opacity: 1;
}

.margin-note-quote {
  padding-left: 8px;
  border-left: 2px solid;
  font-size: 11.5px;
  line-height: 1.45;
}

.margin-note-comment {
  font-size: 12.5px;
  line-height: 1.6;
}

.margin-note-suggestion {
  padding: 6px 8px;
  border-radius: 5px;
  font-size: 12px;
  line-height: 1.5;
}

.margin-note-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  margin-top: 2px;
}

.margin-note-pop-enter-active,
.margin-note-pop-leave-active {
  transition:
    opacity 180ms cubic-bezier(0.16, 1, 0.3, 1),
    transform 180ms cubic-bezier(0.16, 1, 0.3, 1);
}

.margin-note-pop-enter-from,
.margin-note-pop-leave-to {
  opacity: 0;
  transform: translateY(-94%) scale(0.97);
}
</style>
