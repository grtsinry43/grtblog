<script setup lang="ts">
import { computed } from 'vue'

import { toRefsPreferencesStore } from '@/stores'

import Action from './action/index.vue'
import Logo from './logo/index.vue'
import Navigation from './navigation/index.vue'

defineOptions({
  name: 'HeaderLayout',
})

const { navigationMode, backgroundImage } = toRefsPreferencesStore()

const isGlassActive = computed(() =>
  backgroundImage.value.show && backgroundImage.value.url && backgroundImage.value.glassEffect.enable,
)
</script>
<template>
  <header
    class="flex bg-naive-card transition-[background-color]"
    :style="{
      backdropFilter: isGlassActive ? `blur(${backgroundImage.glassEffect.blur}px)` : undefined,
    }"
  >
    <Logo />
    <div
      class="flex flex-1 items-center border-l px-4 py-3.5 transition-[border-color]"
      :class="navigationMode === 'sidebar' ? 'border-naive-border' : 'border-transparent'"
    >
      <Navigation />
      <Action class="gap-x-3 pl-4" />
    </div>
  </header>
</template>
