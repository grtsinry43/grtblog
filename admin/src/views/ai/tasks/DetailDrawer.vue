<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { NDrawer, NDrawerContent, NDescriptions, NDescriptionsItem, NTag, NCode, NButton } from 'naive-ui'
import { getAITaskLog } from '@/services/ai'
import type { AITaskLog } from '@/services/ai'
import { useWindowSize } from '@vueuse/core'

const props = defineProps<{
  show: boolean
  taskLog?: AITaskLog
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
}>()

const { width } = useWindowSize()
const drawerWidth = computed(() => (width.value < 640 ? '100%' : 600))

const detail = ref<AITaskLog | undefined>(undefined)
const loadingDetail = ref(false)

watch(
  () => [props.show, props.taskLog?.id],
  async () => {
    if (props.show && props.taskLog?.id) {
      loadingDetail.value = true
      try {
        detail.value = await getAITaskLog(props.taskLog.id)
      } catch {
        detail.value = props.taskLog
      } finally {
        loadingDetail.value = false
      }
    }
  },
)

const displayItem = computed(() => detail.value ?? props.taskLog)

function statusTagType(status?: string) {
  switch (status) {
    case 'completed': return 'success'
    case 'failed': return 'error'
    case 'interrupted': return 'warning'
    case 'running': return 'warning'
    case 'pending': return 'info'
    default: return 'default'
  }
}

function statusLabel(status?: string) {
  switch (status) {
    case 'pending': return '等待中'
    case 'running': return '运行中'
    case 'completed': return '已完成'
    case 'failed': return '失败'
    case 'interrupted': return '已中断'
    default: return status ?? '-'
  }
}

function taskTypeLabel(type_?: string) {
  switch (type_) {
    case 'comment_moderation': return '评论审核'
    case 'title_generation': return '标题生成'
    case 'content_rewrite': return '内容改写'
    case 'summary_generation': return '摘要生成'
    default: return type_ ?? '-'
  }
}

function triggerLabel(trigger?: string) {
  switch (trigger) {
    case 'manual': return '手动'
    case 'auto': return '自动'
    default: return trigger ?? '-'
  }
}

function formatDuration(ms?: number) {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(1)}s`
}

function handleClose() {
  emit('update:show', false)
}
</script>

<template>
  <NDrawer :show="show" :width="drawerWidth" @update:show="(val) => emit('update:show', val)">
    <NDrawerContent title="AI 任务详情" closable>
      <div v-if="displayItem" class="space-y-6">
        <NDescriptions bordered :column="1" label-placement="left" title="基本信息">
          <NDescriptionsItem label="ID">{{ displayItem.id }}</NDescriptionsItem>
          <NDescriptionsItem label="任务类型">
            <NTag type="info" size="small">{{ taskTypeLabel(displayItem.taskType) }}</NTag>
          </NDescriptionsItem>
          <NDescriptionsItem label="模型">{{ displayItem.modelName || '-' }}</NDescriptionsItem>
          <NDescriptionsItem label="提供商">{{ displayItem.providerName || '-' }}</NDescriptionsItem>
          <NDescriptionsItem label="状态">
            <NTag :type="statusTagType(displayItem.status)" size="small">{{ statusLabel(displayItem.status) }}</NTag>
          </NDescriptionsItem>
          <NDescriptionsItem label="触发来源">
            <NTag :type="displayItem.triggerSource === 'auto' ? 'warning' : 'default'" size="small">{{ triggerLabel(displayItem.triggerSource) }}</NTag>
          </NDescriptionsItem>
          <NDescriptionsItem label="耗时">{{ formatDuration(displayItem.durationMs) }}</NDescriptionsItem>
          <NDescriptionsItem label="创建时间">{{ new Date(displayItem.createdAt).toLocaleString() }}</NDescriptionsItem>
        </NDescriptions>

        <div v-if="displayItem.inputText">
          <h3 class="mb-2 font-bold">输入文本</h3>
          <NCode :code="displayItem.inputText" language="text" word-wrap />
        </div>

        <div v-if="displayItem.outputText">
          <h3 class="mb-2 font-bold">输出文本</h3>
          <NCode :code="displayItem.outputText" language="text" word-wrap />
        </div>

        <div v-if="displayItem.errorMessage">
          <h3 class="mb-2 font-bold text-red-500">错误信息</h3>
          <NCode :code="displayItem.errorMessage" language="text" word-wrap />
        </div>
      </div>
      <template #footer>
        <NButton @click="handleClose">关闭</NButton>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>
