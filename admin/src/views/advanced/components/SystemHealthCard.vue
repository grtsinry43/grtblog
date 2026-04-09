<script setup lang="ts">
import { NEmpty, NTag } from 'naive-ui'

defineProps<{
  componentHealths: Array<{
    name: string
    healthy: boolean
    status: string
    version?: string
    checkedAt?: string
  }>
  componentTagType: (item: { healthy: boolean; status: string }) => string
}>()
</script>

<template>
  <div
    class="rounded border border-naive-border bg-naive-card p-5 transition-[background-color,border-color]"
  >
    <div class="mb-4 text-base font-medium text-neutral-600 dark:text-neutral-300">
      组件健康状态
    </div>
    <NEmpty
      v-if="!componentHealths.length"
      description="暂无组件健康数据"
    />
    <div
      v-else
      class="space-y-3"
    >
      <div class="flex flex-wrap gap-2">
        <NTag
          v-for="item in componentHealths"
          :key="item.name"
          :type="componentTagType(item) as any"
          size="small"
          round
        >
          {{ item.name }} · {{ item.status }} · v{{ item.version || 'n/a' }}
        </NTag>
      </div>
      <div class="text-xs text-neutral-500">
        检查时间：{{
          componentHealths[0]?.checkedAt
            ? new Date(componentHealths[0].checkedAt).toLocaleString()
            : '-'
        }}
      </div>
    </div>
  </div>
</template>
