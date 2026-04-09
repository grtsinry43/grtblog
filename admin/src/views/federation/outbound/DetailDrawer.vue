<script setup lang="ts">
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { useWindowSize } from '@vueuse/core'
import {
  NDrawer,
  NDrawerContent,
  NDescriptions,
  NDescriptionsItem,
  NTag,
  NCode,
  NButton,
  useMessage,
} from 'naive-ui'
import { ref, computed } from 'vue'

import { retryFederationOutboundLog } from '@/services/federation-admin'

import type { FederationOutboundDeliveryResp } from '@/types/federation'

const props = defineProps<{
  // We use a modelValue to control visibility
  show: boolean
  delivery?: FederationOutboundDeliveryResp
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
}>()

const message = useMessage()
const queryClient = useQueryClient()
const { width } = useWindowSize()

const drawerWidth = computed(() => {
  return width.value < 640 ? '100%' : 600
})

const { mutate: retry, isPending: isRetrying } = useMutation({
  mutationFn: retryFederationOutboundLog,
  onSuccess: () => {
    message.success('已加入重试队列')
    queryClient.invalidateQueries({ queryKey: ['federation-outbound-logs'] })
    emit('update:show', false)
  },
  onError: (err: any) => {
    message.error('重试失败: ' + (err.message || 'Unknown error'))
  },
})

function handleRetry() {
  if (props.delivery?.id) {
    retry(props.delivery.id)
  }
}

function handleClose() {
  emit('update:show', false)
}
</script>

<template>
  <NDrawer
    :show="show"
    @update:show="(val) => emit('update:show', val)"
    :width="drawerWidth"
  >
    <NDrawerContent
      title="出站投递详情"
      closable
    >
      <div
        v-if="delivery"
        class="space-y-6"
      >
        <div
          class="flex justify-end"
          v-if="delivery.status !== 'success'"
        >
          <NButton
            type="warning"
            size="small"
            :loading="isRetrying"
            @click="handleRetry"
          >
            立即重试
          </NButton>
        </div>

        <NDescriptions
          bordered
          :column="1"
          label-placement="left"
          title="基本信息"
        >
          <NDescriptionsItem label="ID">{{ delivery.id }}</NDescriptionsItem>
          <NDescriptionsItem label="Request ID">{{ delivery.request_id }}</NDescriptionsItem>
          <NDescriptionsItem label="类型">
            <NTag>{{ delivery.type }}</NTag>
          </NDescriptionsItem>
          <NDescriptionsItem label="状态">
            <NTag :type="delivery.status === 'success' ? 'success' : 'error'">{{
              delivery.status
            }}</NTag>
          </NDescriptionsItem>
          <NDescriptionsItem label="目标">{{ delivery.target_instance_url }}</NDescriptionsItem>
          <NDescriptionsItem label="创建时间">{{
            new Date(delivery.created_at).toLocaleString()
          }}</NDescriptionsItem>
        </NDescriptions>

        <NDescriptions
          bordered
          :column="1"
          label-placement="left"
          title="投递状态"
        >
          <NDescriptionsItem label="尝试次数"
            >{{ delivery.attempt_count }} / {{ delivery.max_attempts }}</NDescriptionsItem
          >
          <NDescriptionsItem label="下次重试">{{
            delivery.next_retry_at ? new Date(delivery.next_retry_at).toLocaleString() : '-'
          }}</NDescriptionsItem>
          <NDescriptionsItem label="HTTP 状态码">{{
            delivery.http_status || '-'
          }}</NDescriptionsItem>
        </NDescriptions>

        <div v-if="delivery.error_message">
          <h3 class="mb-2 font-bold">错误信息</h3>
          <NCode
            :code="delivery.error_message"
            language="text"
            word-wrap
            class="rounded bg-red-50 p-2"
          />
        </div>

        <div v-if="delivery.response_body">
          <h3 class="mb-2 font-bold">响应体</h3>
          <NCode
            :code="delivery.response_body"
            language="json"
            word-wrap
            class="rounded bg-gray-100 p-2 dark:bg-gray-800"
          />
        </div>
      </div>
    </NDrawerContent>
  </NDrawer>
</template>
