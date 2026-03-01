import * as echarts from 'echarts'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { getVisitorInsights } from '@/services/visitors'

import type { ECharts } from 'echarts'
import type { VisitorInsights } from '@/types/visitors'

export function useVisitorInsights(message: { error: (m: string) => void }) {
  const insightDays = ref<number>(30)
  const sourceTab = ref<'platform' | 'browser' | 'location'>('platform')
  const insightsLoading = ref(false)
  const insights = ref<VisitorInsights | null>(null)

  const sourceChartRef = ref<HTMLDivElement | null>(null)
  const trendChartRef = ref<HTMLDivElement | null>(null)
  const funnelChartRef = ref<HTMLDivElement | null>(null)
  let sourceChart: ECharts | null = null
  let trendChart: ECharts | null = null
  let funnelChart: ECharts | null = null

  const daysOptions = [
    { label: '最近 7 天', value: 7 },
    { label: '最近 30 天', value: 30 },
    { label: '最近 90 天', value: 90 },
  ]

  const sourceSeries = computed(() => {
    if (!insights.value) return []
    if (sourceTab.value === 'platform') return insights.value.platformTop
    if (sourceTab.value === 'browser') return insights.value.browserTop
    return insights.value.locationTop
  })

  const dataSourceLabel = computed(() => {
    if (!insights.value) return '-'
    return insights.value.dataSource === 'api' ? '用户行为埋点聚合' : '浏览埋点聚合'
  })

  function toPercent(value: number) {
    return `${(value * 100).toFixed(1)}%`
  }

  function renderSourceChart() {
    if (!sourceChartRef.value) return
    sourceChart?.dispose()
    sourceChart = echarts.init(sourceChartRef.value)
    sourceChart.setOption({
      tooltip: { trigger: 'item' },
      legend: { top: 4 },
      series: [{ type: 'pie', radius: ['38%', '66%'], center: ['50%', '58%'], data: sourceSeries.value.map((item) => ({ name: item.name, value: item.count })) }],
    })
  }

  function renderTrendChart() {
    if (!trendChartRef.value || !insights.value) return
    trendChart?.dispose()
    trendChart = echarts.init(trendChartRef.value)
    trendChart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { top: 4 },
      grid: { left: 28, right: 20, top: 30, bottom: 20, containLabel: true },
      xAxis: { type: 'category', data: insights.value.trend.map((item) => item.date.slice(5)) },
      yAxis: { type: 'value' },
      series: [
        { name: '活跃访客', type: 'line', smooth: true, data: insights.value.trend.map((item) => item.activeVisitors) },
        { name: '浏览', type: 'line', smooth: true, data: insights.value.trend.map((item) => item.views) },
        { name: '点赞', type: 'line', smooth: true, data: insights.value.trend.map((item) => item.likes) },
        { name: '评论', type: 'line', smooth: true, data: insights.value.trend.map((item) => item.comments) },
      ],
    })
  }

  function renderFunnelChart() {
    if (!funnelChartRef.value || !insights.value) return
    funnelChart?.dispose()
    funnelChart = echarts.init(funnelChartRef.value)
    funnelChart.setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: ['浏览访客', '点赞访客', '评论访客'] },
      yAxis: { type: 'value' },
      series: [{
        type: 'bar',
        data: [insights.value.funnel.viewVisitors, insights.value.funnel.likeVisitors, insights.value.funnel.commentVisitors],
      }],
    })
  }

  function renderCharts() {
    renderSourceChart()
    renderTrendChart()
    renderFunnelChart()
  }

  async function loadInsights() {
    insightsLoading.value = true
    try {
      insights.value = await getVisitorInsights(insightDays.value)
      await nextTick()
      renderCharts()
    } catch (error: any) {
      message.error(error?.message || '获取访客统计失败')
    } finally {
      insightsLoading.value = false
    }
  }

  watch(insightDays, () => loadInsights())
  watch(sourceTab, async () => {
    if (!insights.value) return
    await nextTick()
    renderSourceChart()
  })

  onMounted(async () => {
    await loadInsights()
    window.addEventListener('resize', renderCharts)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', renderCharts)
    sourceChart?.dispose()
    trendChart?.dispose()
    funnelChart?.dispose()
  })

  return {
    insightDays,
    sourceTab,
    insightsLoading,
    insights,
    sourceChartRef,
    trendChartRef,
    funnelChartRef,
    daysOptions,
    dataSourceLabel,
    toPercent,
  }
}
