<script setup lang="ts">
import { NButton, NImage, NInput, NInputGroup } from 'naive-ui'
import { ref } from 'vue'

import ImagePickerModal from './ImagePickerModal.vue'

const props = defineProps<{
  value: string | null
}>()

const emit = defineEmits<{
  'update:value': [value: string | null]
}>()

const showPicker = ref(false)

function handleInput(val: string | null) {
  emit('update:value', val || null)
}

function handleSelect(url: string) {
  emit('update:value', url)
}
</script>

<template>
  <div class="w-full">
    <NInputGroup>
      <NInput
        :value="value"
        placeholder="图片 URL"
        @update:value="handleInput"
      >
        <template #prefix><div class="iconify ph--image" /></template>
      </NInput>
      <NButton
        @click="showPicker = true"
        style="padding: 0 10px"
      >
        <template #icon><div class="iconify ph--folder-open" /></template>
      </NButton>
    </NInputGroup>

    <div
      v-if="value"
      class="mt-2 overflow-hidden rounded-lg border border-current/10"
    >
      <NImage
        :src="value"
        :alt="value"
        object-fit="contain"
        :img-props="{ style: 'max-height: 128px; width: 100%; object-fit: contain;' }"
      />
    </div>

    <ImagePickerModal
      v-model:show="showPicker"
      @select="handleSelect"
    />
  </div>
</template>
