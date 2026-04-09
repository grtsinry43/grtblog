<script setup lang="ts">
import { useWindowSize } from '@vueuse/core'
import {
  NDrawer,
  NDrawerContent,
  NDescriptions,
  NDescriptionsItem,
  NTag,
  NCode,
  NButton,
  NTabs,
  NTabPane,
} from 'naive-ui'
import { computed, ref, watch } from 'vue'

import { getEmailOutboxDetail } from '@/services/email'

import type { EmailOutbox } from '@/services/email'

const props = defineProps<{
  show: boolean
  outbox?: EmailOutbox
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
}>()

const { width } = useWindowSize()
const drawerWidth = computed(() => (width.value < 640 ? '100%' : 600))

const detail = ref<EmailOutbox | undefined>(undefined)
const loadingDetail = ref(false)

watch(
  () => [props.show, props.outbox?.id],
  async () => {
    if (props.show && props.outbox?.id) {
      loadingDetail.value = true
      try {
        detail.value = await getEmailOutboxDetail(props.outbox.id)
      } catch {
        detail.value = props.outbox
      } finally {
        loadingDetail.value = false
      }
    }
  },
)

const displayItem = computed(() => detail.value ?? props.outbox)

function statusTagType(status?: string) {
  switch (status) {
    case 'sent':
      return 'success'
    case 'failed':
      return 'error'
    case 'sending':
      return 'warning'
    case 'pending':
      return 'info'
    default:
      return 'default'
  }
}

function statusLabel(status?: string) {
  switch (status) {
    case 'pending':
      return '等待中'
    case 'sending':
      return '发送中'
    case 'sent':
      return '已发送'
    case 'failed':
      return '失败'
    default:
      return status ?? '-'
  }
}

function handleClose() {
  emit('update:show', false)
}
</script>

<template>
  <NDrawer
    :show="show"
    :width="drawerWidth"
    @update:show="(val) => emit('update:show', val)"
  >
    <NDrawerContent
      title="邮件出站详情"
      closable
    >
      <div
        v-if="displayItem"
        class="space-y-6"
      >
        <NDescriptions
          bordered
          :column="1"
          label-placement="left"
          title="基本信息"
        >
          <NDescriptionsItem label="ID">{{ displayItem.id }}</NDescriptionsItem>
          <NDescriptionsItem label="主题">{{ displayItem.subject }}</NDescriptionsItem>
          <NDescriptionsItem label="模板编码">{{ displayItem.templateCode }}</NDescriptionsItem>
          <NDescriptionsItem label="事件名">
            <NTag
              type="info"
              size="small"
              >{{ displayItem.eventName }}</NTag
            >
          </NDescriptionsItem>
          <NDescriptionsItem label="收件人">{{
            displayItem.toEmails?.join(', ') || '-'
          }}</NDescriptionsItem>
          <NDescriptionsItem label="状态">
            <NTag
              :type="statusTagType(displayItem.status)"
              size="small"
              >{{ statusLabel(displayItem.status) }}</NTag
            >
          </NDescriptionsItem>
          <NDescriptionsItem label="创建时间">{{
            new Date(displayItem.createdAt).toLocaleString()
          }}</NDescriptionsItem>
          <NDescriptionsItem label="发送时间">{{
            displayItem.sentAt ? new Date(displayItem.sentAt).toLocaleString() : '-'
          }}</NDescriptionsItem>
        </NDescriptions>

        <NDescriptions
          bordered
          :column="1"
          label-placement="left"
          title="重试信息"
        >
          <NDescriptionsItem label="重试次数">{{ displayItem.retryCount }}</NDescriptionsItem>
          <NDescriptionsItem label="下次重试">{{
            displayItem.nextRetryAt ? new Date(displayItem.nextRetryAt).toLocaleString() : '-'
          }}</NDescriptionsItem>
        </NDescriptions>

        <div v-if="displayItem.lastError">
          <h3 class="mb-2 font-bold">错误信息</h3>
          <NCode
            :code="displayItem.lastError"
            language="text"
            word-wrap
          />
        </div>

        <div v-if="displayItem.htmlBody || displayItem.textBody">
          <NTabs
            type="line"
            size="small"
          >
            <NTabPane
              v-if="displayItem.htmlBody"
              name="html"
              tab="HTML 正文"
            >
              <div
                class="max-h-80 overflow-auto rounded border p-2"
                v-html="displayItem.htmlBody"
              />
            </NTabPane>
            <NTabPane
              v-if="displayItem.textBody"
              name="text"
              tab="纯文本"
            >
              <NCode
                :code="displayItem.textBody"
                language="text"
                word-wrap
              />
            </NTabPane>
          </NTabs>
        </div>
      </div>
      <template #footer>
        <NButton @click="handleClose">关闭</NButton>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>
