<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NDivider,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NPopconfirm,
  NSelect,
  NSpace,
  NSwitch,
  NTag,
  useMessage,
} from 'naive-ui'
import { computed, h, onMounted, ref } from 'vue'

import {
  createAIModel,
  createAIProvider,
  deleteAIModel,
  deleteAIProvider,
  listAIModels,
  listAIProviders,
  updateAIModel,
  updateAIProvider,
} from '@/services/ai'
import { listSysConfigs, updateSysConfigs } from '@/services/sysconfig'

import ConfigPanel from '../ConfigPanel'

import type { AIModel, AIProvider } from '@/services/ai'
import type { DataTableColumns, SelectOption } from 'naive-ui'

const emit = defineEmits<{ 'dirty-change': [dirty: boolean] }>()
const message = useMessage()

// ── Providers ──

const providers = ref<AIProvider[]>([])
const providerLoading = ref(false)
const providerModalVisible = ref(false)
const providerSaving = ref(false)
const editingProvider = ref<AIProvider | null>(null)
const providerForm = ref({
  name: '',
  type: 'openai' as string,
  apiUrl: '',
  apiKey: '',
  isActive: true,
})

const providerTypeOptions: SelectOption[] = [
  { label: 'OpenAI 兼容', value: 'openai' },
  { label: 'OpenRouter', value: 'openrouter' },
  { label: 'Google Gemini', value: 'gemini' },
]

const providerColumns = computed<DataTableColumns<AIProvider>>(() => [
  { title: 'ID', key: 'id', width: 60 },
  { title: '名称', key: 'name', minWidth: 120 },
  {
    title: '类型',
    key: 'type',
    width: 130,
    render: (row) => {
      const map: Record<string, string> = {
        openai: 'OpenAI 兼容',
        openrouter: 'OpenRouter',
        gemini: 'Gemini',
      }
      return map[row.type] || row.type
    },
  },
  { title: 'API 地址', key: 'apiUrl', ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'isActive',
    width: 80,
    render: (row) =>
      h(
        NTag,
        { type: row.isActive ? 'success' : 'default', size: 'small', bordered: false },
        { default: () => (row.isActive ? '启用' : '禁用') },
      ),
  },
  {
    title: '操作',
    key: 'actions',
    width: 130,
    render: (row) =>
      h(NSpace, { size: 'small' }, () => [
        h(
          NButton,
          { size: 'small', tertiary: true, onClick: () => openEditProvider(row) },
          { default: () => '编辑' },
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => handleDeleteProvider(row) },
          {
            trigger: () =>
              h(
                NButton,
                { size: 'small', type: 'error', tertiary: true },
                { default: () => '删除' },
              ),
            default: () => '确认删除该提供商？关联的模型也会被删除。',
          },
        ),
      ]),
  },
])

async function fetchProviders() {
  providerLoading.value = true
  try {
    providers.value = await listAIProviders()
  } finally {
    providerLoading.value = false
  }
}

function openCreateProvider() {
  editingProvider.value = null
  providerForm.value = { name: '', type: 'openai', apiUrl: '', apiKey: '', isActive: true }
  providerModalVisible.value = true
}

function openEditProvider(row: AIProvider) {
  editingProvider.value = row
  providerForm.value = {
    name: row.name,
    type: row.type,
    apiUrl: row.apiUrl,
    apiKey: '',
    isActive: row.isActive,
  }
  providerModalVisible.value = true
}

async function handleSaveProvider() {
  if (!providerForm.value.name.trim()) {
    message.error('名称不能为空')
    return
  }
  providerSaving.value = true
  try {
    if (editingProvider.value) {
      const data: Record<string, unknown> = {
        name: providerForm.value.name,
        type: providerForm.value.type,
        apiUrl: providerForm.value.apiUrl,
        isActive: providerForm.value.isActive,
      }
      if (providerForm.value.apiKey) {
        data.apiKey = providerForm.value.apiKey
      }
      await updateAIProvider(editingProvider.value.id, data)
      message.success('提供商更新成功')
    } else {
      await createAIProvider(providerForm.value)
      message.success('提供商创建成功')
    }
    providerModalVisible.value = false
    await fetchProviders()
    await fetchModels()
  } finally {
    providerSaving.value = false
  }
}

async function handleDeleteProvider(row: AIProvider) {
  await deleteAIProvider(row.id)
  message.success('提供商删除成功')
  await fetchProviders()
  await fetchModels()
}

// ── Models ──

const models = ref<AIModel[]>([])
const modelLoading = ref(false)
const modelModalVisible = ref(false)
const modelSaving = ref(false)
const editingModel = ref<AIModel | null>(null)
const modelForm = ref({
  providerId: null as number | null,
  name: '',
  modelId: '',
  isActive: true,
})

const providerSelectOptions = computed<SelectOption[]>(() =>
  providers.value.map((p) => ({ label: `${p.name} (${p.type})`, value: p.id })),
)

const modelColumns = computed<DataTableColumns<AIModel>>(() => [
  { title: 'ID', key: 'id', width: 60 },
  { title: '模型名称', key: 'name', minWidth: 120 },
  { title: '模型 ID', key: 'modelId', minWidth: 140, ellipsis: { tooltip: true } },
  {
    title: '提供商',
    key: 'providerName',
    width: 150,
    render: (row) => row.providerName || `#${row.providerId}`,
  },
  {
    title: '状态',
    key: 'isActive',
    width: 80,
    render: (row) =>
      h(
        NTag,
        { type: row.isActive ? 'success' : 'default', size: 'small', bordered: false },
        { default: () => (row.isActive ? '启用' : '禁用') },
      ),
  },
  {
    title: '操作',
    key: 'actions',
    width: 130,
    render: (row) =>
      h(NSpace, { size: 'small' }, () => [
        h(
          NButton,
          { size: 'small', tertiary: true, onClick: () => openEditModel(row) },
          { default: () => '编辑' },
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => handleDeleteModel(row) },
          {
            trigger: () =>
              h(
                NButton,
                { size: 'small', type: 'error', tertiary: true },
                { default: () => '删除' },
              ),
            default: () => '确认删除该模型？',
          },
        ),
      ]),
  },
])

async function fetchModels() {
  modelLoading.value = true
  try {
    models.value = await listAIModels()
  } finally {
    modelLoading.value = false
  }
}

function openCreateModel() {
  editingModel.value = null
  modelForm.value = { providerId: null, name: '', modelId: '', isActive: true }
  modelModalVisible.value = true
}

function openEditModel(row: AIModel) {
  editingModel.value = row
  modelForm.value = {
    providerId: row.providerId,
    name: row.name,
    modelId: row.modelId,
    isActive: row.isActive,
  }
  modelModalVisible.value = true
}

async function handleSaveModel() {
  if (!modelForm.value.name.trim()) {
    message.error('模型名称不能为空')
    return
  }
  if (!modelForm.value.modelId.trim()) {
    message.error('模型 ID 不能为空')
    return
  }
  if (!modelForm.value.providerId) {
    message.error('请选择提供商')
    return
  }
  modelSaving.value = true
  try {
    if (editingModel.value) {
      await updateAIModel(editingModel.value.id, {
        providerId: modelForm.value.providerId,
        name: modelForm.value.name,
        modelId: modelForm.value.modelId,
        isActive: modelForm.value.isActive,
      })
      message.success('模型更新成功')
    } else {
      await createAIModel({
        providerId: modelForm.value.providerId!,
        name: modelForm.value.name,
        modelId: modelForm.value.modelId,
        isActive: modelForm.value.isActive,
      })
      message.success('模型创建成功')
    }
    modelModalVisible.value = false
    await fetchModels()
  } finally {
    modelSaving.value = false
  }
}

async function handleDeleteModel(row: AIModel) {
  await deleteAIModel(row.id)
  message.success('模型删除成功')
  await fetchModels()
}

// ── Init ──

onMounted(async () => {
  await Promise.all([fetchProviders(), fetchModels()])
})
</script>

<template>
  <!-- 提供商管理 -->
  <NCard
    title="AI 提供商"
    class="mb-4"
  >
    <template #header-extra>
      <NButton
        type="primary"
        size="small"
        @click="openCreateProvider"
        >新增提供商</NButton
      >
    </template>
    <NDataTable
      :loading="providerLoading"
      :columns="providerColumns"
      :data="providers"
      :bordered="false"
      size="small"
    />
  </NCard>

  <!-- 模型管理 -->
  <NCard
    title="AI 模型"
    class="mb-4"
  >
    <template #header-extra>
      <NButton
        type="primary"
        size="small"
        @click="openCreateModel"
        >新增模型</NButton
      >
    </template>
    <NDataTable
      :loading="modelLoading"
      :columns="modelColumns"
      :data="models"
      :bordered="false"
      size="small"
    />
  </NCard>

  <!-- 任务配置（提示词和模型分配） -->
  <NDivider />
  <ConfigPanel
    :list-fn="listSysConfigs"
    :update-fn="updateSysConfigs"
    title="AI 任务配置"
    description="启用开关、任务模型分配和提示词配置"
    :filter-groups="['ai', 'ai/task', 'ai/prompt']"
    :on-dirty-change="(dirty: boolean) => emit('dirty-change', dirty)"
  />

  <!-- 提供商编辑弹窗 -->
  <NModal
    v-model:show="providerModalVisible"
    preset="card"
    :title="editingProvider ? '编辑提供商' : '新增提供商'"
    style="width: 560px"
  >
    <NForm
      label-placement="left"
      label-width="96"
    >
      <NFormItem
        label="名称"
        required
      >
        <NInput
          v-model:value="providerForm.name"
          placeholder="如：OpenAI 官方"
        />
      </NFormItem>
      <NFormItem
        label="类型"
        required
      >
        <NSelect
          v-model:value="providerForm.type"
          :options="providerTypeOptions"
        />
      </NFormItem>
      <NFormItem label="API 地址">
        <NInput
          v-model:value="providerForm.apiUrl"
          placeholder="留空使用默认地址"
        />
      </NFormItem>
      <NFormItem :label="editingProvider ? 'API Key（留空不修改）' : 'API Key'">
        <NInput
          v-model:value="providerForm.apiKey"
          type="password"
          show-password-on="click"
          placeholder="sk-..."
        />
      </NFormItem>
      <NFormItem label="启用">
        <NSwitch v-model:value="providerForm.isActive" />
      </NFormItem>
    </NForm>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="providerModalVisible = false">取消</NButton>
        <NButton
          type="primary"
          :loading="providerSaving"
          @click="handleSaveProvider"
          >保存</NButton
        >
      </NSpace>
    </template>
  </NModal>

  <!-- 模型编辑弹窗 -->
  <NModal
    v-model:show="modelModalVisible"
    preset="card"
    :title="editingModel ? '编辑模型' : '新增模型'"
    style="width: 560px"
  >
    <NForm
      label-placement="left"
      label-width="96"
    >
      <NFormItem
        label="提供商"
        required
      >
        <NSelect
          v-model:value="modelForm.providerId"
          :options="providerSelectOptions"
          placeholder="请选择提供商"
        />
      </NFormItem>
      <NFormItem
        label="显示名称"
        required
      >
        <NInput
          v-model:value="modelForm.name"
          placeholder="如：GPT-4o"
        />
      </NFormItem>
      <NFormItem
        label="模型 ID"
        required
      >
        <NInput
          v-model:value="modelForm.modelId"
          placeholder="如：gpt-4o、gemini-2.0-flash"
        />
      </NFormItem>
      <NFormItem label="启用">
        <NSwitch v-model:value="modelForm.isActive" />
      </NFormItem>
    </NForm>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="modelModalVisible = false">取消</NButton>
        <NButton
          type="primary"
          :loading="modelSaving"
          @click="handleSaveModel"
          >保存</NButton
        >
      </NSpace>
    </template>
  </NModal>
</template>
