<script setup lang="ts">
import { NButton } from 'naive-ui'

import FeatureToggleList from '../FeatureToggleList.vue'

import type { UpgradeGuideVersion } from '../registry'

defineProps<{
  featureGuide: UpgradeGuideVersion
  primaryColorRgb: string
  submitting: boolean
}>()

const states = defineModel<Record<string, boolean>>('states', { required: true })

defineEmits<{
  back: []
  finish: []
}>()
</script>

<template>
  <div>
    <div class="mb-8">
      <div class="text-xs font-semibold tracking-wide text-neutral-400 uppercase">可选设置</div>
      <h2 class="mt-2 text-2xl font-bold tracking-tight">按你的需要启用新功能</h2>
      <p class="mt-2 text-[13px] leading-relaxed text-neutral-500">
        所有选项都可以稍后在设置中修改；未选择的功能会保持关闭。
      </p>
    </div>

    <FeatureToggleList
      v-if="featureGuide.features.length > 0"
      v-model:states="states"
      :guides="[featureGuide]"
      :primary-color-rgb="primaryColorRgb"
    />

    <div
      class="mt-8 flex items-center justify-between border-t border-neutral-100 pt-6 dark:border-neutral-800"
    >
      <NButton
        quaternary
        :disabled="submitting"
        @click="$emit('back')"
      >
        返回
      </NButton>
      <NButton
        type="primary"
        :loading="submitting"
        @click="$emit('finish')"
      >
        保存并完成
      </NButton>
    </div>
  </div>
</template>
