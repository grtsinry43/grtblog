<script setup lang="ts">
import { ComingSoon, ScrollContainer } from '@/components'

defineOptions({
  name: 'PluginList',
})
</script>

<template>
  <ScrollContainer wrapper-class="p-4">
    <ComingSoon
      title="插件与云函数"
      description="用于扩展站点能力与自动化任务接入，预留插件与云函数管理入口。"
      icon="ph--puzzle-piece"
    />
  </ScrollContainer>
</template>
