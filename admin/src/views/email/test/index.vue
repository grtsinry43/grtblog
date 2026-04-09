<script setup lang="ts">
import { NButton, NCard, NForm, NFormItem, NInput, NSelect, useMessage } from 'naive-ui'
import { onMounted, reactive, ref } from 'vue'

import { ScrollContainer } from '@/components'
import TemplateEditor from '@/components/template-editor/TemplateEditor.vue'
import { usePreviewData } from '@/composables/email/use-preview-data'
import { listEmailTemplates, testEmailTemplate } from '@/services/email'
import { getEventCatalogItem } from '@/services/events'

import type { EmailTemplate } from '@/services/email'

const message = useMessage()

const loading = ref(false)
const templateOptions = ref<{ label: string; value: string }[]>([])
const templatesList = ref<EmailTemplate[]>([])

const { generatePreviewData } = usePreviewData()
const currentEventName = ref('')

const form = reactive({
  code: '',
  toEmail: '',
  variables: '{\n  "Name": "Test User"\n}',
})

async function fetchTemplates() {
  const list = await listEmailTemplates()
  templatesList.value = list

  templateOptions.value = list.map((t) => ({ label: `${t.name} (${t.code})`, value: t.code }))
  const first = list[0]
  if (first) {
    form.code = first.code
    handleTemplateChange(form.code)
  }
}

async function handleTemplateChange(code: string) {
  const t = templatesList.value.find((i) => i.code === code)
  if (t && t.eventName) {
    currentEventName.value = t.eventName
    try {
      const details = await getEventCatalogItem(t.eventName)
      if (details && details.fields) {
        form.variables = generatePreviewData(details.fields)
      }
    } catch (e) {
      // ignore
    }
  }
}

async function handleSend() {
  if (!form.code || !form.toEmail) {
    message.error('请填写必要信息')
    return
  }
  loading.value = true
  try {
    let variables = {}
    try {
      if (form.variables) {
        variables = JSON.parse(form.variables)
      }
    } catch (e) {
      message.error('JSON 格式错误')
      loading.value = false
      return
    }

    await testEmailTemplate(form.code, {
      toEmails: [form.toEmail],
      variables: variables,
    })
    message.success('邮件发送任务已提交')
  } catch (err) {
    //
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchTemplates()
})
</script>

<template>
  <ScrollContainer wrapper-class="p-4">
    <NCard
      title="邮件发送测试"
      class="mx-auto max-w-2xl"
    >
      <div class="mb-4 text-sm text-[var(--text-color-3)]">
        此处可测试 SMTP 配置连通性及模版渲染效果。
      </div>

      <NForm
        label-placement="top"
        :disabled="loading"
      >
        <NFormItem label="选择模版">
          <NSelect
            v-model:value="form.code"
            :options="templateOptions"
            filterable
            placeholder="请选择"
            @update:value="handleTemplateChange"
          />
        </NFormItem>
        <NFormItem label="收件人">
          <NInput
            v-model:value="form.toEmail"
            placeholder="target@example.com"
          />
        </NFormItem>
        <NFormItem label="模版变量 (JSON)">
          <TemplateEditor v-model="form.variables" />
        </NFormItem>
        <NButton
          type="primary"
          block
          :loading="loading"
          @click="handleSend"
        >
          发送测试邮件
        </NButton>
      </NForm>
    </NCard>
  </ScrollContainer>
</template>
