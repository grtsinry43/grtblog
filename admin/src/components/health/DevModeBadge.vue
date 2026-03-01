<script setup lang="ts">
const isDev = import.meta.env.MODE === 'development'
</script>

<template>
  <span
    v-if="isDev"
    class="inline-flex items-center rounded-full bg-violet-500/15 px-2 py-0.5 text-[10px] font-bold tracking-wider text-violet-500"
  >
    DEV
  </span>
</template>
