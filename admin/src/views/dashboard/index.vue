<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { NNumberAnimation, NSkeleton } from 'naive-ui'
import { computed } from 'vue'

import { ChartPanel, ScrollContainer } from '@/components'
import { getDashboardStats, getHitokoto } from '@/services/stats'
import { toRefsUserStore } from '@/stores/user'

import { useDashboardCharts } from './composables/use-dashboard-charts'

defineOptions({
  name: 'Dashboard',
})

const { user } = toRefsUserStore()

const { data: stats, isLoading } = useQuery({
  queryKey: ['dashboard-stats'],
  queryFn: getDashboardStats,
  refetchInterval: 60000,
})

const { data: hitokoto, isLoading: isHitokotoLoading } = useQuery({
  queryKey: ['hitokoto'],
  queryFn: getHitokoto,
  staleTime: 1000 * 60 * 60,
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

const cardList = computed(() => {
  const s = stats.value
  if (!s) return Array.from({ length: 4 }).map(() => ({ loading: true }))
  return [
    {
      title: '用户总数',
      value: s.overview.users,
      precision: 0,
      iconClass: 'iconify ph--users-bold text-indigo-50 dark:text-indigo-150',
      iconBgClass:
        'text-indigo-500/5 bg-indigo-400 ring-4 ring-indigo-200 dark:bg-indigo-650 dark:ring-indigo-500/30 transition-all',
      description: '注册用户总数',
    },
    {
      title: '总访问量',
      value: s.interaction.viewsTotal,
      precision: 0,
      iconClass: 'iconify ph--eye-bold text-blue-50 dark:text-blue-150',
      iconBgClass:
        'text-blue-500/5 bg-blue-400 ring-4 ring-blue-200 dark:bg-blue-650 dark:ring-blue-500/30 transition-all',
      description: '全站内容总浏览',
    },
    {
      title: '在线峰值',
      value: s.todayPeakOnline,
      precision: 0,
      iconClass: 'iconify ph--lightning-bold text-amber-50 dark:text-amber-150',
      iconBgClass:
        'text-amber-500/5 bg-amber-400 ring-4 ring-amber-200 dark:bg-amber-650 dark:ring-amber-500/30 transition-all',
      description: '今日最高在线',
    },
    {
      title: '待办事项',
      value: s.pending.unviewedComments + s.pending.friendLinkApplications,
      precision: 0,
      iconClass: 'iconify ph--list-checks-bold text-orange-50 dark:text-orange-150',
      iconBgClass:
        'text-orange-500/5 bg-orange-400 ring-4 ring-orange-200 dark:bg-orange-650 dark:ring-orange-500/30 transition-all',
      description: '待审核评论与友链',
    },
  ]
})

const {
  mainTrendTab,
  distributionTab,
  sourceTab,
  topContentTab,
  mainTrendChart,
  distributionChart,
  sourceChart,
  topContentChart,
} = useDashboardCharts(stats, isLoading)

const mainTrendTabs = [
  { label: '流量', value: 'traffic' },
  { label: '在线', value: 'online' },
  { label: '发布', value: 'publishing' },
]
const distributionTabs = [
  { label: '分类', value: 'category' },
  { label: '专栏', value: 'column' },
  { label: '字数', value: 'words' },
]
const sourceTabs = [
  { label: '系统', value: 'platform' },
  { label: '浏览器', value: 'browser' },
  { label: '地区', value: 'location' },
]
const topContentTabs = [
  { label: '文章', value: 'articles' },
  { label: '动态', value: 'moments' },
  { label: '页面', value: 'pages' },
  { label: '思考', value: 'thinkings' },
]
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-y-2 max-sm:gap-y-2">
    <!-- Welcome Section -->
    <div class="relative mt-4 mb-4 max-w-6xl">
      <div class="relative z-10 flex flex-col gap-2 md:flex-row md:items-end md:gap-12">
        <div class="flex shrink-0 flex-col gap-y-1">
          <div
            class="flex items-center gap-x-2 text-xs font-medium tracking-wider text-neutral-500 uppercase dark:text-neutral-400"
          >
            <span
              >今天是{{
                new Date().toLocaleDateString('zh-CN', {
                  month: 'long',
                  day: 'numeric',
                  weekday: 'long',
                })
              }}</span
            >
          </div>
          <h2 class="text-2xl font-light text-neutral-800 dark:text-neutral-100">
            {{ greeting }}，<span class="font-normal">{{ user.nickname || user.username }}</span>
          </h2>
        </div>
        <div class="relative max-w-2xl pl-8 md:pl-0">
          <NSkeleton
            v-if="isHitokotoLoading"
            text
            style="width: 200px"
          />
          <template v-else-if="hitokoto">
            <div
              class="absolute -top-2 left-2 text-neutral-200 md:-top-3 md:-left-1 dark:text-neutral-700"
            >
              <span class="iconify text-2xl opacity-50 ph--quotes-fill" />
            </div>
            <p
              class="relative z-10 font-serif text-sm leading-relaxed text-neutral-700 dark:text-neutral-300"
            >
              {{ hitokoto.sentence.hitokoto }}
              <span
                class="ml-2 font-sans text-xs font-medium tracking-wider text-neutral-400 uppercase dark:text-neutral-500"
              >
                —— {{ hitokoto.sentence.from_who ? hitokoto.sentence.from_who + ' ' : ''
                }}{{ hitokoto.sentence.from ? `《${hitokoto.sentence.from}》` : '' }}
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
              <NNumberAnimation
                :to="item.value"
                show-separator
                :precision="item.precision"
              />
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

    <!-- Row 2: Main Trend & Distribution -->
    <div class="grid grid-cols-1 gap-4 overflow-hidden max-sm:gap-2 lg:grid-cols-12">
      <div class="col-span-1 lg:col-span-8">
        <ChartPanel
          title="趋势分析"
          :tabs="mainTrendTabs"
          v-model:active-tab="mainTrendTab"
          :loading="isLoading && !stats"
        >
          <div
            ref="mainTrendChart"
            class="h-full w-full"
          />
        </ChartPanel>
      </div>
      <div class="col-span-1 lg:col-span-4">
        <ChartPanel
          title="内容构成"
          :tabs="distributionTabs"
          v-model:active-tab="distributionTab"
          :loading="isLoading && !stats"
        >
          <div
            ref="distributionChart"
            class="h-full w-full"
          />
        </ChartPanel>
      </div>
    </div>

    <!-- Row 3: Source & Top Content -->
    <div class="grid grid-cols-1 gap-4 overflow-hidden max-sm:gap-2 lg:grid-cols-12">
      <div class="col-span-1 lg:col-span-5">
        <ChartPanel
          title="访问来源"
          :tabs="sourceTabs"
          v-model:active-tab="sourceTab"
          :height="380"
          :loading="isLoading && !stats"
        >
          <template #header-extra>
            <span
              class="rounded bg-amber-100 px-2 py-0.5 text-[11px] text-amber-700 dark:bg-amber-900/40 dark:text-amber-300"
              >浏览埋点聚合</span
            >
          </template>
          <div
            ref="sourceChart"
            class="h-full w-full"
          />
        </ChartPanel>
      </div>
      <div class="col-span-1 lg:col-span-7">
        <ChartPanel
          title="热门内容"
          :tabs="topContentTabs"
          v-model:active-tab="topContentTab"
          :height="380"
          :loading="isLoading && !stats"
        >
          <div
            ref="topContentChart"
            class="h-full w-full"
          />
        </ChartPanel>
      </div>
    </div>
  </ScrollContainer>
</template>
