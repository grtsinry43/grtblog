<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { NDropdown } from 'naive-ui'
import { computed, h } from 'vue'
import { useRouter } from 'vue-router'

import { getOwnerStatus } from '@/services/owner-status'
import { useRealtimeStore, useUserStore } from '@/stores'

import type { DropdownProps } from 'naive-ui'

interface UserDropdownProps extends /** @vue-ignore */ DropdownProps {}

defineProps<UserDropdownProps>()

defineOptions({
  inheritAttrs: false,
})

const router = useRouter()
const { cleanup } = useUserStore()
const realtimeStore = useRealtimeStore()

const { data: ownerStatusData } = useQuery({
  queryKey: ['owner-status', 'user-dropdown'],
  queryFn: () => getOwnerStatus(),
  refetchInterval: 30_000,
  refetchOnWindowFocus: true,
})

const ownerOnline = computed(() => (ownerStatusData.value?.ok ?? 0) === 1)
const panelOnline = computed(() => ownerStatusData.value?.adminPanelOnline === true)
const wsConnected = computed(() => realtimeStore.realtimeWsConnected)
const appText = computed(() => ownerStatusData.value?.process?.trim() || '-')
const mediaText = computed(() => {
  const media = ownerStatusData.value?.media
  if (!media?.title?.trim()) return '-'
  const title = media.title.trim()
  const artist = media.artist?.trim()
  return artist ? `${title} · ${artist}` : title
})

const userDropdownOptions = computed(() => [
  {
    icon: () =>
      h('span', {
        class: `iconify ph--circle-fill size-3 ${ownerOnline.value ? 'text-emerald-500' : 'text-gray-400'}`,
      }),
    key: 'owner-status',
    label: `站长状态：${ownerOnline.value ? '在线' : '离线'} / 面板${panelOnline.value ? '在线' : '离线'}`,
    disabled: true,
  },
  {
    icon: () =>
      h('span', {
        class: `iconify ph--circle-fill size-3 ${wsConnected.value ? 'text-emerald-500' : 'text-orange-500'}`,
      }),
    key: 'ws-status',
    label: `实时 WS：${wsConnected.value ? '已连接' : '重连中'}`,
    disabled: true,
  },
  {
    icon: () => h('span', { class: 'iconify ph--app-window size-4 text-gray-400' }),
    key: 'owner-app',
    label: `应用：${appText.value}`,
    disabled: true,
  },
  {
    icon: () => h('span', { class: 'iconify ph--music-note-simple size-4 text-gray-400' }),
    key: 'owner-media',
    label: `媒体：${mediaText.value}`,
    disabled: true,
  },
  {
    key: 'divider-1',
    type: 'divider',
  },
  {
    icon: () => h('span', { class: 'iconify ph--user size-5' }),
    key: 'user',
    label: '个人中心',
  },
  {
    icon: () => h('span', { class: 'iconify ph--sign-out size-5' }),
    key: 'signOut',
    label: '退出登录',
  },
])

const onUserDropdownSelected = (key: string) => {
  switch (key) {
    case 'user':
      router.push({ name: 'userCenter' })
      break
    case 'signOut':
      cleanup()
      break
  }
}
</script>
<template>
  <NDropdown
    trigger="click"
    :options="userDropdownOptions"
    show-arrow
    @select="onUserDropdownSelected"
    v-bind="$attrs"
  >
    <slot />
  </NDropdown>
</template>
