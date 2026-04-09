<script setup lang="ts">
import { NCard, NPopover } from 'naive-ui'

defineProps<{
  idle: boolean
  cursorLine: number
  cursorColumn: number
  readingMinutes: number
  charCount: number
  chineseCharCount: number
  wordCount: number
  totalCharCount: number
  paragraphCount: number
  selectionTotal: number
  selectionChars: number
}>()
</script>

<template>
  <div
    class="pointer-events-none absolute right-3 bottom-3 z-10 transition-opacity duration-200"
    :class="idle ? 'opacity-75 hover:opacity-100' : 'opacity-0'"
  >
    <NCard
      size="small"
      class="pointer-events-auto shadow-sm"
      content-style="padding: 6px 8px;"
    >
      <div class="flex items-center gap-3 text-[13px]">
        <NPopover
          trigger="hover"
          :disabled="!idle"
          content-style="padding: 4px 6px;"
        >
          <template #trigger>
            <span class="cursor-help">字数 {{ charCount }}</span>
          </template>
          <div class="flex flex-col gap-0.5 text-[11px] leading-tight">
            <span v-if="selectionTotal">选中 {{ selectionChars }}</span>
            <span>中文 {{ chineseCharCount }}</span>
            <span>英文词 {{ wordCount }}</span>
            <span>字符 {{ totalCharCount }}</span>
            <span>段落 {{ paragraphCount }}</span>
          </div>
        </NPopover>
        <span v-if="selectionTotal">选中 {{ selectionChars }} 字</span>
        <span>{{ cursorLine }}:{{ cursorColumn }}</span>
        <span>预计阅读 {{ readingMinutes }} 分钟</span>
      </div>
    </NCard>
  </div>
</template>
