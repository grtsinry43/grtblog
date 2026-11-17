import { onUnmounted, ref, watch, type Ref } from 'vue'

import { checkWriting, completeTextStream } from '@/services/ai'

import { clearGhostText, ghostTextField, setGhostText } from './extensions/ai-ghost-text'
import {
  clearLintMatches,
  flashRangeEffect,
  lintMatchesField,
  removeLintMatch,
  setLintMatches,
  type LintMatch,
} from './extensions/ai-inline-lint'

import type { UpdateHook } from './use-codemirror'
import type { EditorView, ViewUpdate } from '@codemirror/view'

export type RealtimeAIStatus = 'idle' | 'thinking' | 'suggesting'

export interface ActiveLintPopover {
  match: LintMatch
  pos: { top: number; left: number }
}

interface UseRealtimeAIOptions {
  view: Ref<EditorView | undefined>
  onViewUpdate: (callback: UpdateHook) => () => void
  enabled: Ref<boolean>
  enableCompletion?: Ref<boolean>
  enableLint?: Ref<boolean>
  getTitle?: () => string | undefined
}

function isInsideCodeFence(docText: string, pos: number): boolean {
  const before = docText.slice(0, pos)
  const fences = before.match(/^```/gm)
  return !!fences && fences.length % 2 === 1
}

export function useRealtimeAI(options: UseRealtimeAIOptions) {
  const { view, onViewUpdate, enabled, enableCompletion, enableLint, getTitle } = options
  const isCompletionOn = () => enabled.value && (enableCompletion ? enableCompletion.value : true)
  const isLintOn = () => enabled.value && (enableLint ? enableLint.value : true)

  const status = ref<RealtimeAIStatus>('idle')
  const activeLint = ref<ActiveLintPopover | null>(null)

  let completionTimer: ReturnType<typeof setTimeout> | null = null
  let lintTimer: ReturnType<typeof setTimeout> | null = null
  let completionAbort: AbortController | null = null
  let lintAbort: AbortController | null = null
  let flashTimer: ReturnType<typeof setTimeout> | null = null

  const checkedParagraphs = new Set<string>()
  const ignoredOriginals = new Set<string>()

  function clearTimers() {
    if (completionTimer) {
      clearTimeout(completionTimer)
      completionTimer = null
    }
    if (lintTimer) {
      clearTimeout(lintTimer)
      lintTimer = null
    }
  }

  function abortRequests() {
    if (completionAbort) {
      completionAbort.abort()
      completionAbort = null
    }
    if (lintAbort) {
      lintAbort.abort()
      lintAbort = null
    }
  }

  function resetAll() {
    clearTimers()
    abortRequests()
    status.value = 'idle'
    activeLint.value = null
    const currentView = view.value
    if (currentView) {
      currentView.dispatch({
        effects: [clearGhostText.of(null), clearLintMatches.of(null)],
      })
    }
  }

  async function runCompletion() {
    const currentView = view.value
    if (!currentView || !isCompletionOn() || !currentView.hasFocus) return

    const { state } = currentView
    const sel = state.selection.main
    if (!sel.empty) return

    const line = state.doc.lineAt(sel.head)
    const afterCursorOnLine = line.text.slice(sel.head - line.from)
    if (afterCursorOnLine.trim().length > 0) return

    const fullText = state.doc.toString()
    if (isInsideCodeFence(fullText, sel.head)) return

    const prefix = state.doc.sliceString(Math.max(0, sel.head - 1200), sel.head)
    if (prefix.trim().length < 12) return

    const suffix = state.doc.sliceString(sel.head, Math.min(state.doc.length, sel.head + 400))

    completionAbort?.abort()
    completionAbort = new AbortController()
    const signal = completionAbort.signal
    const anchorPos = sel.head

    status.value = 'thinking'
    let accumulated = ''

    try {
      await completeTextStream(
        {
          prefix,
          suffix: suffix.trim() ? suffix : undefined,
          title: getTitle?.(),
        },
        (chunk) => {
          if (signal.aborted) return
          const v = view.value
          if (!v || v.state.selection.main.head !== anchorPos) return

          const prevLen = accumulated.length
          accumulated += chunk
          // 去掉模型偶尔带出的前导换行过多或包裹引号
          const cleaned = accumulated.replace(/^\n{2,}/, '\n')
          if (!cleaned) return

          status.value = 'suggesting'
          v.dispatch({
            effects: setGhostText.of({
              pos: anchorPos,
              text: cleaned,
              animFrom: Math.min(prevLen, cleaned.length),
            }),
          })
        },
        signal,
      )
    } catch {
      // AbortError 或网络波动静默处理，避免打断写作心流
    } finally {
      if (!signal.aborted) {
        const v = view.value
        const hasGhost = !!v?.state.field(ghostTextField, false)
        status.value = hasGhost ? 'suggesting' : 'idle'
      }
    }
  }

  async function runInlineLint(centerPos: number) {
    const currentView = view.value
    if (!currentView || !isLintOn()) return

    const { doc } = currentView.state
    const centerLine = doc.lineAt(Math.min(centerPos, doc.length))
    const startLineNum = Math.max(1, centerLine.number - 1)
    const endLineNum = Math.min(doc.lines, centerLine.number + 1)

    const windowFrom = doc.line(startLineNum).from
    const windowTo = doc.line(endLineNum).to
    const snippet = doc.sliceString(windowFrom, windowTo)

    if (snippet.trim().length < 10) return
    if (isInsideCodeFence(doc.toString(), windowFrom)) return
    if (checkedParagraphs.has(snippet)) return

    lintAbort?.abort()
    lintAbort = new AbortController()
    const signal = lintAbort.signal

    try {
      const resp = await checkWriting(snippet, signal)
      if (signal.aborted) return

      if (checkedParagraphs.size > 30) {
        checkedParagraphs.clear()
      }
      checkedParagraphs.add(snippet)

      const v = view.value
      if (!v) return
      const latestWindowText = v.state.doc.sliceString(
        windowFrom,
        Math.min(v.state.doc.length, windowTo),
      )
      if (latestWindowText !== snippet) return

      const existing = v.state.field(lintMatchesField, false) || []
      const outsideWindow = existing.filter((m) => m.to <= windowFrom || m.from >= windowTo)

      const newMatches: LintMatch[] = []
      for (const issue of resp.issues || []) {
        if (ignoredOriginals.has(issue.original)) continue
        const idx = snippet.indexOf(issue.original)
        if (idx === -1) continue
        newMatches.push({
          id: `${windowFrom + idx}:${issue.original}`,
          from: windowFrom + idx,
          to: windowFrom + idx + issue.original.length,
          original: issue.original,
          suggestion: issue.suggestion,
          reason: issue.reason,
          severity: issue.severity,
        })
      }

      v.dispatch({
        effects: setLintMatches.of([...outsideWindow, ...newMatches]),
      })
    } catch {
      // 静默失败，不弹出错误提示打扰写作
    }
  }

  function updateActiveLintAtPos(pos: number) {
    const currentView = view.value
    if (!currentView) {
      activeLint.value = null
      return
    }
    const matches = currentView.state.field(lintMatchesField, false) || []
    const hit = matches.find((m) => pos >= m.from && pos <= m.to)
    if (!hit) {
      activeLint.value = null
      return
    }
    const coords = currentView.coordsAtPos(hit.from)
    if (!coords) {
      activeLint.value = null
      return
    }
    activeLint.value = {
      match: hit,
      pos: {
        top: coords.top - 8,
        left: coords.left,
      },
    }
  }

  function acceptLint(match: LintMatch) {
    const currentView = view.value
    if (!currentView) return

    const newTo = match.from + match.suggestion.length
    currentView.dispatch({
      changes: { from: match.from, to: match.to, insert: match.suggestion },
      effects: [
        removeLintMatch.of(match.id),
        flashRangeEffect.of({ from: match.from, to: newTo }),
      ],
    })
    activeLint.value = null
    currentView.focus()

    if (flashTimer) clearTimeout(flashTimer)
    flashTimer = setTimeout(() => {
      view.value?.dispatch({ effects: flashRangeEffect.of(null) })
    }, 560)
  }

  function dismissLint(match: LintMatch) {
    ignoredOriginals.add(match.original)
    view.value?.dispatch({ effects: removeLintMatch.of(match.id) })
    activeLint.value = null
    view.value?.focus()
  }

  function closeLint() {
    activeLint.value = null
  }

  const stopUpdateListener = onViewUpdate((update: ViewUpdate) => {
    if (!enabled.value) return

    if (update.docChanged) {
      activeLint.value = null
      completionAbort?.abort()
      status.value = 'idle'

      if (completionTimer) clearTimeout(completionTimer)
      if (isCompletionOn()) {
        completionTimer = setTimeout(() => {
          void runCompletion()
        }, 750)
      }

      const head = update.state.selection.main.head
      if (lintTimer) clearTimeout(lintTimer)
      if (isLintOn()) {
        lintTimer = setTimeout(() => {
          void runInlineLint(head)
        }, 1800)
      }
    } else if (update.selectionSet) {
      const sel = update.state.selection.main
      const hasGhost = !!update.state.field(ghostTextField, false)
      if (!hasGhost && status.value === 'suggesting') {
        status.value = 'idle'
      }
      if (sel.empty) {
        updateActiveLintAtPos(sel.head)
        // 光标停在行尾或空行时，停顿 1.1s 也可触发幽灵续写
        if (!hasGhost && isCompletionOn()) {
          if (completionTimer) clearTimeout(completionTimer)
          completionTimer = setTimeout(() => {
            void runCompletion()
          }, 1100)
        }
      } else {
        activeLint.value = null
      }
    }
  })

  watch(enabled, (isEnabled) => {
    if (!isEnabled) {
      resetAll()
    }
  })

  if (enableCompletion) {
    watch(enableCompletion, (on) => {
      if (!on) {
        if (completionTimer) clearTimeout(completionTimer)
        completionAbort?.abort()
        status.value = 'idle'
        view.value?.dispatch({ effects: clearGhostText.of(null) })
      }
    })
  }

  if (enableLint) {
    watch(enableLint, (on) => {
      if (!on) {
        if (lintTimer) clearTimeout(lintTimer)
        lintAbort?.abort()
        activeLint.value = null
        view.value?.dispatch({ effects: clearLintMatches.of(null) })
      }
    })
  }

  onUnmounted(() => {
    stopUpdateListener()
    clearTimers()
    abortRequests()
    if (flashTimer) clearTimeout(flashTimer)
  })

  return {
    status,
    activeLint,
    acceptLint,
    dismissLint,
    closeLint,
  }
}
