<script setup lang="ts">
import {
  NButton,
  NCard,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSpace,
  NSwitch,
  useThemeVars,
  useMessage,
} from 'naive-ui'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

import EditorAIToolbar from '@/components/markdown-editor/EditorAIToolbar.vue'
import EditorCitationPicker from '@/components/markdown-editor/EditorCitationPicker.vue'
import EditorFloatingMenu from '@/components/markdown-editor/EditorFloatingMenu.vue'
import EditorLintPopover from '@/components/markdown-editor/EditorLintPopover.vue'
import EditorMarginNotePopover from '@/components/markdown-editor/EditorMarginNotePopover.vue'
import EditorMentionPicker from '@/components/markdown-editor/EditorMentionPicker.vue'
import { type MarginNote } from '@/composables/markdown-editor/extensions/ai-margin-notes'
import { useAIToolbar } from '@/composables/markdown-editor/use-ai-toolbar'
import { useCitationPicker } from '@/composables/markdown-editor/use-citation-picker'
import { useCodeMirror } from '@/composables/markdown-editor/use-codemirror'
import { useComponentInserter } from '@/composables/markdown-editor/use-component-inserter.ts'
import { useFloatingMenu } from '@/composables/markdown-editor/use-floating-menu'
import { useMentionPicker } from '@/composables/markdown-editor/use-mention-picker'
import {
  useRealtimeAI,
  type CompletionPace,
  type RealtimeAIStatus,
} from '@/composables/markdown-editor/use-realtime-ai'
import { uploadFile } from '@/services/uploads'
import { cah } from '@/utils/chromaHelper'

const props = defineProps<{
  modelValue: string
  realtimeAi?: boolean
  realtimeAiCompletion?: boolean
  realtimeAiCompletionPace?: CompletionPace
  realtimeAiMarginNotes?: boolean
  realtimeAiLint?: boolean
  articleTitle?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'ai-status-change', value: RealtimeAIStatus): void
  (
    e: 'cursor-change',
    value: {
      line: number
      column: number
      selectionChars: number
      selectionTotal: number
    },
  ): void
}>()

const editorRef = ref<HTMLElement>()
const themeVars = useThemeVars()
const message = useMessage()

// 样式定义
const editorStyle = computed(() => ({
  '--cm-bg': themeVars.value.cardColor,
  '--cm-fg': themeVars.value.textColor1,
  '--cm-fg-muted': themeVars.value.textColor3,
  '--cm-border': themeVars.value.borderColor,
  '--cm-gutter-fg': themeVars.value.textColor3,
  '--cm-gutter-bg': themeVars.value.cardColor,
  '--cm-selection': cah(themeVars.value.primaryColor, 0.22),
  '--cm-active-line': cah(themeVars.value.primaryColor, 0.02),
  '--cm-cursor': themeVars.value.primaryColor,
  '--cm-tooltip-bg': themeVars.value.popoverColor,
  '--cm-tooltip-border': themeVars.value.borderColor,
  '--cm-tooltip-shadow': themeVars.value.boxShadow2,
  '--cm-tooltip-hover': themeVars.value.buttonColor2Hover,
  '--cm-tooltip-selected': themeVars.value.buttonColor2Pressed,
  '--cm-block-highlight': `color-mix(in srgb, ${themeVars.value.primaryColor} 10%, transparent)`,
  '--cm-accent': themeVars.value.primaryColor,
  '--cm-syntax-keyword': themeVars.value.primaryColor,
  '--cm-syntax-string': themeVars.value.successColor,
  '--cm-syntax-number': themeVars.value.warningColor,
  '--cm-syntax-property': themeVars.value.infoColor,
  '--cm-syntax-function': themeVars.value.infoColor,
  '--cm-syntax-type': themeVars.value.primaryColor,
  '--cm-syntax-tag': themeVars.value.errorColor,
  '--cm-syntax-invalid': themeVars.value.errorColor,
  '--cm-syntax-comment': themeVars.value.textColor3,
  '--cm-syntax-variable': themeVars.value.textColor2,
  '--cm-syntax-operator': themeVars.value.textColor2,
  '--cm-syntax-punctuation': themeVars.value.textColor3,
}))

// 1. 初始化 CodeMirror
const { view, onViewUpdate } = useCodeMirror(editorRef, {
  initialDoc: props.modelValue,
  onChange: (val) => emit('update:modelValue', val),
  // 点击图标时，触发 inserter
  onComponentEdit: (payload) => inserter.open(payload),
  onUploadImage: async (file) => {
    try {
      const result = await uploadFile(file, 'picture')
      return result.publicUrl
    } catch (error) {
      console.error(error)
      message.error('图片上传失败')
      throw error
    }
  },
})

// 2. 挂载组件插入逻辑
const inserter = useComponentInserter(view)

// 3. 挂载浮动菜单逻辑
const { isVisible, menuPos, activeFormats, selectedText, executeCommand, formatPangu } =
  useFloatingMenu({
    view,
    onViewUpdate,
  })

// 4. AI 改写工具栏
const aiToolbar = useAIToolbar(() => view.value)

// 4.5 实时 AI 帮写（幽灵续写 + 纸边眉批 + 错字校对）
const realtimeAIEnabled = computed(() => !!props.realtimeAi)
const realtimeAICompletion = computed(() => props.realtimeAiCompletion ?? true)
const realtimeAICompletionPace = computed<CompletionPace>(
  () => props.realtimeAiCompletionPace ?? 'restrained',
)
const realtimeAIMarginNotes = computed(() => props.realtimeAiMarginNotes ?? true)
const realtimeAILint = computed(() => props.realtimeAiLint ?? true)
const realtimeAI = useRealtimeAI({
  view,
  onViewUpdate,
  enabled: realtimeAIEnabled,
  enableCompletion: realtimeAICompletion,
  completionPace: realtimeAICompletionPace,
  enableMarginNotes: realtimeAIMarginNotes,
  enableLint: realtimeAILint,
  getTitle: () => props.articleTitle,
})

watch(
  () => realtimeAI.status.value,
  (status) => {
    emit('ai-status-change', status)
  },
  { immediate: true },
)

// 5. Federation pickers
const mentionPicker = useMentionPicker(view)
const citationPicker = useCitationPicker(view)

function onAIRewriteTrigger() {
  aiToolbar.open()
}

function onPanguFormatTrigger() {
  const changed = formatPangu(true)
  if (changed) {
    message.success('已完成全文盘古格式化')
  } else {
    message.info('当前排版已很规范，无需调整')
  }
}

function handlePanguFormat() {
  const changed = formatPangu(false)
  if (changed) {
    message.success('已完成盘古格式化')
  } else {
    message.info('选中内容排版已很规范，无需调整')
  }
}

function onMentionTrigger() {
  mentionPicker.open()
}

function onCitationTrigger() {
  citationPicker.open()
}

async function handleAIExecute(customInstruction?: string) {
  try {
    await aiToolbar.execute(customInstruction)
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : 'AI 改写失败')
  }
}

function handleFloatingAIAction(instruction?: string) {
  if (instruction) {
    aiToolbar.openAndExecute(instruction)
  } else {
    aiToolbar.open()
  }
}

function handleMarginNoteRewrite(note: MarginNote) {
  const hint = note.suggestion
    ? `参考这条眉批进行微调（保持原意与个人文风）：${note.comment}。参考方向：${note.suggestion}`
    : `参考这条眉批进行微调（保持原意与个人文风）：${note.comment}`
  realtimeAI.closeMarginNote()
  aiToolbar.openForRange(note.paragraphFrom, note.paragraphTo, hint)
}

onMounted(() => {
  editorRef.value?.addEventListener('ai-rewrite-trigger', onAIRewriteTrigger)
  editorRef.value?.addEventListener('pangu-format-trigger', onPanguFormatTrigger)
  editorRef.value?.addEventListener('federation-mention-trigger', onMentionTrigger)
  editorRef.value?.addEventListener('federation-citation-trigger', onCitationTrigger)
})
onUnmounted(() => {
  editorRef.value?.removeEventListener('ai-rewrite-trigger', onAIRewriteTrigger)
  editorRef.value?.removeEventListener('pangu-format-trigger', onPanguFormatTrigger)
  editorRef.value?.removeEventListener('federation-mention-trigger', onMentionTrigger)
  editorRef.value?.removeEventListener('federation-citation-trigger', onCitationTrigger)
})

// 5. 光标与选择更新事件
onViewUpdate((update) => {
  const pos = update.state.selection.main.head
  const selection = update.state.selection.main

  // 让 AI 工具栏持续追踪选区
  aiToolbar.trackSelection(selection.from, selection.to)

  const selectionText =
    selection.from === selection.to
      ? ''
      : update.state.doc.sliceString(selection.from, selection.to)

  const line = update.state.doc.lineAt(pos)

  emit('cursor-change', {
    line: line.number,
    column: pos - line.from + 1,
    selectionChars: selectionText.replace(/\s/g, '').length,
    selectionTotal: selectionText.length,
  })
})

// 5. 外部 modelValue 变更同步
watch(
  () => props.modelValue,
  (newVal) => {
    if (view.value && view.value.state.doc.toString() !== newVal) {
      view.value.dispatch({
        changes: { from: 0, to: view.value.state.doc.length, insert: newVal },
      })
    }
  },
)

/** 承接选题灵感：在空文档顶部植入注释种子并自动打开 AI 工具栏生成注释大纲 */
function seedAndOpenAIOutline(seedComment: string, instruction: string) {
  const currentView = view.value
  if (!currentView) return

  const currentDoc = currentView.state.doc.toString()
  if (!currentDoc.trim()) {
    currentView.dispatch({
      changes: { from: 0, to: currentView.state.doc.length, insert: seedComment },
      selection: { anchor: seedComment.length },
    })
    aiToolbar.openForRange(0, seedComment.length, instruction)
  } else {
    const insertText = `${seedComment}\n\n`
    currentView.dispatch({
      changes: { from: 0, to: 0, insert: insertText },
      selection: { anchor: seedComment.length },
    })
    aiToolbar.openForRange(0, seedComment.length, instruction)
  }
}

defineExpose({
  seedAndOpenAIOutline,
})
</script>

<template>
  <div class="relative h-full">
    <div
      ref="editorRef"
      class="codemirror-wrapper h-full w-full overflow-visible rounded-md"
      :style="editorStyle"
    ></div>

    <NModal
      v-model:show="inserter.state.show"
      style="width: 520px; max-width: 90vw"
    >
      <NCard
        :title="inserter.componentMeta.value?.label ?? '组件参数'"
        size="small"
      >
        <NForm
          label-placement="left"
          label-width="110px"
        >
          <NFormItem
            v-for="attr in inserter.componentMeta.value?.attrs ?? []"
            :key="attr.key"
            :label="attr.label"
          >
            <NSwitch
              v-if="attr.inputType === 'switch'"
              :value="(inserter.state.attrs[attr.key] ?? attr.defaultValue ?? 'false') === 'true'"
              @update:value="(val) => inserter.updateAttr(attr.key, val)"
            />
            <NInput
              v-else
              :value="inserter.state.attrs[attr.key] ?? attr.defaultValue ?? ''"
              :placeholder="attr.placeholder"
              @update:value="(val) => inserter.updateAttr(attr.key, val)"
            />
          </NFormItem>
          <NFormItem
            v-if="inserter.componentMeta.value?.body"
            :label="inserter.componentMeta.value?.body?.label || '内容'"
          >
            <NInput
              type="textarea"
              :autosize="{ minRows: 4, maxRows: 12 }"
              :value="inserter.state.body"
              :placeholder="inserter.componentMeta.value?.body?.placeholder"
              @update:value="(val) => inserter.updateBody(val)"
            />
          </NFormItem>
        </NForm>
        <NSpace justify="end">
          <NButton @click="inserter.close">关闭</NButton>
        </NSpace>
      </NCard>
    </NModal>

    <EditorFloatingMenu
      :visible="isVisible"
      :pos="menuPos"
      :active-formats="activeFormats"
      :realtime-ai="realtimeAIEnabled"
      :selected-text="selectedText"
      @command="executeCommand"
      @pangu="handlePanguFormat"
      @ai-action="handleFloatingAIAction"
    />

    <EditorLintPopover
      :match="realtimeAI.activeLint.value?.match ?? null"
      :pos="realtimeAI.activeLint.value?.pos ?? null"
      @accept="realtimeAI.acceptLint"
      @dismiss="realtimeAI.dismissLint"
    />

    <EditorMarginNotePopover
      :note="realtimeAI.activeMarginNote.value?.note ?? null"
      :pos="realtimeAI.activeMarginNote.value?.pos ?? null"
      @rewrite="handleMarginNoteRewrite"
      @dismiss="realtimeAI.dismissMarginNote"
      @close="realtimeAI.closeMarginNote"
    />

    <EditorAIToolbar
      :visible="aiToolbar.visible.value"
      :instruction="aiToolbar.instruction.value"
      :loading="aiToolbar.loading.value"
      :result-content="aiToolbar.resultContent.value"
      :show-result="aiToolbar.showResult.value"
      :original-content="aiToolbar.originalContent.value"
      :anchor-pos="aiToolbar.anchorPos.value"
      @update:instruction="(v) => (aiToolbar.instruction.value = v)"
      @execute="handleAIExecute"
      @accept="aiToolbar.accept()"
      @reject="aiToolbar.reject()"
      @close="aiToolbar.close()"
    />

    <EditorMentionPicker
      :show="mentionPicker.state.show"
      :query="mentionPicker.state.query"
      :results="mentionPicker.state.results"
      :loading="mentionPicker.state.loading"
      @update:show="
        (v) => {
          if (!v) mentionPicker.close()
        }
      "
      @search="mentionPicker.search"
      @select="mentionPicker.insert"
      @insertRaw="mentionPicker.insertRaw"
    />

    <EditorCitationPicker
      :show="citationPicker.state.show"
      :step="citationPicker.state.step"
      :url-input="citationPicker.state.urlInput"
      :url-valid="citationPicker.state.urlValid"
      :url-error="citationPicker.state.urlError"
      :instances="citationPicker.state.instances"
      :instances-loading="citationPicker.state.instancesLoading"
      :posts="citationPicker.state.posts"
      :posts-loading="citationPicker.state.postsLoading"
      :search-query="citationPicker.state.searchQuery"
      :page="citationPicker.state.page"
      :total="citationPicker.state.total"
      :page-size="citationPicker.state.pageSize"
      :resolved-u-r-l="citationPicker.state.resolvedURL"
      :resolved-name="citationPicker.state.resolvedName"
      @update:show="
        (v: boolean) => {
          if (!v) citationPicker.close()
        }
      "
      @url-input="citationPicker.onURLInput"
      @submit-u-r-l="citationPicker.submitURL"
      @select-instance="citationPicker.selectInstance"
      @search-posts="citationPicker.searchPosts"
      @go-to-page="citationPicker.goToPage"
      @select="citationPicker.insert"
      @back="citationPicker.back"
      @insert-raw="citationPicker.insertRaw"
    />
  </div>
</template>

<style scoped>
:deep(.cm-editor) {
  height: 100%;
}
:deep(.cm-content) {
  padding-bottom: 2rem;
}
</style>
