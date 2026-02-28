<script setup lang="ts">
import { ref } from 'vue'

import {
  listActivityPubConfigs,
  listFederationConfigs,
  updateActivityPubConfigs,
  updateFederationConfigs,
} from '@/services/sysconfig'

import ConfigPanel from '../ConfigPanel'

const emit = defineEmits<{ 'dirty-change': [dirty: boolean] }>()

const federationDirty = ref(false)
const activityPubDirty = ref(false)

function updateDirty(source: 'federation' | 'activitypub', dirty: boolean) {
  if (source === 'federation') federationDirty.value = dirty
  else activityPubDirty.value = dirty
  emit('dirty-change', federationDirty.value || activityPubDirty.value)
}
</script>

<template>
  <div class="space-y-6">
    <ConfigPanel
      :list-fn="listFederationConfigs"
      :update-fn="updateFederationConfigs"
      title="Federation 联合"
      description="启用后系统会自动生成密钥，仅需填写基础信息即可"
      :on-dirty-change="(dirty: boolean) => updateDirty('federation', dirty)"
    />

    <ConfigPanel
      :list-fn="listActivityPubConfigs"
      :update-fn="updateActivityPubConfigs"
      title="ActivityPub"
      description="兼容功能独立配置，启用后将使用 ActivityPub 专用密钥"
      :on-dirty-change="(dirty: boolean) => updateDirty('activitypub', dirty)"
    />
  </div>
</template>
