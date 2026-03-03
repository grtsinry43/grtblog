<script setup lang="ts">
import {
  NButton,
  NCard,
  NDivider,
  NGi,
  NGrid,
  NStatistic,
  NTabPane,
  NTabs,
  NTag,
  NUpload,
  useMessage,
} from 'naive-ui'

import { ScrollContainer, UserAvatar } from '@/components'
import { useProfile } from './composables/use-profile'
import ProfileForm from './components/ProfileForm.vue'
import PasswordForm from './components/PasswordForm.vue'
import OAuthBindings from './components/OAuthBindings.vue'
import AvatarCropper from './components/AvatarCropper.vue'

defineOptions({ name: 'UserCenter' })

const message = useMessage()

const {
  user,
  profileFormRef,
  passwordFormRef,
  oauthLoading,
  oauthBindings,
  oauthProviders,
  profileForm,
  passwordForm,
  showCropper,
  cropperImg,
  isUploading,
  profileRules,
  passwordRules,
  registrationDays,
  handleProfileSubmit,
  handlePasswordSubmit,
  handleCopy,
  onBeforeUpload,
  handleConfirmCrop,
} = useProfile(message)
</script>

<template>
  <ScrollContainer wrapper-class="p-4 md:p-6">
    <NGrid x-gap="24" y-gap="24" cols="1 800:12">
      <!-- Left: User Info Card -->
      <NGi span="1 800:4 1200:3">
        <div class="flex flex-col gap-4">
          <NCard :bordered="false">
            <div class="flex flex-col items-center py-4">
              <div class="relative mb-6">
                <UserAvatar :size="100" :src="user.avatar" />
                <div class="absolute -bottom-2 -right-2">
                  <NUpload :show-file-list="false" accept="image/*" @before-upload="onBeforeUpload">
                    <NButton circle type="primary" size="small">
                      <template #icon><span class="iconify ph--camera-bold" /></template>
                    </NButton>
                  </NUpload>
                </div>
              </div>
              <div class="text-center">
                <div class="text-xl font-medium">{{ user.nickname || '未设置昵称' }}</div>
                <div class="text-sm text-neutral-500">@{{ user.username }}</div>
              </div>
              <div class="mt-4 flex flex-wrap justify-center gap-2">
                <NTag v-if="user.id" type="success" size="small" round>已激活</NTag>
                <NTag v-if="user.isAdmin" type="primary" size="small" round>管理员</NTag>
              </div>
              <NDivider />
              <div class="flex w-full justify-around">
                <NStatistic label="注册天数" tabular-nums>{{ registrationDays }}</NStatistic>
              </div>
            </div>
          </NCard>

          <NCard title="基本信息" size="small" :bordered="false">
            <div class="space-y-3 text-sm">
              <div class="flex justify-between">
                <span class="text-neutral-400">UID</span>
                <span class="cursor-pointer font-mono hover:text-primary" @click="handleCopy(String(user.id))">
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
        <NCard :bordered="false" content-style="padding: 0;">
          <NTabs
            type="line"
            size="large"
            class="ml-4"
            animated
            justify-content="start"
            pane-style="padding: 32px; min-height: 540px;"
          >
            <NTabPane name="profile" tab="个人资料">
              <ProfileForm
                :form-ref="profileFormRef"
                :form="profileForm"
                :rules="profileRules"
                @update:form-ref="profileFormRef = $event"
                @submit="handleProfileSubmit"
              />
            </NTabPane>

            <NTabPane name="security" tab="安全设置">
              <PasswordForm
                :form-ref="passwordFormRef"
                :form="passwordForm"
                :rules="passwordRules"
                @update:form-ref="passwordFormRef = $event"
                @submit="handlePasswordSubmit"
              />
            </NTabPane>

            <NTabPane name="binding" tab="账号绑定">
              <OAuthBindings :loading="oauthLoading" :bindings="oauthBindings" :providers="oauthProviders" />
            </NTabPane>
          </NTabs>
        </NCard>
      </NGi>
    </NGrid>

    <AvatarCropper
      :visible="showCropper"
      :cropper-img="cropperImg"
      :is-uploading="isUploading"
      @update:visible="showCropper = $event"
      @confirm="handleConfirmCrop"
    />
  </ScrollContainer>
</template>
