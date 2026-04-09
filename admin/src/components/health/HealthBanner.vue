<script setup lang="ts">
import { computed, ref } from 'vue'

import { useHealthStore } from '@/stores'

const healthStore = useHealthStore()

const expanded = ref(false)

const modeConfig = computed(() => {
  switch (healthStore.mode) {
    case 'maintenance':
      return {
        icon: 'ph--wrench-bold',
        label: '维护模式',
        desc: '站点当前处于维护状态',
        bg: 'bg-amber-500/10 border-amber-500/30',
        text: 'text-amber-600 dark:text-amber-400',
        iconColor: 'text-amber-500',
      }
    case 'degraded':
      return {
        icon: 'ph--warning-bold',
        label: '服务降级',
        desc: '部分非核心组件异常，功能可能受限',
        bg: 'bg-orange-500/10 border-orange-500/30',
        text: 'text-orange-600 dark:text-orange-400',
        iconColor: 'text-orange-500',
      }
    case 'critical':
      return {
        icon: 'ph--warning-octagon-bold',
        label: '严重故障',
        desc: '数据库不可用，服务严重受限',
        bg: 'bg-red-500/10 border-red-500/30',
        text: 'text-red-600 dark:text-red-400',
        iconColor: 'text-red-500',
      }
    case 'outage':
      return {
        icon: 'ph--x-circle-bold',
        label: '完全宕机',
        desc: 'Backend 或 Nginx 不可用',
        bg: 'bg-red-600/10 border-red-600/30',
        text: 'text-red-700 dark:text-red-300',
        iconColor: 'text-red-600',
      }
    default:
      return null
  }
})

const stateBinaryStr = computed(() => {
  return healthStore.state.toString(2).padStart(6, '0')
})

const componentList = computed(() => {
  const c = healthStore.components
  return [
    { name: 'Nginx', ok: c.nginx },
    { name: 'Backend', ok: c.backend },
    { name: 'Database', ok: c.database },
    { name: 'Redis', ok: c.redis },
    { name: 'Renderer', ok: c.renderer },
  ]
})
</script>

<template>
  <Transition
    enter-active-class="transition-all duration-300 ease-out"
    enter-from-class="max-h-0 -translate-y-2 opacity-0"
    enter-to-class="max-h-40 translate-y-0 opacity-100"
    leave-active-class="transition-all duration-200 ease-in"
    leave-from-class="max-h-40 translate-y-0 opacity-100"
    leave-to-class="max-h-0 -translate-y-2 opacity-0"
  >
    <div
      v-if="healthStore.showBanner && modeConfig"
      class="overflow-hidden border-b px-4 py-2 text-sm transition-[background-color,border-color]"
      :class="[modeConfig.bg]"
    >
      <div class="flex items-center justify-between">
        <div
          class="flex items-center gap-2"
          :class="modeConfig.text"
        >
          <span
            class="iconify text-base"
            :class="modeConfig.icon"
          />
          <span class="font-medium">{{ modeConfig.label }}</span>
          <span class="opacity-70">—</span>
          <span class="opacity-70">{{ modeConfig.desc }}</span>
        </div>
        <button
          class="flex items-center gap-1 rounded px-2 py-0.5 text-xs opacity-60 transition-opacity hover:opacity-100"
          :class="modeConfig.text"
          @click="expanded = !expanded"
        >
          {{ expanded ? '收起' : '详情' }}
          <span
            class="iconify transition-transform"
            :class="expanded ? 'ph--caret-up-bold' : 'ph--caret-down-bold'"
          />
        </button>
      </div>

      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="max-h-0 opacity-0"
        enter-to-class="max-h-20 opacity-100"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="max-h-20 opacity-100"
        leave-to-class="max-h-0 opacity-0"
      >
        <div
          v-if="expanded"
          class="mt-2 overflow-hidden"
          :class="modeConfig.text"
        >
          <div class="flex items-center gap-4 text-xs">
            <span class="font-mono opacity-70">状态码: {{ stateBinaryStr }}</span>
            <span class="opacity-70">模式: {{ healthStore.mode }}</span>
          </div>
          <div class="mt-1 flex flex-wrap gap-x-3 gap-y-0.5 text-xs">
            <span
              v-for="comp in componentList"
              :key="comp.name"
              class="flex items-center gap-1"
            >
              <span
                class="iconify text-xs"
                :class="
                  comp.ok
                    ? 'text-green-500 ph--check-circle-bold'
                    : 'text-red-500 ph--x-circle-bold'
                "
              />
              {{ comp.name }}
            </span>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>
