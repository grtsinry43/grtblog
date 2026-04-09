<script setup lang="ts">
import { ArrowClockwise24Regular } from '@vicons/fluent'
import {
  NButton,
  NDescriptions,
  NDescriptionsItem,
  NEmpty,
  NIcon,
  NNumberAnimation,
  NProgress,
  NSelect,
  NTag,
  NSkeleton,
} from 'naive-ui'
import { watch } from 'vue'

import { PageHeader, ScrollContainer } from '@/components'

import AlertTimeline from './components/AlertTimeline.vue'
import FederationChart from './components/FederationChart.vue'
import SystemHealthCard from './components/SystemHealthCard.vue'
import TrafficChart from './components/TrafficChart.vue'
import { useObservability } from './composables/use-observability'

defineOptions({
  name: 'AdvancedInfo',
})

const {
  lastRefreshAt,
  timelineWindow,
  windowOptions,
  loading,
  controlData,
  realtimeData,
  federationData,
  storageData,
  alertsData,
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
} = useObservability()

watch(timelineWindow, () => queryClient.invalidateQueries({ queryKey: ['obs-timeline'] }))
</script>

<template>
  <ScrollContainer wrapper-class="p-4 md:p-6 space-y-4">
    <PageHeader
      title="高级信息 / Observability"
      icon="ph--desktop"
      class="mb-4"
    >
      <template #badge>
        <NTag
          size="small"
          type="info"
          round
          >实时</NTag
        >
      </template>
      <template #actions>
        <span class="mr-2 text-xs whitespace-nowrap text-neutral-400">
          最后刷新：{{ lastRefreshAt.toLocaleTimeString() }}
        </span>
        <NSelect
          v-model:value="timelineWindow"
          :options="windowOptions"
          size="small"
          class="w-32"
        />
        <NButton
          size="small"
          secondary
          :loading="loading"
          @click="refreshAll"
        >
          <template #icon><NIcon :component="ArrowClockwise24Regular" /></template>
          刷新
        </NButton>
      </template>
    </PageHeader>

    <!-- Top Cards -->
    <div class="grid grid-cols-1 gap-4 max-sm:gap-2 md:grid-cols-2 lg:grid-cols-4">
      <div
        v-for="(item, index) in cardStats"
        :key="index"
        class="flex items-center justify-between gap-x-4 overflow-hidden rounded border border-naive-border bg-naive-card p-6 transition-[background-color,border-color]"
      >
        <template v-if="!loading || item.value">
          <div class="flex-1">
            <span class="text-sm font-medium text-neutral-450">{{ item.title }}</span>
            <div class="mt-1 mb-1.5 flex gap-x-1 text-2xl text-neutral-700 dark:text-neutral-400">
              <NNumberAnimation
                :to="item.value"
                show-separator
                :precision="item.precision || 0"
              />
              <span class="mb-1 self-end text-xs text-neutral-400">{{ item.suffix }}</span>
            </div>
            <div class="flex items-center">
              <span class="text-xs text-neutral-500 dark:text-neutral-400">{{
                item.description
              }}</span>
            </div>
          </div>
          <div>
            <div
              class="grid place-items-center rounded-full p-3"
              :class="item.iconBgClass"
            >
              <span
                class="size-7"
                :class="item.iconClass"
              />
            </div>
          </div>
        </template>
        <template v-else>
          <div class="flex w-full gap-4">
            <div class="flex-1 space-y-2">
              <NSkeleton
                text
                style="width: 40%"
              />
              <NSkeleton
                text
                style="width: 80%; height: 28px"
              />
              <NSkeleton
                text
                style="width: 60%"
              />
            </div>
            <NSkeleton
              circle
              size="medium"
              style="width: 48px; height: 48px"
            />
          </div>
        </template>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="grid grid-cols-1 gap-4 overflow-hidden max-sm:gap-2 lg:grid-cols-12">
      <div class="col-span-1 lg:col-span-8">
        <TrafficChart :traffic-series="trafficSeries" />
      </div>
      <div class="col-span-1 lg:col-span-4">
        <FederationChart :outbound-by-status="federationData?.outboundByStatus" />
      </div>
    </div>

    <!-- Info Row -->
    <div class="grid grid-cols-1 gap-4 overflow-hidden max-sm:gap-2 lg:grid-cols-12">
      <!-- Control Plane -->
      <div class="col-span-1 lg:col-span-6">
        <div
          class="flex flex-col rounded border border-naive-border bg-naive-card p-5 transition-[background-color,border-color]"
        >
          <div class="mb-4 text-base font-medium text-neutral-600 dark:text-neutral-300">
            控制平面概况
          </div>
          <NDescriptions
            :column="2"
            size="small"
          >
            <NDescriptionsItem label="RPS">{{
              controlData?.api.rps?.toFixed(2)
            }}</NDescriptionsItem>
            <NDescriptionsItem label="P95 延迟"
              >{{ controlData?.api.p95LatencyMs?.toFixed(1) }} ms</NDescriptionsItem
            >
            <NDescriptionsItem label="API 错误率">{{
              formatPercent(controlData?.api.errorRate)
            }}</NDescriptionsItem>
            <NDescriptionsItem label="Go Goroutines">{{
              controlData?.goRuntime.numGoroutine
            }}</NDescriptionsItem>
            <NDescriptionsItem label="DB 连接状态">
              <NTag
                :type="controlData?.database.status === 'connected' ? 'success' : 'error'"
                size="small"
                round
              >
                {{ controlData?.database.status || 'unknown' }}
              </NTag>
            </NDescriptionsItem>
            <NDescriptionsItem label="DB 等待">{{
              controlData?.database.waitCount
            }}</NDescriptionsItem>
          </NDescriptions>
        </div>
      </div>

      <!-- Realtime & Storage -->
      <div class="col-span-1 lg:col-span-6">
        <div
          class="flex flex-col rounded border border-naive-border bg-naive-card p-5 transition-[background-color,border-color]"
        >
          <div class="mb-4 text-base font-medium text-neutral-600 dark:text-neutral-300">
            实时与存储
          </div>
          <NDescriptions
            :column="2"
            size="small"
          >
            <NDescriptionsItem label="WS 在线">{{
              realtimeData?.snapshot.currentOnline
            }}</NDescriptionsItem>
            <NDescriptionsItem label="WS 广播错误率">{{
              formatPercent(realtimeData?.snapshot.broadcastErrorRate)
            }}</NDescriptionsItem>
            <NDescriptionsItem label="WS Fanout P95"
              >{{ realtimeData?.snapshot.broadcastP95Ms?.toFixed(1) }} ms</NDescriptionsItem
            >
            <NDescriptionsItem label="平均接收人数">{{
              realtimeData?.snapshot.avgRecipients?.toFixed(2)
            }}</NDescriptionsItem>
            <NDescriptionsItem label="HTML 存储">{{
              formatBytes(storageData?.storageHtml.size)
            }}</NDescriptionsItem>
            <NDescriptionsItem label="日志存储">{{
              formatBytes(storageData?.storageLogs.size)
            }}</NDescriptionsItem>
          </NDescriptions>
          <div class="mt-4 border-t border-neutral-100 pt-4 dark:border-neutral-800">
            <div class="mb-1 flex items-center justify-between">
              <span class="text-xs text-neutral-500">Redis 队列深度</span>
              <span class="text-xs text-neutral-400">{{
                storageData?.redis.analyticsQueueDepth || 0
              }}</span>
            </div>
            <NProgress
              type="line"
              status="info"
              :percentage="
                Math.min(((storageData?.redis.analyticsQueueDepth || 0) / 1000) * 100, 100)
              "
              :show-indicator="false"
              processing
            />
          </div>
        </div>
      </div>
    </div>

    <SystemHealthCard
      :component-healths="componentHealths"
      :component-tag-type="componentTagType"
    />

    <!-- Update Check -->
    <div
      class="rounded border border-naive-border bg-naive-card p-5 transition-[background-color,border-color]"
    >
      <div class="mb-4 flex items-center justify-between">
        <div class="text-base font-medium text-neutral-600 dark:text-neutral-300">更新检查</div>
        <NTag
          size="small"
          :type="updateTagType() as any"
          round
        >
          {{
            updateInfo?.status === 'error'
              ? '检查失败'
              : updateInfo?.status === 'disabled'
                ? '已关闭'
                : updateInfo?.hasUpdate
                  ? '可更新'
                  : '已最新'
          }}
        </NTag>
      </div>
      <NEmpty
        v-if="!updateInfo"
        description="暂无更新信息"
      />
      <NDescriptions
        v-else
        :column="2"
        size="small"
      >
        <NDescriptionsItem label="当前版本">{{ updateInfo.currentVersion }}</NDescriptionsItem>
        <NDescriptionsItem label="更新通道">{{ updateInfo.channel }}</NDescriptionsItem>
        <NDescriptionsItem label="目标版本">
          {{ updateInfo.targetRelease?.tag || '-' }}
          <NTag
            v-if="updateInfo.targetRelease?.prerelease"
            class="ml-2"
            size="tiny"
            type="warning"
            round
            >测试版</NTag
          >
        </NDescriptionsItem>
        <NDescriptionsItem label="检查时间">{{
          updateInfo.checkedAt ? new Date(updateInfo.checkedAt).toLocaleString() : '-'
        }}</NDescriptionsItem>
      </NDescriptions>
      <div class="mt-3 flex items-center justify-between gap-2">
        <span class="text-xs text-neutral-500">{{
          updateInfo?.message || '版本来源：GitHub Releases / Git tags'
        }}</span>
        <NButton
          v-if="updateInfo?.releaseNotesUrl || updateInfo?.upgradeUrl"
          size="small"
          secondary
          tag="a"
          :href="updateInfo.releaseNotesUrl || updateInfo.upgradeUrl"
          target="_blank"
        >
          查看说明
        </NButton>
      </div>
    </div>

    <AlertTimeline :alerts="alertsData?.items" />
  </ScrollContainer>
</template>
