<script setup lang="ts">
import {
  NButton,
  NConfigProvider,
  NForm,
  NFormItem,
  NGlobalStyle,
  NH1,
  NH2,
  NIcon,
  NInput,
  NResult,
  NSpace,
  NSpin,
  NStep,
  NSteps,
  NText,
  darkTheme,
  useMessage,
  useOsTheme,
} from 'naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'
import {
  ArrowForwardOutline,
  DesktopOutline,
  EarthOutline,
  KeyOutline,
  LinkOutline,
  LockClosedOutline,
  MailOutline,
  PersonOutline,
  PlanetOutline,
  RocketOutline,
} from '@vicons/ionicons5'

import router from '@/router'
import { getSetupState, login, register } from '@/services/auth'
import { ApiError } from '@/services/http'
import { updateWebsiteInfo } from '@/services/website-info'
import { useUserStore } from '@/stores'

import type { FormItemRule, GlobalThemeOverrides } from 'naive-ui'

defineOptions({
  name: 'InitPage',
})

const message = useMessage()
const userStore = useUserStore()
const osTheme = useOsTheme()

const theme = computed(() => (osTheme.value === 'dark' ? darkTheme : null))

// Custom Theme Overrides for "Exquisite" Feel
const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#18a058',
    primaryColorHover: '#36ad6a',
    primaryColorPressed: '#0c7a43',
    borderRadius: '8px', // Slightly softer corners
  },
  Input: {
    paddingLarge: '12px 16px',
    fontSizeLarge: '16px',
    borderRadius: '8px',
  },
  Button: {
    heightLarge: '46px',
    fontSizeLarge: '16px',
    fontWeight: '600',
    borderRadius: '8px',
  },
  Card: {
    borderRadius: '16px',
    boxShadow1: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
  },
}

const loadingState = ref(true)
const submitting = ref(false)
const setupState = ref<Awaited<ReturnType<typeof getSetupState>> | null>(null)
const formRef = ref<InstanceType<typeof NForm> | null>(null)
const currentStep = ref(1)

const form = reactive({
  username: '',
  nickname: '',
  email: '',
  password: '',
  confirmPassword: '',
  websiteName: 'grtBlog',
  publicUrl: '',
  description: '',
  keywords: '',
})

const rules: Record<string, FormItemRule[]> = {
  username: [{ required: true, message: '请输入管理员账号', trigger: ['input', 'blur'] }],
  password: [{ required: true, message: '请输入密码', trigger: ['input', 'blur'] }],
  confirmPassword: [
    { required: true, message: '请再次输入密码', trigger: ['input', 'blur'] },
    {
      validator: () => form.password === form.confirmPassword,
      message: '两次输入的密码不一致',
      trigger: ['input', 'blur'],
    },
  ],
  websiteName: [{ required: true, message: '请输入站点名称', trigger: ['input', 'blur'] }],
  publicUrl: [{ required: true, message: '请输入站点公开地址', trigger: ['input', 'blur'] }],
}

const needsAccountSetup = computed(() => !setupState.value?.hasUser)
const needsWebsiteSetup = computed(() => !setupState.value?.websiteInfoReady)

function normalizePublicURL(url: string) {
  const trimmed = url.trim()
  if (!trimmed) return ''
  return trimmed.replace(/\/+$/, '')
}

async function loadSetupState() {
  loadingState.value = true
  try {
    const state = await getSetupState()
    setupState.value = state
    if (!state.needsSetup) {
      await router.replace({ name: 'signIn' })
      return
    }
    if (!state.hasUser) {
      form.publicUrl = window.location.origin
    }
  } catch (error) {
    if (!(error instanceof ApiError)) {
      message.error('获取初始化状态失败，请稍后重试')
    }
  } finally {
    loadingState.value = false
  }
}

async function handleNextStep() {
  try {
    await formRef.value?.validate()
    if (currentStep.value === 1) {
      currentStep.value = 2
    } else {
      await submitSetup()
    }
  } catch (e) {
    // Validation failed
  }
}

async function submitSetup() {
  submitting.value = true
  try {
    if (needsAccountSetup.value) {
      await register({
        username: form.username.trim(),
        nickname: form.nickname.trim() || form.username.trim(),
        email: form.email.trim() || undefined,
        password: form.password,
      })
      const loginResp = await login({
        credential: form.username.trim(),
        password: form.password,
      })
      userStore.setAuth({
        token: loginResp.token,
        user: {
          id: loginResp.user.id,
          username: loginResp.user.username,
          nickname: loginResp.user.nickname,
          email: loginResp.user.email,
          avatar: loginResp.user.avatar,
          isAdmin: loginResp.user.isAdmin,
          roles: loginResp.roles,
          permissions: loginResp.permissions,
          createdAt: loginResp.user.createdAt,
          updatedAt: loginResp.user.updatedAt,
        },
      })
    }

    if (needsWebsiteSetup.value) {
      const tasks: Promise<unknown>[] = []
      const websiteName = form.websiteName.trim()
      const publicURL = normalizePublicURL(form.publicUrl)
      const description = form.description.trim()
      const keywords = form.keywords.trim()
      if (websiteName) {
        tasks.push(updateWebsiteInfo('website_name', { value: websiteName }))
      }
      if (publicURL) {
        tasks.push(updateWebsiteInfo('public_url', { value: publicURL }))
        tasks.push(updateWebsiteInfo('api_url', { value: `${publicURL}/api/v2` }))
      }
      if (description) {
        tasks.push(updateWebsiteInfo('description', { value: description }))
      }
      if (keywords) {
        tasks.push(updateWebsiteInfo('keywords', { value: keywords }))
      }
      await Promise.all(tasks)
    }

    message.success('初始化完成')
    await router.replace({ path: '/' })
  } catch (error) {
    if (error instanceof ApiError) return
    message.error('初始化失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadSetupState()
})
</script>

<template>
  <NConfigProvider :theme="theme" :theme-overrides="themeOverrides">
    <NGlobalStyle />
    <div class="init-layout">
      <!-- Loading State -->
      <NSpin :show="loadingState" class="loading-spin" size="large" v-if="loadingState">
        <template #description>正在加载环境...</template>
      </NSpin>

      <template v-else-if="setupState">
        <!-- New Setup Split Layout -->
        <div v-if="!setupState.hasUser" class="split-container">
          <!-- Left: Brand -->
          <div class="brand-panel">
            <div class="brand-content">
              <div class="logo-area">
                <NIcon size="48" color="#18a058">
                  <RocketOutline />
                </NIcon>
                <NH1 class="brand-title">grtBlog</NH1>
              </div>
              <div class="brand-message">
                <NH2>开启您的创作之旅</NH2>
                <NText class="brand-desc" depth="3">
                  只需几步，即可构建属于您的现代化博客。
                  <br />
                  极致体验，原生交互，即刻出发。
                </NText>
              </div>
              <div class="brand-footer">
                <NText depth="3" class="version-text">v2.0.0</NText>
              </div>
            </div>
            <!-- Decorative Circle -->
            <div class="decoration-circle" />
          </div>

          <!-- Right: Form -->
          <div class="form-panel">
            <div class="form-container">
              <div class="form-header">
                <NH2 class="step-title">
                  {{ currentStep === 1 ? '创建管理员' : '站点基本信息' }}
                </NH2>
                <NSteps :current="currentStep" size="small" class="form-steps">
                  <NStep title="账户" />
                  <NStep title="站点" />
                </NSteps>
              </div>

              <NForm
                ref="formRef"
                :model="form"
                :rules="rules"
                label-placement="top"
                :show-require-mark="false"
                class="main-form"
              >
                <!-- Step 1: Admin Account -->
                <Transition name="slide-fade" mode="out-in">
                  <div v-if="currentStep === 1" key="step1" class="form-step-content">
                    <NFormItem label="账号" path="username">
                      <NInput v-model:value="form.username" placeholder="请输入管理员账号" size="large">
                        <template #prefix>
                          <NIcon :component="PersonOutline" />
                        </template>
                      </NInput>
                    </NFormItem>
                    <NFormItem label="昵称">
                      <NInput v-model:value="form.nickname" placeholder="显示的昵称 (可选)" size="large">
                        <template #prefix>
                          <NIcon :component="DesktopOutline" />
                        </template>
                      </NInput>
                    </NFormItem>
                    <NFormItem label="密码" path="password">
                      <NInput
                        v-model:value="form.password"
                        type="password"
                        show-password-on="click"
                        placeholder="请输入密码"
                        size="large"
                      >
                        <template #prefix>
                          <NIcon :component="KeyOutline" />
                        </template>
                      </NInput>
                    </NFormItem>
                    <NFormItem label="确认密码" path="confirmPassword">
                      <NInput
                        v-model:value="form.confirmPassword"
                        type="password"
                        show-password-on="click"
                        placeholder="请再次输入密码"
                        size="large"
                      >
                        <template #prefix>
                          <NIcon :component="LockClosedOutline" />
                        </template>
                      </NInput>
                    </NFormItem>
                    <NFormItem label="邮箱">
                      <NInput v-model:value="form.email" placeholder="通知邮箱 (可选)" size="large">
                        <template #prefix>
                          <NIcon :component="MailOutline" />
                        </template>
                      </NInput>
                    </NFormItem>
                  </div>
                  <!-- Step 2: Site Info -->
                  <div v-else key="step2" class="form-step-content">
                    <NFormItem label="站点名称" path="websiteName">
                      <NInput v-model:value="form.websiteName" placeholder="例如：My Awesome Blog" size="large">
                        <template #prefix>
                          <NIcon :component="PlanetOutline" />
                        </template>
                      </NInput>
                    </NFormItem>
                    <NFormItem label="公开地址" path="publicUrl">
                      <NInput v-model:value="form.publicUrl" placeholder="例如：https://grtsinry.com" size="large">
                        <template #prefix>
                          <NIcon :component="LinkOutline" />
                        </template>
                      </NInput>
                    </NFormItem>
                    <NFormItem label="站点描述">
                      <NInput
                        v-model:value="form.description"
                        type="textarea"
                        placeholder="简单介绍一下您的博客..."
                        :rows="3"
                        size="large"
                      />
                    </NFormItem>
                    <NFormItem label="关键词">
                      <NInput v-model:value="form.keywords" placeholder="技术, 生活, 随笔 (逗号分隔)" size="large">
                        <template #prefix>
                          <NIcon :component="EarthOutline" />
                        </template>
                      </NInput>
                    </NFormItem>
                  </div>
                </Transition>
              </NForm>

              <div class="form-actions">
                <NSpace justify="space-between" align="center">
                  <NButton v-if="currentStep > 1" text @click="currentStep--">
                    <template #icon>
                      <NIcon><ArrowForwardOutline style="transform: rotate(180deg)" /></NIcon>
                    </template>
                    返回上一步
                  </NButton>
                  <div v-else></div> <!-- Spacer -->

                  <NButton
                    type="primary"
                    size="large"
                    :loading="submitting"
                    @click="handleNextStep"
                    class="next-btn"
                  >
                    {{ currentStep === 2 ? '完成初始化' : '下一步' }}
                    <template #icon v-if="currentStep === 1">
                      <NIcon><ArrowForwardOutline /></NIcon>
                    </template>
                  </NButton>
                </NSpace>
              </div>
            </div>
          </div>
        </div>

        <!-- Existing User State (Full Centered) -->
        <div v-else class="existing-state-container">
          <div class="existing-content">
            <NResult
              status="info"
              title="准备就绪"
              description="系统检测到管理员账户已存在，无需重复初始化。"
              size="large"
            >
              <template #footer>
                <NButton type="primary" size="large" @click="router.replace({ name: 'signIn' })">
                  立即登录
                </NButton>
              </template>
            </NResult>
          </div>
        </div>
      </template>
    </div>
  </NConfigProvider>
</template>

<style scoped>
.init-layout {
  min-height: 100vh;
  width: 100vw;
  display: flex;
  background-color: var(--n-color);
  transition: background-color 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.loading-spin {
  margin: auto;
}

/* Split Container */
.split-container {
  display: flex;
  width: 100%;
  height: 100vh;
  overflow: hidden;
}

/* Brand Panel (Left) */
.brand-panel {
  flex: 0 0 40%;
  background-color: rgba(24, 160, 88, 0.05); /* Light Green Tint */
  display: flex;
  flex-direction: column;
  justify-content: center;
  position: relative;
  overflow: hidden;
  padding: 60px;
}

/* Dark mode adjustment for brand panel */
:deep(.n-config-provider--theme-dark) .brand-panel {
  background-color: rgba(0, 0, 0, 0.2);
}

.brand-content {
  position: relative;
  z-index: 10;
  max-width: 480px;
  margin-left: auto;
  margin-right: 60px;
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 40px;
}

.brand-title {
  margin: 0;
  font-size: 32px;
  font-weight: 800;
  letter-spacing: -1px;
}

.brand-message h2 {
  font-size: 40px;
  font-weight: 700;
  line-height: 1.2;
  margin-bottom: 24px;
}

.brand-desc {
  font-size: 18px;
  line-height: 1.6;
  opacity: 0.8;
}

.brand-footer {
  margin-top: 80px;
}

.decoration-circle {
  position: absolute;
  top: -20%;
  right: -20%;
  width: 80%;
  padding-bottom: 80%;
  background: radial-gradient(circle, rgba(24, 160, 88, 0.1) 0%, rgba(0, 0, 0, 0) 70%);
  border-radius: 50%;
  z-index: 1;
}

/* Form Panel (Right) */
.form-panel {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  background-color: var(--n-color);
}

.form-container {
  width: 100%;
  max-width: 520px;
}

.form-header {
  margin-bottom: 40px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.step-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.form-steps {
  width: 160px;
}

.main-form {
  min-height: 400px; /* Prevent layout jump */
}

/* Form Actions */
.form-actions {
  margin-top: 40px;
}

.next-btn {
  padding-left: 32px;
  padding-right: 32px;
}

/* Existing State */
.existing-state-container {
  display: flex;
  width: 100%;
  height: 100vh;
  align-items: center;
  justify-content: center;
}

.existing-content {
  max-width: 600px;
  padding: 60px;
  text-align: center;
}

/* Transitions */
.slide-fade-enter-active,
.slide-fade-leave-active {
  transition: all 0.3s ease-out;
}

.slide-fade-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.slide-fade-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

/* Mobile Responsive */
@media (max-width: 1024px) {
  .brand-panel {
    display: none; /* Hide branding on tablet/mobile for focus */
  }
  
  .form-panel {
    padding: 20px;
  }
}
</style>
