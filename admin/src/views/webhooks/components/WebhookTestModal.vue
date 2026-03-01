<script setup lang="ts">
import { NModal, NSelect } from 'naive-ui'

import type { SelectOption } from 'naive-ui'

defineProps<{
  visible: boolean
  testEventName: string | null
  eventOptions: SelectOption[]
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'update:testEventName': [value: string | null]
  confirm: []
}>()
</script>

<template>
  <NModal
    :show="visible"
    preset="dialog"
    title="测试 Webhook"
    positive-text="发送"
    negative-text="取消"
    @positive-click="emit('confirm')"
    @update:show="emit('update:visible', $event)"
  >
    <div class="flex flex-col gap-3 py-2">
      <div class="text-xs text-[var(--text-color-3)]">
        选择一个事件用于测试投递。
      </div>
      <NSelect
        :value="testEventName"
        :options="eventOptions"
        placeholder="选择事件"
        clearable
        @update:value="emit('update:testEventName', $event)"
      />
    </div>
  </NModal>
</template>
