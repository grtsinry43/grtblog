import { ref } from 'vue'

import { rewriteContentStream } from '@/services/ai'

import type { EditorView } from '@codemirror/view'

export interface DiffLine {
  type: 'added' | 'removed' | 'unchanged'
  text: string
}

/** LCS-based line-level diff, O(mn), no external dependencies */
export function computeLineDiff(oldText: string, newText: string): DiffLine[] {
  const oldLines = oldText.split('\n')
  const newLines = newText.split('\n')
  const m = oldLines.length
  const n = newLines.length

  // Build LCS table
  const dp: number[][] = Array.from({ length: m + 1 }, () => Array(n + 1).fill(0))
  for (let i = 1; i <= m; i++) {
    for (let j = 1; j <= n; j++) {
      dp[i]![j] = oldLines[i - 1] === newLines[j - 1]
        ? dp[i - 1]![j - 1]! + 1
        : Math.max(dp[i - 1]![j]!, dp[i]![j - 1]!)
    }
  }

  // Backtrack to produce diff
  const result: DiffLine[] = []
  let i = m
  let j = n
  while (i > 0 || j > 0) {
    if (i > 0 && j > 0 && oldLines[i - 1] === newLines[j - 1]) {
      result.push({ type: 'unchanged', text: oldLines[i - 1]! })
      i--
      j--
    } else if (j > 0 && (i === 0 || dp[i]![j - 1]! >= dp[i - 1]![j]!)) {
      result.push({ type: 'added', text: newLines[j - 1]! })
      j--
    } else {
      result.push({ type: 'removed', text: oldLines[i - 1]! })
      i--
    }
  }

  return result.reverse()
}

export function useAIToolbar(getView: () => EditorView | undefined) {
  const visible = ref(false)
  const instruction = ref('')
  const loading = ref(false)
  const resultContent = ref('')
  const showResult = ref(false)
  const selectionRange = ref<{ from: number; to: number } | null>(null)
  const originalContent = ref('')

  // 记录最近一次非空选区，供 slash 命令触发时恢复
  let lastNonEmptySelection: { from: number; to: number } | null = null

  /** 由 MarkdownEditor 的 onViewUpdate 调用，持续追踪选区 */
  function trackSelection(from: number, to: number) {
    if (from !== to) {
      lastNonEmptySelection = { from, to }
    }
  }

  function open() {
    const view = getView()
    if (!view) return
    const { from, to } = view.state.selection.main
    if (from !== to) {
      // 当前有选区：直接使用
      selectionRange.value = { from, to }
    } else if (lastNonEmptySelection) {
      // slash 命令会先替换选区再触发 open，此时选区已折叠，
      // 但 lastNonEmptySelection 仍保留了替换前的范围。
      // 需要校验范围仍在文档内（slash 删除了 "/" 可能导致偏移）
      const docLen = view.state.doc.length
      const savedFrom = Math.min(lastNonEmptySelection.from, docLen)
      const savedTo = Math.min(lastNonEmptySelection.to, docLen)
      if (savedFrom !== savedTo) {
        selectionRange.value = { from: savedFrom, to: savedTo }
      } else {
        selectionRange.value = { from: 0, to: docLen }
      }
    } else {
      // 无任何选区记录：改写全文
      selectionRange.value = { from: 0, to: view.state.doc.length }
    }
    // 用完即清，避免下次误用旧选区
    lastNonEmptySelection = null
    visible.value = true
    instruction.value = ''
    resultContent.value = ''
    showResult.value = false
  }

  async function execute() {
    const view = getView()
    if (!view || !selectionRange.value) return
    const { from, to } = selectionRange.value
    const content = view.state.doc.sliceString(from, to)
    if (!content.trim()) return
    originalContent.value = content
    loading.value = true
    showResult.value = true
    resultContent.value = ''
    try {
      await rewriteContentStream(content, instruction.value, (chunk) => {
        resultContent.value += chunk
      })
    } finally {
      loading.value = false
    }
  }

  function accept() {
    const view = getView()
    if (!view || !selectionRange.value) return
    const { from, to } = selectionRange.value
    view.dispatch({ changes: { from, to, insert: resultContent.value } })
    close()
  }

  function reject() {
    close()
  }

  function close() {
    visible.value = false
    resultContent.value = ''
    showResult.value = false
    selectionRange.value = null
    originalContent.value = ''
  }

  return { visible, instruction, loading, resultContent, showResult, originalContent, trackSelection, open, execute, accept, reject, close }
}
