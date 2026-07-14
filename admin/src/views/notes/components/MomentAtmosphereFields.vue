<script setup lang="ts">
import { NButton } from 'naive-ui'

import { momentMoodOptions, momentWeatherOptions } from '../model/moment-atmosphere'

import type { MomentMood, MomentWeather } from '@/types/ext-info'

const weather = defineModel<MomentWeather | null>('weather', { required: true })
const mood = defineModel<MomentMood | null>('mood', { required: true })
</script>

<template>
  <div class="space-y-5">
    <div class="space-y-2">
      <div class="flex items-center justify-between gap-3">
        <span class="text-sm">天气</span>
        <NButton
          v-if="weather"
          text
          size="tiny"
          @click="weather = null"
        >
          清除
        </NButton>
      </div>
      <div
        class="flex flex-wrap gap-2"
        role="radiogroup"
        aria-label="天气"
      >
        <NButton
          v-for="option in momentWeatherOptions"
          :key="option.value"
          size="small"
          :type="weather === option.value ? 'primary' : 'default'"
          :secondary="weather === option.value"
          :aria-pressed="weather === option.value"
          @click="weather = weather === option.value ? null : option.value"
        >
          <template #icon>
            <span :class="['iconify', option.iconClass]" />
          </template>
          {{ option.label }}
        </NButton>
      </div>
    </div>

    <div class="space-y-2">
      <div class="flex items-center justify-between gap-3">
        <span class="text-sm">心情</span>
        <NButton
          v-if="mood"
          text
          size="tiny"
          @click="mood = null"
        >
          清除
        </NButton>
      </div>
      <div
        class="flex flex-wrap gap-2"
        role="radiogroup"
        aria-label="心情"
      >
        <NButton
          v-for="option in momentMoodOptions"
          :key="option.value"
          size="small"
          :type="mood === option.value ? 'primary' : 'default'"
          :secondary="mood === option.value"
          :aria-pressed="mood === option.value"
          @click="mood = mood === option.value ? null : option.value"
        >
          <template #icon>
            <span :class="['iconify', option.iconClass]" />
          </template>
          {{ option.label }}
        </NButton>
      </div>
    </div>

    <p class="text-xs opacity-50">均为可选项；再次点击已选项也可以取消。</p>
  </div>
</template>
