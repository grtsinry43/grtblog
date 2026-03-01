<script setup lang="ts">
import { NButton, NModal } from 'naive-ui'
import VuePictureCropper from 'vue-picture-cropper'

defineProps<{
  visible: boolean
  cropperImg: string
  isUploading: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  confirm: []
}>()
</script>

<template>
  <NModal
    :show="visible"
    preset="card"
    style="max-width: 600px"
    title="裁剪头像"
    :mask-closable="false"
    :closable="!isUploading"
    @update:show="emit('update:visible', $event)"
  >
    <div class="h-80 w-full overflow-hidden rounded bg-neutral-100 dark:bg-neutral-900">
      <VuePictureCropper
        :box-style="{
          width: '100%',
          height: '100%',
          backgroundColor: '#f8f8f8',
          margin: 'auto',
        }"
        :img="cropperImg"
        :options="{
          viewMode: 1,
          dragMode: 'move',
          aspectRatio: 1,
          cropBoxResizable: false,
        }"
      />
    </div>
    <template #footer>
      <div class="flex justify-end gap-2">
        <NButton :disabled="isUploading" @click="emit('update:visible', false)">取消</NButton>
        <NButton type="primary" :loading="isUploading" @click="emit('confirm')">确认并上传</NButton>
      </div>
    </template>
  </NModal>
</template>
