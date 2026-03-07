<script setup lang="ts">
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { isEmpty } from 'lodash-es'
import { NScrollbar, useDialog } from 'naive-ui'
import { computed, defineAsyncComponent, h, onMounted, onUnmounted, watch } from 'vue'

import texturePng from '@/assets/texture.png'
import { CollapseTransition, EmptyPlaceholder } from '@/components'
import HealthBanner from '@/components/health/HealthBanner.vue'
import { useInjection } from '@/composables'
import { createMarkdownIt } from '@/composables/markdown-it/core'
import { mediaQueryInjectionKey, layoutInjectionKey } from '@/injection'
import router from '@/router'
import { adminRealtimeWSCore } from '@/services/realtime-ws'
import { getSystemUpdateCheck } from '@/services/system'
import { DEFAULT_PREFERENCES_OPTIONS, toRefsPreferencesStore, toRefsTabsStore, toRefsUserStore, useRealtimeStore, useHealthStore, useUserStore } from '@/stores'

import FooterLayout from './footer/index.vue'
import HeaderLayout from './header/index.vue'
import MainLayout from './main/index.vue'
import Tabs from './tabs/index.vue'

import type { HealthWSPayload } from '@/services/health'
import type { OwnerStatusPayload } from '@/services/owner-status'

defineOptions({
  name: 'Layout',
})

const {
  preferences,
  sidebarMenu,
  navigationMode,
  showFooter,
  tabs: tabsOptions,
  backgroundImage,
} = toRefsPreferencesStore()
const { token, user } = toRefsUserStore()
const userStore = useUserStore()
const realtimeStore = useRealtimeStore()
const healthStore = useHealthStore()
const queryClient = useQueryClient()
const dialog = useDialog()
const releaseNotesMd = createMarkdownIt({
  options: {
    html: false,
    linkify: true,
    breaks: true,
  },
})

const UPDATE_DIALOG_ACK_KEY = 'grtblog:update-dialog:last-seen'

const { data: updateInfo } = useQuery({
  queryKey: ['system-update-check'],
  queryFn: () => getSystemUpdateCheck(false),
  staleTime: 30 * 60 * 1000,
  refetchOnWindowFocus: false,
})

function getUpdateDialogVersionKey() {
  const info = updateInfo.value
  return info?.targetRelease?.tag || info?.latestRelease?.tag || ''
}

function markUpdateDialogSeen() {
  if (typeof window === 'undefined') return
  const versionKey = getUpdateDialogVersionKey()
  if (!versionKey) return
  window.sessionStorage.setItem(UPDATE_DIALOG_ACK_KEY, versionKey)
}

function shouldShowUpdateDialog() {
  const info = updateInfo.value
  if (!info?.hasUpdate) return false
  const versionKey = getUpdateDialogVersionKey()
  if (!versionKey || typeof window === 'undefined') return false
  return window.sessionStorage.getItem(UPDATE_DIALOG_ACK_KEY) !== versionKey
}

function escapeHtml(input: string) {
  return input
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}

function buildUpgradeCommands(targetVersion: string) {
  const bare = targetVersion.replace(/^v/, '')
  const tag = `v${bare}`

  const prebuilt = [
    `# deploy/.env`,
    `APP_VERSION=${bare}`,
    `docker compose pull server renderer`,
    `docker compose up -d server renderer`,
  ].join('\n')

  const localBuild = [
    `git fetch --tags`,
    `git checkout ${tag}`,
    `docker compose up -d --build server renderer`,
  ].join('\n')

  return {
    localBuild,
    prebuilt,
  }
}

function renderUpdateDialogContent() {
  const info = updateInfo.value
  const targetVersion = info?.targetRelease?.tag || info?.latestRelease?.tag || '最新版本'
  const releaseBody = info?.targetRelease?.body?.trim() || info?.latestRelease?.body?.trim() || ''
  const { prebuilt, localBuild } = buildUpgradeCommands(targetVersion)
  const releaseHtml = releaseBody
    ? releaseNotesMd.render(releaseBody)
    : `<p>${info?.message || `当前版本 ${info?.currentVersion}，检测到新版本 ${targetVersion}。`}</p>`

  return () => h('div', { class: 'space-y-3' }, [
    h('div', { class: 'space-y-1 text-sm text-neutral-500 dark:text-neutral-400' }, [
      h('div', `当前版本 ${info?.currentVersion || '-'} → 目标版本 ${targetVersion}`),
      h('div', `更新通道 ${info?.channel || '-'} / 来源 ${info?.source || '-'}`),
    ]),
    h(NScrollbar, { style: 'max-height: 360px' }, {
      default: () => h('div', {
        class: 'rounded-lg bg-neutral-50 p-4 text-sm leading-6 text-neutral-700 dark:bg-neutral-900 dark:text-neutral-200',
        innerHTML: releaseHtml,
      }),
    }),
    h('div', { class: 'space-y-2' }, [
      h('div', { class: 'text-xs font-medium text-neutral-500 dark:text-neutral-400' }, '预构建镜像升级'),
      h('pre', {
        class: 'overflow-x-auto rounded-lg bg-neutral-950 p-3 text-xs leading-5 text-neutral-100',
        innerHTML: escapeHtml(prebuilt),
      }),
      h('div', { class: 'text-xs font-medium text-neutral-500 dark:text-neutral-400' }, '本地构建升级'),
      h('pre', {
        class: 'overflow-x-auto rounded-lg bg-neutral-950 p-3 text-xs leading-5 text-neutral-100',
        innerHTML: escapeHtml(localBuild),
      }),
    ]),
  ])
}

function openUpdateDialog() {
  const info = updateInfo.value
  if (!info || !shouldShowUpdateDialog()) return

  const tag = info.targetRelease?.tag || info.latestRelease?.tag || ''
  const releaseNotesUrl = tag ? `https://grtblog.js.org/releases/${tag}` : ''

  dialog.info({
    title: `发现新版本 ${tag}`.trim(),
    content: renderUpdateDialogContent(),
    positiveText: releaseNotesUrl ? '查看说明' : '知道了',
    negativeText: releaseNotesUrl ? '稍后再说' : undefined,
    maskClosable: false,
    style: 'width: min(720px, calc(100vw - 32px));',
    onPositiveClick: () => {
      markUpdateDialogSeen()
      if (releaseNotesUrl && typeof window !== 'undefined') {
        window.open(releaseNotesUrl, '_blank', 'noopener,noreferrer')
      }
    },
    onNegativeClick: () => {
      markUpdateDialogSeen()
    },
    onClose: () => {
      markUpdateDialogSeen()
    },
  })
}

watch(updateInfo, (info) => {
  if (!info?.hasUpdate) return
  openUpdateDialog()
}, { immediate: true })

const AsyncMobileHeader = defineAsyncComponent(() => import('./mobile/MobileHeader.vue'))
const AsyncMobileLeftAside = defineAsyncComponent(() => import('./mobile/MobileLeftAside.vue'))
const AsyncMobileRightAside = defineAsyncComponent(() => import('./mobile/MobileRightAside.vue'))
const AsyncAsideLayout = defineAsyncComponent({
  loader: () => import('./aside/index.vue'),
  loadingComponent: () => {
    const { minWidth, width, collapsed } = sidebarMenu.value
    const { minWidth: defaultMinWidth, width: defaultWidth } =
      DEFAULT_PREFERENCES_OPTIONS.sidebarMenu
    const mergedMinWidth = minWidth || defaultMinWidth
    const mergedWidth = width || defaultWidth
    const finalWidth = collapsed ? mergedMinWidth : mergedWidth

    return h('div', {
      style: {
        width: `${finalWidth + 1}px`,
      },
    })
  },
  delay: 0,
})

const { tabs } = toRefsTabsStore()

const { isMaxSm } = useInjection(mediaQueryInjectionKey)

const {
  layoutSlideDirection,
  setLayoutSlideDirection,
  mobileLeftAsideWidth,
  mobileRightAsideWidth,
} = useInjection(layoutInjectionKey)

const layoutTranslateOffset = computed(() => {
  return layoutSlideDirection.value === 'right'
    ? mobileLeftAsideWidth.value || 0
    : layoutSlideDirection.value === 'left'
      ? -(mobileRightAsideWidth.value || 0)
      : 0
})

const showBgImage = computed(() => backgroundImage.value.show && backgroundImage.value.url)

const bgImageStyle = computed(() => {
  if (!showBgImage.value) return {}
  const bg = backgroundImage.value
  return {
    backgroundImage: `url(${bg.url})`,
    backgroundSize: 'cover',
    backgroundPosition: 'center',
    opacity: bg.opacity / 100,
    filter: bg.blur > 0 ? `blur(${bg.blur}px)` : undefined,
  }
})

const stopRealtimeConnectionListener = adminRealtimeWSCore.onConnection((connected) => {
  realtimeStore.setRealtimeWsConnected(connected)
})

const stopRealtimeMessageListener = adminRealtimeWSCore.onMessage((payload) => {
  if (payload && typeof payload === 'object' && (payload as Record<string, unknown>).type === 'system.health.state') {
    healthStore.handleWSMessage(payload as HealthWSPayload)
    return
  }

  const ownerStatus = normalizeOwnerStatusPayload(payload)
  if (!ownerStatus) return
  queryClient.setQueryData(['owner-status', 'user-dropdown'], ownerStatus)
})

const stopAuthFailureListener = adminRealtimeWSCore.onAuthFailure(() => {
  const currentPath = router.currentRoute.value?.fullPath
  userStore.cleanup(currentPath)
})

watch(isMaxSm, (isMaxSm) => {
  if (isMaxSm) {
    preferences.value.sidebarMenu.collapsed = false
    setLayoutSlideDirection(null)
  }
})

watch(
  [token, () => user.value.isAdmin],
  ([nextToken, isAdmin]) => {
    const jwt = nextToken?.trim() || null
    if (!jwt) {
      adminRealtimeWSCore.stop()
      realtimeStore.setRealtimeWsConnected(false)
      return
    }

    adminRealtimeWSCore.updateToken(jwt)
    adminRealtimeWSCore.setPanelHeartbeat(isAdmin === true)
    adminRealtimeWSCore.start()
  },
  { immediate: true },
)

onUnmounted(() => {
  stopRealtimeConnectionListener()
  stopRealtimeMessageListener()
  stopAuthFailureListener()
  adminRealtimeWSCore.stop()
  realtimeStore.setRealtimeWsConnected(false)
  healthStore.stopPolling()
})

onMounted(() => {
  healthStore.startPolling()
})

function normalizeOwnerStatusPayload(payload: unknown): OwnerStatusPayload | null {
  if (!payload || typeof payload !== 'object') return null
  const raw = payload as Record<string, unknown>
  if (raw.type !== 'owner.status') return null

  const mediaRaw = raw.media
  const media =
    mediaRaw && typeof mediaRaw === 'object'
      ? {
          title: typeof (mediaRaw as Record<string, unknown>).title === 'string' ? (mediaRaw as Record<string, unknown>).title as string : undefined,
          artist: typeof (mediaRaw as Record<string, unknown>).artist === 'string' ? (mediaRaw as Record<string, unknown>).artist as string : undefined,
          thumbnail:
            typeof (mediaRaw as Record<string, unknown>).thumbnail === 'string'
              ? (mediaRaw as Record<string, unknown>).thumbnail as string
              : undefined,
        }
      : null

  return {
    ok: raw.ok === 1 ? 1 : 0,
    process: typeof raw.process === 'string' ? raw.process : undefined,
    extend: typeof raw.extend === 'string' ? raw.extend : undefined,
    media,
    timestamp: typeof raw.timestamp === 'number' ? raw.timestamp : undefined,
    adminPanelOnline: raw.adminPanelOnline === true,
  }
}
</script>

<template>
  <div
    class="relative h-svh overflow-hidden"
    :style="{ backgroundImage: `url(${texturePng})` }"
  >
    <div
      v-if="showBgImage"
      class="pointer-events-none absolute inset-0 z-0 transition-[opacity,filter]"
      :style="bgImageStyle"
    />
    <AsyncMobileLeftAside v-if="isMaxSm" />

    <div
      class="relative z-[1] flex h-full flex-col max-sm:bg-naive-card/50"
      :class="{
        'border-naive-border transition-[background-color,border-color,rounded,transform]': isMaxSm,
        'rounded-xl border pb-2': isMaxSm && layoutTranslateOffset,
      }"
      :style="
        isMaxSm &&
        layoutSlideDirection && {
          transform: `translate(${layoutTranslateOffset}px) scale(0.88)`,
        }
      "
    >
      <HealthBanner />
      <HeaderLayout v-if="!isMaxSm" />
      <AsyncMobileHeader v-else />
      <div class="flex flex-1 overflow-hidden">
        <CollapseTransition
          v-if="!isMaxSm"
          :display="navigationMode === 'sidebar'"
          content-class="min-h-0"
        >
          <AsyncAsideLayout />
        </CollapseTransition>
        <div
          class="relative flex flex-1 flex-col overflow-hidden border-t border-naive-border transition-[border-color]"
        >
          <CollapseTransition
            v-if="!isMaxSm"
            :display="!isEmpty(tabs) && tabsOptions.show"
            direction="horizontal"
            :render-content="false"
          >
            <Tabs />
          </CollapseTransition>
          <main class="relative flex-1 overflow-hidden">
            <MainLayout />
          </main>
          <EmptyPlaceholder
            :show="isEmpty(tabs)"
            description="空标签页"
            size="huge"
          >
            <template #icon>
              <div class="flex items-center justify-center">
                <span class="iconify ph--rectangle" />
              </div>
            </template>
          </EmptyPlaceholder>
          <CollapseTransition
            v-if="!isMaxSm"
            :display="showFooter"
            direction="horizontal"
            :render-content="false"
          >
            <FooterLayout />
          </CollapseTransition>
        </div>
      </div>
      <div
        v-if="isMaxSm && layoutSlideDirection"
        class="absolute inset-0"
        style="z-index: 9997"
        @click="setLayoutSlideDirection(null)"
      />
    </div>
    <AsyncMobileRightAside v-if="isMaxSm" />
  </div>
</template>
