<script setup lang="ts">
import { NButton, NDivider, NForm, NFormItem, NInput } from 'naive-ui'

import type { FormInst, FormItemRule } from 'naive-ui'

defineProps<{
  formRef: FormInst | null
  form: { oldPassword: string; newPassword: string; confirmPassword: string }
  rules: Record<string, FormItemRule[]>
}>()

const emit = defineEmits<{
  'update:formRef': [value: FormInst | null]
  submit: []
}>()
</script>

<template>
  <div class="mx-auto max-w-lg pt-4">
    <div class="mb-8">
      <div class="text-lg font-medium">修改账户密码</div>
      <div class="text-sm text-neutral-400">为了您的账户安全，建议定期更换高强度密码</div>
    </div>

    <NForm
      :ref="(el: any) => emit('update:formRef', el)"
      :model="form"
      :rules="rules"
      label-placement="top"
    >
      <NFormItem label="当前密码" path="oldPassword">
        <NInput
          v-model:value="form.oldPassword"
          type="password"
          show-password-on="click"
          placeholder="输入旧密码进行身份验证"
        />
      </NFormItem>
      <NDivider />
      <NFormItem label="新密码" path="newPassword">
        <NInput
          v-model:value="form.newPassword"
          type="password"
          show-password-on="click"
          placeholder="设置您的新密码"
        />
      </NFormItem>
      <NFormItem label="确认新密码" path="confirmPassword">
        <NInput
          v-model:value="form.confirmPassword"
          type="password"
          show-password-on="click"
          placeholder="再次输入新密码"
        />
      </NFormItem>
      <div class="mt-6">
        <NButton type="primary" block size="large" @click="emit('submit')">确认更改密码</NButton>
      </div>
    </NForm>
  </div>
</template>
