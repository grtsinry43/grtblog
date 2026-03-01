<script setup lang="ts">
import { CloudArrowUp24Regular, Document24Regular, Image24Regular } from '@vicons/fluent'
import { NButton, NIcon, NRadioButton, NRadioGroup, NSpace, NUpload } from 'naive-ui'

import type { FileType } from '@/services/uploads'
import type { UploadFileInfo } from 'naive-ui'

defineProps<{
  uploadType: FileType
  uploading: boolean
}>()

const emit = defineEmits<{
  'update:uploadType': [value: FileType]
  upload: [payload: { file: UploadFileInfo }]
}>()
</script>

<template>
  <NSpace align="center">
    <NRadioGroup :value="uploadType" size="small" @update:value="emit('update:uploadType', $event)">
      <NRadioButton value="picture">
        <NIcon><Image24Regular /></NIcon>
        图片
      </NRadioButton>
      <NRadioButton value="file">
        <NIcon><Document24Regular /></NIcon>
        文件
      </NRadioButton>
    </NRadioGroup>
    <NUpload :show-file-list="false" :custom-request="(p: any) => emit('upload', p)" :disabled="uploading">
      <NButton type="primary" :loading="uploading">
        <template #icon><NIcon><CloudArrowUp24Regular /></NIcon></template>
        上传文件
      </NButton>
    </NUpload>
  </NSpace>
</template>
