<script setup lang="ts">
import { NAlert, NSwitch } from 'naive-ui'

import type { UpgradeGuideVersion } from './registry'

defineProps<{
  /** One or more guide versions to render */
  guides: UpgradeGuideVersion[]
  /** Primary color CSS var value, e.g. "24, 120, 215" */
  primaryColorRgb?: string
}>()

/** Record of feature id → enabled boolean, shared via v-model */
const states = defineModel<Record<string, boolean>>('states', { required: true })

function toggle(featureId: string, value: boolean) {
  states.value = { ...states.value, [featureId]: value }
}
</script>

<template>
  <div class="space-y-6">
    <template
      v-for="guide in guides"
      :key="guide.version"
    >
      <!-- Version section header (only shown when there are multiple guides) -->
      <div
        v-if="guides.length > 1"
        class="mb-2"
      >
        <div class="text-xs font-semibold tracking-wide text-neutral-400 uppercase">
          {{ guide.tag }}
        </div>
        <div class="mt-1 text-sm font-medium text-neutral-700 dark:text-neutral-200">
          {{ guide.title }}
        </div>
        <p
          v-if="guide.description"
          class="mt-1 text-xs text-neutral-500"
        >
          {{ guide.description }}
        </p>
      </div>

      <!-- Feature cards -->
      <div class="space-y-4">
        <div
          v-for="feature in guide.features"
          :key="feature.id"
          class="rounded-lg border border-neutral-100 p-5 transition-colors dark:border-neutral-800"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="flex-1">
              <div class="flex items-center gap-2">
                <span
                  class="iconify text-lg"
                  :class="feature.icon"
                  :style="primaryColorRgb ? { color: `rgb(${primaryColorRgb})` } : undefined"
                ></span>
                <span class="text-sm font-medium text-neutral-800 dark:text-neutral-100">
                  {{ feature.label }}
                </span>
              </div>
              <p class="mt-2 text-xs leading-relaxed text-neutral-500">
                {{ feature.description }}
              </p>
            </div>
            <NSwitch
              :value="states[feature.id] ?? false"
              @update:value="(v: boolean) => toggle(feature.id, v)"
            />
          </div>
        </div>
      </div>

      <!-- Hint -->
      <NAlert
        v-if="guide.hint"
        type="info"
        :show-icon="false"
        class="text-xs"
      >
        {{ guide.hint }}
      </NAlert>
    </template>
  </div>
</template>
