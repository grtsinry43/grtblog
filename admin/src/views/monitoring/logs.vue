<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { ArrowClockwise24Regular, Desktop24Regular, Pause16Filled, Play16Filled } from '@vicons/fluent'
import { useIntervalFn, useScroll } from '@vueuse/core'
import { NButton, NCard, NEmpty, NIcon, NLog, NSpin, NTag } from 'naive-ui'
import { nextTick, ref, watch } from 'vue'

import { ScrollContainer } from '@/components'
import { getSystemLogs } from '@/services/system'
import { toRefsPreferencesStore } from '@/stores'

const { isDark } = toRefsPreferencesStore()
const logInstRef = ref<HTMLElement | null>(null)
const isAutoScroll = ref(true)

const { data: logs, isLoading, isError, refetch } = useQuery({
  queryKey: ['systemLogs'],
  queryFn: getSystemLogs,
  refetchOnWindowFocus: false,
})

// Auto-refresh every 10 seconds
const { pause, resume, isActive } = useIntervalFn(() => {
  refetch()
}, 10000)

// Watch logs to auto-scroll
watch(logs, () => {
  if (isAutoScroll.value) {
    nextTick(() => {
      scrollToBottom()
    })
  }
})

function scrollToBottom() {
  if (logInstRef.value) {
    const el = logInstRef.value as any
    if (el?.scrollTo) {
      el.scrollTo({ top: 999999, behavior: 'smooth' })
    }
  }
}

function handleScroll(e: Event) {
  const target = e.target as HTMLElement
  // Simple check: if we are near bottom (within 50px), enable auto-scroll, otherwise disable.
  const isAtBottom = target.scrollHeight - target.scrollTop - target.clientHeight < 50
  
  if (isAtBottom) {
    isAutoScroll.value = true
  } else {
    isAutoScroll.value = false
  }
}

watch(logInstRef, (val) => {
  if (val) {
    const scrollContainer = (val as any).$el?.querySelector('.n-scrollbar-container') || (val as any).$el
    if (scrollContainer) {
      scrollContainer.addEventListener('scroll', handleScroll)
    }
  }
})
</script>

<template>
  <ScrollContainer wrapper-class="p-4 md:p-6 space-y-4">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <NIcon :component="Desktop24Regular" class="text-primary text-xl" />
        <!-- User requested "fine" (thin) font -->
        <div class="text-lg font-light">系统日志 / System Logs</div>
      </div>
      <div class="flex items-center gap-2">
         <NTag size="small" :type="isAutoScroll ? 'success' : 'warning'" round class="cursor-pointer" @click="isAutoScroll = !isAutoScroll">
          <template #icon>
            <NIcon :component="isAutoScroll ? Play16Filled : Pause16Filled" />
          </template>
          {{ isAutoScroll ? '自动追踪' : '已暂停追踪' }}
        </NTag>
        <div class="h-4 w-[1px] bg-neutral-200 dark:bg-neutral-800 mx-1"></div>
        <NButton size="small" secondary :loading="isLoading" @click="() => refetch()">
          <template #icon><NIcon :component="ArrowClockwise24Regular" /></template>
          刷新日志
        </NButton>
      </div>
    </div>

    <NCard :bordered="false" content-style="padding: 0;" class="overflow-hidden rounded-lg bg-white dark:bg-[#18181c]">
      <div class="h-[75vh] relative group">
        <!-- Loading State -->
        <div
          v-if="isLoading && !logs"
          class="absolute inset-0 flex items-center justify-center z-10 bg-white/80 dark:bg-[#18181c]/80 backdrop-blur-sm transition-opacity"
        >
          <NSpin size="large" />
        </div>

        <!-- Error State -->
        <div v-if="isError" class="absolute inset-0 flex flex-col items-center justify-center text-rose-500 gap-2">
          <div class="text-lg">加载日志失败</div>
          <NButton size="small" secondary type="error" @click="() => refetch()">重试</NButton>
        </div>

        <!-- Empty State -->
        <NEmpty
          v-else-if="logs && logs.length === 0"
          description="暂无日志数据"
          class="absolute inset-0 flex items-center justify-center"
        />

        <!-- Log Viewer -->
        <div v-else class="h-full relative">
          <NLog
            ref="logInstRef"
            :log="logs?.join('\n') || ''"
            :rows="30"
            class="h-full text-xs font-mono p-4"
            :class="[
              isDark 
                ? 'bg-[#101014] text-gray-300' 
                : 'bg-gray-50 text-gray-700 border border-gray-100'
            ]"
          />
          
          <!-- Scroll to Bottom Button (Visible when not auto-scrolling) -->
           <div 
            v-if="!isAutoScroll"
            class="absolute bottom-4 right-8 z-20"
           >
             <NButton size="small" type="primary" secondary round @click="scrollToBottom(); isAutoScroll = true">
                <template #icon><NIcon :component="Play16Filled" /></template>
                回到底部
             </NButton>
           </div>
        </div>
      </div>
    </NCard>
  </ScrollContainer>
</template>

<style scoped>
:deep(.n-log) {
  height: 100%;
}
/* Ensure code font looks good */
:deep(.n-log pre) {
  font-family: 'JetBrains Mono', 'Fira Code', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace !important;
}
</style>
