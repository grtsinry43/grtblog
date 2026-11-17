<script setup lang="ts">
import { Sparkle20Regular } from '@vicons/fluent'
import { NButton, NIcon, NTooltip } from 'naive-ui'
import { computed } from 'vue'

const props = defineProps<{
  loading: boolean
  hasContent: boolean
  hasExistingTitle?: boolean
}>()

const emit = defineEmits<{
  generate: []
}>()

const tooltipText = computed(() => {
  if (!props.hasContent) {
    return '请先输入正文内容，再由 AI 生成标题与短链接'
  }
  if (props.hasExistingTitle) {
    return '根据正文重新生成标题与短链接（将覆盖当前值）'
  }
  return '根据正文自动生成标题与短链接'
})
</script>

<template>
  <NTooltip trigger="hover">
    <template #trigger>
      <NButton
        quaternary
        size="small"
        :loading="loading"
        :disabled="!hasContent"
        @click="emit('generate')"
      >
        <template #icon>
          <NIcon :component="Sparkle20Regular" />
        </template>
        {{ loading ? '生成中…' : '生成标题' }}
      </NButton>
    </template>
    <span>{{ tooltipText }}</span>
  </NTooltip>
</template>
