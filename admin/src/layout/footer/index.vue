<script setup lang="ts">
import { ref } from 'vue'

import { getSystemStatus } from '@/services/system'

defineOptions({
  name: 'FooterLayout',
})

const APP_NAME = import.meta.env.VITE_APP_NAME
const version = ref('')
const commit = ref('')

getSystemStatus().then((res) => {
  version.value = res.app.version
  commit.value = res.app.commit ?? ''
}).catch(() => {
  version.value = 'unknown'
})
</script>

<template>
  <footer
    class="min-h-0 border-t border-naive-border bg-naive-card transition-[background-color,border-color]"
  >
    <div class="flex items-center justify-center overflow-hidden py-1.5 text-xs text-neutral-500 dark:text-neutral-400">
      <span>{{ APP_NAME }} {{ version }}{{ commit ? ` (${commit})` : '' }}</span>
    </div>
  </footer>
</template>
