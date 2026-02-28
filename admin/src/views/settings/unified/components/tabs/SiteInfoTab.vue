<script setup lang="ts">
import { listSysConfigs, updateSysConfigs } from '@/services/sysconfig'

import ConfigPanel from '../ConfigPanel'

const emit = defineEmits<{ 'dirty-change': [dirty: boolean] }>()
</script>

<template>
  <ConfigPanel
    :list-fn="listSysConfigs"
    :update-fn="updateSysConfigs"
    title="站点信息"
    description="网站基础信息与 OG 默认值"
    :filter-groups="['site']"
    :exclude-item-keys="['site.theme_extend_info']"
    :on-dirty-change="(dirty: boolean) => emit('dirty-change', dirty)"
  />
</template>
