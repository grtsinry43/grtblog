<script setup lang="ts">
import { listSysConfigs, updateSysConfigs } from '@/services/sysconfig'

import ConfigPanel from '../ConfigPanel'

const emit = defineEmits<{ 'dirty-change': [dirty: boolean] }>()
</script>

<template>
  <ConfigPanel
    :list-fn="listSysConfigs"
    :update-fn="updateSysConfigs"
    title="邮件配置"
    description="SMTP 服务器和发送参数"
    :filter-groups="['notification/email']"
    :on-dirty-change="(dirty: boolean) => emit('dirty-change', dirty)"
  />
</template>
