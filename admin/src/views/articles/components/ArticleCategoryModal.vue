<script setup lang="ts">
import { NButton, NCard, NForm, NFormItem, NInput, NModal } from 'naive-ui'

import type { NewCategoryModalState } from '../composables/use-taxonomy-select'

const modal = defineModel<NewCategoryModalState>('modal', { required: true })

defineEmits<{
  create: []
}>()
</script>

<template>
  <NModal
    v-model:show="modal.show"
    style="width: 420px; max-width: 90vw"
  >
    <NCard
      title="新建分类"
      size="small"
    >
      <NForm
        label-placement="top"
        label-width="auto"
        class="space-y-3"
      >
        <NFormItem
          label="名称"
          :show-feedback="false"
        >
          <NInput
            v-model:value="modal.name"
            placeholder="例如：随笔"
          />
        </NFormItem>
        <NFormItem
          label="短链接"
          :show-feedback="false"
        >
          <NInput
            v-model:value="modal.slug"
            placeholder="例如：notes"
          />
        </NFormItem>
      </NForm>
      <div class="mt-4 flex justify-end gap-2">
        <NButton
          quaternary
          @click="modal.show = false"
          >取消</NButton
        >
        <NButton
          type="primary"
          :loading="modal.loading"
          @click="$emit('create')"
          >创建并选择</NButton
        >
      </div>
    </NCard>
  </NModal>
</template>
