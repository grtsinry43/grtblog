<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NInput,
  NPopconfirm,
  NSpace,
  NSwitch,
  NTag,
  useMessage,
} from 'naive-ui'
import { computed, h, onMounted, reactive, ref } from 'vue'

import TemplateEditor from '@/components/template-editor/TemplateEditor.vue'
import {
  createOAuthProvider,
  deleteOAuthProvider,
  listOAuthProviders,
  updateOAuthProvider,
} from '@/services/oauth-providers'
import { listSysConfigs, updateSysConfigs } from '@/services/sysconfig'

import ConfigPanel from '../ConfigPanel'

import type { DataTableColumns } from 'naive-ui'
import type { AdminOAuthProvider, OAuthProviderPayload } from '@/services/oauth-providers'

const emit = defineEmits<{ 'dirty-change': [dirty: boolean] }>()

const message = useMessage()
const oauthLoading = ref(false)
const oauthSaving = ref(false)
const providers = ref<AdminOAuthProvider[]>([])

const configDirty = ref(false)

function handleConfigDirty(dirty: boolean) {
  configDirty.value = dirty
  emit('dirty-change', dirty)
}

const formVisible = ref(false)
const editing = ref<AdminOAuthProvider | null>(null)
const extraJsonError = ref<string | null>(null)

const form = reactive({
  key: '',
  displayName: '',
  clientId: '',
  clientSecret: '',
  authorizationEndpoint: '',
  tokenEndpoint: '',
  userinfoEndpoint: '',
  redirectUriTemplate: '',
  scopes: '',
  issuer: '',
  jwksUri: '',
  pkceRequired: false,
  enabled: true,
  extraParams: '{}',
})

const formTitle = computed(() => (editing.value ? '编辑 OAuth Provider' : '新增 OAuth Provider'))
const formActionLabel = computed(() => (editing.value ? '保存' : '创建'))

const presets: Array<{
  name: string
  payload: Partial<OAuthProviderPayload>
}> = [
  {
    name: 'GitHub',
    payload: {
      key: 'github',
      displayName: 'GitHub',
      authorizationEndpoint: 'https://github.com/login/oauth/authorize',
      tokenEndpoint: 'https://github.com/login/oauth/access_token',
      userinfoEndpoint: 'https://api.github.com/user',
      redirectUriTemplate: 'https://your-domain.com/auth/providers/{provider}/callback',
      scopes: 'read:user user:email',
      pkceRequired: false,
      enabled: true,
      extraParams: {},
    },
  },
  {
    name: 'Google',
    payload: {
      key: 'google',
      displayName: 'Google',
      authorizationEndpoint: 'https://accounts.google.com/o/oauth2/v2/auth',
      tokenEndpoint: 'https://oauth2.googleapis.com/token',
      userinfoEndpoint: 'https://openidconnect.googleapis.com/v1/userinfo',
      redirectUriTemplate: 'https://your-domain.com/auth/providers/{provider}/callback',
      scopes: 'openid profile email',
      pkceRequired: true,
      enabled: true,
      extraParams: {},
    },
  },
]

const columns = computed<DataTableColumns<AdminOAuthProvider>>(() => [
  {
    title: 'Key',
    key: 'key',
    width: 160,
    render: (row) => h('div', { class: 'font-medium' }, row.key),
  },
  {
    title: '显示名',
    key: 'displayName',
    width: 180,
  },
  {
    title: '状态',
    key: 'enabled',
    width: 100,
    render: (row) =>
      h(
        NTag,
        { type: row.enabled ? 'success' : 'default', size: 'small' },
        { default: () => (row.enabled ? '启用' : '停用') },
      ),
  },
  {
    title: 'Auth URL',
    key: 'authorizationEndpoint',
    render: (row) =>
      h('div', { class: 'text-xs text-[var(--text-color-3)]' }, row.authorizationEndpoint),
  },
  {
    title: '操作',
    key: 'actions',
    width: 140,
    render: (row) =>
      h(
        NSpace,
        { size: 8 },
        {
          default: () => [
            h(
              NButton,
              { size: 'tiny', tertiary: true, onClick: () => openEdit(row) },
              { default: () => '编辑' },
            ),
            h(
              NPopconfirm,
              { onPositiveClick: () => handleDelete(row) },
              {
                default: () => '确认删除该 Provider？',
                trigger: () =>
                  h(
                    NButton,
                    { size: 'tiny', type: 'error', tertiary: true },
                    { default: () => '删除' },
                  ),
              },
            ),
          ],
        },
      ),
  },
])

async function fetchProviders() {
  oauthLoading.value = true
  try {
    providers.value = (await listOAuthProviders()) || []
  } catch (err) {
    message.error(err instanceof Error ? err.message : '加载 OAuth Providers 失败')
  } finally {
    oauthLoading.value = false
  }
}

function resetForm() {
  form.key = ''
  form.displayName = ''
  form.clientId = ''
  form.clientSecret = ''
  form.authorizationEndpoint = ''
  form.tokenEndpoint = ''
  form.userinfoEndpoint = ''
  form.redirectUriTemplate = ''
  form.scopes = ''
  form.issuer = ''
  form.jwksUri = ''
  form.pkceRequired = false
  form.enabled = true
  form.extraParams = '{}'
  extraJsonError.value = null
}

function applyPreset(preset: { name: string; payload: Partial<OAuthProviderPayload> }) {
  editing.value = null
  resetForm()
  form.key = preset.payload.key ?? ''
  form.displayName = preset.payload.displayName ?? ''
  form.authorizationEndpoint = preset.payload.authorizationEndpoint ?? ''
  form.tokenEndpoint = preset.payload.tokenEndpoint ?? ''
  form.userinfoEndpoint = preset.payload.userinfoEndpoint ?? ''
  form.redirectUriTemplate = preset.payload.redirectUriTemplate ?? ''
  form.scopes = preset.payload.scopes ?? ''
  form.issuer = preset.payload.issuer ?? ''
  form.jwksUri = preset.payload.jwksUri ?? ''
  form.pkceRequired = !!preset.payload.pkceRequired
  form.enabled = preset.payload.enabled ?? true
  form.extraParams = JSON.stringify(preset.payload.extraParams ?? {}, null, 2)
  formVisible.value = true
}

function openCreate() {
  editing.value = null
  resetForm()
  formVisible.value = true
}

function openEdit(row: AdminOAuthProvider) {
  editing.value = row
  form.key = row.key
  form.displayName = row.displayName || ''
  form.clientId = row.clientId || ''
  form.clientSecret = ''
  form.authorizationEndpoint = row.authorizationEndpoint || ''
  form.tokenEndpoint = row.tokenEndpoint || ''
  form.userinfoEndpoint = row.userinfoEndpoint || ''
  form.redirectUriTemplate = row.redirectUriTemplate || ''
  form.scopes = row.scopes || ''
  form.issuer = row.issuer || ''
  form.jwksUri = row.jwksUri || ''
  form.pkceRequired = !!row.pkceRequired
  form.enabled = !!row.enabled
  form.extraParams = JSON.stringify(row.extraParams ?? {}, null, 2)
  extraJsonError.value = null
  formVisible.value = true
}

function buildPayload(): OAuthProviderPayload {
  let extraParams: Record<string, unknown> | undefined
  if (form.extraParams.trim()) {
    try {
      extraParams = JSON.parse(form.extraParams)
      extraJsonError.value = null
    } catch (err) {
      extraJsonError.value = err instanceof Error ? err.message : 'Extra Params JSON 格式不正确'
      throw new Error(extraJsonError.value)
    }
  }

  return {
    key: form.key.trim(),
    displayName: form.displayName.trim(),
    clientId: form.clientId.trim(),
    clientSecret: form.clientSecret.trim() || undefined,
    authorizationEndpoint: form.authorizationEndpoint.trim(),
    tokenEndpoint: form.tokenEndpoint.trim(),
    userinfoEndpoint: form.userinfoEndpoint.trim(),
    redirectUriTemplate: form.redirectUriTemplate.trim(),
    scopes: form.scopes.trim(),
    issuer: form.issuer.trim(),
    jwksUri: form.jwksUri.trim(),
    pkceRequired: form.pkceRequired,
    enabled: form.enabled,
    extraParams,
  }
}

async function handleSave() {
  if (oauthSaving.value) return
  const required = ['key', 'authorizationEndpoint', 'tokenEndpoint', 'redirectUriTemplate']
  const emptyKey = required.find((field) => !(form as any)[field]?.trim())
  if (emptyKey) {
    message.error('Key / Auth URL / Token URL / Redirect URI 不能为空')
    return
  }

  oauthSaving.value = true
  try {
    const payload = buildPayload()
    if (editing.value) {
      await updateOAuthProvider(editing.value.key, payload)
      message.success('已更新')
    } else {
      await createOAuthProvider(payload)
      message.success('已创建')
    }
    formVisible.value = false
    await fetchProviders()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '保存失败')
  } finally {
    oauthSaving.value = false
  }
}

async function handleDelete(row: AdminOAuthProvider) {
  try {
    await deleteOAuthProvider(row.key)
    message.success('已删除')
    await fetchProviders()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '删除失败')
  }
}

onMounted(fetchProviders)
</script>

<template>
  <div class="space-y-6">
    <!-- Turnstile config -->
    <ConfigPanel
      :list-fn="listSysConfigs"
      :update-fn="updateSysConfigs"
      title="Turnstile 人机验证"
      description="Cloudflare Turnstile 验证配置"
      :filter-groups="['security/turnstile']"
      :on-dirty-change="handleConfigDirty"
    />

    <!-- OAuth Providers -->
    <NCard>
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <div class="text-base font-semibold">OAuth 登录方式</div>
            <div class="text-xs text-neutral-500">管理 OAuth 登录提供方</div>
          </div>
          <div class="flex items-center gap-2">
            <NButton size="small" secondary :loading="oauthLoading" @click="fetchProviders">
              刷新
            </NButton>
            <NButton size="small" type="primary" @click="openCreate">新增 Provider</NButton>
            <NButton
              v-for="preset in presets"
              :key="preset.name"
              size="small"
              secondary
              @click="applyPreset(preset)"
            >
              {{ preset.name }} 预设
            </NButton>
          </div>
        </div>
      </template>

      <NDataTable :columns="columns" :data="providers" :loading="oauthLoading" :bordered="false" />
    </NCard>
  </div>

  <NDrawer v-model:show="formVisible" placement="right" :width="520">
    <NDrawerContent :title="formTitle" closable>
      <NForm label-placement="top" class="space-y-2">
        <NFormItem label="Key" required>
          <NInput v-model:value="form.key" :disabled="!!editing" placeholder="github / google" />
        </NFormItem>
        <NFormItem label="显示名">
          <NInput v-model:value="form.displayName" placeholder="GitHub" />
        </NFormItem>
        <NFormItem label="Client ID">
          <NInput v-model:value="form.clientId" placeholder="" />
        </NFormItem>
        <NFormItem label="Client Secret" :feedback="editing ? '留空则保持原值' : ''">
          <NInput v-model:value="form.clientSecret" type="password" placeholder="" />
        </NFormItem>
        <NFormItem label="Authorization Endpoint" required>
          <NInput
            v-model:value="form.authorizationEndpoint"
            placeholder="https://.../authorize"
          />
        </NFormItem>
        <NFormItem label="Token Endpoint" required>
          <NInput v-model:value="form.tokenEndpoint" placeholder="https://.../token" />
        </NFormItem>
        <NFormItem label="Userinfo Endpoint">
          <NInput v-model:value="form.userinfoEndpoint" placeholder="https://.../userinfo" />
        </NFormItem>
        <NFormItem label="Redirect URI Template" required>
          <NInput
            v-model:value="form.redirectUriTemplate"
            placeholder="https://.../auth/providers/{provider}/callback"
          />
        </NFormItem>
        <NFormItem label="Scopes">
          <NInput v-model:value="form.scopes" placeholder="openid profile email" />
        </NFormItem>
        <NFormItem label="Issuer">
          <NInput v-model:value="form.issuer" placeholder="" />
        </NFormItem>
        <NFormItem label="JWKS URI">
          <NInput v-model:value="form.jwksUri" placeholder="" />
        </NFormItem>
        <NFormItem label="PKCE Required">
          <NSwitch v-model:value="form.pkceRequired" />
        </NFormItem>
        <NFormItem label="启用">
          <NSwitch v-model:value="form.enabled" />
        </NFormItem>
        <NFormItem
          label="Extra Params"
          :feedback="extraJsonError || ''"
          :validation-status="extraJsonError ? 'error' : undefined"
        >
          <TemplateEditor v-model="form.extraParams" />
        </NFormItem>
        <div class="flex justify-end gap-2 pt-4">
          <NButton secondary @click="formVisible = false">取消</NButton>
          <NButton type="primary" :loading="oauthSaving" @click="handleSave">
            {{ formActionLabel }}
          </NButton>
        </div>
      </NForm>
    </NDrawerContent>
  </NDrawer>
</template>
