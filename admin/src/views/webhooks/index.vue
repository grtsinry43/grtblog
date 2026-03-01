<script setup lang="ts">
import {
  NButton,
  NCard,
  NDivider,
  NGi,
  NGrid,
  NSpace,
  NStatistic,
  NTabPane,
  NTag,
  NTabs,
  useMessage,
} from 'naive-ui'
import { onMounted, ref } from 'vue'

import { ScrollContainer } from '@/components'
import { useWebhookForm } from './composables/use-webhook-form'

import WebhookTable from './components/WebhookTable.vue'
import WebhookFormDrawer from './components/WebhookFormDrawer.vue'
import WebhookHistoryPanel from './components/WebhookHistoryPanel.vue'
import WebhookHistoryDrawer from './components/WebhookHistoryDrawer.vue'
import WebhookTestModal from './components/WebhookTestModal.vue'

defineOptions({
  name: 'WebhookList',
})

const message = useMessage()
const activeTab = ref('list')

const {
  webhooks,
  eventGroups,
  loading,
  history,
  historyLoading,
  historyPage,
  historyPageSize,
  historyTotal,
  form,
  formDrawerVisible,
  saving,
  currentEventFields,
  testModalVisible,
  testEventName,
  historyDrawerVisible,
  activeHistory,
  listFilters,
  historyFilters,
  eventOptions,
  webhookOptions,
  statusOptions,
  webhookMap,
  formTitle,
  formActionLabel,
  totalWebhooks,
  enabledCount,
  disabledCount,
  historyFailureCount,
  latestHistoryStatus,
  latestHistoryMeta,
  isTestOnly,
  detailStatus,
  validVariables,
  openCreate,
  openEdit,
  addHeaderRow,
  removeHeaderRow,
  fetchWebhooks,
  fetchHistory,
  applyHistoryFilters,
  resetHistoryFilters,
  resetListFilters,
  handleHistoryPageChange,
  handleHistoryPageSizeChange,
  handleSave,
  handleFormatPayload,
  handleDelete,
  openTest,
  handleTest,
  openHistory,
  handleReplay,
  formatHeaders,
  formatBody,
  init,
} = useWebhookForm(message)

onMounted(() => init())
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-4">
    <NCard>
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div class="space-y-1">
          <div class="text-lg font-semibold">Webhook 管理</div>
          <div class="text-xs text-[var(--text-color-3)]">配置事件推送、测试与投递记录。</div>
        </div>
        <NSpace align="center">
          <NButton secondary @click="fetchWebhooks">刷新</NButton>
          <NButton type="primary" @click="openCreate">新建 Webhook</NButton>
        </NSpace>
      </div>
      <NDivider class="my-4" />
      <NGrid cols="1 640:2 900:4" x-gap="16" y-gap="12">
        <NGi><NStatistic label="Webhook 总数" tabular-nums>{{ totalWebhooks }}</NStatistic></NGi>
        <NGi><NStatistic label="启用中" tabular-nums>{{ enabledCount }}</NStatistic></NGi>
        <NGi><NStatistic label="已停用" tabular-nums>{{ disabledCount }}</NStatistic></NGi>
        <NGi>
          <NStatistic label="最近投递">
            <NTag
              size="small"
              :bordered="false"
              :type="latestHistoryStatus.type === 'default' ? undefined : latestHistoryStatus.type"
            >
              {{ latestHistoryStatus.label }}
            </NTag>
          </NStatistic>
          <div class="mt-1 text-xs text-[var(--text-color-3)]">{{ latestHistoryMeta }}</div>
        </NGi>
      </NGrid>
    </NCard>

    <NTabs v-model:value="activeTab" type="line" animated>
      <NTabPane name="list" tab="Webhook 列表">
        <WebhookTable
          :webhooks="webhooks"
          :loading="loading"
          :list-filters="listFilters"
          :event-options="eventOptions"
          :status-options="statusOptions"
          @edit="openEdit"
          @test="openTest"
          @delete="handleDelete"
          @reset-filters="resetListFilters"
        />
      </NTabPane>

      <NTabPane name="history" tab="投递历史">
        <WebhookHistoryPanel
          :history="history"
          :history-loading="historyLoading"
          :history-page="historyPage"
          :history-page-size="historyPageSize"
          :history-total="historyTotal"
          :history-failure-count="historyFailureCount"
          :webhook-map="webhookMap"
          :history-filters="historyFilters"
          :is-test-only="isTestOnly"
          :webhook-options="webhookOptions"
          :event-options="eventOptions"
          @update:history-page="handleHistoryPageChange"
          @update:history-page-size="handleHistoryPageSizeChange"
          @update:is-test-only="isTestOnly = $event"
          @apply-filters="applyHistoryFilters"
          @reset-filters="resetHistoryFilters"
          @refresh="fetchHistory"
          @view-detail="openHistory"
          @replay="handleReplay"
        />
      </NTabPane>
    </NTabs>

    <WebhookFormDrawer
      :visible="formDrawerVisible"
      :title="formTitle"
      :action-label="formActionLabel"
      :saving="saving"
      :form="form"
      :event-groups="eventGroups"
      :valid-variables="validVariables"
      @update:visible="formDrawerVisible = $event"
      @save="handleSave"
      @format-payload="handleFormatPayload"
      @add-header="addHeaderRow"
      @remove-header="removeHeaderRow"
    />

    <WebhookTestModal
      :visible="testModalVisible"
      :test-event-name="testEventName"
      :event-options="eventOptions"
      @update:visible="testModalVisible = $event"
      @update:test-event-name="testEventName = $event"
      @confirm="handleTest"
    />

    <WebhookHistoryDrawer
      :visible="historyDrawerVisible"
      :active-history="activeHistory"
      :detail-status="detailStatus"
      :webhook-map="webhookMap"
      :format-headers="formatHeaders"
      :format-body="formatBody"
      @update:visible="historyDrawerVisible = $event"
    />
  </ScrollContainer>
</template>

<style scoped>
.template-code {
  font-family: 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,
    'Liberation Mono', 'Courier New', monospace;
}
</style>
