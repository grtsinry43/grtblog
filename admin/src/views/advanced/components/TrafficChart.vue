<script setup lang="ts">
import chroma from 'chroma-js'
import * as echarts from 'echarts'
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { toRefsPreferencesStore } from '@/stores'
import twc from '@/utils/tailwindColor'

import type { ECharts } from 'echarts'

const props = defineProps<{
  trafficSeries: {
    xAxis: string[]
    pv: number[]
    online: number[]
    outbound: number[]
  }
}>()

const { isDark, themeColor } = toRefsPreferencesStore()
const chartEl = ref<HTMLDivElement | null>(null)
let chart: ECharts | null = null

const createTooltipConfig = () => ({
  trigger: 'axis',
  backgroundColor: isDark.value ? twc.neutral[750] : '#fff',
  borderWidth: 1,
  borderColor: isDark.value ? twc.neutral[700] : twc.neutral[150],
  padding: 8,
  extraCssText: 'box-shadow: none;',
  textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600], fontSize: 12 },
  axisPointer: { type: 'none' },
})

function render() {
  if (!chartEl.value) return
  if (!chart) chart = echarts.init(chartEl.value)
  const data = props.trafficSeries
  const color = themeColor.value

  chart.setOption({
    tooltip: createTooltipConfig(),
    legend: {
      data: ['PV', '在线峰值', '联合出站'],
      right: 0,
      top: 0,
      textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] },
    },
    grid: { left: 12, right: 16, top: 36, bottom: 8, containLabel: true },
    xAxis: {
      type: 'category',
      data: data.xAxis,
      axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] },
      axisLine: { show: false },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] },
      splitLine: { lineStyle: { color: isDark.value ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.08)' } },
    },
    series: [
      {
        name: 'PV',
        type: 'line',
        smooth: true,
        data: data.pv,
        lineStyle: { width: 3, color },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: chroma(color).alpha(0.2).hex() },
            { offset: 1, color: chroma(color).alpha(0.02).hex() },
          ]),
        },
        itemStyle: { color },
      },
      {
        name: '在线峰值',
        type: 'line',
        smooth: true,
        data: data.online,
        lineStyle: { width: 2, color: twc.amber[500] },
        itemStyle: { color: twc.amber[500] },
      },
      {
        name: '联邦出站',
        type: 'bar',
        data: data.outbound,
        itemStyle: { color: twc.emerald[500] },
      },
    ],
  })
}

watch(() => props.trafficSeries, () => nextTick(render), { deep: true })
watch([isDark, themeColor], () => nextTick(render))

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
      <span class="text-base font-medium text-neutral-600 dark:text-neutral-300">全链路趋势</span>
    </div>
    <div class="flex-1 px-4 pt-2 pb-4">
      <div ref="chartEl" class="h-full w-full" />
    </div>
  </div>
</template>
