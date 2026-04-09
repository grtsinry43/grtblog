<script setup lang="ts">
import { NButton, NForm, NFormItem, NGi, NGrid, NInput } from 'naive-ui'

import type { FormInst, FormItemRule } from 'naive-ui'

const form = defineModel<{ nickname: string; email: string; avatar: string }>('form', {
  required: true,
})

defineProps<{
  formRef: FormInst | null
  rules: Record<string, FormItemRule[]>
}>()

const emit = defineEmits<{
  'update:formRef': [value: FormInst | null]
  submit: []
}>()
</script>

<template>
  <div class="max-w-2xl">
    <NForm
      :ref="(el: any) => emit('update:formRef', el)"
      :model="form"
      :rules="rules"
      label-placement="top"
    >
      <NGrid
        cols="1 m:2"
        x-gap="24"
      >
        <NGi>
          <NFormItem
            label="昵称"
            path="nickname"
          >
            <NInput
              v-model:value="form.nickname"
              placeholder="请输入您的昵称"
            />
          </NFormItem>
        </NGi>
        <NGi>
          <NFormItem
            label="电子邮箱"
            path="email"
          >
            <NInput
              v-model:value="form.email"
              placeholder="请输入电子邮箱"
            />
          </NFormItem>
        </NGi>
      </NGrid>
      <NFormItem label="头像地址 (URL)">
        <NInput
          v-model:value="form.avatar"
          type="textarea"
          :rows="2"
          placeholder="如果您有外部头像链接，也可以直接填入此处"
        />
      </NFormItem>
      <div class="mt-4">
        <NButton
          type="primary"
          size="large"
          strong
          @click="emit('submit')"
          >保存基本信息</NButton
        >
      </div>
    </NForm>
  </div>
</template>
