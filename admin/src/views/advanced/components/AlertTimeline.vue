<script setup lang="ts">
import { NEmpty, NTimeline, NTimelineItem } from 'naive-ui'

defineProps<{
  alerts?: Array<{
    id: number
    isRead: boolean
    title: string
    createdAt: string
    type: string
    content: string
  }>
}>()
</script>

<template>
  <div
    class="rounded border border-naive-border bg-naive-card p-5 transition-[background-color,border-color]"
  >
    <div class="mb-4 text-base font-medium text-neutral-600 dark:text-neutral-300">系统告警流</div>
    <NEmpty
      v-if="!alerts?.length"
      description="暂无告警"
    />
    <NTimeline v-else>
      <NTimelineItem
        v-for="item in alerts"
        :key="item.id"
        :type="item.isRead ? 'default' : 'warning'"
        :title="item.title"
        :time="new Date(item.createdAt).toLocaleString()"
      >
        <div class="mb-1 text-xs text-neutral-500">{{ item.type }}</div>
        <div class="text-sm text-neutral-700 dark:text-neutral-300">{{ item.content }}</div>
      </NTimelineItem>
    </NTimeline>
  </div>
</template>
