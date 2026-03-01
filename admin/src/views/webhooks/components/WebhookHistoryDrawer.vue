<script setup lang="ts">
import {
  NAlert,
  NCard,
  NCode,
  NDrawer,
  NDrawerContent,
  NEmpty,
  NSpace,
  NTabPane,
  NTable,
  NTag,
  NTabs,
} from 'naive-ui'
import { formatDate } from '@/utils/format'

import { ScrollContainer } from '@/components'

import type { WebhookHistoryItem } from '@/services/webhooks'
import type { StatusTagType } from '../composables/use-webhook-form'

defineProps<{
  visible: boolean
  activeHistory: WebhookHistoryItem | null
  detailStatus: { label: string; type: StatusTagType }
  webhookMap: Map<number, string>
  formatHeaders: (headers?: Record<string, string>) => string
  formatBody: (body?: string) => string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()
</script>

<template>
  <NDrawer
    :show="visible"
    placement="right"
    width="640"
    @update:show="emit('update:visible', $event)"
  >
    <NDrawerContent
      title="投递详情"
      closable
      header-style="padding: 20px 24px"
      body-style="padding: 0"
    >
      <ScrollContainer wrapper-class="flex flex-col gap-4">
        <NEmpty v-if="!activeHistory" description="暂无投递详情" />
        <template v-else>
          <NCard title="概览">
            <NTable size="small" :bordered="false" :single-line="false">
              <tbody>
                <tr>
                  <th class="w-24 text-xs text-[var(--text-color-3)]">事件</th>
                  <td class="font-medium">{{ activeHistory.eventName || '-' }}</td>
                </tr>
                <tr>
                  <th class="text-xs text-[var(--text-color-3)]">Webhook</th>
                  <td>{{ webhookMap.get(activeHistory.webhookId) || `#${activeHistory.webhookId}` }}</td>
                </tr>
                <tr>
                  <th class="text-xs text-[var(--text-color-3)]">请求 URL</th>
                  <td class="break-words font-mono text-xs">{{ activeHistory.requestUrl || '-' }}</td>
                </tr>
                <tr>
                  <th class="text-xs text-[var(--text-color-3)]">状态</th>
                  <td>
                    <NTag
                      size="small"
                      :bordered="false"
                      :type="detailStatus.type === 'default' ? undefined : detailStatus.type"
                    >
                      {{ detailStatus.label }}
                    </NTag>
                  </td>
                </tr>
                <tr>
                  <th class="text-xs text-[var(--text-color-3)]">测试</th>
                  <td>
                    <NTag v-if="activeHistory.isTest" size="small" type="warning" :bordered="false">是</NTag>
                    <span v-else>否</span>
                  </td>
                </tr>
                <tr>
                  <th class="text-xs text-[var(--text-color-3)]">时间</th>
                  <td>{{ formatDate(activeHistory.createdAt) }}</td>
                </tr>
              </tbody>
            </NTable>
          </NCard>

          <NCard>
            <NTabs type="segment" animated>
              <NTabPane name="request" tab="请求">
                <NSpace vertical size="large">
                  <NCard size="small" title="Headers">
                    <NCode :code="formatHeaders(activeHistory.requestHeaders)" word-wrap />
                  </NCard>
                  <NCard size="small" title="Body">
                    <NCode :code="formatBody(activeHistory.requestBody)" language="json" word-wrap />
                  </NCard>
                </NSpace>
              </NTabPane>
              <NTabPane name="response" tab="响应">
                <NSpace vertical size="large">
                  <NCard size="small" title="Headers">
                    <NCode :code="formatHeaders(activeHistory.responseHeaders)" word-wrap />
                  </NCard>
                  <NCard size="small" title="Body">
                    <NCode :code="formatBody(activeHistory.responseBody)" language="json" word-wrap />
                  </NCard>
                  <NAlert v-if="activeHistory.errorMessage" type="error" :show-icon="false">
                    {{ activeHistory.errorMessage }}
                  </NAlert>
                  <NAlert v-else type="info" :show-icon="false">无错误信息</NAlert>
                </NSpace>
              </NTabPane>
            </NTabs>
          </NCard>
        </template>
      </ScrollContainer>
    </NDrawerContent>
  </NDrawer>
</template>
