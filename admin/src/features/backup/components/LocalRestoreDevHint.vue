<script setup lang="ts">
import { NAlert } from 'naive-ui'

const isDev = import.meta.env.MODE === 'development'
</script>

<template>
  <NAlert
    v-if="isDev"
    type="info"
    title="DEV · 本地恢复测试"
    class="mb-4"
  >
    <p>
      直接运行后端时，提交恢复会先让 API 退出。确认后端进程退出后，请在
      <code class="font-mono">server</code> 目录执行：
    </p>
    <code class="mt-2 block w-fit rounded bg-black/5 px-2 py-1 font-mono text-xs dark:bg-white/10"
      >make restore-and-run</code
    >
    <p class="mt-2 text-xs opacity-75">使用 Docker Compose 开发时无需手动执行。</p>
  </NAlert>
</template>
