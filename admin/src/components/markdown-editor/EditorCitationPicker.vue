<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  NModal,
  NCard,
  NInput,
  NSpin,
  NButton,
  NTag,
  NScrollbar,
  NCollapse,
  NCollapseItem,
  NTooltip,
  useThemeVars,
} from 'naive-ui'
import type { FederationInstanceResp, FederationCachedPostResp } from '@/types/federation'

const props = defineProps<{
  show: boolean
  step: 'instance' | 'post'
  instances: FederationInstanceResp[]
  instanceFilter: string
  selectedInstance: FederationInstanceResp | null
  posts: FederationCachedPostResp[]
  searchQuery: string
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'selectInstance', inst: FederationInstanceResp): void
  (e: 'searchPosts', query: string): void
  (e: 'filterInstances', query: string): void
  (e: 'select', post: FederationCachedPostResp): void
  (e: 'back'): void
  (e: 'insertRaw', instance: string, postId: string): void
}>()

const themeVars = useThemeVars()

const infoBadgeBg = computed(() => `color-mix(in srgb, ${themeVars.value.infoColor} 12%, transparent)`)
const infoIconColor = computed(() => themeVars.value.infoColor)
const hoverBg = computed(() => themeVars.value.hoverColor)
const borderRadius = computed(() => themeVars.value.borderRadius)
const borderColor = computed(() => themeVars.value.borderColor)
const textColor1 = computed(() => themeVars.value.textColor1)
const textColor3 = computed(() => themeVars.value.textColor3)
const codeBg = computed(() => themeVars.value.codeColor)
const contextBarBg = computed(() => `color-mix(in srgb, ${themeVars.value.infoColor} 6%, ${themeVars.value.cardColor})`)

const filteredInstances = computed(() => {
  const kw = props.instanceFilter.toLowerCase()
  if (!kw) return props.instances
  return props.instances.filter(
    (i) =>
      (i.name ?? '').toLowerCase().includes(kw) ||
      i.base_url.toLowerCase().includes(kw),
  )
})

const manualInstance = ref('')
const manualPostId = ref('')

function handleManualInsert() {
  if (manualInstance.value.trim() && manualPostId.value.trim()) {
    emit('insertRaw', manualInstance.value.trim(), manualPostId.value.trim())
    manualInstance.value = ''
    manualPostId.value = ''
  }
}

function formatDate(iso: string) {
  try {
    return new Date(iso).toLocaleDateString()
  } catch {
    return iso
  }
}

function extractHost(url: string): string {
  try {
    return new URL(url).hostname
  } catch {
    return url.replace(/^https?:\/\//, '').replace(/\/.*$/, '')
  }
}

function getStatusType(status: string): 'success' | 'error' | 'default' {
  if (status === 'active') return 'success'
  if (status === 'banned') return 'error'
  return 'default'
}

function getStatusLabel(status: string): string {
  if (status === 'active') return '在线'
  if (status === 'banned') return '已封禁'
  return status
}
</script>

<template>
  <NModal
    :show="show"
    style="width: 580px; max-width: 90vw"
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
          <!-- Quote badge -->
          <div
            class="grid size-8 shrink-0 place-items-center rounded-lg"
            :style="{ background: infoBadgeBg }"
          >
            <span
              class="iconify ph--quotes-bold text-base"
              :style="{ color: infoIconColor }"
            />
          </div>
          <!-- Breadcrumb -->
          <div class="flex items-center gap-1.5 text-sm">
            <span
              :class="step === 'instance'
                ? 'font-medium'
                : 'cursor-pointer transition-colors duration-150'"
              :style="{
                color: step === 'instance' ? textColor1 : textColor3,
              }"
              @click="step === 'post' ? emit('back') : undefined"
              @mouseenter="($event.target as HTMLElement).style.color = step === 'post' ? themeVars.infoColor : ''"
              @mouseleave="($event.target as HTMLElement).style.color = step === 'post' ? textColor3 : textColor1"
            >
              选择实例
            </span>
            <template v-if="step === 'post'">
              <span :style="{ color: textColor3 }">›</span>
              <span
                class="max-w-[200px] truncate font-medium"
                :style="{ color: textColor1 }"
              >
                {{ selectedInstance?.name || extractHost(selectedInstance?.base_url ?? '') }}
              </span>
            </template>
          </div>
        </div>
      </template>

      <!-- ========== Step 1: Instance selection ========== -->
      <template v-if="step === 'instance'">
        <div class="flex flex-col gap-3">
          <!-- Search -->
          <NInput
            :value="instanceFilter"
            placeholder="搜索实例名称或域名..."
            clearable
            @update:value="emit('filterInstances', $event)"
          >
            <template #prefix>
              <span class="iconify ph--magnifying-glass text-base" :style="{ color: textColor3 }" />
            </template>
          </NInput>

          <!-- Instance list -->
          <NSpin :show="loading">
            <NScrollbar style="max-height: 340px">
              <!-- Empty -->
              <div
                v-if="!loading && filteredInstances.length === 0"
                class="flex flex-col items-center justify-center gap-2 py-12"
              >
                <span
                  class="iconify ph--plugs-connected text-3xl"
                  :style="{ color: textColor3 }"
                />
                <span class="text-sm" :style="{ color: textColor3 }">
                  暂无已联合的实例
                </span>
              </div>

              <!-- List -->
              <div v-else class="flex flex-col gap-0.5">
                <div
                  v-for="inst in filteredInstances"
                  :key="inst.id"
                  class="group flex cursor-pointer items-center gap-3 px-2.5 py-2.5 transition-colors duration-150"
                  :style="{ borderRadius }"
                  @click="emit('selectInstance', inst)"
                >
                  <!-- Globe icon box -->
                  <div
                    class="grid size-9 shrink-0 place-items-center rounded-lg"
                    :style="{ background: infoBadgeBg }"
                  >
                    <span
                      class="iconify ph--globe-simple text-lg"
                      :style="{ color: infoIconColor }"
                    />
                  </div>
                  <!-- Info -->
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center gap-2">
                      <span class="truncate text-sm font-medium" :style="{ color: textColor1 }">
                        {{ inst.name || extractHost(inst.base_url) }}
                      </span>
                      <NTag
                        size="tiny"
                        round
                        :type="getStatusType(inst.status)"
                      >
                        {{ getStatusLabel(inst.status) }}
                      </NTag>
                    </div>
                    <div class="truncate text-xs" :style="{ color: textColor3 }">
                      {{ extractHost(inst.base_url) }}
                    </div>
                    <div
                      v-if="inst.description"
                      class="mt-0.5 truncate text-xs"
                      :style="{ color: textColor3 }"
                    >
                      {{ inst.description }}
                    </div>
                  </div>
                  <!-- Arrow -->
                  <span
                    class="iconify ph--caret-right shrink-0 text-base opacity-40 transition-opacity duration-150 group-hover:opacity-80"
                    :style="{ color: textColor3 }"
                  />
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
                <NInput
                  v-model:value="manualInstance"
                  placeholder="实例域名，如 blog.example.com"
                  size="small"
                >
                  <template #prefix>
                    <span class="iconify ph--globe text-sm" :style="{ color: textColor3 }" />
                  </template>
                </NInput>
                <NInput
                  v-model:value="manualPostId"
                  placeholder="文章 ID，如 my-post-slug"
                  size="small"
                >
                  <template #prefix>
                    <span class="iconify ph--article text-sm" :style="{ color: textColor3 }" />
                  </template>
                </NInput>
                <div class="flex items-center justify-between">
                  <code
                    class="rounded px-2 py-1 text-xs"
                    :style="{
                      background: codeBg,
                      fontFamily: '\'Fira Code\', \'SFMono-Regular\', monospace',
                      color: textColor1,
                    }"
                  >
                    &lt;cite:{{ manualInstance || 'instance' }}|{{ manualPostId || 'post-id' }}&gt;
                  </code>
                  <NButton
                    size="small"
                    type="primary"
                    :disabled="!manualInstance.trim() || !manualPostId.trim()"
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
      </template>

      <!-- ========== Step 2: Post selection ========== -->
      <template v-else>
        <div class="flex flex-col gap-3">
          <!-- Instance context bar -->
          <div
            class="flex items-center gap-2.5 rounded-lg px-3 py-2"
            :style="{
              background: contextBarBg,
              border: `1px solid ${borderColor}`,
            }"
          >
            <div
              class="grid size-7 shrink-0 place-items-center rounded"
              :style="{ background: infoBadgeBg }"
            >
              <span
                class="iconify ph--globe-simple text-sm"
                :style="{ color: infoIconColor }"
              />
            </div>
            <div class="min-w-0 flex-1">
              <span class="text-sm font-medium" :style="{ color: textColor1 }">
                {{ selectedInstance?.name || extractHost(selectedInstance?.base_url ?? '') }}
              </span>
              <span class="ml-1.5 text-xs" :style="{ color: textColor3 }">
                {{ extractHost(selectedInstance?.base_url ?? '') }}
              </span>
            </div>
            <NButton
              size="tiny"
              quaternary
              @click="emit('back')"
            >
              <template #icon>
                <span class="iconify ph--arrow-left" />
              </template>
              切换
            </NButton>
          </div>

          <!-- Search -->
          <NInput
            :value="searchQuery"
            placeholder="搜索文章标题..."
            clearable
            @update:value="emit('searchPosts', $event)"
          >
            <template #prefix>
              <span class="iconify ph--magnifying-glass text-base" :style="{ color: textColor3 }" />
            </template>
          </NInput>

          <!-- Post list -->
          <NSpin :show="loading">
            <NScrollbar style="max-height: 360px">
              <!-- Empty: search yielded nothing -->
              <div
                v-if="!loading && posts.length === 0 && searchQuery.trim()"
                class="flex flex-col items-center justify-center gap-2 py-12"
              >
                <span
                  class="iconify ph--magnifying-glass text-3xl"
                  :style="{ color: textColor3 }"
                />
                <span class="text-sm" :style="{ color: textColor3 }">
                  未找到匹配的文章
                </span>
              </div>

              <!-- Empty: no cached posts at all -->
              <div
                v-else-if="!loading && posts.length === 0 && !searchQuery.trim()"
                class="flex flex-col items-center justify-center gap-2 py-12"
              >
                <span
                  class="iconify ph--article text-3xl"
                  :style="{ color: textColor3 }"
                />
                <span class="text-sm" :style="{ color: textColor3 }">
                  该实例暂无缓存文章
                </span>
              </div>

              <!-- Post list -->
              <div v-else class="flex flex-col gap-0.5">
                <div
                  v-for="post in posts"
                  :key="post.id"
                  class="group flex cursor-pointer items-start gap-3 px-2.5 py-2.5 transition-colors duration-150"
                  :style="{ borderRadius }"
                  @click="emit('select', post)"
                >
                  <!-- Cover image or placeholder -->
                  <div
                    v-if="post.coverImage"
                    class="shrink-0 overflow-hidden rounded-md"
                    style="width: 72px; height: 54px"
                  >
                    <img
                      :src="post.coverImage"
                      alt=""
                      class="size-full object-cover"
                    />
                  </div>
                  <div
                    v-else
                    class="grid shrink-0 place-items-center rounded-md"
                    style="width: 72px; height: 54px"
                    :style="{ background: infoBadgeBg }"
                  >
                    <span
                      class="iconify ph--article text-xl"
                      :style="{ color: infoIconColor }"
                    />
                  </div>

                  <!-- Content -->
                  <div class="min-w-0 flex-1">
                    <!-- Title row -->
                    <div class="flex items-center gap-1.5">
                      <span
                        class="flex-1 truncate text-sm font-medium"
                        :style="{ color: textColor1 }"
                      >
                        {{ post.title }}
                      </span>
                      <NTooltip v-if="!post.allowCitation" trigger="hover">
                        <template #trigger>
                          <NTag size="tiny" type="warning" round>
                            <template #icon>
                              <span class="iconify ph--warning text-xs" />
                            </template>
                            不可引用
                          </NTag>
                        </template>
                        该文章作者未允许被引用
                      </NTooltip>
                    </div>
                    <!-- Meta row -->
                    <div class="mt-0.5 flex items-center gap-3 text-xs" :style="{ color: textColor3 }">
                      <span v-if="post.authorName" class="flex items-center gap-1">
                        <span class="iconify ph--user text-xs" />
                        {{ post.authorName }}
                      </span>
                      <span class="flex items-center gap-1">
                        <span class="iconify ph--calendar-blank text-xs" />
                        {{ formatDate(post.publishedAt) }}
                      </span>
                    </div>
                    <!-- Summary -->
                    <div
                      v-if="post.summary"
                      class="post-summary mt-1 text-xs leading-relaxed"
                      :style="{ color: textColor3 }"
                    >
                      {{ post.summary }}
                    </div>
                  </div>
                </div>
              </div>
            </NScrollbar>
          </NSpin>
        </div>
      </template>
    </NCard>
  </NModal>
</template>

<style scoped>
.group:hover {
  background: v-bind(hoverBg);
}

.post-summary {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
