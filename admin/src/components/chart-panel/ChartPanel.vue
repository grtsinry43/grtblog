<script setup lang="ts">
import { NSkeleton } from 'naive-ui'

defineOptions({
  name: 'ChartPanel',
})

const props = withDefaults(defineProps<{
  title: string
  height?: number
  tabs?: { label: string, value: string }[]
  activeTab?: string
  loading?: boolean
}>(), {
  height: 420,
  tabs: () => [],
  activeTab: '',
  loading: false,
})

const emit = defineEmits<{
  'update:activeTab': [value: string]
}>()

function selectTab(value: string) {
  emit('update:activeTab', value)
}
</script>

<template>
  <div
    class="flex flex-col rounded border border-naive-border bg-naive-card transition-[background-color,border-color]"
    :style="{ height: `${props.height}px` }"
  >
    <div class="flex items-center justify-between px-5 pt-4">
      <div class="flex items-center gap-2">
        <span class="text-base font-medium text-neutral-600 dark:text-neutral-300">{{ title }}</span>
        <slot name="header-extra" />
      </div>
      <div v-if="tabs.length" class="flex items-center gap-x-1 rounded bg-neutral-100 p-0.5 dark:bg-neutral-800">
        <button
          v-for="tab in tabs"
          :key="tab.value"
          class="rounded-xs px-3 py-1 text-xs transition-all"
          :class="activeTab === tab.value ? 'bg-white text-neutral-700 shadow-sm dark:bg-neutral-700 dark:text-neutral-200' : 'text-neutral-400 hover:text-neutral-600 dark:hover:text-neutral-300'"
          @click="selectTab(tab.value)"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>
    <div class="flex-1 px-4 pb-4 pt-2">
      <slot v-if="!loading" />
      <slot v-else name="loading">
        <div class="flex h-full items-center justify-center">
          <NSkeleton text class="h-full w-full" />
        </div>
      </slot>
    </div>
  </div>
</template>
