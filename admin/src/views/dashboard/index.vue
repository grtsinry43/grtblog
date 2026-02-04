<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { watchDebounced } from '@vueuse/core'
import chroma from 'chroma-js'
import * as echarts from 'echarts'
import { NNumberAnimation, NSkeleton } from 'naive-ui'
import { onMounted, watch, ref, computed, onUnmounted, nextTick } from 'vue'

import { ScrollContainer } from '@/components'
import { getDashboardStats, getHitokoto } from '@/services/stats'
import { toRefsPreferencesStore } from '@/stores'
import { toRefsUserStore } from '@/stores/user'
import twc from '@/utils/tailwindColor'

import type { ECharts } from 'echarts'

defineOptions({
  name: 'Dashboard',
})

const { sidebarMenu, navigationMode, themeColor, isDark } = toRefsPreferencesStore()
const { user } = toRefsUserStore()

// --- Data Fetching ---
const { data: stats, isLoading } = useQuery({
  queryKey: ['dashboard-stats'],
  queryFn: getDashboardStats,
  refetchInterval: 60000, 
})

const { data: hitokoto, isLoading: isHitokotoLoading } = useQuery({
  queryKey: ['hitokoto'],
  queryFn: getHitokoto,
  staleTime: 1000 * 60 * 60, // Cache for 1 hour
})

const greeting = computed(() => {
    const hour = new Date().getHours()
    if (hour < 6) return '夜深了'
    if (hour < 9) return '早上好'
    if (hour < 12) return '上午好'
    if (hour < 14) return '中午好'
    if (hour < 17) return '下午好'
    if (hour < 19) return '傍晚好'
    return '晚上好'
})

// --- Tabs State ---
const mainTrendTab = ref('traffic') // traffic | online | publishing
const distributionTab = ref('category') // category | column | words
const sourceTab = ref('platform') // platform | browser | location
const topContentTab = ref('articles') // articles | moments

// --- Computed Data Mappings ---
const cardList = computed(() => {
  const s = stats.value
  
  if (!s) return Array.from({ length: 4 }).map(() => ({ loading: true }))

  return [
    {
      title: '用户总数',
      value: s.overview.users,
      iconClass: 'iconify ph--users-bold text-indigo-50 dark:text-indigo-150',
      iconBgClass:
        'text-indigo-500/5 bg-indigo-400 ring-4 ring-indigo-200 dark:bg-indigo-650 dark:ring-indigo-500/30 transition-all',
      description: '注册用户总数',
      precision: 0,
    },
    {
      title: '总访问量',
      value: s.interaction.viewsTotal,
      iconClass: 'iconify ph--eye-bold text-blue-50 dark:text-blue-150',
      iconBgClass:
        'text-blue-500/5 bg-blue-400 ring-4 ring-blue-200 dark:bg-blue-650 dark:ring-blue-500/30 transition-all',
      description: '全站内容总浏览',
      precision: 0,
    },
    {
        title: '在线峰值',
        value: s.todayPeakOnline,
        iconClass: 'iconify ph--lightning-bold text-amber-50 dark:text-amber-150',
        iconBgClass: 'text-amber-500/5 bg-amber-400 ring-4 ring-amber-200 dark:bg-amber-650 dark:ring-amber-500/30 transition-all',
        description: '今日最高在线',
        precision: 0,
    },
    {
      title: '待办事项',
      value: s.pending.unviewedComments + s.pending.friendLinkApplications,
      iconClass: 'iconify ph--list-checks-bold text-orange-50 dark:text-orange-150',
      iconBgClass:
        'text-orange-500/5 bg-orange-400 ring-4 ring-orange-200 dark:bg-orange-650 dark:ring-orange-500/30 transition-all',
      description: '待审核评论与友链',
      precision: 0,
    },
  ]
})

// --- Chart Refs ---
const mainTrendChart = ref<HTMLDivElement | null>(null)
let mainTrendChartInstance: ECharts | null = null
let mainTrendChartResizeHandler: (() => void) | null = null

const distributionChart = ref<HTMLDivElement | null>(null)
let distributionChartInstance: ECharts | null = null
let distributionChartResizeHandler: (() => void) | null = null

const sourceChart = ref<HTMLDivElement | null>(null)
let sourceChartInstance: ECharts | null = null
let sourceChartResizeHandler: (() => void) | null = null

const topContentChart = ref<HTMLDivElement | null>(null)
let topContentChartInstance: ECharts | null = null
let topContentChartResizeHandler: (() => void) | null = null

// --- Chart Helpers ---
const createTooltipConfig = (formatter?: any) => ({
  trigger: 'axis',
  backgroundColor: isDark.value ? twc.neutral[750] : '#fff',
  borderWidth: 1,
  borderColor: isDark.value ? twc.neutral[700] : twc.neutral[150],
  padding: 8,
  extraCssText: 'box-shadow: none;',
  textStyle: {
    color: isDark.value ? twc.neutral[400] : twc.neutral[600],
    fontSize: 12,
  },
  axisPointer: {
    type: 'none',
  },
  ...(formatter && { formatter }),
})

// --- Main Trend Chart ---
function initMainTrendChart() {
    if (!mainTrendChart.value || !stats.value) return

    if (mainTrendChartInstance) {
        mainTrendChartInstance.dispose()
    }
    const chart = echarts.init(mainTrendChart.value)
    
    let option: any = {}
    const color = themeColor.value

    if (mainTrendTab.value === 'traffic') {
        const dates = stats.value.viewTrend.map(d => d.date)
        const views = stats.value.viewTrend.map(d => d.count)
        
        option = {
            tooltip: createTooltipConfig(),
            grid: { left: 20, right: 20, top: 20, bottom: 0, containLabel: true },
            xAxis: {
                type: 'category',
                boundaryGap: false,
                data: dates,
                axisLine: { show: false },
                axisTick: { show: false },
                axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600], fontSize: 11 },
            },
            yAxis: {
                type: 'value',
                axisLine: { show: false },
                axisTick: { show: false },
                axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600], fontSize: 11 },
                splitLine: { show: true, lineStyle: { color: isDark.value ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)', width: 1 } },
            },
            series: [{
                name: '访问量',
                type: 'line',
                smooth: true,
                symbol: 'none',
                data: views,
                lineStyle: { width: 3, color: color },
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        { offset: 0, color: chroma(color).alpha(0.2).hex() },
                        { offset: 1, color: chroma(color).alpha(0.02).hex() }
                    ])
                },
                itemStyle: { color: color }
            }]
        }
    } else if (mainTrendTab.value === 'online') {
         const data = stats.value.online24h
         const hours = data.map(d => d.hour.split(' ')[1])
         const peaks = data.map(d => d.peak)
         const avgs = data.map(d => Math.round(d.avg))

         option = {
            tooltip: createTooltipConfig(),
            legend: { 
                data: ['峰值', '平均'], 
                right: 0, top: 0,
                textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] } 
            },
            grid: { left: 20, right: 20, top: 30, bottom: 0, containLabel: true },
            xAxis: {
                type: 'category',
                data: hours,
                boundaryGap: false,
                axisLine: { show: false }, axisTick: { show: false },
                axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600], fontSize: 11 },
            },
            yAxis: {
                type: 'value',
                axisLine: { show: false }, axisTick: { show: false },
                splitLine: { show: true, lineStyle: { color: isDark.value ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)' } },
            },
            series: [
                {
                    name: '峰值',
                    type: 'line',
                    smooth: true,
                    showSymbol: false,
                    data: peaks,
                    lineStyle: { width: 2, color: twc.amber[500] },
                    itemStyle: { color: twc.amber[500] }
                },
                {
                    name: '平均',
                    type: 'line',
                    smooth: true,
                    showSymbol: false,
                    data: avgs,
                    lineStyle: { width: 2, color: twc.blue[500], type: 'dashed' },
                    itemStyle: { color: twc.blue[500] },
                     areaStyle: {
                        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                            { offset: 0, color: chroma(twc.blue[500]).alpha(0.1).hex() },
                            { offset: 1, color: chroma(twc.blue[500]).alpha(0.02).hex() }
                        ])
                    },
                }
            ]
         }
    } else if (mainTrendTab.value === 'publishing') {
         const data = stats.value.trend
         const dates = data.map(d => d.date)
         
         option = {
             tooltip: createTooltipConfig(),
             legend: { 
                 data: ['文章', '动态', '思考'],
                 right: 0, top: 0,
                 textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] }
             },
             grid: { left: 20, right: 20, top: 30, bottom: 0, containLabel: true },
             xAxis: {
                 type: 'category',
                 data: dates,
                 axisLine: { show: false }, axisTick: { show: false },
                 axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600], fontSize: 11 },
             },
             yAxis: {
                type: 'value',
                axisLine: { show: false }, axisTick: { show: false },
                splitLine: { show: true, lineStyle: { color: isDark.value ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)' } },
             },
             series: [
                 { name: '文章', type: 'bar', stack: 'total', data: data.map(d => d.articles), itemStyle: { color: twc.emerald[500] } },
                 { name: '动态', type: 'bar', stack: 'total', data: data.map(d => d.moments), itemStyle: { color: twc.sky[500] } },
                  { name: '思考', type: 'bar', stack: 'total', data: data.map(d => d.thinkings), itemStyle: { color: twc.purple[500], borderRadius: [2, 2, 0, 0] } },
             ]
         }
    }

    option.animationDuration = 500
    chart.setOption(option)
    mainTrendChartInstance = chart
    mainTrendChartResizeHandler = () => chart.resize()
    window.addEventListener('resize', mainTrendChartResizeHandler, { passive: true })
}

// --- Distribution Chart ---
function initDistributionChart() {
    if (!distributionChart.value || !stats.value) return
    if (distributionChartInstance) distributionChartInstance.dispose()
    
    const chart = echarts.init(distributionChart.value)
    
    let data: { value: number; name: string; itemStyle?: any }[] = []
    let name = ''

    if (distributionTab.value === 'words') {
        const s = stats.value.words
        name = '字数统计'
        data = [
            { value: s.articles, name: '文章', itemStyle: { color: twc.emerald[500] } },
            { value: s.moments, name: '动态', itemStyle: { color: twc.sky[500] } },
            { value: s.pages, name: '页面', itemStyle: { color: twc.amber[500] } },
            { value: s.thinkings, name: '思考', itemStyle: { color: twc.purple[500] } },
        ]
    } else {
        const sourceData = distributionTab.value === 'category' ? stats.value.categories : stats.value.columns
        // Take top 8 or all if less
        const topData = sourceData.slice(0, 8)
        name = distributionTab.value === 'category' ? '分类分布' : '专栏分布'
        // Generate colors
        const colors = [
             twc.cyan[500], twc.blue[500], twc.indigo[500], twc.violet[500],
             twc.fuchsia[500], twc.pink[500], twc.rose[500], twc.orange[500]
        ]
        data = topData.map((d, i) => ({
            value: d.count,
            name: d.name,
            itemStyle: { color: colors[i % colors.length] }
        }))
    }

    const option = {
        tooltip: { trigger: 'item' },
        legend: { 
            top: '5%', 
            left: 'center', 
            textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] } 
        },
        series: [{
            name: name,
            type: 'pie',
            radius: ['40%', '70%'],
            center: ['50%', '60%'],
            itemStyle: {
                borderRadius: 5,
                borderColor: isDark.value ? twc.neutral[800] : '#fff',
                borderWidth: 2
            },
            label: { show: false },
            emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
            data: data
        }]
    }

    chart.setOption(option)
    distributionChartInstance = chart
    distributionChartResizeHandler = () => chart.resize()
    window.addEventListener('resize', distributionChartResizeHandler, { passive: true })
}

// --- Source Chart ---
function initSourceChart() {
    if (!sourceChart.value || !stats.value) return
    if (sourceChartInstance) sourceChartInstance.dispose()

    const chart = echarts.init(sourceChart.value)
    let data: { value: number; name: string; itemStyle?: any }[] = []
    let name = ''

    if (sourceTab.value === 'platform') {
        data = stats.value.platformTop.map((d, i) => ({ value: d.count, name: d.name, itemStyle: { color: [twc.indigo[500], twc.blue[500], twc.sky[500], twc.cyan[500]][i % 4] } }))
        name = '系统分布'
    } else if (sourceTab.value === 'browser') {
        data = stats.value.browserTop.map((d, i) => ({ value: d.count, name: d.name, itemStyle: { color: [twc.teal[500], twc.emerald[500], twc.green[500], twc.lime[500]][i % 4] } }))
        name = '浏览器分布'
    } else if (sourceTab.value === 'location') {
        data = stats.value.locationTop.map((d, i) => ({ value: d.count, name: d.name, itemStyle: { color: [twc.rose[500], twc.pink[500], twc.fuchsia[500], twc.purple[500]][i % 4] } }))
        name = '地区分布'
    }

    const topData = data.slice(0, 8)

    const option = {
        tooltip: { trigger: 'item' },
        legend: { 
            top: '5%', 
            left: 'center', 
            textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] } 
        },
        series: [{
            name: name,
            type: 'pie',
            radius: ['40%', '70%'],
            center: ['50%', '60%'],
             itemStyle: {
                borderRadius: 5,
                borderColor: isDark.value ? twc.neutral[800] : '#fff',
                borderWidth: 2
            },
            label: { show: false },
            emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
            data: topData
        }]
    }

    chart.setOption(option)
    sourceChartInstance = chart
    sourceChartResizeHandler = () => chart.resize()
    window.addEventListener('resize', sourceChartResizeHandler, { passive: true })
}

// --- Top Content Chart ---
function initTopContentChart() {
    if (!topContentChart.value || !stats.value) return
    if (topContentChartInstance) topContentChartInstance.dispose()

    const chart = echarts.init(topContentChart.value)
    let data: any[] = []
    if (topContentTab.value === 'articles') data = stats.value.topArticles
    else if (topContentTab.value === 'moments') data = stats.value.topMoments
    else if (topContentTab.value === 'pages') data = stats.value.topPages
    else if (topContentTab.value === 'thinkings') data = stats.value.topThinkings
    const topData = data.slice(0, 8)
    
    // Gradient colors for rank 1-3
    const colors = [twc.red[500], twc.orange[500], twc.amber[500]]

    const option = {
        tooltip: createTooltipConfig(),
        grid: { left: 10, right: 30, top: 0, bottom: 0, containLabel: true },
        xAxis: { 
            type: 'value', 
            splitLine: { show: true, lineStyle: { type: 'dashed', color: isDark.value ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.05)' } } 
        },
        yAxis: {
            type: 'category',
            data: topData.map(d => d.title),
            inverse: true,
            axisLine: { show: false }, axisTick: { show: false },
            axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600], width: 180, overflow: 'truncate' },
        },
        series: [{
            name: '浏览',
            type: 'bar',
            data: topData.map(d => d.views),
            barWidth: 16,
            itemStyle: {
                borderRadius: [0, 4, 4, 0],
                color: (params: any) => {
                   return params.dataIndex < 3 ? colors[params.dataIndex] : twc.indigo[400]
                }
            },
            label: { show: true, position: 'right', formatter: '{@score}' }
        }]
    }

    chart.setOption(option)
    topContentChartInstance = chart
    topContentChartResizeHandler = () => chart.resize()
    window.addEventListener('resize', topContentChartResizeHandler, { passive: true })
}


// --- Watchers ---
watch([stats, isDark, themeColor], () => {
    nextTick(() => {
        initMainTrendChart()
        initDistributionChart()
        initSourceChart()
        initTopContentChart()
    })
})

watch(mainTrendTab, () => { nextTick(initMainTrendChart) })
watch(distributionTab, () => { nextTick(initDistributionChart) })
watch(sourceTab, () => { nextTick(initSourceChart) })
watch(topContentTab, () => { nextTick(initTopContentChart) })

watchDebounced([() => sidebarMenu.value, () => navigationMode.value], () => {
    mainTrendChartInstance?.resize()
    distributionChartInstance?.resize()
    sourceChartInstance?.resize()
    topContentChartInstance?.resize()
}, { debounce: 300 })

function disposeAll() {
    mainTrendChartInstance?.dispose(); window.removeEventListener('resize', mainTrendChartResizeHandler!)
    distributionChartInstance?.dispose(); window.removeEventListener('resize', distributionChartResizeHandler!)
    sourceChartInstance?.dispose(); window.removeEventListener('resize', sourceChartResizeHandler!)
    topContentChartInstance?.dispose(); window.removeEventListener('resize', topContentChartResizeHandler!)
}

onMounted(() => {
    if (stats.value) {
        initMainTrendChart()
        initDistributionChart()
        initSourceChart()
        initTopContentChart()
    }
})
onUnmounted(disposeAll)

</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-y-2 max-sm:gap-y-2">
    <!-- Welcome Section -->
    <div class="relative mt-4 mb-4 max-w-6xl">
        <div class="relative z-10 flex flex-col md:flex-row md:items-end gap-2 md:gap-12">
             <!-- Greeting -->
             <div class="flex flex-col gap-y-1 shrink-0">
                <div class="flex items-center gap-x-2 text-xs font-medium tracking-wider text-neutral-500 dark:text-neutral-400 uppercase">
                    <span>今天是{{ new Date().toLocaleDateString('zh-CN', { month: 'long', day: 'numeric', weekday: 'long' }) }}</span>
                </div>
                <h2 class="text-2xl font-light text-neutral-800 dark:text-neutral-100">
                    {{ greeting }}，<span class="font-normal">{{ user.nickname || user.username }}</span>
                </h2>
             </div>

             <!-- Quote -->
            <div class="relative max-w-2xl pl-8 md:pl-0">
                 <NSkeleton v-if="isHitokotoLoading" text style="width: 200px" />
                 <template v-else-if="hitokoto">
                    <div class="absolute left-2 -top-2 md:-left-1 md:-top-3 text-neutral-200 dark:text-neutral-700">
                        <span class="iconify ph--quotes-fill text-2xl opacity-50"></span>
                    </div>
                    <p class="relative z-10 font-serif text-sm leading-relaxed text-neutral-700 dark:text-neutral-300">
                        {{ hitokoto.sentence.hitokoto }}
                        <span class="ml-2 text-xs font-sans font-medium tracking-wider text-neutral-400 dark:text-neutral-500 uppercase">
                            —— {{ hitokoto.sentence.from_who ? hitokoto.sentence.from_who + ' ' : '' }}{{ hitokoto.sentence.from ? `《${hitokoto.sentence.from}》` : '' }}
                        </span>
                    </p>
                 </template>
            </div>
        </div>
    </div>

    <!-- Top Cards -->
    <div class="grid grid-cols-1 gap-4 max-sm:gap-2 md:grid-cols-2 lg:grid-cols-4">
      <div
        v-for="(item, index) in cardList"
        :key="index"
        class="flex items-center justify-between gap-x-4 overflow-hidden rounded border border-naive-border bg-naive-card p-6 transition-[background-color,border-color]"
      >
        <template v-if="!('loading' in item)">
             <div class="flex-1">
                <span class="text-sm font-medium text-neutral-450">{{ item.title }}</span>
                <div class="mt-1 mb-1.5 flex gap-x-4 text-2xl text-neutral-700 dark:text-neutral-400">
                    <NNumberAnimation :to="item.value" show-separator :precision="item.precision" />
                </div>
                <div class="flex items-center">
                    <span class="text-neutral-500 dark:text-neutral-400 text-xs">{{ item.description }}</span>
                </div>
            </div>
            <div>
            <div
                class="grid place-items-center rounded-full p-3"
                :class="item.iconBgClass"
            >
                <span class="size-7" :class="item.iconClass" />
            </div>
            </div>
        </template>
        <template v-else>
             <div class="w-full flex gap-4">
                 <div class="flex-1 space-y-2">
                     <NSkeleton text style="width: 40%" />
                     <NSkeleton text style="width: 80%; height: 28px" />
                     <NSkeleton text style="width: 60%" />
                 </div>
                 <NSkeleton circle size="medium" style="width: 48px; height: 48px" />
             </div>
        </template>
      </div>
    </div>

    <!-- Row 2: Main Trend & Distribution -->
    <div class="grid grid-cols-1 gap-4 overflow-hidden max-sm:gap-2 lg:grid-cols-12">
      <!-- Main Trend -->
      <div class="col-span-1 lg:col-span-8">
        <div class="flex flex-col rounded border border-naive-border bg-naive-card transition-[background-color,border-color]" style="height: 420px">
            <div class="flex items-center justify-between px-5 pt-4">
                 <span class="text-base font-medium text-neutral-600 dark:text-neutral-300">趋势分析</span>
                 <div class="flex items-center gap-x-1 rounded bg-neutral-100 p-0.5 dark:bg-neutral-800">
                    <button 
                        v-for="tab in [
                            { label: '流量', value: 'traffic' },
                            { label: '在线', value: 'online' },
                            { label: '发布', value: 'publishing' }
                        ]"
                        :key="tab.value"
                        @click="mainTrendTab = tab.value"
                        class="px-3 py-1 text-xs transition-all rounded-xs"
                        :class="mainTrendTab === tab.value 
                            ? 'bg-white text-neutral-700 shadow-sm dark:bg-neutral-700 dark:text-neutral-200' 
                            : 'text-neutral-400 hover:text-neutral-600 dark:hover:text-neutral-300'"
                    >
                        {{ tab.label }}
                    </button>
                 </div>
            </div>
            <div class="flex-1 px-4 pb-4 pt-2">
                 <div v-if="isLoading && !stats" class="flex h-full items-center justify-center"><NSkeleton text class="w-full h-full" /></div>
                 <div v-else ref="mainTrendChart" class="h-full w-full" />
            </div>
        </div>
      </div>
      
      <!-- Distribution -->
      <div class="col-span-1 lg:col-span-4">
        <div class="flex flex-col rounded border border-naive-border bg-naive-card transition-[background-color,border-color]" style="height: 420px">
             <div class="flex items-center justify-between px-5 pt-4">
                 <span class="text-base font-medium text-neutral-600 dark:text-neutral-300">内容构成</span>
                 <div class="flex items-center gap-x-1 rounded bg-neutral-100 p-0.5 dark:bg-neutral-800">
                    <button 
                        v-for="tab in [
                            { label: '分类', value: 'category' },
                            { label: '专栏', value: 'column' },
                            { label: '字数', value: 'words' }
                        ]"
                        :key="tab.value"
                        @click="distributionTab = tab.value"
                        class="px-3 py-1 text-xs transition-all rounded-xs"
                        :class="distributionTab === tab.value 
                            ? 'bg-white text-neutral-700 shadow-sm dark:bg-neutral-700 dark:text-neutral-200' 
                            : 'text-neutral-400 hover:text-neutral-600 dark:hover:text-neutral-300'"
                    >
                        {{ tab.label }}
                    </button>
                 </div>
            </div>
             <div class="flex-1 px-4 pb-4 pt-2">
                 <div v-if="isLoading && !stats" class="flex h-full items-center justify-center"><NSkeleton text class="w-full h-full" /></div>
                 <div v-else ref="distributionChart" class="h-full w-full" />
            </div>
        </div>
      </div>
    </div>

    <!-- Row 3: Source & Top Content -->
    <div class="grid grid-cols-1 gap-4 overflow-hidden max-sm:gap-2 lg:grid-cols-12">
      <!-- Source -->
      <div class="col-span-1 lg:col-span-5">
        <div class="flex flex-col rounded border border-naive-border bg-naive-card transition-[background-color,border-color]" style="height: 380px">
            <div class="flex items-center justify-between px-5 pt-4">
                 <span class="text-base font-medium text-neutral-600 dark:text-neutral-300">访问来源</span>
                 <div class="flex items-center gap-x-1 rounded bg-neutral-100 p-0.5 dark:bg-neutral-800">
                    <button 
                        v-for="tab in [
                            { label: '系统', value: 'platform' },
                            { label: '浏览器', value: 'browser' },
                            { label: '地区', value: 'location' }
                        ]"
                        :key="tab.value"
                        @click="sourceTab = tab.value"
                        class="px-3 py-1 text-xs transition-all rounded-xs"
                        :class="sourceTab === tab.value 
                            ? 'bg-white text-neutral-700 shadow-sm dark:bg-neutral-700 dark:text-neutral-200' 
                            : 'text-neutral-400 hover:text-neutral-600 dark:hover:text-neutral-300'"
                    >
                        {{ tab.label }}
                    </button>
                 </div>
            </div>
           <div class="flex-1 px-4 pb-4 pt-2">
                 <div v-if="isLoading && !stats" class="flex h-full items-center justify-center"><NSkeleton text class="w-full h-full" /></div>
                 <div v-else ref="sourceChart" class="h-full w-full" />
            </div>
        </div>
      </div>

      <!-- Top Content -->
      <div class="col-span-1 lg:col-span-7">
        <div class="flex flex-col rounded border border-naive-border bg-naive-card transition-[background-color,border-color]" style="height: 380px">
            <div class="flex items-center justify-between px-5 pt-4">
                 <span class="text-base font-medium text-neutral-600 dark:text-neutral-300">热门内容</span>
                 <div class="flex items-center gap-x-1 rounded bg-neutral-100 p-0.5 dark:bg-neutral-800">
                    <button 
                        v-for="tab in [
                            { label: '文章', value: 'articles' },
                            { label: '动态', value: 'moments' },
                            { label: '页面', value: 'pages' },
                            { label: '思考', value: 'thinkings' }
                        ]"
                        :key="tab.value"
                        @click="topContentTab = tab.value"
                        class="px-3 py-1 text-xs transition-all rounded-xs"
                        :class="topContentTab === tab.value 
                            ? 'bg-white text-neutral-700 shadow-sm dark:bg-neutral-700 dark:text-neutral-200' 
                            : 'text-neutral-400 hover:text-neutral-600 dark:hover:text-neutral-300'"
                    >
                        {{ tab.label }}
                    </button>
                 </div>
            </div>
           <div class="flex-1 px-4 pb-4 pt-2">
                 <div v-if="isLoading && !stats" class="flex h-full items-center justify-center"><NSkeleton text class="w-full h-full" /></div>
                 <div v-else ref="topContentChart" class="h-full w-full" />
            </div>
        </div>
      </div>
    </div>
  </ScrollContainer>
</template>
