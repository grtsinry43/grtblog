<script setup lang="ts">
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { NDataTable, NButton, NTag, useMessage, NModal, NInput, NCard, NSpace } from 'naive-ui'
import { h, ref } from 'vue'

import { ScrollContainer } from '@/components'
import {
  getFederationPendingReviews,
  reviewFederationCitation,
  reviewFederationMention,
} from '@/services/federation-admin'

import type { FederationReviewItemResp } from '@/types/federation'
import type { DataTableColumns } from 'naive-ui'

const message = useMessage()
const queryClient = useQueryClient()

const { data, isPending } = useQuery({
  queryKey: ['federation-pending-reviews'],
  queryFn: getFederationPendingReviews,
})

const columns: DataTableColumns<FederationReviewItemResp> = [
  {
    title: '类型',
    key: 'type',
    width: 100,
    render(row) {
      return h(NTag, { type: 'info', bordered: false }, { default: () => row.type })
    },
  },
  { title: '摘要', key: 'summary', ellipsis: { tooltip: true } },
  {
    title: '申请时间',
    key: 'requested_at',
    width: 180,
    render: (row) => new Date(row.requested_at).toLocaleString(),
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render(row) {
      return h(
        NSpace,
        {},
        {
          default: () => [
            h(
              NButton,
              {
                size: 'small',
                type: 'success',
                onClick: () => handleApprove(row),
              },
              { default: () => '批准' },
            ),
            h(
              NButton,
              {
                size: 'small',
                type: 'error',
                onClick: () => openRejectModal(row),
              },
              { default: () => '拒绝' },
            ),
          ],
        },
      )
    },
  },
]

// Approve Logic
const { mutate: approveMutation } = useMutation({
  mutationFn: ({ id, type }: { id: number; type: string }) => {
    if (type === 'citation') return reviewFederationCitation(id, { status: 'approved' })
    if (type === 'mention') return reviewFederationMention(id, { status: 'approved' })
    return Promise.reject(new Error('Unknown type'))
  },
  onSuccess: () => {
    message.success('已批准')
    queryClient.invalidateQueries({ queryKey: ['federation-pending-reviews'] })
  },
  onError: (err: any) => {
    message.error('批准失败: ' + (err.message || 'Unknown error'))
  },
})

function handleApprove(row: FederationReviewItemResp) {
  approveMutation({ id: row.id, type: row.type })
}

// Reject Logic
const showRejectModal = ref(false)
const rejectReason = ref('')
const currentRejectItem = ref<FederationReviewItemResp | null>(null)

function openRejectModal(row: FederationReviewItemResp) {
  currentRejectItem.value = row
  rejectReason.value = ''
  showRejectModal.value = true
}

const { mutate: rejectMutation, isPending: isRejecting } = useMutation({
  mutationFn: ({ id, type, reason }: { id: number; type: string; reason: string }) => {
    if (type === 'citation') return reviewFederationCitation(id, { status: 'rejected', reason })
    if (type === 'mention') return reviewFederationMention(id, { status: 'rejected', reason })
    return Promise.reject(new Error('Unknown type'))
  },
  onSuccess: () => {
    message.success('已拒绝')
    queryClient.invalidateQueries({ queryKey: ['federation-pending-reviews'] })
    showRejectModal.value = false
  },
  onError: (err: any) => {
    message.error('拒绝失败: ' + (err.message || 'Unknown error'))
  },
})

function confirmReject() {
  if (currentRejectItem.value) {
    rejectMutation({
      id: currentRejectItem.value.id,
      type: currentRejectItem.value.type,
      reason: rejectReason.value,
    })
  }
}
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-y-4">
    <NCard :bordered="false">
      <div class="flex items-center justify-between">
        <div class="text-lg font-medium">待审核列表</div>
        <div class="flex items-center gap-2">
          <NButton
            secondary
            size="small"
            :loading="isPending"
            @click="queryClient.invalidateQueries({ queryKey: ['federation-pending-reviews'] })"
          >
            刷新
          </NButton>
        </div>
      </div>
    </NCard>

    <NCard
      :bordered="false"
      content-style="padding: 0;"
    >
      <NDataTable
        :columns="columns"
        :data="data?.items || []"
        :loading="isPending"
        :bordered="false"
        :row-key="(row: FederationReviewItemResp) => row.id"
      />
    </NCard>

    <NModal
      v-model:show="showRejectModal"
      preset="dialog"
      title="拒绝申请"
      positive-text="确认拒绝"
      negative-text="取消"
      @positive-click="confirmReject"
      @negative-click="showRejectModal = false"
      :loading="isRejecting"
    >
      <div class="py-4">
        <NInput
          v-model:value="rejectReason"
          type="textarea"
          placeholder="拒绝理由 (选填)"
        />
      </div>
    </NModal>
  </ScrollContainer>
</template>
