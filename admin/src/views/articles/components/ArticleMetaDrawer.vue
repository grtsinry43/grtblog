<script setup lang="ts">
import {
  NAlert,
  NAutoComplete,
  NButton,
  NDivider,
  NDrawer,
  NDrawerContent,
  NDynamicTags,
  NEmpty,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NPopconfirm,
  NSelect,
  NSkeleton,
  NSwitch,
  NTag,
} from 'naive-ui'

import ImageInput from '@/components/image-picker/ImageInput.vue'
import AiSummaryAssist from '@/views/shared/content-editor/components/AiSummaryAssist.vue'

import type { ArticleEditorForm } from '../composables/use-article-form'
import type { NewCategoryModalState } from '../composables/use-taxonomy-select'
import type { ArticleDetail } from '@/services/articles'
import type { FederationOutboundInteractionResp } from '@/types/federation'
import type { SelectOption } from 'naive-ui'

type FederationSignalType = 'mention' | 'citation'

export interface FederationSignalRow {
  key: string
  type: FederationSignalType
  instance: string
  target: string
  marker: string
  inContent: boolean
  deliveredAt: string | null
  outbound: FederationOutboundInteractionResp | null
}

const show = defineModel<boolean>('show', { required: true })
const form = defineModel<ArticleEditorForm>('form', { required: true })

defineProps<{
  categoryOptions: SelectOption[]
  dynamicTags: string[]
  tagSearchValue: string
  autoCompleteOptions: { label: string; value: string }[]
  newCategoryModal: NewCategoryModalState
  aiSummaryLoading: boolean
  aiSummaryResult: string
  aiSummaryDone: boolean
  isYearSummary: boolean
  yearSummaryYear: number
  apStatusText: string
  apLastPublishedAtText: string
  loadedArticle: ArticleDetail | null
  canRepublishToActivityPub: boolean
  apPublishing: boolean
  isCreating: boolean
  federationInteractionsError: string
  federationInteractionsLoading: boolean
  federationSignalRows: FederationSignalRow[]
  resetAllFederationLoading: boolean
  resetSignalLoadingKeys: Record<string, boolean>
  formatDateTime: (value?: string | null) => string
  signalStatusText: (row: FederationSignalRow) => string
}>()

const emit = defineEmits<{
  'update:tagSearchValue': [value: string]
  'update:isYearSummary': [value: boolean]
  'update:yearSummaryYear': [value: number | null]
  openCategoryModal: []
  tagsChange: [value: string[]]
  addTag: [value: string]
  generateAiSummary: []
  adoptAiSummary: []
  dismissAiSummary: []
  republishActivityPub: []
  resetAllFederationSignals: []
  resetFederationSignal: [row: FederationSignalRow]
}>()

function onTagEnter(event: KeyboardEvent, value: string) {
  if (event.key === 'Enter') {
    event.preventDefault()
    emit('addTag', value)
  }
}

function outboundStatusTagType(status?: string | null) {
  const normalized = (status || '').trim().toLowerCase()
  if (normalized === 'approved' || normalized === 'accepted') return 'success'
  if (normalized === 'queued' || normalized === 'sending') return 'warning'
  if (
    normalized === 'rejected' ||
    normalized === 'failed' ||
    normalized === 'timeout' ||
    normalized === 'dead'
  ) {
    return 'error'
  }
  return 'default'
}

function emitResetRow(row: FederationSignalRow) {
  emit('resetFederationSignal', row)
}
</script>

<template>
  <NDrawer
    v-model:show="show"
    placement="right"
    width="400"
  >
    <NDrawerContent
      title="文章设置"
      :native-scrollbar="false"
      closable
      header-style="padding: 24px;"
      body-style="padding: 24px;"
    >
      <div class="flex flex-col gap-6">
        <div class="space-y-4">
          <div class="flex items-center gap-2 text-sm font-medium">
            <div class="iconify ph--tag" />
            <span>分类与标签</span>
          </div>
          <NForm
            label-placement="top"
            label-width="auto"
            class="space-y-4"
          >
            <NFormItem
              label="分类"
              :show-feedback="false"
            >
              <div class="flex w-full items-center gap-2">
                <NSelect
                  v-model:value="form.categoryId"
                  :options="categoryOptions"
                  placeholder="选择分类"
                  clearable
                  filterable
                  class="flex-1"
                />
                <NButton
                  quaternary
                  size="small"
                  @click="$emit('openCategoryModal')"
                  >新建</NButton
                >
              </div>
            </NFormItem>
            <NFormItem
              label="标签"
              :show-feedback="false"
            >
              <div class="flex w-full flex-col gap-2">
                <NDynamicTags
                  :value="dynamicTags"
                  @update:value="$emit('tagsChange', $event)"
                />
                <div class="flex items-center gap-2">
                  <NAutoComplete
                    :value="tagSearchValue"
                    :options="autoCompleteOptions"
                    placeholder="搜索或创建标签"
                    class="flex-1"
                    :input-props="{
                      onKeydown: (event: KeyboardEvent) => onTagEnter(event, tagSearchValue),
                    }"
                    @update:value="$emit('update:tagSearchValue', $event)"
                    @select="$emit('addTag', $event)"
                  />
                  <NButton
                    quaternary
                    size="small"
                    @click="$emit('addTag', tagSearchValue)"
                    >添加</NButton
                  >
                </div>
              </div>
            </NFormItem>
          </NForm>
        </div>

        <NDivider style="margin: 0" />

        <div class="space-y-4">
          <div class="flex items-center gap-2 text-sm font-medium">
            <div class="iconify ph--article" />
            <span>元信息</span>
          </div>
          <NForm
            label-placement="top"
            label-width="auto"
            class="space-y-4"
          >
            <NFormItem :show-feedback="false">
              <template #label>
                <span>摘要</span>
                <span class="ml-1 text-xs opacity-50">用于外显描述、OG 信息、SEO</span>
              </template>
              <NInput
                v-model:value="form.summary"
                type="textarea"
                placeholder="文章外显摘要，用于列表卡片、网页描述、社交分享..."
                :autosize="{ minRows: 2, maxRows: 4 }"
              />
            </NFormItem>
            <AiSummaryAssist
              :model-value="form.aiSummary"
              :loading="aiSummaryLoading"
              :result="aiSummaryResult"
              :done="aiSummaryDone"
              placeholder="AI 生成的内容导读，展示在正文之前..."
              :disabled="!form.content?.trim()"
              @update:model-value="form.aiSummary = $event"
              @generate="$emit('generateAiSummary')"
              @adopt="$emit('adoptAiSummary')"
              @dismiss="$emit('dismissAiSummary')"
            />
            <NFormItem
              label="引言"
              :show-feedback="false"
            >
              <NInput
                v-model:value="form.leadIn"
                type="textarea"
                placeholder="文章引言..."
                :autosize="{ minRows: 2, maxRows: 4 }"
              />
            </NFormItem>
            <NFormItem
              label="封面图"
              :show-feedback="false"
            >
              <ImageInput v-model:value="form.cover" />
            </NFormItem>
          </NForm>
        </div>

        <NDivider style="margin: 0" />

        <div class="space-y-4">
          <div class="flex items-center gap-2 text-sm font-medium">
            <div class="iconify ph--toggle-left" />
            <span>属性</span>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="flex items-center justify-between rounded-lg px-4 py-3">
              <span class="text-sm">置顶</span>
              <NSwitch
                v-model:value="form.isTop"
                size="small"
              />
            </div>
            <div class="flex items-center justify-between rounded-lg px-4 py-3">
              <span class="text-sm">允许评论</span>
              <NSwitch
                v-model:value="form.allowComment"
                size="small"
              />
            </div>
            <div class="flex items-center justify-between rounded-lg px-4 py-3">
              <span class="text-sm">原创</span>
              <NSwitch
                v-model:value="form.isOriginal"
                size="small"
              />
            </div>
            <div class="col-span-2 rounded-lg px-4 py-3">
              <div class="flex items-center justify-between gap-3">
                <span class="text-sm">这是年终总结</span>
                <NSwitch
                  :value="isYearSummary"
                  size="small"
                  @update:value="$emit('update:isYearSummary', $event)"
                />
              </div>
              <div
                v-if="isYearSummary"
                class="mt-3"
              >
                <NInputNumber
                  :value="yearSummaryYear"
                  :min="1900"
                  :max="3000"
                  :precision="0"
                  class="w-full"
                  placeholder="输入年份，例如 2024"
                  @update:value="$emit('update:yearSummaryYear', $event)"
                />
              </div>
            </div>
            <div class="col-span-2 rounded-lg px-4 py-3">
              <div class="flex items-start justify-between gap-4">
                <div class="min-w-0 space-y-1">
                  <div class="text-sm">ActivityPub：{{ apStatusText }}</div>
                  <div class="text-xs opacity-70">最近发布：{{ apLastPublishedAtText }}</div>
                  <div
                    v-if="loadedArticle?.activityPubObjectId"
                    class="text-xs break-all opacity-70"
                  >
                    {{ loadedArticle.activityPubObjectId }}
                  </div>
                </div>
                <NButton
                  size="small"
                  secondary
                  :loading="apPublishing"
                  :disabled="!canRepublishToActivityPub || apPublishing"
                  @click="$emit('republishActivityPub')"
                >
                  手动补发
                </NButton>
              </div>
            </div>
          </div>
        </div>

        <NDivider style="margin: 0" />

        <div class="space-y-4">
          <div class="flex items-center justify-between gap-3">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--circles-three" />
              <span>联合条目状态</span>
            </div>
            <NPopconfirm
              trigger="click"
              @positive-click="$emit('resetAllFederationSignals')"
            >
              <template #trigger>
                <NButton
                  size="small"
                  secondary
                  :loading="resetAllFederationLoading"
                  :disabled="
                    isCreating || federationSignalRows.length === 0 || resetAllFederationLoading
                  "
                >
                  全部重置
                </NButton>
              </template>
              将重置全部联合条目状态，并尝试重新触发出站。
            </NPopconfirm>
          </div>

          <NAlert
            v-if="isCreating"
            type="info"
            :show-icon="false"
          >
            新建文章保存后可查看联合条目状态。
          </NAlert>

          <NAlert
            v-else-if="federationInteractionsError"
            type="warning"
            :show-icon="false"
          >
            {{ federationInteractionsError }}
          </NAlert>

          <div
            v-else-if="federationInteractionsLoading"
            class="space-y-3"
          >
            <NSkeleton
              text
              :repeat="2"
            />
            <NSkeleton
              text
              :repeat="2"
            />
          </div>

          <NEmpty
            v-else-if="federationSignalRows.length === 0"
            size="small"
            description="当前未识别到联合条目"
          />

          <div
            v-else
            class="space-y-3"
          >
            <div
              v-for="row in federationSignalRows"
              :key="row.key"
              class="rounded-lg border border-current/10 p-3"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0 space-y-1">
                  <div class="flex flex-wrap items-center gap-2">
                    <NTag
                      size="small"
                      :bordered="false"
                      :type="row.type === 'mention' ? 'info' : 'warning'"
                    >
                      {{ row.type === 'mention' ? '提及' : '引用' }}
                    </NTag>
                    <NTag
                      size="small"
                      :bordered="false"
                      :type="outboundStatusTagType(row.outbound?.status)"
                    >
                      {{ signalStatusText(row) }}
                    </NTag>
                    <NTag
                      v-if="!row.inContent"
                      size="small"
                      :bordered="false"
                      type="default"
                    >
                      已不在正文
                    </NTag>
                  </div>
                  <div class="font-mono text-xs break-all opacity-80">
                    {{ row.marker }}
                  </div>
                  <div class="text-xs opacity-70">目标：{{ row.instance }} / {{ row.target }}</div>
                  <div
                    v-if="row.deliveredAt"
                    class="text-xs opacity-70"
                  >
                    已记录时间：{{ formatDateTime(row.deliveredAt) }}
                  </div>
                  <div
                    v-if="row.outbound?.updated_at"
                    class="text-xs opacity-70"
                  >
                    队列更新时间：{{ formatDateTime(row.outbound.updated_at) }}
                  </div>
                </div>

                <NPopconfirm
                  trigger="click"
                  @positive-click="emitResetRow(row)"
                >
                  <template #trigger>
                    <NButton
                      size="tiny"
                      secondary
                      :loading="!!resetSignalLoadingKeys[row.key]"
                      :disabled="!!resetSignalLoadingKeys[row.key]"
                    >
                      重置
                    </NButton>
                  </template>
                  将重置该条目状态，并尝试重新触发一次出站。
                </NPopconfirm>
              </div>
            </div>
          </div>
        </div>
      </div>
    </NDrawerContent>
  </NDrawer>
</template>
