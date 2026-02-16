<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'

import {
  NButton,
  NCard,
  NDivider,
  NForm,
  NFormItem,
  NGi,
  NGrid,
  NInput,
  NModal,
  NStatistic,
  NTabPane,
  NTabs,
  NTag,
  NUpload,
  useMessage,
} from 'naive-ui'
import VuePictureCropper, { cropper } from 'vue-picture-cropper'

import { ScrollContainer, UserAvatar } from '@/components'
import {
  changePassword,
  getAccessInfo,
  getOAuthBindings,
  updateProfile,
} from '@/services/auth'
import { uploadFile } from '@/services/uploads'
import { toRefsUserStore, useUserStore } from '@/stores'

import type { OAuthBinding } from '@/services/auth'
import type { FormInst, FormItemRule, UploadCustomRequestOptions } from 'naive-ui'

defineOptions({ name: 'UserCenter' })

const userStore = useUserStore()
const { user, token } = toRefsUserStore()
const message = useMessage()

// --- State ---
const profileFormRef = ref<FormInst | null>(null)
const passwordFormRef = ref<FormInst | null>(null)
const oauthLoading = ref(false)
const oauthBindings = ref<OAuthBinding[]>([])

const profileForm = reactive({
  nickname: '',
  email: '',
  avatar: '',
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

// --- Cropper State ---
const showCropper = ref(false)
const cropperImg = ref('')
const isUploading = ref(false)

// --- Rules ---
const profileRules: Record<string, FormItemRule[]> = {
  nickname: [{ required: true, message: '请输入昵称', trigger: ['blur', 'input'] }],
  email: [{ type: 'email', message: '请输入有效邮箱', trigger: ['blur', 'input'] }],
}

const passwordRules: Record<string, FormItemRule[]> = {
  oldPassword: [{ required: true, message: '请输入旧密码', trigger: ['blur', 'input'] }],
  newPassword: [{ required: true, message: '请输入新密码', trigger: ['blur', 'input'] }],
  confirmPassword: [
    {
      required: true,
      trigger: ['blur', 'input'],
      validator: (_rule, value) => value === passwordForm.newPassword,
      message: '两次输入的密码不一致',
    },
  ],
}

// --- Actions ---
async function loadAccessInfo() {
  const data = await getAccessInfo()
  userStore.setAuth({
    token: token.value || '',
    user: {
      id: data.user.id,
      username: data.user.username,
      nickname: data.user.nickname,
      email: data.user.email,
      avatar: data.user.avatar,
      roles: data.roles,
      permissions: data.permissions,
      createdAt: data.user.createdAt,
      updatedAt: data.user.updatedAt,
      isAdmin: data.user.isAdmin,
    },
  })
  profileForm.nickname = data.user.nickname
  profileForm.email = data.user.email
  profileForm.avatar = data.user.avatar
}

async function handleProfileSubmit() {
  profileFormRef.value?.validate(async (errors) => {
    if (errors) return
    const updated = await updateProfile({
      nickname: profileForm.nickname,
      email: profileForm.email,
      avatar: profileForm.avatar,
    })
    userStore.setAuth({
      token: token.value || '',
      user: {
        ...user.value,
        nickname: updated.nickname,
        email: updated.email,
        avatar: updated.avatar,
        updatedAt: updated.updatedAt,
      } as any,
    })
    message.success('个人信息更新成功')
  })
}

async function handlePasswordSubmit() {
  passwordFormRef.value?.validate(async (errors) => {
    if (errors) return
    await changePassword({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword,
    })
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
    message.success('密码修改成功')
  })
}

async function loadOAuthBindings() {
  oauthLoading.value = true
  try {
    oauthBindings.value = await getOAuthBindings()
  } finally {
    oauthLoading.value = false
  }
}

function handleCopy(text: string) {
  navigator.clipboard.writeText(text)
  message.success('已复制到剪贴板')
}

// --- Avatar & Cropper Logic ---
const onBeforeUpload = async (options: { file: { file: File | null } }) => {
  const file = options.file.file
  if (!file) return false

  // Check file size (e.g., 2MB)
  if (file.size > 2 * 1024 * 1024) {
    message.error('图片大小不能超过 2MB')
    return false
  }

  const reader = new FileReader()
  reader.readAsDataURL(file)
  reader.onload = (e) => {
    cropperImg.value = e.target?.result as string
    showCropper.value = true
  }
  return false // Prevent auto upload
}

const handleConfirmCrop = async () => {
  if (!cropper) return
  const result = cropper.getFile()
  if (!result) return

  isUploading.value = true
  try {
    // vue-picture-cropper's getFile might return a Promise
    const file = await result
    if (!file) return
    const res = await uploadFile(file, 'picture')
    profileForm.avatar = res.publicUrl
    showCropper.value = false
    message.success('头像处理成功，请保存设置以生效')
  } catch (err: any) {
    message.error('上传失败: ' + err.message)
  } finally {
    isUploading.value = false
  }
}

const registrationDays = computed(() => {
  if (!user.value.createdAt) return 0
  return Math.floor((Date.now() - new Date(user.value.createdAt).getTime()) / (1000 * 60 * 60 * 24))
})

onMounted(() => {
  profileForm.nickname = user.value.nickname
  profileForm.email = user.value.email
  profileForm.avatar = user.value.avatar
  loadAccessInfo()
  loadOAuthBindings()
})
</script>

<template>
  <ScrollContainer wrapper-class="p-4 md:p-6">
    <NGrid
      x-gap="24"
      y-gap="24"
      cols="1 800:12"
    >
      <!-- Left: User Info Card -->
      <NGi span="1 800:4 1200:3">
        <div class="flex flex-col gap-4">
          <NCard :bordered="false">
            <div class="flex flex-col items-center py-4">
              <div class="relative mb-6">
                <UserAvatar
                  :size="100"
                  :src="user.avatar"
                />
                <div class="absolute -bottom-2 -right-2">
                  <NUpload
                    :show-file-list="false"
                    accept="image/*"
                    @before-upload="onBeforeUpload"
                  >
                    <NButton
                      circle
                      type="primary"
                      size="small"
                    >
                      <template #icon>
                        <span class="iconify ph--camera-bold" />
                      </template>
                    </NButton>
                  </NUpload>
                </div>
              </div>
              <div class="text-center">
                <div class="text-xl font-medium">
                  {{ user.nickname || '未设置昵称' }}
                </div>
                <div class="text-sm text-neutral-500">
                  @{{ user.username }}
                </div>
              </div>

              <div class="mt-4 flex flex-wrap justify-center gap-2">
                <NTag
                  v-if="user.id"
                  type="success"
                  size="small"
                  round
                >
                  已激活
                </NTag>
                <NTag
                  v-if="user.isAdmin"
                  type="primary"
                  size="small"
                  round
                >
                  管理员
                </NTag>
              </div>

              <NDivider />

              <div class="flex w-full justify-around">
                <NStatistic
                  label="注册天数"
                  tabular-nums
                >
                  {{ registrationDays }}
                </NStatistic>
              </div>
            </div>
          </NCard>

          <NCard
            title="基本信息"
            size="small"
            :bordered="false"
          >
            <div class="space-y-3 text-sm">
              <div class="flex justify-between">
                <span class="text-neutral-400">UID</span>
                <span
                  class="cursor-pointer font-mono hover:text-primary"
                  @click="handleCopy(String(user.id))"
                >
                  {{ user.id }}
                </span>
              </div>
              <div class="flex justify-between">
                <span class="text-neutral-400">注册日期</span>
                <span>{{ user.createdAt ? new Date(user.createdAt).toLocaleDateString() : '-' }}</span>
              </div>
            </div>
          </NCard>
        </div>
      </NGi>

      <!-- Right: Settings Area -->
      <NGi span="1 800:8 1200:9">
        <NCard
          :bordered="false"
          content-style="padding: 0;"
        >
          <NTabs
            type="line"
            size="large"
            class="ml-4"
            animated
            justify-content="start"
            pane-style="padding: 32px; min-height: 540px;"
          >
            <NTabPane
              name="profile"
              tab="个人资料"
            >
              <div class="max-w-2xl">
                <NForm
                  ref="profileFormRef"
                  :model="profileForm"
                  :rules="profileRules"
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
                          v-model:value="profileForm.nickname"
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
                          v-model:value="profileForm.email"
                          placeholder="请输入电子邮箱"
                        />
                      </NFormItem>
                    </NGi>
                  </NGrid>
                  <NFormItem label="头像地址 (URL)">
                    <NInput
                      v-model:value="profileForm.avatar"
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
                      @click="handleProfileSubmit"
                    >
                      保存基本信息
                    </NButton>
                  </div>
                </NForm>
              </div>
            </NTabPane>

            <NTabPane
              name="security"
              tab="安全设置"
            >
              <div class="mx-auto max-w-lg pt-4">
                <div class="mb-8">
                  <div class="text-lg font-medium">
                    修改账户密码
                  </div>
                  <div class="text-sm text-neutral-400">
                    为了您的账户安全，建议定期更换高强度密码
                  </div>
                </div>

                <NForm
                  ref="passwordFormRef"
                  :model="passwordForm"
                  :rules="passwordRules"
                  label-placement="top"
                >
                  <NFormItem
                    label="当前密码"
                    path="oldPassword"
                  >
                    <NInput
                      v-model:value="passwordForm.oldPassword"
                      type="password"
                      show-password-on="click"
                      placeholder="输入旧密码进行身份验证"
                    />
                  </NFormItem>
                  <NDivider />
                  <NFormItem
                    label="新密码"
                    path="newPassword"
                  >
                    <NInput
                      v-model:value="passwordForm.newPassword"
                      type="password"
                      show-password-on="click"
                      placeholder="设置您的新密码"
                    />
                  </NFormItem>
                  <NFormItem
                    label="确认新密码"
                    path="confirmPassword"
                  >
                    <NInput
                      v-model:value="passwordForm.confirmPassword"
                      type="password"
                      show-password-on="click"
                      placeholder="再次输入新密码"
                    />
                  </NFormItem>
                  <div class="mt-6">
                    <NButton
                      type="primary"
                      block
                      size="large"
                      @click="handlePasswordSubmit"
                    >
                      确认更改密码
                    </NButton>
                  </div>
                </NForm>
              </div>
            </NTabPane>

            <NTabPane
              name="binding"
              tab="账号绑定"
            >
              <div v-if="oauthLoading" class="py-12 text-center text-neutral-400">
                正在加载绑定信息...
              </div>
              <div
                v-else-if="oauthBindings.length === 0"
                class="flex flex-col items-center justify-center py-20"
              >
                <div class="mb-4 text-5xl text-neutral-150 dark:text-neutral-800">
                  <span class="iconify ph--link-break" />
                </div>
                <div class="text-neutral-500">
                  尚未绑定任何第三方账号
                </div>
              </div>
              <NGrid
                v-else
                cols="1 m:2"
                x-gap="16"
                y-gap="16"
              >
                <NGi
                  v-for="item in oauthBindings"
                  :key="item.providerKey + item.oauthID"
                >
                  <NCard
                    size="small"
                    hoverable
                  >
                    <div class="flex items-center gap-4 py-1">
                      <div class="grid h-10 w-10 place-items-center rounded bg-primary/10 text-xl font-bold text-primary">
                        {{ item.providerKey.charAt(0).toUpperCase() }}
                      </div>
                      <div class="flex-1 overflow-hidden">
                        <div class="flex items-center justify-between">
                          <span class="font-medium">{{ item.providerName || item.providerKey }}</span>
                          <NTag
                            type="success"
                            size="tiny"
                            round
                          >
                            已关联
                          </NTag>
                        </div>
                        <div class="truncate text-xs text-neutral-400">
                          ID: {{ item.oauthID }}
                        </div>
                      </div>
                    </div>
                  </NCard>
                </NGi>
              </NGrid>
            </NTabPane>
          </NTabs>
        </NCard>
      </NGi>
    </NGrid>

    <!-- Cropper Modal -->
    <NModal
      v-model:show="showCropper"
      preset="card"
      style="max-width: 600px"
      title="裁剪头像"
      :mask-closable="false"
      :closable="!isUploading"
    >
      <div class="h-80 w-full overflow-hidden rounded bg-neutral-100 dark:bg-neutral-900">
        <VuePictureCropper
          :box-style="{
            width: '100%',
            height: '100%',
            backgroundColor: '#f8f8f8',
            margin: 'auto',
          }"
          :img="cropperImg"
          :options="{
            viewMode: 1,
            dragMode: 'move',
            aspectRatio: 1,
            cropBoxResizable: false,
          }"
        />
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <NButton
            :disabled="isUploading"
            @click="showCropper = false"
          >
            取消
          </NButton>
          <NButton
            type="primary"
            :loading="isUploading"
            @click="handleConfirmCrop"
          >
            确认并上传
          </NButton>
        </div>
      </template>
    </NModal>
  </ScrollContainer>
</template>
