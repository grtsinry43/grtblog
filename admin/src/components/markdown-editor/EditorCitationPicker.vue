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
import type { FederationInstanceResp, FederationRemotePostResp } from '@/types/federation'

const props = defineProps<{
  show: boolean
  step: 'input' | 'posts'
  // URL 输入
  urlInput: string
  urlValid: boolean
  urlError: string
  // 已有实例快捷选项
  instances: FederationInstanceResp[]
  instancesLoading: boolean
  // 文章列表
  posts: FederationRemotePostResp[]
  postsLoading: boolean
  searchQuery: string
  // 分页
  page: number
  total: number
  pageSize: number
  // 当前远端信息
  resolvedURL: string
  resolvedName: string
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'urlInput', value: string): void
  (e: 'submitURL'): void
  (e: 'selectInstance', inst: FederationInstanceResp): void
  (e: 'searchPosts', query: string): void
  (e: 'select', post: FederationRemotePostResp): void
  (e: 'back'): void
  (e: 'goToPage', page: number): void
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

const totalPages = computed(() => props.pageSize > 0 ? Math.ceil(props.total / props.pageSize) : 1)
const hasPrev = computed(() => props.page > 1)
const hasNext = computed(() => props.page < totalPages.value)

const manualInstance = ref('')
const manualPostId = ref('')

function handleManualInsert() {
  if (manualInstance.value.trim() && manualPostId.value.trim()) {
    emit('insertRaw', manualInstance.value.trim(), manualPostId.value.trim())
    manualInstance.value = ''
    manualPostId.value = ''
  }
}

function handleURLKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && props.urlValid) {
    emit('submitURL')
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
    return url.replace(/^https?:\/\//, '').replace(/[:/].*$/, '')
  }
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
          <div
            class="grid size-8 shrink-0 place-items-center rounded-lg"
            :style="{ background: infoBadgeBg }"
          >
            <span
              class="iconify ph--quotes-bold text-base"
              :style="{ color: infoIconColor }"
            />
          </div>
          <div class="flex items-center gap-1.5 text-sm">
            <span
              :class="step === 'input'
                ? 'font-medium'
                : 'cursor-pointer transition-colors duration-150'"
              :style="{
                color: step === 'input' ? textColor1 : textColor3,
              }"
              @click="step === 'posts' ? emit('back') : undefined"
              @mouseenter="($event.target as HTMLElement).style.color = step === 'posts' ? themeVars.infoColor : ''"
              @mouseleave="($event.target as HTMLElement).style.color = step === 'posts' ? textColor3 : textColor1"
            >
              输入远端地址
            </span>
            <template v-if="step === 'posts'">
              <span :style="{ color: textColor3 }">›</span>
              <span
                class="max-w-[200px] truncate font-medium"
                :style="{ color: textColor1 }"
              >
                {{ resolvedName || extractHost(resolvedURL) }}
              </span>
            </template>
          </div>
        </div>
      </template>

      <!-- ========== Step 1: URL input + instance shortcuts ========== -->
      <template v-if="step === 'input'">
        <div class="flex flex-col gap-3">
          <!-- URL input -->
          <div class="flex gap-2">
            <NInput
              :value="urlInput"
              placeholder="输入远端博客地址，如 blog.example.com"
              clearable
              :status="urlError ? 'error' : undefined"
              class="flex-1"
              @update:value="emit('urlInput', $event)"
              @keydown="handleURLKeydown"
            >
              <template #prefix>
                <span class="iconify ph--globe text-base" :style="{ color: textColor3 }" />
              </template>
            </NInput>
            <NButton
              type="primary"
              :disabled="!urlValid"
              @click="emit('submitURL')"
            >
              <template #icon>
                <span class="iconify ph--arrow-right" />
              </template>
              拉取
            </NButton>
          </div>
          <div v-if="urlError" class="text-xs" style="color: var(--error-color, #e88080); margin-top: -6px">
            {{ urlError }}
          </div>

          <!-- Instance shortcuts -->
          <div v-if="instances.length > 0" class="mt-1">
            <div class="mb-2 text-xs" :style="{ color: textColor3 }">
              <span class="iconify ph--lightning inline-block align-text-bottom text-sm" />
              快捷选择已联合实例
            </div>
            <NSpin :show="instancesLoading">
              <NScrollbar style="max-height: 240px">
                <div class="flex flex-col gap-0.5">
                  <div
                    v-for="inst in instances"
                    :key="inst.id"
                    class="group flex cursor-pointer items-center gap-3 px-2.5 py-2 transition-colors duration-150"
                    :style="{ borderRadius }"
                    @click="emit('selectInstance', inst)"
                  >
                    <div
                      class="grid size-8 shrink-0 place-items-center rounded-lg"
                      :style="{ background: infoBadgeBg }"
                    >
                      <span
                        class="iconify ph--globe-simple text-base"
                        :style="{ color: infoIconColor }"
                      />
                    </div>
                    <div class="min-w-0 flex-1">
                      <div class="truncate text-sm font-medium" :style="{ color: textColor1 }">
                        {{ inst.name || extractHost(inst.base_url) }}
                      </div>
                      <div class="truncate text-xs" :style="{ color: textColor3 }">
                        {{ extractHost(inst.base_url) }}
                      </div>
                    </div>
                    <span
                      class="iconify ph--caret-right shrink-0 text-base opacity-40 transition-opacity duration-150 group-hover:opacity-80"
                      :style="{ color: textColor3 }"
                    />
                  </div>
                </div>
              </NScrollbar>
            </NSpin>
          </div>

          <!-- Manual input fallback -->
          <NCollapse arrow-placement="right" class="mt-1">
            <NCollapseItem title="手动输入引用标记" name="manual">
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
          <!-- Context bar -->
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
                {{ resolvedName || extractHost(resolvedURL) }}
              </span>
              <span class="ml-1.5 text-xs" :style="{ color: textColor3 }">
                {{ extractHost(resolvedURL) }}
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
          <NSpin :show="postsLoading">
            <NScrollbar style="max-height: 360px">
              <!-- Empty: search yielded nothing -->
              <div
                v-if="!postsLoading && posts.length === 0 && searchQuery.trim()"
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

              <!-- Empty: no posts from remote -->
              <div
                v-else-if="!postsLoading && posts.length === 0 && !searchQuery.trim()"
                class="flex flex-col items-center justify-center gap-2 py-12"
              >
                <span
                  class="iconify ph--article text-3xl"
                  :style="{ color: textColor3 }"
                />
                <span class="text-sm" :style="{ color: textColor3 }">
                  该远端暂无可用文章，可能不支持联合协议
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
                    v-if="post.cover_image"
                    class="shrink-0 overflow-hidden rounded-md"
                    style="width: 72px; height: 54px"
                  >
                    <img
                      :src="post.cover_image"
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
                    <div class="flex items-center gap-1.5">
                      <span
                        class="flex-1 truncate text-sm font-medium"
                        :style="{ color: textColor1 }"
                      >
                        {{ post.title }}
                      </span>
                      <NTooltip v-if="!post.allow_citation" trigger="hover">
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
                    <div class="mt-0.5 flex items-center gap-3 text-xs" :style="{ color: textColor3 }">
                      <span v-if="post.author?.name" class="flex items-center gap-1">
                        <span class="iconify ph--user text-xs" />
                        {{ post.author.name }}
                      </span>
                      <span class="flex items-center gap-1">
                        <span class="iconify ph--calendar-blank text-xs" />
                        {{ formatDate(post.published_at) }}
                      </span>
                    </div>
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

          <!-- Pager -->
          <div
            v-if="totalPages > 1"
            class="flex items-center justify-between pt-2"
          >
            <span class="text-xs" :style="{ color: textColor3 }">
              共 {{ total }} 篇，第 {{ page }}/{{ totalPages }} 页
            </span>
            <div class="flex gap-1.5">
              <NButton
                size="tiny"
                :disabled="!hasPrev || postsLoading"
                @click="emit('goToPage', page - 1)"
              >
                <template #icon>
                  <span class="iconify ph--caret-left" />
                </template>
              </NButton>
              <NButton
                size="tiny"
                :disabled="!hasNext || postsLoading"
                @click="emit('goToPage', page + 1)"
              >
                <template #icon>
                  <span class="iconify ph--caret-right" />
                </template>
              </NButton>
            </div>
          </div>
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
