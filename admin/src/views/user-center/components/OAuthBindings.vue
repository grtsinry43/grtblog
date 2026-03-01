<script setup lang="ts">
import { NCard, NGi, NGrid, NTag } from 'naive-ui'

import type { OAuthBinding } from '@/services/auth'

defineProps<{
  loading: boolean
  bindings: OAuthBinding[]
}>()
</script>

<template>
  <div v-if="loading" class="py-12 text-center text-neutral-400">
    正在加载绑定信息...
  </div>
  <div
    v-else-if="bindings.length === 0"
    class="flex flex-col items-center justify-center py-20"
  >
    <div class="mb-4 text-5xl text-neutral-150 dark:text-neutral-800">
      <span class="iconify ph--link-break" />
    </div>
    <div class="text-neutral-500">尚未绑定任何第三方账号</div>
  </div>
  <NGrid v-else cols="1 m:2" x-gap="16" y-gap="16">
    <NGi v-for="item in bindings" :key="item.providerKey + item.oauthID">
      <NCard size="small" hoverable>
        <div class="flex items-center gap-4 py-1">
          <div class="grid h-10 w-10 place-items-center rounded bg-primary/10 text-xl font-bold text-primary">
            {{ item.providerKey.charAt(0).toUpperCase() }}
          </div>
          <div class="flex-1 overflow-hidden">
            <div class="flex items-center justify-between">
              <span class="font-medium">{{ item.providerName || item.providerKey }}</span>
              <NTag type="success" size="tiny" round>已关联</NTag>
            </div>
            <div class="truncate text-xs text-neutral-400">ID: {{ item.oauthID }}</div>
          </div>
        </div>
      </NCard>
    </NGi>
  </NGrid>
</template>
