<script setup lang="ts">
import { ChevronDown24Regular, CloudArrowUp24Regular, Document24Regular, Image24Regular } from '@vicons/fluent'
import { NButton, NDropdown, NIcon, NSpace, NUpload } from 'naive-ui'
import { h } from 'vue'

import type { FileType } from '@/services/uploads'
import type { DropdownOption, UploadFileInfo } from 'naive-ui'

defineProps<{
  uploadType: FileType
  uploading: boolean
}>()

const emit = defineEmits<{
  'update:uploadType': [value: FileType]
  upload: [payload: { file: UploadFileInfo }]
}>()

const typeOptions: DropdownOption[] = [
  { label: '图片', key: 'picture', icon: renderIcon(Image24Regular) },
  { label: '文件', key: 'file', icon: renderIcon(Document24Regular) },
]

function renderIcon(icon: typeof Image24Regular | typeof Document24Regular) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

function getTypeLabel(value: FileType) {
  return value === 'picture' ? '图片' : '文件'
}
</script>

<template>
  <NSpace align="center">
    <NDropdown :options="typeOptions" @select="emit('update:uploadType', $event)">
      <NButton secondary>
        <span class="upload-type-option">
          <NIcon><Image24Regular v-if="uploadType === 'picture'" /><Document24Regular v-else /></NIcon>
          <span>{{ getTypeLabel(uploadType) }}</span>
          <NIcon><ChevronDown24Regular /></NIcon>
        </span>
      </NButton>
    </NDropdown>
    <NUpload
      multiple
      :accept="uploadType === 'picture' ? 'image/*' : undefined"
      :show-file-list="false"
      :custom-request="(p: any) => emit('upload', p)"
      :disabled="uploading"
    >
      <NButton type="primary" :loading="uploading">
        <template #icon><NIcon><CloudArrowUp24Regular /></NIcon></template>
        上传{{ getTypeLabel(uploadType) }}
      </NButton>
    </NUpload>
  </NSpace>
</template>

<style scoped>
.upload-type-option {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  line-height: 1;
}
</style>
