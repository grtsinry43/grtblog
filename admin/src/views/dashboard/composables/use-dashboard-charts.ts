import { watchDebounced } from '@vueuse/core'
import chroma from 'chroma-js'
import * as echarts from 'echarts'
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { toRefsPreferencesStore } from '@/stores'
import twc from '@/utils/tailwindColor'

import type { Ref } from 'vue'
import type { ECharts } from 'echarts'

export function useDashboardCharts(stats: Ref<any>, isLoading: Ref<boolean>) {
  const { sidebarMenu, navigationMode, themeColor, isDark } = toRefsPreferencesStore()

  const mainTrendTab = ref('traffic')
  const distributionTab = ref('category')
  const sourceTab = ref('platform')
  const topContentTab = ref('articles')

  const mainTrendChart = ref<HTMLDivElement | null>(null)
  const distributionChart = ref<HTMLDivElement | null>(null)
  const sourceChart = ref<HTMLDivElement | null>(null)
  const topContentChart = ref<HTMLDivElement | null>(null)

  let mainTrendChartInstance: ECharts | null = null
  let distributionChartInstance: ECharts | null = null
  let sourceChartInstance: ECharts | null = null
  let topContentChartInstance: ECharts | null = null

  let mainTrendChartResizeHandler: (() => void) | null = null
  let distributionChartResizeHandler: (() => void) | null = null
  let sourceChartResizeHandler: (() => void) | null = null
  let topContentChartResizeHandler: (() => void) | null = null

  const createTooltipConfig = (formatter?: any) => ({
    trigger: 'axis',
    backgroundColor: isDark.value ? twc.neutral[750] : '#fff',
    borderWidth: 1,
    borderColor: isDark.value ? twc.neutral[700] : twc.neutral[150],
    padding: 8,
    extraCssText: 'box-shadow: none;',
    textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600], fontSize: 12 },
    axisPointer: { type: 'none' },
    ...(formatter && { formatter }),
  })

  function initMainTrendChart() {
    if (!mainTrendChart.value || !stats.value) return
    if (mainTrendChartInstance) mainTrendChartInstance.dispose()
    const chart = echarts.init(mainTrendChart.value)
    let option: any = {}
    const color = themeColor.value

    if (mainTrendTab.value === 'traffic') {
      const dates = stats.value.viewTrend.map((d: any) => d.date)
      const views = stats.value.viewTrend.map((d: any) => d.count)
      option = {
        tooltip: createTooltipConfig(),
        grid: { left: 20, right: 20, top: 20, bottom: 0, containLabel: true },
        xAxis: {
          type: 'category', boundaryGap: false, data: dates,
          axisLine: { show: false }, axisTick: { show: false },
          axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600], fontSize: 11 },
        },
        yAxis: {
          type: 'value', axisLine: { show: false }, axisTick: { show: false },
          axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600], fontSize: 11 },
          splitLine: { show: true, lineStyle: { color: isDark.value ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)', width: 1 } },
        },
        series: [{
          name: '访问量', type: 'line', smooth: true, symbol: 'none', data: views,
          lineStyle: { width: 3, color },
          areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{ offset: 0, color: chroma(color).alpha(0.2).hex() }, { offset: 1, color: chroma(color).alpha(0.02).hex() }]) },
          itemStyle: { color },
        }],
      }
    } else if (mainTrendTab.value === 'online') {
      const data = stats.value.online24h
      const hours = data.map((d: any) => d.hour.split(' ')[1])
      const peaks = data.map((d: any) => d.peak)
      const avgs = data.map((d: any) => Math.round(d.avg))
      option = {
        tooltip: createTooltipConfig(),
        legend: { data: ['峰值', '平均'], right: 0, top: 0, textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] } },
        grid: { left: 20, right: 20, top: 30, bottom: 0, containLabel: true },
        xAxis: { type: 'category', data: hours, boundaryGap: false, axisLine: { show: false }, axisTick: { show: false }, axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600], fontSize: 11 } },
        yAxis: { type: 'value', axisLine: { show: false }, axisTick: { show: false }, splitLine: { show: true, lineStyle: { color: isDark.value ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)' } } },
        series: [
          { name: '峰值', type: 'line', smooth: true, showSymbol: false, data: peaks, lineStyle: { width: 2, color: twc.amber[500] }, itemStyle: { color: twc.amber[500] } },
          { name: '平均', type: 'line', smooth: true, showSymbol: false, data: avgs, lineStyle: { width: 2, color: twc.blue[500], type: 'dashed' }, itemStyle: { color: twc.blue[500] }, areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{ offset: 0, color: chroma(twc.blue[500]).alpha(0.1).hex() }, { offset: 1, color: chroma(twc.blue[500]).alpha(0.02).hex() }]) } },
        ],
      }
    } else if (mainTrendTab.value === 'publishing') {
      const data = stats.value.trend
      const dates = data.map((d: any) => d.date)
      option = {
        tooltip: createTooltipConfig(),
        legend: { data: ['文章', '动态', '思考'], right: 0, top: 0, textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] } },
        grid: { left: 20, right: 20, top: 30, bottom: 0, containLabel: true },
        xAxis: { type: 'category', data: dates, axisLine: { show: false }, axisTick: { show: false }, axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600], fontSize: 11 } },
        yAxis: { type: 'value', axisLine: { show: false }, axisTick: { show: false }, splitLine: { show: true, lineStyle: { color: isDark.value ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)' } } },
        series: [
          { name: '文章', type: 'bar', stack: 'total', data: data.map((d: any) => d.articles), itemStyle: { color: twc.emerald[500] } },
          { name: '动态', type: 'bar', stack: 'total', data: data.map((d: any) => d.moments), itemStyle: { color: twc.sky[500] } },
          { name: '思考', type: 'bar', stack: 'total', data: data.map((d: any) => d.thinkings), itemStyle: { color: twc.purple[500], borderRadius: [2, 2, 0, 0] } },
        ],
      }
    }
    option.animationDuration = 500
    chart.setOption(option)
    mainTrendChartInstance = chart
    mainTrendChartResizeHandler = () => chart.resize()
    window.addEventListener('resize', mainTrendChartResizeHandler, { passive: true })
  }

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
      const topData = sourceData.slice(0, 8)
      name = distributionTab.value === 'category' ? '分类分布' : '专栏分布'
      const colors = [twc.cyan[500], twc.blue[500], twc.indigo[500], twc.violet[500], twc.fuchsia[500], twc.pink[500], twc.rose[500], twc.orange[500]]
      data = topData.map((d: any, i: number) => ({ value: d.count, name: d.name, itemStyle: { color: colors[i % colors.length] } }))
    }
    chart.setOption({
      tooltip: { trigger: 'item' },
      legend: { top: '5%', left: 'center', textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] } },
      series: [{ name, type: 'pie', radius: ['40%', '70%'], center: ['50%', '60%'], itemStyle: { borderRadius: 5, borderColor: isDark.value ? twc.neutral[800] : '#fff', borderWidth: 2 }, label: { show: false }, emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } }, data }],
    })
    distributionChartInstance = chart
    distributionChartResizeHandler = () => chart.resize()
    window.addEventListener('resize', distributionChartResizeHandler, { passive: true })
  }

  function initSourceChart() {
    if (!sourceChart.value || !stats.value) return
    if (sourceChartInstance) sourceChartInstance.dispose()
    const chart = echarts.init(sourceChart.value)
    let data: { value: number; name: string; itemStyle?: any }[] = []
    let name = ''
    if (sourceTab.value === 'platform') {
      data = stats.value.platformTop.map((d: any, i: number) => ({ value: d.count, name: d.name, itemStyle: { color: [twc.indigo[500], twc.blue[500], twc.sky[500], twc.cyan[500]][i % 4] } }))
      name = '系统分布'
    } else if (sourceTab.value === 'browser') {
      data = stats.value.browserTop.map((d: any, i: number) => ({ value: d.count, name: d.name, itemStyle: { color: [twc.teal[500], twc.emerald[500], twc.green[500], twc.lime[500]][i % 4] } }))
      name = '浏览器分布'
    } else if (sourceTab.value === 'location') {
      data = stats.value.locationTop.map((d: any, i: number) => ({ value: d.count, name: d.name, itemStyle: { color: [twc.rose[500], twc.pink[500], twc.fuchsia[500], twc.purple[500]][i % 4] } }))
      name = '地区分布'
    }
    const topData = data.slice(0, 8)
    chart.setOption({
      tooltip: { trigger: 'item' },
      legend: { top: '5%', left: 'center', textStyle: { color: isDark.value ? twc.neutral[400] : twc.neutral[600] } },
      series: [{ name, type: 'pie', radius: ['40%', '70%'], center: ['50%', '60%'], itemStyle: { borderRadius: 5, borderColor: isDark.value ? twc.neutral[800] : '#fff', borderWidth: 2 }, label: { show: false }, emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } }, data: topData }],
    })
    sourceChartInstance = chart
    sourceChartResizeHandler = () => chart.resize()
    window.addEventListener('resize', sourceChartResizeHandler, { passive: true })
  }

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
    const colors = [twc.red[500], twc.orange[500], twc.amber[500]]
    chart.setOption({
      tooltip: createTooltipConfig(),
      grid: { left: 10, right: 30, top: 0, bottom: 0, containLabel: true },
      xAxis: { type: 'value', splitLine: { show: true, lineStyle: { type: 'dashed', color: isDark.value ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.05)' } } },
      yAxis: { type: 'category', data: topData.map((d: any) => d.title), inverse: true, axisLine: { show: false }, axisTick: { show: false }, axisLabel: { color: isDark.value ? twc.neutral[400] : twc.neutral[600], width: 180, overflow: 'truncate' } },
      series: [{ name: '浏览', type: 'bar', data: topData.map((d: any) => d.views), barWidth: 16, itemStyle: { borderRadius: [0, 4, 4, 0], color: (params: any) => params.dataIndex < 3 ? colors[params.dataIndex] : twc.indigo[400] }, label: { show: true, position: 'right', formatter: '{@score}' } }],
    })
    topContentChartInstance = chart
    topContentChartResizeHandler = () => chart.resize()
    window.addEventListener('resize', topContentChartResizeHandler, { passive: true })
  }

  // Watchers
  watch([stats, isDark, themeColor], () => {
    nextTick(() => { initMainTrendChart(); initDistributionChart(); initSourceChart(); initTopContentChart() })
  })
  watch(mainTrendTab, () => nextTick(initMainTrendChart))
  watch(distributionTab, () => nextTick(initDistributionChart))
  watch(sourceTab, () => nextTick(initSourceChart))
  watch(topContentTab, () => nextTick(initTopContentChart))

  watchDebounced([() => sidebarMenu.value, () => navigationMode.value], () => {
    mainTrendChartInstance?.resize()
    distributionChartInstance?.resize()
    sourceChartInstance?.resize()
    topContentChartInstance?.resize()
  }, { debounce: 300 })

  onMounted(() => {
    if (stats.value) {
      initMainTrendChart(); initDistributionChart(); initSourceChart(); initTopContentChart()
    }
  })

  onUnmounted(() => {
    mainTrendChartInstance?.dispose(); if (mainTrendChartResizeHandler) window.removeEventListener('resize', mainTrendChartResizeHandler)
    distributionChartInstance?.dispose(); if (distributionChartResizeHandler) window.removeEventListener('resize', distributionChartResizeHandler)
    sourceChartInstance?.dispose(); if (sourceChartResizeHandler) window.removeEventListener('resize', sourceChartResizeHandler)
    topContentChartInstance?.dispose(); if (topContentChartResizeHandler) window.removeEventListener('resize', topContentChartResizeHandler)
  })

  return {
    mainTrendTab,
    distributionTab,
    sourceTab,
    topContentTab,
    mainTrendChart,
    distributionChart,
    sourceChart,
    topContentChart,
  }
}
