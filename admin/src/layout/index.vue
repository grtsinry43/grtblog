<script setup lang="ts">
import { useQueryClient } from '@tanstack/vue-query'
import { isEmpty } from 'lodash-es'
import { computed, defineAsyncComponent, h, onMounted, onUnmounted, watch } from 'vue'

import texturePng from '@/assets/texture.png'
import { CollapseTransition, EmptyPlaceholder } from '@/components'
import HealthBanner from '@/components/health/HealthBanner.vue'
import DevModeBadge from '@/components/health/DevModeBadge.vue'
import { useInjection } from '@/composables'
import { mediaQueryInjectionKey, layoutInjectionKey } from '@/injection'
import { adminRealtimeWSCore } from '@/services/realtime-ws'
import { getSystemUpdateCheck } from '@/services/system'
import { DEFAULT_PREFERENCES_OPTIONS, toRefsPreferencesStore, toRefsTabsStore, toRefsUserStore, useRealtimeStore, useHealthStore } from '@/stores'

import type { OwnerStatusPayload } from '@/services/owner-status'
import type { HealthWSPayload } from '@/services/health'

import FooterLayout from './footer/index.vue'
import HeaderLayout from './header/index.vue'
import MainLayout from './main/index.vue'
import Tabs from './tabs/index.vue'

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
const realtimeStore = useRealtimeStore()
const healthStore = useHealthStore()
const queryClient = useQueryClient()

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
  // Dispatch health state messages.
  if (payload && typeof payload === 'object' && (payload as Record<string, unknown>).type === 'system.health.state') {
    healthStore.handleWSMessage(payload as HealthWSPayload)
    return
  }

  const ownerStatus = normalizeOwnerStatusPayload(payload)
  if (!ownerStatus) return
  queryClient.setQueryData(['owner-status', 'user-dropdown'], ownerStatus)
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
  adminRealtimeWSCore.stop()
  realtimeStore.setRealtimeWsConnected(false)
  healthStore.stopPolling()
})

onMounted(() => {
  healthStore.startPolling()
  void queryClient.prefetchQuery({
    queryKey: ['system-update-check'],
    queryFn: () => getSystemUpdateCheck(false),
    staleTime: 30 * 60 * 1000,
  })
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
