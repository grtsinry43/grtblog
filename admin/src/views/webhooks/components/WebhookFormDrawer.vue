<script setup lang="ts">
import {
  NButton,
  NCheckbox,
  NCheckboxGroup,
  NCollapse,
  NCollapseItem,
  NDivider,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NGi,
  NGrid,
  NInput,
  NSpace,
  NSwitch,
  NTag,
} from 'naive-ui'

import { ScrollContainer } from '@/components'
import TemplateEditor from '@/components/template-editor/TemplateEditor.vue'

import type { AdminEventGroupResp } from '@/services/events'
import type { HeaderRow } from '../composables/use-webhook-form'

defineProps<{
  visible: boolean
  title: string
  actionLabel: string
  saving: boolean
  form: {
    name: string
    url: string
    events: string[]
    headers: HeaderRow[]
    payloadTemplate: string
    isEnabled: boolean
  }
  eventGroups: AdminEventGroupResp[]
  validVariables: string[]
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  save: []
  formatPayload: []
  addHeader: []
  removeHeader: [index: number]
}>()
</script>

<template>
  <NDrawer
    :show="visible"
    placement="right"
    width="min(680px, 100%)"
    @update:show="emit('update:visible', $event)"
  >
    <NDrawerContent
      :title="title"
      closable
      header-style="padding: 20px 24px"
      body-style="padding: 0"
    >
      <ScrollContainer wrapper-class="flex flex-col gap-5">
        <NForm label-placement="top">
          <NGrid cols="1 640:2" x-gap="16" y-gap="12">
            <NGi>
              <NFormItem label="名称">
                <NInput v-model:value="form.name" placeholder="如：联合站点推送" />
              </NFormItem>
            </NGi>
            <NGi>
              <NFormItem label="URL">
                <NInput v-model:value="form.url" placeholder="https://example.com/webhook" />
              </NFormItem>
            </NGi>
          </NGrid>

          <NDivider class="!mb-0 !-mx-4 max-sm:!-mx-2 !w-[calc(100%+2rem)] max-sm:!w-[calc(100%+1rem)]">事件订阅</NDivider>
          <NFormItem label="订阅事件" :show-label="false">
            <NCheckboxGroup v-model:value="form.events" class="w-full">
              <NCollapse arrow-placement="right" class="mt-4 w-full">
                <NCollapseItem
                  v-for="group in eventGroups"
                  :key="group.category"
                  :title="group.category"
                  :name="group.category"
                >
                  <NGrid cols="1 640:2" x-gap="16" y-gap="8" class="pl-6">
                    <NGi v-for="item in group.events" :key="item">
                      <NCheckbox :value="item" :label="item" />
                    </NGi>
                  </NGrid>
                </NCollapseItem>
              </NCollapse>
            </NCheckboxGroup>
          </NFormItem>
          <NFormItem label="启用">
            <NSwitch v-model:value="form.isEnabled" />
          </NFormItem>

          <NDivider>请求配置</NDivider>
          <NFormItem label="Headers">
            <div class="flex w-full flex-col gap-2">
              <div
                v-for="(row, index) in form.headers"
                :key="`${row.key}-${index}`"
                class="flex items-center gap-2"
              >
                <div class="w-40">
                  <NInput v-model:value="row.key" placeholder="Header Key" class="w-full" />
                </div>
                <div class="min-w-0 flex-1">
                  <NInput v-model:value="row.value" placeholder="Header Value" class="w-full" />
                </div>
                <NButton tertiary size="small" @click="emit('removeHeader', index)">删除</NButton>
              </div>
              <div>
                <NButton size="small" @click="emit('addHeader')">添加 Header</NButton>
              </div>
            </div>
          </NFormItem>
          <NFormItem label="Payload 模板">
            <div class="flex w-full flex-col gap-2">
              <div class="flex justify-end">
                <NButton size="small" tertiary @click="emit('formatPayload')">格式化</NButton>
              </div>
              <TemplateEditor v-model="form.payloadTemplate" :valid-variables="validVariables" />
              <div
                v-if="validVariables.length > 0"
                class="rounded border bg-gray-50 p-3 text-xs dark:border-neutral-700 dark:bg-neutral-800"
              >
                <div class="mb-2 text-gray-500">可用变量：</div>
                <NSpace size="small">
                  <NTag
                    v-for="v in validVariables"
                    :key="v"
                    size="small"
                    type="info"
                    dashed
                    :bordered="false"
                    class="cursor-pointer select-all bg-white dark:bg-neutral-900"
                  >
                    {{ v }}
                  </NTag>
                </NSpace>
              </div>
              <div
                v-else
                class="rounded border bg-gray-50 p-3 text-xs text-gray-400 dark:border-neutral-700 dark:bg-neutral-800"
              >
                * 选择上方事件以查看特定事件变量
              </div>
            </div>
          </NFormItem>
        </NForm>
        <div class="flex justify-end gap-2">
          <NButton @click="emit('update:visible', false)">取消</NButton>
          <NButton type="primary" :loading="saving" @click="emit('save')">
            {{ actionLabel }}
          </NButton>
        </div>
      </ScrollContainer>
    </NDrawerContent>
  </NDrawer>
</template>
