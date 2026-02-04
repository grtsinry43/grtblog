<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { getSystemLogs } from '@/services/system'
import { NButton, NCard, NEmpty, NSpin, NLog } from 'naive-ui'

const { data: logs, isLoading, isError, refetch } = useQuery({
  queryKey: ['systemLogs'],
  queryFn: getSystemLogs,
  refetchOnWindowFocus: false,
})
</script>

<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-gray-100">系统日志</h1>
      <NButton type="primary" size="small" :loading="isLoading" @click="() => refetch()">
        刷新日志
      </NButton>
    </div>

    <NCard title="" content-style="padding: 0;">
      <div class="h-[70vh] bg-[#1e1e1e] p-4 rounded-md overflow-hidden relative">
        <div v-if="isLoading && !logs" class="absolute inset-0 flex items-center justify-center z-10 bg-[#1e1e1e]/80">
          <NSpin size="large" />
        </div>
        
        <div v-if="isError" class="flex items-center justify-center h-full text-red-400">
          加载日志失败，请稍后重试
        </div>

        <NEmpty v-else-if="logs && logs.length === 0" description="暂无日志" class="h-full flex items-center justify-center text-gray-400" />

        <NLog v-else :log="logs?.join('\n') || ''" :rows="30" class="h-full text-xs font-mono" />
      </div>
    </NCard>
  </div>
</template>

<style scoped>
:deep(.n-log) {
  height: 100%;
}
</style>
