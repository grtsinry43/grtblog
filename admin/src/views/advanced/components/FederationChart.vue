<script setup lang="ts">
import * as echarts from 'echarts'
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { toRefsPreferencesStore } from '@/stores'
import twc from '@/utils/tailwindColor'

import type { ECharts } from 'echarts'

const props = defineProps<{
  outboundByStatus?: Record<string, number>
}>()

const { isDark } = toRefsPreferencesStore()
const chartEl = ref<HTMLDivElement | null>(null)
let chart: ECharts | null = null

function render() {
  if (!chartEl.value) return
  if (!chart) chart = echarts.init(chartEl.value)
  const statusMap = props.outboundByStatus ?? {}
  const pieData = Object.entries(statusMap).map(([name, value]) => ({ name, value }))
  chart.setOption({
    tooltip: { trigger: 'item' },
    legend: {
      bottom: 0,
      textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] },
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '72%'],
        center: ['50%', '45%'],
        itemStyle: {
          borderRadius: 5,
          borderColor: isDark.value ? twc.neutral[800] : '#fff',
          borderWidth: 2,
        },
        label: {
          formatter: '{b}: {d}%',
          color: isDark.value ? twc.neutral[400] : twc.neutral[600],
        },
        data: pieData,
      },
    ],
  })
}

watch(() => props.outboundByStatus, () => nextTick(render), { deep: true })
watch(isDark, () => nextTick(render))

onMounted(() => {
  nextTick(render)
  window.addEventListener('resize', render)
})

onUnmounted(() => {
  window.removeEventListener('resize', render)
  chart?.dispose()
})
</script>

<template>
  <div class="flex flex-col rounded border border-naive-border bg-naive-card transition-[background-color,border-color]" style="height: 420px">
    <div class="flex items-center justify-between px-5 pt-4">
      <span class="text-base font-medium text-neutral-600 dark:text-neutral-300">联邦出站状态分布</span>
    </div>
    <div class="flex-1 px-4 pt-2 pb-4">
      <div ref="chartEl" class="h-full w-full" />
    </div>
  </div>
</template>
