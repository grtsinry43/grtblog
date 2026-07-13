<script setup lang="ts">
import { NButton, NTag } from 'naive-ui'

import type { UpgradeGuideVersion } from '../registry'

defineProps<{
  guide: UpgradeGuideVersion
  submitting: boolean
}>()

defineEmits<{
  continue: []
  skip: []
}>()
</script>

<template>
  <div>
    <NTag
      size="small"
      type="success"
      round
      :bordered="false"
    >
      {{ guide.tag }}
    </NTag>
    <h2 class="mt-3 text-2xl font-bold tracking-tight">{{ guide.title }}</h2>
    <p class="mt-2 text-[13px] leading-relaxed text-neutral-500">
      {{ guide.description }}
    </p>

    <div class="mt-8 space-y-4">
      <div
        v-for="feature in guide.features"
        :key="feature.id"
        class="rounded-lg border border-neutral-100 p-5 dark:border-neutral-800"
      >
        <div class="flex items-center gap-2">
          <span
            class="iconify text-lg"
            :class="feature.icon"
          ></span>
          <span class="text-sm font-medium">{{ feature.label }}</span>
        </div>
        <p class="mt-2 text-xs leading-relaxed text-neutral-500">
          {{ feature.description }}
        </p>
      </div>
    </div>

    <div
      class="mt-8 flex items-center justify-between border-t border-neutral-100 pt-6 dark:border-neutral-800"
    >
      <NButton
        quaternary
        :disabled="submitting"
        @click="$emit('skip')"
      >
        稍后再说
      </NButton>
      <NButton
        type="primary"
        @click="$emit('continue')"
      >
        开始配置
      </NButton>
    </div>
  </div>
</template>
