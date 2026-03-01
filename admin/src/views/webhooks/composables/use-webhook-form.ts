import { computed, reactive, ref, watch } from 'vue'
import { formatTemplateJson } from '@/composables/template-editor/json-lint'
import { formatDate } from '@/utils/format'
import {
  createWebhook,
  deleteWebhook,
  listWebhooks,
  listWebhookHistory,
  replayWebhookHistory,
  testWebhook,
  updateWebhook,
} from '@/services/webhooks'
import { listEvents, getEventCatalogItem } from '@/services/events'

import type { SelectOption } from 'naive-ui'
import type { WebhookHistoryItem, WebhookItem } from '@/services/webhooks'
import type { AdminEventGroupResp, AdminEventFieldResp } from '@/services/events'

export type HeaderRow = {
  key: string
  value: string
}

export type StatusTagType = 'default' | 'success' | 'error'

export function useWebhookForm(message: { error: (msg: string) => void; success: (msg: string) => void }) {
  // --- Core data ---
  const webhooks = ref<WebhookItem[]>([])
  const eventGroups = ref<AdminEventGroupResp[]>([])
  const loading = ref(false)

  // --- History ---
  const history = ref<WebhookHistoryItem[]>([])
  const historyLoading = ref(false)
  const historyPage = ref(1)
  const historyPageSize = ref(10)
  const historyTotal = ref(0)

  // --- Form drawer ---
  const formDrawerVisible = ref(false)
  const saving = ref(false)
  const editingWebhook = ref<WebhookItem | null>(null)

  const form = reactive({
    name: '',
    url: '',
    events: [] as string[],
    headers: [] as HeaderRow[],
    payloadTemplate: '',
    isEnabled: true,
  })

  const currentEventFields = ref<AdminEventFieldResp[]>([])

  watch(
    () => form.events,
    async (newEvents) => {
      if (newEvents && newEvents.length > 0) {
        const lastEvent = newEvents[newEvents.length - 1]
        if (!lastEvent) {
          currentEventFields.value = []
          return
        }
        try {
          const item = await getEventCatalogItem(lastEvent)
          currentEventFields.value = item.fields
        } catch (e) {
          console.error(e)
        }
      } else {
        currentEventFields.value = []
      }
    },
    { deep: true },
  )

  // --- Test modal ---
  const testModalVisible = ref(false)
  const testingWebhook = ref<WebhookItem | null>(null)
  const testEventName = ref<string | null>(null)

  // --- History detail drawer ---
  const historyDrawerVisible = ref(false)
  const activeHistory = ref<WebhookHistoryItem | null>(null)

  // --- Filters ---
  const listFilters = reactive({
    keyword: '',
    status: 'all' as 'all' | 'enabled' | 'disabled',
    event: null as string | null,
  })

  const historyFilters = reactive({
    webhookId: null as number | null,
    eventName: null as string | null,
    isTest: null as boolean | null,
  })

  // --- Computed ---
  const eventOptions = computed<SelectOption[]>(() =>
    eventGroups.value.map((group) => ({
      type: 'group',
      label: group.category.toUpperCase(),
      key: group.category,
      children: group.events.map((e) => ({
        label: e,
        value: e,
      })),
    })),
  )

  const webhookOptions = computed<SelectOption[]>(() =>
    webhooks.value.map((item) => ({
      label: item.name,
      value: item.id,
    })),
  )

  const statusOptions: SelectOption[] = [
    { label: '全部状态', value: 'all' },
    { label: '启用中', value: 'enabled' },
    { label: '已停用', value: 'disabled' },
  ]

  const webhookMap = computed(() => new Map(webhooks.value.map((item) => [item.id, item.name])))

  const formTitle = computed(() => (editingWebhook.value ? '编辑 Webhook' : '新建 Webhook'))
  const formActionLabel = computed(() => (editingWebhook.value ? '保存' : '创建'))

  const totalWebhooks = computed(() => webhooks.value.length)
  const enabledCount = computed(() => webhooks.value.filter((item) => item.isEnabled).length)
  const disabledCount = computed(() => totalWebhooks.value - enabledCount.value)
  const historyFailureCount = computed(
    () => history.value.filter((item) => item.responseStatus < 200 || item.responseStatus >= 300).length,
  )

  const latestHistory = computed(() => history.value[0] ?? null)
  const latestHistoryStatus = computed<{ label: string; type: StatusTagType }>(() => {
    const entry = latestHistory.value
    if (!entry) return { label: '暂无', type: 'default' as const }
    const success = entry.responseStatus >= 200 && entry.responseStatus < 300
    return { label: success ? '成功' : '失败', type: success ? 'success' : 'error' }
  })

  const latestHistoryMeta = computed(() => {
    const entry = latestHistory.value
    if (!entry) return '暂无投递记录'
    const hookName = webhookMap.value.get(entry.webhookId) || `#${entry.webhookId}`
    return `${hookName} · ${formatDate(entry.createdAt)}`
  })

  const isTestOnly = computed({
    get: () => historyFilters.isTest === true,
    set: (value) => { historyFilters.isTest = value ? true : null },
  })

  const detailStatus = computed<{ label: string; type: StatusTagType }>(() => {
    const entry = activeHistory.value
    if (!entry) return { label: '-', type: 'default' as const }
    const status = entry.responseStatus
    if (!status) return { label: '未知', type: 'default' as const }
    const success = status >= 200 && status < 300
    return { label: success ? `成功 ${status}` : `失败 ${status}`, type: success ? 'success' : 'error' }
  })

  const validVariables = computed(() => {
    const vars = ['eventName', 'OccurredAt']
    currentEventFields.value.forEach((f) => vars.push(f.name))
    return vars
  })

  const filteredWebhooks = computed(() => {
    const keyword = listFilters.keyword.trim().toLowerCase()
    return webhooks.value.filter((item) => {
      if (listFilters.status === 'enabled' && !item.isEnabled) return false
      if (listFilters.status === 'disabled' && item.isEnabled) return false
      if (listFilters.event && !item.events?.includes(listFilters.event)) return false
      if (!keyword) return true
      return item.name.toLowerCase().includes(keyword) || item.url.toLowerCase().includes(keyword)
    })
  })

  // --- Actions ---
  function resetForm() {
    form.name = ''
    form.url = ''
    form.events = []
    form.headers = []
    form.payloadTemplate = ''
    form.isEnabled = true
  }

  function ensureHeaderRow() {
    if (form.headers.length === 0) {
      form.headers.push({ key: '', value: '' })
    }
  }

  function openCreate() {
    editingWebhook.value = null
    resetForm()
    ensureHeaderRow()
    formDrawerVisible.value = true
    currentEventFields.value = []
  }

  function openEdit(item: WebhookItem) {
    editingWebhook.value = item
    form.name = item.name
    form.url = item.url
    form.events = [...(item.events || [])]
    form.payloadTemplate = item.payloadTemplate || ''
    form.isEnabled = item.isEnabled
    form.headers = Object.entries(item.headers || {}).map(([key, value]) => ({ key, value }))
    ensureHeaderRow()
    formDrawerVisible.value = true
  }

  function addHeaderRow() {
    form.headers.push({ key: '', value: '' })
  }

  function removeHeaderRow(index: number) {
    form.headers.splice(index, 1)
    ensureHeaderRow()
  }

  function buildHeaderPayload() {
    const headers: Record<string, string> = {}
    form.headers.forEach((row) => {
      const key = row.key.trim()
      if (!key) return
      headers[key] = row.value
    })
    return headers
  }

  async function fetchWebhooks() {
    loading.value = true
    try {
      webhooks.value = await listWebhooks()
    } finally {
      loading.value = false
    }
  }

  async function fetchEvents() {
    const { groups } = await listEvents('webhook')
    eventGroups.value = groups
  }

  async function fetchHistory() {
    historyLoading.value = true
    try {
      const response = await listWebhookHistory({
        page: historyPage.value,
        pageSize: historyPageSize.value,
        webhookId: historyFilters.webhookId ?? undefined,
        eventName: historyFilters.eventName ?? undefined,
        isTest: historyFilters.isTest ?? undefined,
      })
      history.value = response.items
      historyTotal.value = response.total
    } finally {
      historyLoading.value = false
    }
  }

  function applyHistoryFilters() {
    historyPage.value = 1
    fetchHistory()
  }

  function resetHistoryFilters() {
    historyFilters.webhookId = null
    historyFilters.eventName = null
    historyFilters.isTest = null
    historyPage.value = 1
    fetchHistory()
  }

  function resetListFilters() {
    listFilters.keyword = ''
    listFilters.status = 'all'
    listFilters.event = null
  }

  function handleHistoryPageChange(value: number) {
    historyPage.value = value
    fetchHistory()
  }

  function handleHistoryPageSizeChange(value: number) {
    historyPageSize.value = value
    historyPage.value = 1
    fetchHistory()
  }

  async function handleSave() {
    if (!form.name.trim()) { message.error('请填写名称'); return }
    if (!form.url.trim()) { message.error('请填写 URL'); return }
    if (form.events.length === 0) { message.error('请选择订阅事件'); return }

    saving.value = true
    try {
      const payload = {
        name: form.name.trim(),
        url: form.url.trim(),
        events: form.events,
        headers: buildHeaderPayload(),
        payloadTemplate: form.payloadTemplate,
        isEnabled: form.isEnabled,
      }
      if (editingWebhook.value) {
        await updateWebhook(editingWebhook.value.id, payload)
      } else {
        await createWebhook(payload)
      }
      formDrawerVisible.value = false
      await fetchWebhooks()
    } finally {
      saving.value = false
    }
  }

  function handleFormatPayload() {
    try {
      form.payloadTemplate = formatTemplateJson(form.payloadTemplate)
      message.success('已格式化')
    } catch (err) {
      const reason = err instanceof Error ? err.message : 'JSON 格式不正确'
      message.error(`格式化失败：${reason}`)
    }
  }

  async function handleDelete(item: WebhookItem) {
    await deleteWebhook(item.id)
    fetchWebhooks()
  }

  function openTest(item: WebhookItem) {
    testingWebhook.value = item
    testEventName.value = item.events?.[0] || null
    testModalVisible.value = true
  }

  async function handleTest() {
    if (!testingWebhook.value) return
    await testWebhook(testingWebhook.value.id, testEventName.value)
    testModalVisible.value = false
    fetchHistory()
  }

  function openHistory(item: WebhookHistoryItem) {
    activeHistory.value = item
    historyDrawerVisible.value = true
  }

  async function handleReplay(item: WebhookHistoryItem) {
    await replayWebhookHistory(item.id)
    fetchHistory()
  }

  function formatHeaders(headers?: Record<string, string>) {
    if (!headers || Object.keys(headers).length === 0) return '-'
    return Object.entries(headers)
      .map(([key, value]) => `${key}: ${value}`)
      .join('\n')
  }

  function formatBody(body?: string) {
    if (!body) return '-'
    try {
      const parsed = JSON.parse(body)
      return JSON.stringify(parsed, null, 2)
    } catch {
      return body
    }
  }

  async function init() {
    await Promise.all([fetchWebhooks(), fetchEvents(), fetchHistory()])
  }

  return {
    // Data
    webhooks,
    eventGroups,
    loading,
    history,
    historyLoading,
    historyPage,
    historyPageSize,
    historyTotal,
    // Form
    form,
    formDrawerVisible,
    saving,
    editingWebhook,
    currentEventFields,
    // Test
    testModalVisible,
    testingWebhook,
    testEventName,
    // History detail
    historyDrawerVisible,
    activeHistory,
    // Filters
    listFilters,
    historyFilters,
    // Computed
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
    filteredWebhooks,
    // Actions
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
  }
}

