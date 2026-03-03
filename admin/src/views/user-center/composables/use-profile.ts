import { computed, onMounted, reactive, ref } from 'vue'
import {
  changePassword,
  getAccessInfo,
  getOAuthBindings,
  updateProfile,
} from '@/services/auth'
import { listOAuthProviders } from '@/services/oauth-providers'
import { uploadFile } from '@/services/uploads'
import { toRefsUserStore, useUserStore } from '@/stores'
import { cropper } from 'vue-picture-cropper'

import type { OAuthBinding } from '@/services/auth'
import type { AdminOAuthProvider } from '@/services/oauth-providers'
import type { FormInst, FormItemRule } from 'naive-ui'

export function useProfile(message: { error: (m: string) => void; success: (m: string) => void; warning?: (m: string) => void }) {
  const userStore = useUserStore()
  const { user, token } = toRefsUserStore()

  const profileFormRef = ref<FormInst | null>(null)
  const passwordFormRef = ref<FormInst | null>(null)
  const oauthLoading = ref(false)
  const oauthBindings = ref<OAuthBinding[]>([])
  const oauthProviders = ref<AdminOAuthProvider[]>([])

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

  const showCropper = ref(false)
  const cropperImg = ref('')
  const isUploading = ref(false)

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

  const registrationDays = computed(() => {
    if (!user.value.createdAt) return 0
    return Math.floor((Date.now() - new Date(user.value.createdAt).getTime()) / (1000 * 60 * 60 * 24))
  })

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
      const [bindings, providers] = await Promise.all([
        getOAuthBindings(),
        listOAuthProviders(),
      ])
      oauthBindings.value = bindings
      oauthProviders.value = providers.filter(p => p.enabled)
    } finally {
      oauthLoading.value = false
    }
  }

  function handleCopy(text: string) {
    navigator.clipboard.writeText(text)
    message.success('已复制到剪贴板')
  }

  const onBeforeUpload = async (options: { file: { file: File | null } }) => {
    const file = options.file.file
    if (!file) return false
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
    return false
  }

  const handleConfirmCrop = async () => {
    if (!cropper) return
    const result = cropper.getFile()
    if (!result) return
    isUploading.value = true
    try {
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

  onMounted(() => {
    profileForm.nickname = user.value.nickname
    profileForm.email = user.value.email
    profileForm.avatar = user.value.avatar
    loadAccessInfo()
    loadOAuthBindings()
  })

  return {
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
  }
}
