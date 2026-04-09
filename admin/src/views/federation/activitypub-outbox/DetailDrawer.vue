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
  NDataTable,
  useMessage,
} from 'naive-ui'
import { computed, h, ref, watch } from 'vue'

import { getActivityPubOutboxDetail, retryActivityPubOutbox } from '@/services/federation-admin'

import type { ActivityPubDeliveryDetailResp, ActivityPubOutboxItemResp } from '@/types/federation'
import type { DataTableColumns } from 'naive-ui'

const props = defineProps<{ show: boolean; item?: ActivityPubOutboxItemResp }>()
const emit = defineEmits<{ (e: 'update:show', value: boolean): void; (e: 'refresh'): void }>()
const message = useMessage()

const { width } = useWindowSize()
const drawerWidth = computed(() => (width.value < 640 ? '100%' : 860))

const detail = ref<ActivityPubOutboxItemResp | undefined>()
const loading = ref(false)
const retrying = ref(false)

watch(
  () => [props.show, props.item?.id],
  async () => {
    if (props.show && props.item?.id) {
      loading.value = true
      try {
        detail.value = await getActivityPubOutboxDetail(props.item.id)
      } catch {
        detail.value = props.item
      } finally {
        loading.value = false
      }
    }
  },
)

const displayItem = computed(() => detail.value ?? props.item)
const canRetry = computed(
  () => displayItem.value?.status === 'failed' || displayItem.value?.status === 'partial',
)
const activityCode = computed(() => {
  const raw = displayItem.value?.activity
  if (!raw) return '{}'
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
})

function statusTagType(status?: string) {
  switch (status) {
    case 'completed':
      return 'success'
    case 'partial':
      return 'warning'
    case 'failed':
      return 'error'
    case 'sending':
      return 'info'
    default:
      return 'default'
  }
}

const columns: DataTableColumns<ActivityPubDeliveryDetailResp> = [
  { title: 'Inbox', key: 'inbox', minWidth: 220, ellipsis: { tooltip: true } },
  { title: 'ActorID', key: 'actor_id', minWidth: 220, ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render(row) {
      return h(
        NTag,
        { size: 'small', type: row.status === 'success' ? 'success' : 'error' },
        { default: () => row.status || '-' },
      )
    },
  },
  { title: 'HTTP', key: 'http_status', width: 80 },
  { title: '错误', key: 'error', minWidth: 220, ellipsis: { tooltip: true } },
  {
    title: '投递时间',
    key: 'delivered_at',
    width: 170,
    render(row) {
      return row.delivered_at ? new Date(row.delivered_at).toLocaleString() : '-'
    },
  },
]

async function handleRetry() {
  if (!displayItem.value?.id) return
  retrying.value = true
  try {
    detail.value = await retryActivityPubOutbox(displayItem.value.id)
    message.success('重试完成')
    emit('refresh')
  } catch (err: any) {
    message.error(err?.message || '重试失败')
  } finally {
    retrying.value = false
  }
}
</script>

<template>
  <NDrawer
    :show="show"
    :width="drawerWidth"
    @update:show="(v) => emit('update:show', v)"
  >
    <NDrawerContent
      title="ActivityPub 出站详情"
      closable
    >
      <div
        v-if="displayItem"
        class="space-y-6"
      >
        <NDescriptions
          bordered
          :column="2"
          label-placement="left"
          title="基本信息"
        >
          <NDescriptionsItem label="ID">{{ displayItem.id }}</NDescriptionsItem>
          <NDescriptionsItem label="状态"
            ><NTag
              :type="statusTagType(displayItem.status)"
              size="small"
              >{{ displayItem.status }}</NTag
            ></NDescriptionsItem
          >
          <NDescriptionsItem label="SourceType">{{ displayItem.source_type }}</NDescriptionsItem>
          <NDescriptionsItem label="TriggerSource">{{
            displayItem.trigger_source
          }}</NDescriptionsItem>
          <NDescriptionsItem label="ActivityID">{{ displayItem.activity_id }}</NDescriptionsItem>
          <NDescriptionsItem label="ObjectID">{{ displayItem.object_id }}</NDescriptionsItem>
          <NDescriptionsItem label="统计"
            >{{ displayItem.success_count }}/{{ displayItem.total_targets }}</NDescriptionsItem
          >
          <NDescriptionsItem label="耗时(ms)">{{
            displayItem.duration_ms ?? '-'
          }}</NDescriptionsItem>
          <NDescriptionsItem label="开始时间">{{
            displayItem.started_at ? new Date(displayItem.started_at).toLocaleString() : '-'
          }}</NDescriptionsItem>
          <NDescriptionsItem label="结束时间">{{
            displayItem.finished_at ? new Date(displayItem.finished_at).toLocaleString() : '-'
          }}</NDescriptionsItem>
        </NDescriptions>

        <div>
          <div class="mb-2 flex items-center justify-between">
            <h3 class="font-medium">投递明细</h3>
            <NButton
              v-if="canRetry"
              type="warning"
              :loading="retrying"
              @click="handleRetry"
              >重试失败投递</NButton
            >
          </div>
          <NDataTable
            :columns="columns"
            :data="displayItem.deliveries || []"
            :loading="loading"
            :scroll-x="1000"
          />
        </div>

        <div>
          <h3 class="mb-2 font-medium">Activity JSON</h3>
          <NCode
            :code="activityCode"
            language="json"
            word-wrap
          />
        </div>
      </div>
      <template #footer>
        <NButton @click="emit('update:show', false)">关闭</NButton>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>
