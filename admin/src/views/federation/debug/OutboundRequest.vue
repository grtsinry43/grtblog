<script setup lang="ts">
import { useMutation } from '@tanstack/vue-query'
import {
  NCard,
  NTabs,
  NTabPane,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NInputNumber,
  NAlert,
  useMessage,
} from 'naive-ui'
import { ref } from 'vue'

import { ScrollContainer } from '@/components'
import {
  checkFederationRemote,
  requestFederationCitation,
  notifyFederationMention,
} from '@/services/federation-admin'

import type { FederationAdminCitationReq, FederationAdminMentionReq } from '@/types/federation'

const message = useMessage()

// Remote Check
const remoteCheckUrl = ref('')
const remoteCheckResult = ref<any>(null)
const { mutate: checkRemote, isPending: isCheckingRemote } = useMutation({
  mutationFn: checkFederationRemote,
  onSuccess: (data) => {
    remoteCheckResult.value = data
    message.success('Remote check successful')
  },
  onError: (err: any) => {
    message.error('Remote check failed: ' + (err.message || 'Unknown error'))
  },
})

// Citation Request
const citationPayload = ref<FederationAdminCitationReq>({
  target_instance_url: '',
  target_post_id: '',
  source_article_id: undefined,
  source_short_url: '',
  citation_context: '',
  citation_type: '',
})
const { mutate: sendCitation, isPending: isSendingCitation } = useMutation({
  mutationFn: requestFederationCitation,
  onSuccess: (data) => {
    message.success('Citation request sent. Delivery ID: ' + data.delivery_id)
  },
  onError: (err: any) => {
    message.error('Failed to send citation request: ' + (err.message || 'Unknown error'))
  },
})

// Mention Notify
const mentionPayload = ref<FederationAdminMentionReq>({
  target_instance_url: '',
  mentioned_user: '',
  source_article_id: undefined,
  source_short_url: '',
  mention_context: '',
  mention_type: '',
})
const { mutate: sendMention, isPending: isSendingMention } = useMutation({
  mutationFn: notifyFederationMention,
  onSuccess: (data) => {
    message.success('Mention notification sent. Delivery ID: ' + data.delivery_id)
  },
  onError: (err: any) => {
    message.error('Failed to send mention notification: ' + (err.message || 'Unknown error'))
  },
})
</script>

<template>
  <ScrollContainer wrapper-class="p-4 flex flex-col gap-4">
    <NAlert
      type="warning"
      title="Test Page"
      show-icon
    >
      本页面仅供测试，请在开发者指导下使用
    </NAlert>

    <NCard title="Remote Check">
      <div class="mb-4 flex gap-4">
        <NInput
          v-model:value="remoteCheckUrl"
          placeholder="Enter Target URL (e.g., https://example.com/)"
        />
        <NButton
          type="primary"
          :loading="isCheckingRemote"
          @click="checkRemote(remoteCheckUrl)"
        >
          Check
        </NButton>
      </div>
      <div
        v-if="remoteCheckResult"
        class="max-h-60 overflow-auto rounded bg-gray-100 p-4 font-mono text-xs dark:bg-gray-800"
      >
        <pre>{{ JSON.stringify(remoteCheckResult, null, 2) }}</pre>
      </div>
    </NCard>

    <NCard title="Outbound Requests">
      <NTabs
        type="line"
        animated
      >
        <NTabPane
          name="citation"
          tab="Citation Request (Test)"
        >
          <NForm
            label-placement="left"
            label-width="160"
          >
            <NFormItem
              label="Target Instance"
              required
            >
              <NInput
                v-model:value="citationPayload.target_instance_url"
                placeholder="https://target.com/"
              />
            </NFormItem>
            <NFormItem
              label="Target Post ID"
              required
            >
              <NInput
                v-model:value="citationPayload.target_post_id"
                placeholder="Remote Post ID"
              />
            </NFormItem>
            <NFormItem label="Source Article ID">
              <NInputNumber
                v-model:value="citationPayload.source_article_id"
                placeholder="Local Article ID"
                class="w-full"
              />
            </NFormItem>
            <NFormItem label="Source Short URL">
              <NInput
                v-model:value="citationPayload.source_short_url"
                placeholder="https://mysite.com/s/xyz"
              />
            </NFormItem>
            <NFormItem label="Citation Content">
              <NInput
                v-model:value="citationPayload.citation_context"
                type="textarea"
              />
            </NFormItem>
            <NButton
              type="primary"
              :loading="isSendingCitation"
              @click="sendCitation(citationPayload)"
            >
              Send Citation
            </NButton>
          </NForm>
        </NTabPane>

        <NTabPane
          name="mention"
          tab="Mention Request (Test)"
        >
          <NForm
            label-placement="left"
            label-width="160"
          >
            <NFormItem
              label="Target Instance"
              required
            >
              <NInput
                v-model:value="mentionPayload.target_instance_url"
                placeholder="https://target.com/"
              />
            </NFormItem>
            <NFormItem
              label="Mentioned User"
              required
            >
              <NInput
                v-model:value="mentionPayload.mentioned_user"
                placeholder="username"
              />
            </NFormItem>
            <NFormItem label="Source Article ID">
              <NInputNumber
                v-model:value="mentionPayload.source_article_id"
                placeholder="Local Article ID"
                class="w-full"
              />
            </NFormItem>
            <NFormItem label="Source Short URL">
              <NInput
                v-model:value="mentionPayload.source_short_url"
                placeholder="https://mysite.com/s/xyz"
              />
            </NFormItem>
            <NFormItem label="Mention Context">
              <NInput
                v-model:value="mentionPayload.mention_context"
                type="textarea"
              />
            </NFormItem>
            <NButton
              type="primary"
              :loading="isSendingMention"
              @click="sendMention(mentionPayload)"
            >
              Notify Mention
            </NButton>
          </NForm>
        </NTabPane>
      </NTabs>
    </NCard>
  </ScrollContainer>
</template>
