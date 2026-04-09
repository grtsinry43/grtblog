<script setup lang="ts">
import {
  NButton,
  NConfigProvider,
  NH1,
  NH2,
  NSpin,
  NStep,
  NSteps,
  NTag,
  useMessage,
  type GlobalThemeOverrides,
} from 'naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'

import noiseBg from '@/assets/noise.png'
import { getConfigProviderProps } from '@/composables'
import ThemeModePopover from '@/layout/header/action/ThemeModePopover.vue'
import router from '@/router'
import { getSetupState } from '@/services/auth'
import { ApiError } from '@/services/http'
import { completeUpgradeGuide } from '@/services/system'
import { listWebsiteInfo } from '@/services/website-info'
import { usePreferencesStore } from '@/stores'
import ThemeColorPopover from '@/views/sign-in/components/ThemeColorPopover.vue'

import { applyEnabledFeatures } from './apply-features'
import FeatureToggleList from './FeatureToggleList.vue'
import { getPendingGuides } from './registry'

import type { UpgradeGuideVersion } from './registry'

defineOptions({
  name: 'UpgradeGuidePage',
})

const message = useMessage()
const preferencesStore = usePreferencesStore()
const configProviderProps = getConfigProviderProps()

function hexToRgb(hex: string) {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex)
  if (!result || !result[1] || !result[2] || !result[3]) return '0, 0, 0'
  return `${parseInt(result[1], 16)}, ${parseInt(result[2], 16)}, ${parseInt(result[3], 16)}`
}

const primaryColorRgb = computed(() => hexToRgb(preferencesStore.themeColor))

const themeOverrides: GlobalThemeOverrides = {
  common: { fontWeightStrong: '600' },
  Button: { heightMedium: '34px', fontSizeMedium: '13px', fontWeight: '500' },
  Steps: { indicatorSizeSmall: '20px', headerFontSizeSmall: '13px' },
}

const loading = ref(true)
const submitting = ref(false)
const sitePublicUrl = ref('')

// Resolved from the API + registry
const guides = ref<UpgradeGuideVersion[]>([])
const currentStepIndex = ref(0)
const currentGuide = computed(() => guides.value[currentStepIndex.value])

// Feature toggle states per guide step, keyed by feature id
const featureStates = reactive<Record<string, boolean>>({})

const hasAnyEnabled = computed(
  () => currentGuide.value?.features.some((f) => featureStates[f.id]) ?? false,
)

// ── Lifecycle ────────────────────────────────────────────────────────────────

async function checkState() {
  loading.value = true
  try {
    const state = await getSetupState()
    const pending = state.pendingUpgradeGuides ?? []
    if (pending.length === 0) {
      await router.replace({ path: '/' })
      return
    }
    guides.value = getPendingGuides(pending)
    if (guides.value.length === 0) {
      // Backend knows versions that the frontend registry doesn't — skip.
      await router.replace({ path: '/' })
      return
    }

    // Fetch public_url for auto-filling instanceURL
    try {
      const items = await listWebsiteInfo()
      const publicUrl = items.find((i) => i.key === 'public_url')
      if (publicUrl?.value) {
        sitePublicUrl.value = publicUrl.value.replace(/\/+$/, '')
      }
    } catch {
      // Non-critical
    }
  } catch (error) {
    if (!(error instanceof ApiError)) {
      message.error('获取状态失败')
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  checkState()
})

// ── Actions ──────────────────────────────────────────────────────────────────

function handleAuthError(error: unknown): boolean {
  if (error instanceof ApiError && error.status === 401) {
    router.replace({ name: 'signIn' })
    return true
  }
  return false
}

async function finishCurrentStep(skip: boolean) {
  const guide = currentGuide.value
  if (!guide) return

  submitting.value = true
  try {
    // Apply enabled feature configs unless skipping
    if (!skip) {
      try {
        await applyEnabledFeatures([guide], featureStates, sitePublicUrl.value)
      } catch {
        message.warning('部分功能配置失败，可在设置中手动配置')
      }
    }

    // Mark this guide version as completed
    await completeUpgradeGuide(guide.version)

    // Move to next step or finish
    if (currentStepIndex.value < guides.value.length - 1) {
      // Clear toggle states so the next step starts fresh
      for (const key of Object.keys(featureStates)) {
        delete featureStates[key]
      }
      currentStepIndex.value++
    } else {
      message.success('升级引导已完成')
      await router.replace({ path: '/' })
    }
  } catch (error) {
    if (handleAuthError(error)) return
    if (!(error instanceof ApiError)) {
      message.error('操作失败，请稍后重试')
    }
  } finally {
    submitting.value = false
  }
}

// ── Left panel helpers ───────────────────────────────────────────────────────

const leftPanelFeatures = computed(() => {
  const guide = currentGuide.value
  if (!guide) return []
  return guide.features.map((f, i) => ({
    index: i + 1,
    label: `${f.label} — ${f.description.length > 30 ? f.description.slice(0, 30) + '...' : f.description}`,
  }))
})

const latestVersion = computed(() => {
  if (guides.value.length === 0) return ''
  return guides.value[guides.value.length - 1]!.version
})
</script>

<template>
  <NConfigProvider
    v-bind="configProviderProps"
    :theme-overrides="themeOverrides"
  >
    <div
      class="relative flex min-h-screen w-screen bg-neutral-50 font-sans text-neutral-900 transition-colors dark:bg-neutral-950 dark:text-neutral-100"
      :style="{ '--primary-color-rgb': primaryColorRgb }"
    >
      <!-- Theme controls -->
      <div class="absolute top-0 right-0 z-100 flex items-center gap-4 p-8">
        <ThemeColorPopover />
        <ThemeModePopover />
      </div>

      <!-- Loading -->
      <NSpin
        v-if="loading"
        :show="loading"
        class="m-auto"
        size="large"
      >
        <template #description>正在加载...</template>
      </NSpin>

      <template v-else-if="guides.length > 0 && currentGuide">
        <div class="flex h-screen w-full overflow-hidden">
          <!-- Left: Brand -->
          <div
            class="brand-panel relative hidden flex-[0_0_45%] flex-col justify-center overflow-hidden px-20 lg:flex"
            :style="{
              background: `linear-gradient(135deg, rgba(var(--primary-color-rgb), 0.05) 0%, rgba(var(--primary-color-rgb), 0.02) 100%)`,
            }"
          >
            <div
              class="absolute inset-0 z-0 opacity-[0.03] mix-blend-multiply dark:mix-blend-overlay"
              :style="{ backgroundImage: `url(${noiseBg})` }"
            ></div>
            <div
              class="absolute -top-[10%] -left-[10%] z-0 h-[600px] w-[600px] rounded-full bg-white opacity-40 blur-3xl dark:opacity-5"
            ></div>

            <div class="relative z-10 max-w-lg">
              <div class="mb-10 flex items-center gap-3 opacity-60">
                <div
                  class="h-1 w-10 rounded-full"
                  :style="{ background: `rgb(var(--primary-color-rgb))` }"
                ></div>
                <span
                  class="text-[10px] font-bold tracking-[0.2em] text-neutral-500 uppercase dark:text-neutral-400"
                  >What's new</span
                >
              </div>

              <NH1
                class="mb-6 text-4xl leading-tight font-bold tracking-tight text-neutral-900 dark:text-white"
              >
                欢迎来到
                <br />
                <span :style="{ color: `rgb(var(--primary-color-rgb))` }"
                  >V{{ latestVersion }}</span
                >
              </NH1>

              <div
                class="text-base leading-relaxed font-light text-neutral-500 dark:text-neutral-400"
              >
                <p class="mb-3">此次更新带来了一些新功能。</p>
                <p>以下向导将帮助您快速了解和配置这些新特性。</p>
              </div>

              <!-- Feature bullet list from current guide -->
              <div class="mt-16 space-y-3 text-sm text-neutral-400 dark:text-neutral-500">
                <div
                  v-for="item in leftPanelFeatures"
                  :key="item.index"
                  class="flex items-center gap-3"
                >
                  <span
                    class="inline-flex h-5 w-5 items-center justify-center rounded-full text-xs font-semibold"
                    :style="{
                      background: `rgba(var(--primary-color-rgb), 0.1)`,
                      color: `rgb(var(--primary-color-rgb))`,
                    }"
                    >{{ item.index }}</span
                  >
                  <span>{{ item.label }}</span>
                </div>
              </div>

              <div
                class="mt-20 flex items-center gap-4 text-[10px] font-medium tracking-widest text-neutral-400 uppercase"
              >
                <span>GRTBLOG V{{ latestVersion }}</span>
                <span class="h-0.5 w-0.5 rounded-full bg-neutral-300"></span>
                <span>UPGRADE GUIDE</span>
              </div>
            </div>
          </div>

          <!-- Right: Form -->
          <div
            class="flex flex-1 overflow-y-auto bg-white p-8 transition-colors sm:p-12 dark:bg-neutral-900"
          >
            <div class="mx-auto flex min-h-full w-full max-w-[420px] flex-col justify-center py-4">
              <!-- Step indicator (multi-guide) -->
              <div
                v-if="guides.length > 1"
                class="mb-6 flex items-center justify-between"
              >
                <div
                  class="text-[10px] font-bold tracking-widest whitespace-nowrap text-neutral-400 uppercase"
                >
                  Step {{ currentStepIndex + 1 }} / {{ guides.length }}
                </div>
                <NSteps
                  :current="currentStepIndex + 1"
                  size="small"
                  class="ml-4"
                  :style="{ width: `${guides.length * 48}px` }"
                >
                  <NStep
                    v-for="g in guides"
                    :key="g.version"
                  />
                </NSteps>
              </div>

              <!-- Header -->
              <div class="mb-8">
                <NTag
                  size="small"
                  type="success"
                  round
                  :bordered="false"
                >
                  {{ currentGuide.tag }}
                </NTag>
                <NH2 class="mt-3 mb-0 text-2xl font-bold tracking-tight">
                  {{ currentGuide.title }}
                </NH2>
                <p class="mt-2 text-[13px] leading-relaxed text-neutral-500">
                  {{ currentGuide.description }}
                </p>
              </div>

              <!-- Dynamic feature toggles -->
              <FeatureToggleList
                :guides="[currentGuide]"
                :primary-color-rgb="primaryColorRgb"
                v-model:states="featureStates"
              />

              <!-- Actions -->
              <div
                class="mt-8 flex items-center justify-between border-t border-neutral-100 pt-6 dark:border-neutral-800"
              >
                <NButton
                  quaternary
                  size="medium"
                  :disabled="submitting"
                  @click="finishCurrentStep(true)"
                >
                  暂时跳过
                </NButton>

                <NButton
                  type="primary"
                  size="medium"
                  :loading="submitting"
                  @click="finishCurrentStep(false)"
                  class="min-w-25 shadow-sm"
                >
                  {{ hasAnyEnabled ? '启用并继续' : '继续' }}
                </NButton>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </NConfigProvider>
</template>
