<script setup lang="ts">
import {
  NButton,
  NCard,
  NCode,
  NCollapse,
  NCollapseItem,
  NDescriptions,
  NDescriptionsItem,
  NEmpty,
  NPopconfirm,
  NScrollbar,
  NSpace,
  NSpin,
  NStatistic,
  NSwitch,
  NTag,
  NTimeline,
  NTimelineItem,
  useMessage,
} from 'naive-ui'
import { computed, onMounted, ref } from 'vue'

import { listSysConfigs, updateSysConfigs } from '@/services/sysconfig'
import {
  getTelemetrySnapshot,
  getTelemetryReportHistory,
  triggerTelemetryReport,
  resetTelemetryErrors,
} from '@/services/telemetry'
import { formatDate } from '@/utils/format'

import type {
  TelemetrySnapshot,
  TelemetryReportRecord,
} from '@/services/telemetry'

const message = useMessage()
const loading = ref(true)
const snapshot = ref<TelemetrySnapshot | null>(null)
const reportHistory = ref<TelemetryReportRecord[]>([])
const reporting = ref(false)
const enabled = ref(false)
const loadingToggle = ref(false)

async function fetchData() {
  loading.value = true
  try {
    const [snap, history, configTree] = await Promise.all([
      getTelemetrySnapshot(),
      getTelemetryReportHistory(),
      listSysConfigs(['telemetry.enabled']),
    ])
    snapshot.value = snap
    reportHistory.value = history.history || []
    // listSysConfigs returns a tree — find the key in root items or nested group items.
    const enabledCfg = configTree.items?.find((c) => c.key === 'telemetry.enabled')
      ?? (configTree.groups ?? []).flatMap((g) => g.items ?? []).find((c) => c.key === 'telemetry.enabled')
    enabled.value = enabledCfg?.value === 'true' || enabledCfg?.value === true
  } catch {
    message.error('加载遥测数据失败')
  } finally {
    loading.value = false
  }
}

async function toggleEnabled(val: boolean) {
  const prev = enabled.value
  loadingToggle.value = true
  enabled.value = val // optimistic update
  try {
    await updateSysConfigs([{ key: 'telemetry.enabled', value: String(val) }])
    message.success(val ? '遥测已启用' : '遥测已禁用')
    await fetchData()
  } catch {
    enabled.value = prev // rollback on failure
    message.error('切换失败')
  } finally {
    loadingToggle.value = false
  }
}

async function doReportNow() {
  reporting.value = true
  try {
    const rec = await triggerTelemetryReport()
    if (rec.status === 'success') {
      message.success('上报成功')
    } else {
      message.warning(`上报结果: ${rec.status} — ${rec.message}`)
    }
    await fetchData()
  } catch {
    message.error('上报请求失败')
  } finally {
    reporting.value = false
  }
}

async function doReset() {
  try {
    await resetTelemetryErrors()
    message.success('错误数据已清空')
    await fetchData()
  } catch {
    message.error('清空失败')
  }
}

const featureList = computed(() => {
  if (!snapshot.value) return []
  const f = snapshot.value.instance.features
  return [
    { label: '联合协议', enabled: f.federationEnabled },
    { label: 'ActivityPub', enabled: f.activityPubEnabled },
    { label: '邮件通知', enabled: f.emailEnabled },
    { label: 'Turnstile', enabled: f.turnstileEnabled },
    { label: '评论（关闭）', enabled: f.commentsDisabled },
  ]
})

function statusType(status: string): 'success' | 'error' | 'warning' | 'default' {
  switch (status) {
    case 'success': return 'success'
    case 'failed': return 'error'
    case 'skipped': return 'warning'
    default: return 'default'
  }
}

onMounted(() => fetchData())
</script>

<template>
  <div class="space-y-4">
    <!-- Header -->
    <NCard size="small" title="匿名遥测">
      <template #header-extra>
        <NSpace align="center" :size="16">
          <NSwitch
            :value="enabled"
            :loading="loadingToggle"
            @update:value="toggleEnabled"
          >
            <template #checked>启用上报</template>
            <template #unchecked>已禁用</template>
          </NSwitch>
          <NButton
            size="small"
            type="primary"
            :loading="reporting"
            :disabled="!enabled"
            @click="doReportNow"
          >
            立即上报
          </NButton>
          <NButton size="small" quaternary @click="fetchData">
            刷新
          </NButton>
        </NSpace>
      </template>
      <p class="text-sm opacity-60">
        匿名收集脱敏后的错误摘要和运行指标，帮助改进 GrtBlog。不包含任何个人信息、文章内容或访客数据。
        您可以在下方预览将要上报的完整数据。GrtBlog 是开源项目，遥测相关的所有代码均可在 GitHub 上查看、审计和提出问题。
      </p>
    </NCard>

    <NSpin v-if="loading && !snapshot" class="py-8" />

    <!-- Summary stats -->
    <div v-if="snapshot" class="grid grid-cols-2 gap-3 sm:grid-cols-4">
      <NCard size="small">
        <NStatistic label="唯一错误" :value="snapshot.summary.uniqueErrors" />
      </NCard>
      <NCard size="small">
        <NStatistic label="错误总数" :value="snapshot.summary.totalErrors" />
      </NCard>
      <NCard size="small">
        <NStatistic label="唯一 Panic" :value="snapshot.summary.uniquePanics" />
      </NCard>
      <NCard size="small">
        <NStatistic label="Panic 总数" :value="snapshot.summary.totalPanics" />
      </NCard>
    </div>

    <!-- Data preview -->
    <NCollapse v-if="snapshot">
      <NCollapseItem title="实例信息" name="instance">
        <NDescriptions :column="2" label-placement="left" bordered size="small">
          <NDescriptionsItem label="Instance ID">{{ snapshot.instance.instanceId }}</NDescriptionsItem>
          <NDescriptionsItem label="版本">{{ snapshot.instance.version }}</NDescriptionsItem>
          <NDescriptionsItem label="Go 版本">{{ snapshot.instance.goVersion }}</NDescriptionsItem>
          <NDescriptionsItem label="系统">{{ snapshot.instance.os }}/{{ snapshot.instance.arch }}</NDescriptionsItem>
          <NDescriptionsItem label="部署模式">{{ snapshot.instance.deployMode }}</NDescriptionsItem>
          <NDescriptionsItem label="运行时间">{{ Math.floor(snapshot.instance.uptimeSeconds / 3600) }}h</NDescriptionsItem>
        </NDescriptions>
        <div class="mt-2">
          <NSpace :size="6">
            <NTag
              v-for="f in featureList"
              :key="f.label"
              :type="f.enabled ? 'success' : 'default'"
              size="small"
            >
              {{ f.label }}
            </NTag>
          </NSpace>
        </div>
      </NCollapseItem>

      <NCollapseItem title="运行指标" name="metrics">
        <NDescriptions :column="2" label-placement="left" bordered size="small">
          <NDescriptionsItem label="文章数">{{ snapshot.metrics.content.articlesTotal }}</NDescriptionsItem>
          <NDescriptionsItem label="手记数">{{ snapshot.metrics.content.momentsTotal }}</NDescriptionsItem>
          <NDescriptionsItem label="评论数">{{ snapshot.metrics.content.commentsTotal }}</NDescriptionsItem>
          <NDescriptionsItem label="友链数">{{ snapshot.metrics.content.friendLinksTotal }}</NDescriptionsItem>
          <NDescriptionsItem label="请求总数 (24h)">{{ snapshot.metrics.traffic.requestTotal }}</NDescriptionsItem>
          <NDescriptionsItem label="5xx 错误率">{{ (snapshot.metrics.traffic.errorRate5xx * 100).toFixed(2) }}%</NDescriptionsItem>
          <NDescriptionsItem label="P95 延迟">{{ snapshot.metrics.traffic.p95LatencyMs.toFixed(0) }}ms</NDescriptionsItem>
          <NDescriptionsItem label="ISR 渲染">{{ snapshot.metrics.isr.renderSuccess }}/{{ snapshot.metrics.isr.renderTotal }}</NDescriptionsItem>
          <NDescriptionsItem label="联合投递 (24h)">{{ snapshot.metrics.federation.outboundTotal }}</NDescriptionsItem>
          <NDescriptionsItem label="WS 连接数">{{ snapshot.metrics.realtime.wsConnectionsCurrent }}</NDescriptionsItem>
        </NDescriptions>
      </NCollapseItem>

      <NCollapseItem title="错误摘要" name="errors">
        <NSpace vertical>
          <NPopconfirm @positive-click="doReset">
            <template #trigger>
              <NButton size="tiny" quaternary type="warning">清空错误数据</NButton>
            </template>
            确定清空所有已收集的错误数据？
          </NPopconfirm>
          <template v-if="snapshot.errors?.length">
            <NCard v-for="e in snapshot.errors" :key="e.fingerprint" size="small" class="!text-xs">
              <div class="flex items-center justify-between gap-2">
                <NTag size="tiny" :type="e.bizCode === 'SERVER_ERROR' ? 'error' : 'warning'">{{ e.bizCode || e.kind }}</NTag>
                <span class="opacity-50">x{{ e.count }}</span>
              </div>
              <div class="mt-1 font-mono text-xs opacity-70">{{ e.location }}</div>
              <div class="mt-1 text-xs opacity-50">{{ e.sampleMessage }}</div>
            </NCard>
          </template>
          <NEmpty v-else description="暂无错误" />
          <template v-if="snapshot.panics?.length">
            <div class="mt-2 text-sm font-medium">Panics</div>
            <NCard v-for="p in snapshot.panics" :key="p.fingerprint" size="small" class="!text-xs">
              <div class="flex items-center justify-between gap-2">
                <NTag size="tiny" type="error">PANIC</NTag>
                <span class="opacity-50">x{{ p.count }}</span>
              </div>
              <div class="mt-1 font-mono text-xs opacity-70">{{ p.location }}</div>
              <div class="mt-1 text-xs opacity-50">{{ p.sampleMessage }}</div>
            </NCard>
          </template>
        </NSpace>
      </NCollapseItem>

      <NCollapseItem title="原始 JSON" name="raw">
        <NScrollbar style="max-height: 400px">
          <NCode :code="JSON.stringify(snapshot, null, 2)" language="json" />
        </NScrollbar>
      </NCollapseItem>
    </NCollapse>

    <!-- Report history -->
    <NCard size="small" title="上报历史">
      <NTimeline v-if="reportHistory.length">
        <NTimelineItem
          v-for="(rec, idx) in reportHistory"
          :key="idx"
          :type="statusType(rec.status)"
          :title="rec.status.toUpperCase()"
          :time="formatDate(rec.timestamp)"
        >
          <span class="text-xs opacity-70">
            {{ rec.message }}
            <template v-if="rec.durationMs"> · {{ rec.durationMs }}ms</template>
            <template v-if="rec.statusCode"> · HTTP {{ rec.statusCode }}</template>
          </span>
        </NTimelineItem>
      </NTimeline>
      <NEmpty v-else description="暂无上报记录" />
    </NCard>
  </div>
</template>
