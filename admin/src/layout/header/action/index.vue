<script setup lang="ts">
import { defineAsyncComponent, h } from 'vue'

import DevModeBadge from '@/components/health/DevModeBadge.vue'
import { useInjection } from '@/composables'
import { mediaQueryInjectionKey } from '@/injection'
import { toRefsPreferencesStore } from '@/stores'

import FullScreen from './FullScreen.vue'
import Notification from './Notification.vue'
import PreferencesDrawer from './PreferencesDrawer.vue'
import SignOut from './SignOut.vue'
import ThemeModePopover from './ThemeModePopover.vue'
defineOptions({
  name: 'Actions',
})

const AsyncAvatarDropdown = defineAsyncComponent({
  loader: () => import('./AvatarDropdown.vue'),
  loadingComponent: () => h('div', { style: { width: '34px', marginLeft: '4px' } }),
  delay: 0,
})

const { isMaxSm } = useInjection(mediaQueryInjectionKey)
const { navigationMode } = toRefsPreferencesStore()
</script>
<template>
  <div class="flex items-center">
    <DevModeBadge />
    <Notification />
    <FullScreen />
    <ThemeModePopover />
    <PreferencesDrawer />
    <SignOut />
    <AsyncAvatarDropdown v-if="!isMaxSm && navigationMode === 'horizontal'" />
  </div>
</template>
