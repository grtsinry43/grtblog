<script setup lang="ts">
import {
  NButton,
  NCard,
  NForm,
  NFormItem,
  NInput,
  NSwitch,
  useMessage,
} from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { createThinking, getThinking, updateThinking } from '@/services/thinking'
import type { FormInst } from 'naive-ui'

defineOptions({ name: 'ThinkingEdit' })

const route = useRoute()
const router = useRouter()
const message = useMessage()

const formRef = ref<FormInst | null>(null)
const formValue = ref({
  content: '',
  allowComment: true,
})
const saving = ref(false)

const id = computed(() => route.params.id as string | undefined)
const isCreating = computed(() => !id.value)

onMounted(() => {
  if (id.value) {
    getThinking(Number(id.value)).then((res) => {
      formValue.value.content = res.content
      formValue.value.allowComment = res.allowComment
    })
  }
})

async function handleSave() {
  try {
    saving.value = true
    if (isCreating.value) {
      await createThinking({
        content: formValue.value.content,
        allowComment: formValue.value.allowComment,
      })
      message.success('创建成功')
    } else {
      await updateThinking(Number(id.value), {
        content: formValue.value.content,
        allowComment: formValue.value.allowComment,
      })
      message.success('更新成功')
    }
    router.push({ name: 'thinkingList' })
  } catch (error) {
    message.error((error as Error).message)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="flex h-full min-h-0 flex-col p-10">
    <NCard>
      <NForm ref="formRef" :model="formValue" label-placement="top">
        <NFormItem label="内容" path="content">
          <NInput
            v-model:value="formValue.content"
            type="textarea"
            placeholder="分享一些思考..."
            :autosize="{ minRows: 5, maxRows: 15 }"
          />
        </NFormItem>
        <NFormItem label="选项">
           <div class="flex items-center gap-2">
            <span class="text-sm">允许评论</span>
            <NSwitch v-model:value="formValue.allowComment" />
           </div>
        </NFormItem>
      </NForm>
      <template #footer>
        <div class="flex justify-end">
          <NButton type="primary" :loading="saving" @click="handleSave">
            {{ isCreating ? '创建' : '更新' }}
          </NButton>
        </div>
      </template>
    </NCard>
  </div>
</template>
