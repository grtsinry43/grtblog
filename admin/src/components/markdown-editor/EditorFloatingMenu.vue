<script setup lang="ts">
import {
  AlignSpaceBetweenHorizontal20Regular as IconPangu,
  Sparkle20Filled as IconSparkle,
  TextBold20Regular as IconBold,
  TextItalic20Regular as IconItalic,
} from '@vicons/fluent'
import { CodeSlash as IconCode, RemoveCircleOutline as IconStrike } from '@vicons/ionicons5'
import { NButton, NCard, NIcon, useThemeVars } from 'naive-ui'
import { computed } from 'vue'

import type { FormatType } from '@/composables/markdown-editor/utils/ast'

interface Props {
  visible: boolean
  pos: { top: number; left: number }
  activeFormats: Set<FormatType>
  realtimeAi?: boolean
  selectedText?: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'command', type: FormatType): void
  (e: 'pangu'): void
  (e: 'ai-action', instruction?: string): void
}>()

const themeVars = useThemeVars()

const style = computed(() => {
  const clampedLeft = Math.max(140, Math.min(props.pos.left, window.innerWidth - 140))
  const flipBelow = props.pos.top < 72
  return {
    top: `${flipBelow ? Math.max(60, props.pos.top + 36) : props.pos.top}px`,
    left: `${clampedLeft}px`,
    transform: flipBelow ? 'translate(-50%, 0)' : 'translate(-50%, -120%)',
    position: 'fixed' as const,
    zIndex: 9999,
  }
})

const menuContentStyle = computed(() => ({
  backgroundColor: themeVars.value.popoverColor,
  borderRadius: themeVars.value.borderRadius,
  boxShadow: themeVars.value.boxShadow2,
  borderColor: themeVars.value.dividerColor,
}))

const buttons = [
  { type: 'bold', icon: IconBold, tip: '加粗' },
  { type: 'italic', icon: IconItalic, tip: '斜体' },
  { type: 'strike', icon: IconStrike, tip: '删除线' },
  { type: 'code', icon: IconCode, tip: '行内代码' },
] as const

const handleMousedown = () => {
  // 阻止默认行为
}
</script>

<template>
  <Teleport to="body">
    <Transition name="fade-scale">
      <div
        v-if="visible"
        class="floating-menu"
        :style="style"
        @mousedown.prevent="handleMousedown"
      >
        <NCard
          size="small"
          :style="menuContentStyle"
          content-style="padding: 4px;"
        >
          <div class="menu-content">
            <n-button
              v-for="btn in buttons"
              :key="btn.type"
              quaternary
              size="small"
              class="action-btn"
              :class="{ 'is-active': activeFormats.has(btn.type as FormatType) }"
              :title="btn.tip"
              @click="emit('command', btn.type as FormatType)"
            >
              <template #icon>
                <n-icon :component="btn.icon" />
              </template>
            </n-button>

            <span
              class="menu-divider"
              :style="{ backgroundColor: themeVars.dividerColor }"
            />

            <n-button
              quaternary
              size="small"
              class="action-btn"
              title="盘古格式化（中英文间自动加空格）"
              @click="emit('pangu')"
            >
              <template #icon>
                <n-icon :component="IconPangu" />
              </template>
            </n-button>

            <n-button
              quaternary
              size="small"
              class="action-btn ai-trigger-btn"
              :class="{ 'is-active': props.realtimeAi }"
              :style="{ color: themeVars.primaryColor }"
              title="AI 改写或扩写"
              @click="emit('ai-action')"
            >
              <template #icon>
                <n-icon :component="IconSparkle" />
              </template>
            </n-button>
          </div>
        </NCard>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.floating-menu {
  pointer-events: auto;
}

.menu-content {
  display: flex;
  align-items: center;
  gap: 2px;
}

.action-btn {
  /* 按钮圆角跟随全局设置 */
  border-radius: v-bind('themeVars.borderRadius') !important;
  color: v-bind('themeVars.textColor2');
  transition: all 0.2s var(--n-bezier);
}

/* * 激活状态
 * 依然使用 color-mix 是为了让激活色看起来“不刺眼”，
 * 这里是把主色调和背景色混合，而不是和透明混合，保证了没有透明度。
 */
.action-btn.is-active {
  /* 在 srgb 空间混合：15% 的主色 + 85% 的 Popover 背景色 = 看起来像淡色背景，其实是实色 */
  background-color: color-mix(
    in srgb,
    v-bind('themeVars.primaryColor'),
    v-bind('themeVars.popoverColor') 85%
  );
  color: v-bind('themeVars.primaryColor');
}

/* 悬停状态 */
.action-btn:not(.is-active):hover {
  background-color: v-bind('themeVars.hoverColor');
  color: v-bind('themeVars.textColor1');
}

.menu-divider {
  width: 1px;
  height: 16px;
  margin: 0 3px;
  background-color: v-bind('themeVars.dividerColor');
}

.ai-trigger-btn {
  color: v-bind('themeVars.primaryColor');
}

/* 动画保持不变，依旧使用 Naive 标准曲线 */
.fade-scale-enter-active,
.fade-scale-leave-active {
  transition:
    opacity 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    transform 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  transform-origin: bottom center;
}

.fade-scale-enter-from,
.fade-scale-leave-to {
  opacity: 0;
  transform: translate(-50%, -100%) scale(0.95);
}
</style>
