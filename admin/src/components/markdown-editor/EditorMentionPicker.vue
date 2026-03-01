<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  NModal,
  NCard,
  NInput,
  NInputGroup,
  NSpin,
  NButton,
  NTag,
  NScrollbar,
  NCollapse,
  NCollapseItem,
  useThemeVars,
} from 'naive-ui'
import type { FederationAuthorResp } from '@/types/federation'

const props = defineProps<{
  show: boolean
  query: string
  results: FederationAuthorResp[]
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'search', query: string): void
  (e: 'select', author: FederationAuthorResp): void
  (e: 'insertRaw', user: string, instance: string): void
}>()

const themeVars = useThemeVars()

const badgeBg = computed(() => `color-mix(in srgb, ${themeVars.value.primaryColor} 12%, transparent)`)
const avatarBg = computed(() => `color-mix(in srgb, ${themeVars.value.primaryColor} 14%, transparent)`)
const avatarColor = computed(() => themeVars.value.primaryColor)
const hoverBg = computed(() => themeVars.value.hoverColor)
const borderRadius = computed(() => themeVars.value.borderRadius)
const textColor1 = computed(() => themeVars.value.textColor1)
const textColor3 = computed(() => themeVars.value.textColor3)
const codeBg = computed(() => themeVars.value.codeColor)

const manualUser = ref('')
const manualInstance = ref('')

function handleManualInsert() {
  if (manualUser.value.trim() && manualInstance.value.trim()) {
    emit('insertRaw', manualUser.value.trim(), manualInstance.value.trim())
    manualUser.value = ''
    manualInstance.value = ''
  }
}

function getInitial(name: string): string {
  return (name || '?').charAt(0).toUpperCase()
}

function extractHost(url: string): string {
  try {
    return new URL(url).hostname
  } catch {
    return url.replace(/^https?:\/\//, '').replace(/\/.*$/, '')
  }
}

const hasSearched = computed(() => props.query.trim().length > 0)
</script>

<template>
  <NModal
    :show="show"
    style="width: 480px; max-width: 90vw"
    @update:show="emit('update:show', $event)"
  >
    <NCard
      size="small"
      :closable="true"
      :bordered="true"
      @close="emit('update:show', false)"
    >
      <template #header>
        <div class="flex items-center gap-2.5">
          <div
            class="grid size-8 shrink-0 place-items-center rounded-lg"
            :style="{ background: badgeBg }"
          >
            <span
              class="iconify ph--at-bold text-base"
              :style="{ color: avatarColor }"
            />
          </div>
          <span class="text-base font-medium" :style="{ color: textColor1 }">
            选择提及对象
          </span>
        </div>
      </template>

      <!-- Search -->
      <div class="flex flex-col gap-3">
        <div>
          <NInput
            :value="query"
            placeholder="输入作者名称搜索..."
            clearable
            @update:value="emit('search', $event)"
          >
            <template #prefix>
              <span class="iconify ph--magnifying-glass text-base" :style="{ color: textColor3 }" />
            </template>
          </NInput>
          <p class="mt-1.5 text-xs" :style="{ color: textColor3 }">
            搜索已联合实例中的作者，点击即可插入提及语法
          </p>
        </div>

        <!-- Results list -->
        <NSpin :show="loading">
          <NScrollbar style="max-height: 260px">
            <!-- Empty: not searched yet -->
            <div
              v-if="!loading && results.length === 0 && !hasSearched"
              class="flex flex-col items-center justify-center gap-2 py-10"
            >
              <span
                class="iconify ph--users-three text-3xl"
                :style="{ color: textColor3 }"
              />
              <span class="text-sm" :style="{ color: textColor3 }">
                输入关键词开始搜索作者
              </span>
            </div>

            <!-- Empty: no results -->
            <div
              v-else-if="!loading && results.length === 0 && hasSearched"
              class="flex flex-col items-center justify-center gap-2 py-10"
            >
              <span
                class="iconify ph--magnifying-glass text-3xl"
                :style="{ color: textColor3 }"
              />
              <span class="text-sm" :style="{ color: textColor3 }">
                未找到匹配的作者，可在下方手动输入
              </span>
            </div>

            <!-- Author list -->
            <div v-else class="flex flex-col gap-0.5">
              <div
                v-for="author in results"
                :key="author.name + '@' + author.instanceUrl"
                class="group flex cursor-pointer items-center gap-3 px-2.5 py-2 transition-colors duration-150"
                :style="{ borderRadius }"
                @click="emit('select', author)"
              >
                <!-- Avatar -->
                <div
                  class="grid size-9 shrink-0 place-items-center rounded-full text-sm font-semibold"
                  :style="{ background: avatarBg, color: avatarColor }"
                >
                  {{ getInitial(author.name) }}
                </div>
                <!-- Info -->
                <div class="min-w-0 flex-1">
                  <div class="truncate text-sm font-medium" :style="{ color: textColor1 }">
                    {{ author.name }}
                  </div>
                  <div class="truncate text-xs" :style="{ color: textColor3 }">
                    {{ author.instanceName || extractHost(author.instanceUrl) }}
                  </div>
                </div>
                <!-- Syntax preview on hover -->
                <NTag
                  size="small"
                  :bordered="false"
                  round
                  class="shrink-0 opacity-0 transition-opacity duration-150 group-hover:opacity-100"
                  :style="{ fontFamily: '\'Fira Code\', \'SFMono-Regular\', monospace', background: codeBg }"
                >
                  @{{ author.name }}
                </NTag>
              </div>
            </div>
          </NScrollbar>
        </NSpin>

        <!-- Manual input -->
        <NCollapse arrow-placement="right" class="mt-1">
          <NCollapseItem title="手动输入" name="manual">
            <template #header-extra>
              <span class="iconify ph--keyboard text-base" :style="{ color: textColor3 }" />
            </template>
            <div class="flex flex-col gap-3 pt-1">
              <NInputGroup>
                <NInput
                  v-model:value="manualUser"
                  placeholder="用户名，如 grtsinry43"
                  size="small"
                >
                  <template #prefix>
                    <span class="iconify ph--user text-sm" :style="{ color: textColor3 }" />
                  </template>
                </NInput>
                <NInput
                  v-model:value="manualInstance"
                  placeholder="实例域名，如 blog.example.com"
                  size="small"
                >
                  <template #prefix>
                    <span class="iconify ph--globe text-sm" :style="{ color: textColor3 }" />
                  </template>
                </NInput>
              </NInputGroup>
              <div class="flex items-center justify-between">
                <code
                  class="rounded px-2 py-1 text-xs"
                  :style="{
                    background: codeBg,
                    fontFamily: '\'Fira Code\', \'SFMono-Regular\', monospace',
                    color: textColor1,
                  }"
                >
                  &lt;@{{ manualUser || 'user' }}@{{ manualInstance || 'instance' }}&gt;
                </code>
                <NButton
                  size="small"
                  type="primary"
                  :disabled="!manualUser.trim() || !manualInstance.trim()"
                  @click="handleManualInsert"
                >
                  <template #icon>
                    <span class="iconify ph--arrow-right" />
                  </template>
                  插入
                </NButton>
              </div>
            </div>
          </NCollapseItem>
        </NCollapse>
      </div>
    </NCard>
  </NModal>
</template>

<style scoped>
.group:hover {
  background: v-bind(hoverBg);
}
</style>
