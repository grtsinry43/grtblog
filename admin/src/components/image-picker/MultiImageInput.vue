<script setup lang="ts">
import { NButton, NImage, NInput, NInputGroup } from 'naive-ui'
import { computed, ref } from 'vue'
import ImagePickerModal from './ImagePickerModal.vue'

const props = defineProps<{
  value: string
}>()

const emit = defineEmits<{
  'update:value': [value: string]
}>()

const showPicker = ref(false)
const inputUrl = ref('')

const images = computed(() =>
  props.value
    .split(/\r?\n/)
    .map((s) => s.trim())
    .filter(Boolean),
)

function emitJoined(list: string[]) {
  emit('update:value', list.join('\n'))
}

function addUrl(url: string) {
  const trimmed = url.trim()
  if (!trimmed) return
  emitJoined([...images.value, trimmed])
}

function handleAdd() {
  addUrl(inputUrl.value)
  inputUrl.value = ''
}

function handleSelect(url: string) {
  addUrl(url)
}

function handleRemove(index: number) {
  const list = [...images.value]
  list.splice(index, 1)
  emitJoined(list)
}
</script>

<template>
  <div class="w-full">
    <div
      v-if="images.length > 0"
      class="mb-3 grid grid-cols-3 gap-2"
    >
      <div
        v-for="(url, index) in images"
        :key="index"
        class="group relative overflow-hidden rounded-lg border border-current/10"
      >
        <NImage
          :src="url"
          :alt="`图片 ${index + 1}`"
          object-fit="cover"
          preview-disabled
          :img-props="{ class: 'aspect-square w-full object-cover' }"
        />
        <NButton
          circle
          size="tiny"
          type="error"
          class="absolute top-1 right-1 opacity-0 shadow-sm transition-opacity group-hover:opacity-100"
          @click="handleRemove(index)"
        >
          <template #icon><div class="iconify ph--x text-xs" /></template>
        </NButton>
      </div>
    </div>

    <NInputGroup>
      <NInput
        v-model:value="inputUrl"
        placeholder="输入图片 URL 并添加"
        @keydown.enter="handleAdd"
      >
        <template #prefix><div class="iconify ph--image" /></template>
      </NInput>
      <NButton
        @click="handleAdd"
        :disabled="!inputUrl.trim()"
      >
        添加
      </NButton>
      <NButton
        @click="showPicker = true"
        style="padding: 0 10px"
      >
        <template #icon><div class="iconify ph--folder-open" /></template>
      </NButton>
    </NInputGroup>

    <ImagePickerModal
      v-model:show="showPicker"
      @select="handleSelect"
    />
  </div>
</template>
