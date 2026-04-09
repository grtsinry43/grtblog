<script setup lang="ts">
import { NCard, NGi, NGrid, NTag } from 'naive-ui'
import { computed } from 'vue'

import type { OAuthBinding } from '@/services/auth'
import type { AdminOAuthProvider } from '@/services/oauth-providers'

const props = defineProps<{
  loading: boolean
  bindings: OAuthBinding[]
  providers: AdminOAuthProvider[]
}>()

const boundSet = computed(() => new Set(props.bindings.map((b) => b.providerKey)))
</script>

<template>
  <div
    v-if="loading"
    class="py-12 text-center text-neutral-400"
  >
    正在加载绑定信息...
  </div>
  <div
    v-else-if="providers.length === 0"
    class="flex flex-col items-center justify-center py-20"
  >
    <div class="mb-4 text-5xl text-neutral-150 dark:text-neutral-800">
      <span class="iconify ph--link-break" />
    </div>
    <div class="text-neutral-500">暂无可用的第三方登录方式</div>
  </div>
  <NGrid
    v-else
    cols="1 m:2"
    x-gap="16"
    y-gap="16"
  >
    <NGi
      v-for="provider in providers"
      :key="provider.key"
    >
      <NCard
        size="small"
        hoverable
      >
        <div class="flex items-center gap-4 py-1">
          <div
            class="grid h-10 w-10 place-items-center rounded text-xl font-bold"
            :class="
              boundSet.has(provider.key)
                ? 'bg-primary/10 text-primary'
                : 'bg-neutral-100 text-neutral-400 dark:bg-neutral-800'
            "
          >
            {{ provider.displayName.charAt(0).toUpperCase() }}
          </div>
          <div class="flex-1 overflow-hidden">
            <div class="flex items-center justify-between">
              <span class="font-medium">{{ provider.displayName }}</span>
              <NTag
                v-if="boundSet.has(provider.key)"
                type="success"
                size="tiny"
                round
              >
                已绑定
              </NTag>
              <NTag
                v-else
                type="default"
                size="tiny"
                round
              >
                未绑定
              </NTag>
            </div>
            <div class="truncate text-xs text-neutral-400">
              {{ provider.scopes || provider.key }}
            </div>
          </div>
        </div>
      </NCard>
    </NGi>
  </NGrid>
</template>
