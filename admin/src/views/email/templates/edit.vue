<script setup lang="ts">
import { ArrowLeft24Regular } from '@vicons/fluent'
import {
  NButton,
  NCard,
  NCode,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NModal,
  NSelect,
  NSpace,
  NSwitch,
  NTabPane,
  NTabs,
  useMessage,
  NTag,
} from 'naive-ui'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import HtmlEditor from '@/components/html-editor/HtmlEditor.vue'
import TemplateEditor from '@/components/template-editor/TemplateEditor.vue'
import {
  createEmailTemplate,
  listEmailTemplates,
  updateEmailTemplate,
  previewEmailTemplate,
} from '@/services/email'
import { getEventCatalogItem, listEvents } from '@/services/events'

import type { AdminEventFieldResp } from '@/services/events'

const message = useMessage()
const route = useRoute()
const router = useRouter()

const isEdit = computed(() => !!route.params.code)
const loading = ref(false)
const saving = ref(false)
const isInternal = ref(false)
const eventOptions = ref<{ label: string; value: string }[]>([])
const currentEventFields = ref<AdminEventFieldResp[]>([])

const form = reactive({
  code: '',
  name: '',
  eventName: '',
  subjectTemplate: '',
  htmlTemplate: '',
  textTemplate: '',
  toEmails: [] as string[],
  isEnabled: true,
})

const previewModalVisible = ref(false)
const previewLoading = ref(false)
const previewData = reactive({
  subject: '',
  htmlBody: '',
  textBody: '',
  variables: '{\n  "Name": "Test User",\n  "Event": {}\n}',
})

// Extract valid variable names for the editor
const validVariables = computed(() => {
  return currentEventFields.value.map((f) => f.name)
})

async function fetchEvents() {
  const { groups } = await listEvents('email')
  // Flatten groups for the select options
  const options: { label: string; value: string }[] = []
  for (const group of groups) {
    for (const event of group.events) {
      options.push({ label: event, value: event })
    }
  }
  eventOptions.value = options
}

async function fetchEventDetails(eventName: string) {
  if (!eventName) {
    currentEventFields.value = []
    return
  }
  try {
    const details = await getEventCatalogItem(eventName)
    currentEventFields.value = details.fields
  } catch (e) {
    console.error('Failed to fetch event details', e)
  }
}

watch(
  () => form.eventName,
  (newVal) => {
    fetchEventDetails(newVal)
  },
)

watch(
  () => route.params.code,
  (newCode) => {
    if (!newCode) {
      form.code = ''
      form.name = ''
      form.eventName = ''
      form.subjectTemplate = ''
      form.htmlTemplate = ''
      form.textTemplate = ''
      form.toEmails = []
      form.isEnabled = true
      isInternal.value = false
      currentEventFields.value = []
    } else {
      fetchDetail()
    }
  },
)

async function fetchDetail() {
  if (!isEdit.value) return
  loading.value = true
  try {
    const list = await listEmailTemplates()
    const target = list.find((item) => item.code === route.params.code)
    if (target) {
      form.code = target.code
      form.name = target.name
      form.eventName = target.eventName
      form.subjectTemplate = target.subjectTemplate
      form.htmlTemplate = target.htmlTemplate
      form.textTemplate = target.textTemplate
      form.toEmails = target.toEmails || []
      form.isEnabled = target.isEnabled
      isInternal.value = target.isInternal

      // Trigger fetch details
      if (form.eventName) {
        await fetchEventDetails(form.eventName)
      }
    } else {
      message.error('模版不存在')
      router.push('/email/templates')
    }
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  if (!form.code || !form.name || !form.eventName) {
    message.error('请填写必填项')
    return
  }

  saving.value = true
  try {
    if (isEdit.value) {
      await updateEmailTemplate(form.code, form)
      message.success('更新成功')
    } else {
      await createEmailTemplate(form)
      message.success('创建成功')
      router.replace(`/email/templates/${form.code}`)
    }
  } finally {
    saving.value = false
  }
}

async function handlePreview() {
  previewLoading.value = true
  try {
    let variables = {}
    try {
      if (previewData.variables) {
        variables = JSON.parse(previewData.variables)
      }
    } catch (e) {
      message.error('JSON 格式错误，请检查变量')
      return
    }

    const res = await previewEmailTemplate(form.code, {
      variables: variables,
    })
    previewData.subject = res.subject
    previewData.htmlBody = res.htmlBody
    previewData.textBody = res.textBody
  } catch (err) {
    //
  } finally {
    previewLoading.value = false
  }
}

import { usePreviewData } from '@/composables/email/use-preview-data'

const { generatePreviewData: generate } = usePreviewData()

const defaultVariables = '{\n  "Name": "Test User"\n}'

function generatePreviewData() {
  return generate(currentEventFields.value)
}

watch(previewModalVisible, (visible) => {
  if (visible) {
    // Only regenerate if it looks like default or empty
    const current = previewData.variables.trim()
    if (!current || current === defaultVariables.trim() || current === '{}') {
      previewData.variables = generatePreviewData()
    }
    // If user changed event, we might want to prompt or force update?
    // For now, let's also check if the current valid variables are present in the JSON?
    // Or simpler: just check if we have event fields.
    if (form.eventName && currentEventFields.value.length > 0) {
      // If the current JSON doesn't contain keys from the current event, maybe regenerate?
      // This is tricky. Let's just stick to "if default, regenerate".
      // And maybe an explicit 'Reset' button in the UI could be nice?
      // For now, let's always regenerate if the user hasn't touched it significantly?
      // Let's rely on the check against defaultVariables.

      // Actually, if the user switched events, defaultVariables check might fail if they edited it for PREVIOUS event.
      // Let's regenerate if the eventName changed since last time?
      // Simpler: Just regenerate it. The user can edit it.
      previewData.variables = generatePreviewData()
    }
  }
})

function handleBack() {
  router.back()
}

onMounted(async () => {
  await fetchEvents()
  await fetchDetail()
})
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-4">
    <NCard>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <NButton
            quaternary
            circle
            @click="handleBack"
          >
            <template #icon>
              <NIcon><ArrowLeft24Regular /></NIcon>
            </template>
          </NButton>
          <div class="text-lg font-bold">{{ isEdit ? '编辑模版' : '新建模版' }}</div>
        </div>
        <NSpace>
          <NButton
            v-if="isEdit"
            secondary
            @click="previewModalVisible = true"
          >
            预览 & 调试
          </NButton>
          <NButton
            type="primary"
            :loading="saving"
            @click="handleSave"
          >
            保存
          </NButton>
        </NSpace>
      </div>
    </NCard>

    <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <div class="lg:col-span-2">
        <NCard
          title="模版内容"
          content-style="padding-bottom: 0;"
        >
          <NFormItem label="邮件主题">
            <NInput
              v-model:value="form.subjectTemplate"
              placeholder="支持模版变量，如：Welcome {{.Name}}"
            />
          </NFormItem>

          <div
            v-if="currentEventFields.length > 0"
            class="mb-4"
          >
            <div class="mb-2 text-xs text-gray-500">可用事件变量:</div>
            <NSpace size="small">
              <NTag
                v-for="field in currentEventFields"
                :key="field.name"
                size="small"
                type="info"
                dashed
              >
                {{ field.name }}
                <template
                  #avatar
                  v-if="field.required"
                  ><span class="text-red-500">*</span></template
                >
              </NTag>
            </NSpace>
          </div>

          <NTabs
            type="line"
            animated
          >
            <NTabPane
              name="html"
              tab="HTML 内容"
            >
              <HtmlEditor
                v-model="form.htmlTemplate"
                :valid-variables="validVariables"
                class="min-h-[500px]"
              />
            </NTabPane>
            <NTabPane
              name="text"
              tab="纯文本内容"
            >
              <NInput
                v-model:value="form.textTemplate"
                type="textarea"
                placeholder="纯文本备选内容..."
                :rows="20"
                class="font-mono"
              />
            </NTabPane>
          </NTabs>
        </NCard>
      </div>
      <div>
        <NCard title="基本设置">
          <NForm
            label-placement="top"
            :disabled="loading"
          >
            <NFormItem label="模版编码 (Key)">
              <NInput
                v-model:value="form.code"
                placeholder="唯一标识，如：welcome_email"
                :disabled="isEdit"
              />
            </NFormItem>
            <NFormItem label="模版名称">
              <NInput
                v-model:value="form.name"
                placeholder="便于管理的名称"
              />
            </NFormItem>
            <NFormItem label="触发事件">
              <NSelect
                v-model:value="form.eventName"
                :options="eventOptions"
                filterable
                :disabled="isInternal"
                placeholder="选择关联的系统事件"
              />
              <div
                v-if="isInternal"
                class="mt-1 text-xs text-neutral-400"
              >
                内置模版不允许更改触发事件
              </div>
            </NFormItem>
            <NFormItem label="默认收件人">
              <NSelect
                v-model:value="form.toEmails"
                multiple
                tag
                filterable
                placeholder="输入邮箱后回车..."
              />
            </NFormItem>
            <NFormItem label="启用">
              <NSwitch v-model:value="form.isEnabled" />
            </NFormItem>
          </NForm>
        </NCard>
      </div>
    </div>

    <NModal
      v-model:show="previewModalVisible"
      preset="card"
      style="width: 900px; max-width: 90vw"
      title="预览与调试"
    >
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div class="flex flex-col gap-4">
          <NCard
            title="测试数据 (JSON)"
            size="small"
          >
            <template #header-extra>
              <NButton
                size="small"
                secondary
                @click="handlePreview"
                :loading="previewLoading"
                >渲染预览</NButton
              >
            </template>
            <TemplateEditor v-model="previewData.variables" />
          </NCard>
        </div>
        <div class="flex flex-col gap-4">
          <div class="rounded border p-4">
            <div class="mb-2 border-b pb-2 font-bold">Subject: {{ previewData.subject }}</div>
            <iframe
              class="h-[600px] w-full border-0"
              :srcdoc="previewData.htmlBody"
              sandbox="allow-same-origin allow-scripts"
            ></iframe>
          </div>
          <NCard
            title="HTML Source"
            size="small"
            embedded
          >
            <NCode
              :code="previewData.htmlBody"
              language="html"
              word-wrap
            />
          </NCard>
        </div>
      </div>
    </NModal>
  </ScrollContainer>
</template>
