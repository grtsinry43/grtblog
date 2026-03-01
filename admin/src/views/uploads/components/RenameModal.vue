<script setup lang="ts">
import { NInput, NModal } from 'naive-ui'

defineProps<{
  visible: boolean
  fileName: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'update:fileName': [value: string]
  confirm: []
}>()
</script>

<template>
  <NModal
    :show="visible"
    preset="dialog"
    title="重命名文件"
    positive-text="确认"
    negative-text="取消"
    @positive-click="emit('confirm')"
    @update:show="emit('update:visible', $event)"
  >
    <div style="padding: 16px 0">
      <NInput
        :value="fileName"
        placeholder="请输入新文件名"
        @update:value="emit('update:fileName', $event)"
        @keyup.enter="emit('confirm')"
      />
    </div>
  </NModal>
</template>
