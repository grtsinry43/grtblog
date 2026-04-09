<script setup lang="ts">
import { computed } from 'vue'

import { getCachedFavicon } from '@/utils/document-title'

const faviconUrl = computed(() => getCachedFavicon())
</script>

<template>
  <div class="size-full shrink-0">
    <img
      v-if="faviconUrl"
      :src="faviconUrl"
      alt="Site Logo"
      class="size-full rounded object-contain"
    />
    <div
      v-else
      class="size-full rounded bg-primary/10"
    ></div>
  </div>
</template>
