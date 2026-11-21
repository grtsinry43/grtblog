import { ref, type Ref } from 'vue'

import { settleRangeEffect } from './extensions/ai-ghost-text'
import { getActiveStyles, toggleFormat, type FormatType } from './utils/ast'
import { formatPanguMarkdown } from './utils/pangu'

import type { UpdateHook } from './use-codemirror'
import type { EditorView, ViewUpdate } from '@codemirror/view'

// 定义入参接口
interface EditorContext {
  view: Ref<EditorView | undefined>
  onViewUpdate: (callback: UpdateHook) => void
}

export function useFloatingMenu({ view, onViewUpdate }: EditorContext) {
  const isVisible = ref(false)
  const menuPos = ref({ top: 0, left: 0 })
  const activeFormats = ref<Set<FormatType>>(new Set())
  const selectedText = ref('')

  onViewUpdate((update: ViewUpdate) => {
    const currentView = update.view
    const { selection } = currentView.state

    // 1. 基础显隐判断
    if (selection.main.empty) {
      isVisible.value = false
      selectedText.value = ''
      return
    }

    // 2. 计算坐标（兼容全选 Cmd/Ctrl+A 时起止位置滚出视口的情况）
    const start = currentView.coordsAtPos(selection.main.from)
    const end = currentView.coordsAtPos(selection.main.to)

    if (start && end) {
      menuPos.value = {
        left: (start.left + end.right) / 2,
        top: start.top - 6,
      }
    } else {
      const editorRect = currentView.dom.getBoundingClientRect()
      const anchor = start ?? end
      menuPos.value = {
        left: anchor ? anchor.left : editorRect.left + editorRect.width / 2,
        top: anchor ? anchor.top - 6 : Math.max(76, editorRect.top + 16),
      }
    }

    // 4. 状态回填 (查询 AST)
    activeFormats.value = getActiveStyles(currentView)
    selectedText.value = currentView.state.doc.sliceString(selection.main.from, selection.main.to)
    isVisible.value = true
  })

  // 执行命令
  const executeCommand = (type: FormatType) => {
    if (view.value) {
      toggleFormat(view.value, type)
      view.value.focus() // 保持焦点
    }
  }

  /** 对当前选区（或全选/全文）执行盘古格式化 */
  const formatPangu = (fullDocument = false): boolean => {
    const currentView = view.value
    if (!currentView) return false

    const { state } = currentView
    const sel = state.selection.main
    const from = fullDocument || sel.empty ? 0 : sel.from
    const to = fullDocument || sel.empty ? state.doc.length : sel.to
    if (from >= to) return false

    const original = state.doc.sliceString(from, to)
    const formatted = formatPanguMarkdown(original)
    if (formatted === original) {
      currentView.focus()
      return false
    }

    const newTo = from + formatted.length
    const settleId = Date.now()
    currentView.dispatch({
      changes: { from, to, insert: formatted },
      selection: fullDocument || sel.empty ? undefined : { anchor: from, head: newTo },
      effects: settleRangeEffect.of({ from, to: newTo, id: settleId }),
      userEvent: 'format.pangu',
    })

    window.setTimeout(() => {
      if (!currentView.dom.isConnected) return
      currentView.dispatch({ effects: settleRangeEffect.of(null) })
    }, 540)

    currentView.focus()
    return true
  }

  return {
    isVisible,
    menuPos,
    activeFormats,
    selectedText,
    executeCommand,
    formatPangu,
  }
}
