<script setup lang="ts">
import { NButton, NForm, NModal, NSpace } from 'naive-ui'

defineOptions({
  name: 'FormModal',
})

const props = withDefaults(
  defineProps<{
    show: boolean
    title: string
    loading?: boolean
    width?: number
    labelWidth?: number
    confirmText?: string
    cancelText?: string
  }>(),
  {
    loading: false,
    width: 540,
    labelWidth: 90,
    confirmText: '保存',
    cancelText: '取消',
  },
)

const emit = defineEmits<{
  'update:show': [value: boolean]
  confirm: []
}>()

function close() {
  emit('update:show', false)
}

function onConfirm() {
  emit('confirm')
}
</script>

<template>
  <NModal
    :show="props.show"
    preset="card"
    :title="props.title"
    :style="{ width: `${props.width}px` }"
    @update:show="emit('update:show', $event)"
  >
    <NForm
      label-placement="left"
      :label-width="props.labelWidth"
    >
      <slot />
    </NForm>

    <template #footer>
      <slot
        name="footer"
        :close="close"
      >
        <NSpace justify="end">
          <NButton @click="close">{{ cancelText }}</NButton>
          <NButton
            type="primary"
            :loading="props.loading"
            @click="onConfirm"
            >{{ confirmText }}</NButton
          >
        </NSpace>
      </slot>
    </template>
  </NModal>
</template>
