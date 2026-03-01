import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { useIntervalFn } from '@vueuse/core'
import { computed, ref } from 'vue'
import {
  getObservabilityAlerts,
  getObservabilityControlPlane,
  getObservabilityFederation,
  getObservabilityOverview,
  getObservabilityRealtime,
  getObservabilityStorage,
  getObservabilityTimeline,
} from '@/services/observability'
import { getSystemStatus, getSystemUpdateCheck } from '@/services/system'
import { formatBytes } from '@/utils/format'

export function useObservability() {
  const queryClient = useQueryClient()
  const lastRefreshAt = ref(new Date())
  const timelineWindow = ref<'24h' | '7d'>('24h')

  const windowOptions = [
    { label: '最近 24 小时', value: '24h' },
    { label: '最近 7 天', value: '7d' },
  ]

  const nowISO = () => new Date().toISOString()
  const sinceISO = computed(() => {
    const now = new Date()
    if (timelineWindow.value === '7d') {
      return new Date(now.getTime() - 7 * 24 * 3600 * 1000).toISOString()
    }
    return new Date(now.getTime() - 24 * 3600 * 1000).toISOString()
  })

  const { data: overviewData, isPending: overviewPending } = useQuery({
    queryKey: ['obs-overview'],
    queryFn: getObservabilityOverview,
    refetchInterval: 15000,
  })
  const { data: controlData, isPending: controlPending } = useQuery({
    queryKey: ['obs-control'],
    queryFn: () => getObservabilityControlPlane('5m'),
    refetchInterval: 15000,
  })
  const { data: realtimeData, isPending: realtimePending } = useQuery({
    queryKey: ['obs-realtime'],
    queryFn: getObservabilityRealtime,
    refetchInterval: 8000,
  })
  const { data: federationData, isPending: federationPending } = useQuery({
    queryKey: ['obs-federation'],
    queryFn: () => getObservabilityFederation('24h'),
    refetchInterval: 15000,
  })
  const { data: storageData, isPending: storagePending } = useQuery({
    queryKey: ['obs-storage'],
    queryFn: getObservabilityStorage,
    refetchInterval: 20000,
  })
  const { data: systemData, isPending: systemPending } = useQuery({
    queryKey: ['system-status-advanced'],
    queryFn: getSystemStatus,
    refetchInterval: 15000,
  })
  const { data: updateData, isPending: updatePending } = useQuery({
    queryKey: ['system-update-check'],
    queryFn: () => getSystemUpdateCheck(false),
    staleTime: 30 * 60 * 1000,
    refetchOnWindowFocus: false,
  })
  const { data: alertsData } = useQuery({
    queryKey: ['obs-alerts'],
    queryFn: () => getObservabilityAlerts(12),
    refetchInterval: 20000,
  })
  const { data: timelineData } = useQuery({
    queryKey: ['obs-timeline', timelineWindow],
    queryFn: () =>
      getObservabilityTimeline({
        since: sinceISO.value,
        until: nowISO(),
        group_by: timelineWindow.value === '7d' ? 'day' : 'hour',
      }),
    refetchInterval: 30000,
  })

  const loading = computed(
    () =>
      overviewPending.value ||
      controlPending.value ||
      realtimePending.value ||
      federationPending.value ||
      storagePending.value ||
      systemPending.value ||
      updatePending.value,
  )

  const componentHealths = computed(() => systemData.value?.components ?? [])
  const updateInfo = computed(() => updateData.value)

  function componentTagType(item: { healthy: boolean; status: string }) {
    if (item.healthy) return 'success'
    if (item.status === 'not_configured') return 'warning'
    return 'error'
  }

  function updateTagType() {
    const info = updateInfo.value
    if (!info) return 'default'
    if (info.status === 'error') return 'error'
    if (info.status === 'disabled') return 'warning'
    if (info.hasUpdate) return 'info'
    return 'success'
  }

  function formatPercent(value?: number) {
    if (value == null || Number.isNaN(value)) return '0%'
    return `${(value * 100).toFixed(2)}%`
  }

  const cardStats = computed(() => {
    const ov = overviewData.value
    return [
      {
        title: 'API 请求(5m)',
        value: ov?.api.requests ?? 0,
        suffix: 'req',
        iconClass: 'iconify ph--arrows-left-right-bold text-indigo-50 dark:text-indigo-150',
        iconBgClass: 'text-indigo-500/5 bg-indigo-400 ring-4 ring-indigo-200 dark:bg-indigo-650 dark:ring-indigo-500/30 transition-all',
        description: '最近5分钟请求',
      },
      {
        title: 'API 错误率',
        value: (ov?.api.errorRate ?? 0) * 100,
        suffix: '%',
        precision: 2,
        iconClass: 'iconify ph--warning-circle-bold text-rose-50 dark:text-rose-150',
        iconBgClass: 'text-rose-500/5 bg-rose-400 ring-4 ring-rose-200 dark:bg-rose-650 dark:ring-rose-500/30 transition-all',
        description: '接口调用异常比例',
      },
      {
        title: '在线连接',
        value: ov?.realtime.currentOnline ?? 0,
        suffix: 'ws',
        iconClass: 'iconify ph--users-three-bold text-blue-50 dark:text-blue-150',
        iconBgClass: 'text-blue-500/5 bg-blue-400 ring-4 ring-blue-200 dark:bg-blue-650 dark:ring-blue-500/30 transition-all',
        description: '实时 WebSocket 连接',
      },
      {
        title: '联合成功率(24h)',
        value: (ov?.federation.deliverySuccessRate ?? 0) * 100,
        suffix: '%',
        precision: 2,
        iconClass: 'iconify ph--planet-bold text-emerald-50 dark:text-emerald-150',
        iconBgClass: 'text-emerald-500/5 bg-emerald-400 ring-4 ring-emerald-200 dark:bg-emerald-650 dark:ring-emerald-500/30 transition-all',
        description: '联合投递成功率',
      },
    ]
  })

  const trafficSeries = computed(() => {
    const items = timelineData.value?.series ?? []
    const xSet = new Set<string>()
    const pvMap = new Map<string, number>()
    const onlineMap = new Map<string, number>()
    const outboundMap = new Map<string, number>()
    for (const item of items) {
      const x = new Date(item.timestamp).toLocaleString()
      xSet.add(x)
      if (item.metric === 'pv') pvMap.set(x, item.value)
      if (item.metric === 'online_peak_avg') onlineMap.set(x, item.value)
      if (item.metric === 'federation_outbound_total') outboundMap.set(x, item.value)
    }
    const xAxis = Array.from(xSet).sort((a, b) => new Date(a).getTime() - new Date(b).getTime())
    return {
      xAxis,
      pv: xAxis.map((x) => pvMap.get(x) ?? 0),
      online: xAxis.map((x) => onlineMap.get(x) ?? 0),
      outbound: xAxis.map((x) => outboundMap.get(x) ?? 0),
    }
  })

  function refreshAll() {
    lastRefreshAt.value = new Date()
    queryClient.invalidateQueries({ queryKey: ['obs-overview'] })
    queryClient.invalidateQueries({ queryKey: ['obs-control'] })
    queryClient.invalidateQueries({ queryKey: ['obs-realtime'] })
    queryClient.invalidateQueries({ queryKey: ['obs-federation'] })
    queryClient.invalidateQueries({ queryKey: ['obs-storage'] })
    queryClient.invalidateQueries({ queryKey: ['system-status-advanced'] })
    queryClient.invalidateQueries({ queryKey: ['obs-alerts'] })
    queryClient.invalidateQueries({ queryKey: ['obs-timeline'] })
  }

  useIntervalFn(refreshAll, 30000)

  return {
    lastRefreshAt,
    timelineWindow,
    windowOptions,
    loading,
    overviewData,
    controlData,
    realtimeData,
    federationData,
    storageData,
    alertsData,
    timelineData,
    componentHealths,
    updateInfo,
    cardStats,
    trafficSeries,
    componentTagType,
    updateTagType,
    formatPercent,
    formatBytes,
    refreshAll,
    queryClient,
  }
}
