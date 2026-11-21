import { onUnmounted, ref, watch, type Ref } from 'vue'

import { annotateParagraph, checkWriting, completeTextStream } from '@/services/ai'

import { clearGhostText, ghostTextField, setGhostText } from './extensions/ai-ghost-text'
import {
  clearLintMatches,
  flashRangeEffect,
  lintMatchesField,
  removeLintMatch,
  setLintMatches,
  type LintMatch,
} from './extensions/ai-inline-lint'
import {
  clearMarginNotes,
  marginNotesField,
  removeMarginNote,
  upsertMarginNote,
  type MarginNote,
} from './extensions/ai-margin-notes'

import type { UpdateHook } from './use-codemirror'
import type { EditorView, ViewUpdate } from '@codemirror/view'

export type RealtimeAIStatus = 'idle' | 'thinking' | 'suggesting'
export type CompletionPace = 'restrained' | 'manual' | 'active'

export interface ActiveLintPopover {
  match: LintMatch
  pos: { top: number; left: number }
}

export interface ActiveMarginNotePopover {
  note: MarginNote
  pos: { top: number; left: number }
}

interface UseRealtimeAIOptions {
  view: Ref<EditorView | undefined>
  onViewUpdate: (callback: UpdateHook) => () => void
  enabled: Ref<boolean>
  enableCompletion?: Ref<boolean>
  completionPace?: Ref<CompletionPace>
  enableMarginNotes?: Ref<boolean>
  enableLint?: Ref<boolean>
  getTitle?: () => string | undefined
}

const SKIP_MARKER = '<SKIP>'

function isInsideCodeFence(docText: string, pos: number): boolean {
  const before = docText.slice(0, pos)
  const fences = before.match(/^```/gm)
  return !!fences && fences.length % 2 === 1
}

function isHeadingOrDirectiveLine(lineText: string): boolean {
  const trimmed = lineText.trim()
  return (
    trimmed.startsWith('#') ||
    trimmed.startsWith(':::') ||
    trimmed.startsWith('---') ||
    trimmed.startsWith('```')
  )
}

/** 根据字符类型返回拟人书写的释放节拍（ms） */
function charPaceMs(ch: string, backlog: number): number {
  if (ch === '。' || ch === '！' || ch === '？' || ch === '.' || ch === '!' || ch === '?') {
    return backlog > 36 ? 70 : 115
  }
  if (ch === '，' || ch === '、' || ch === '；' || ch === '：' || ch === ',' || ch === ';' || ch === ':') {
    return backlog > 36 ? 45 : 70
  }
  if (backlog > 36) return 20
  return 30
}

/** 提取某个位置所在的非空 Markdown 段落（起止偏移与文本） */
function extractParagraphAt(
  view: EditorView,
  lineNumber: number,
): { from: number; to: number; text: string; prevContext: string } | null {
  const { doc } = view.state
  if (lineNumber < 1 || lineNumber > doc.lines) return null
  const targetLine = doc.line(lineNumber)
  if (!targetLine.text.trim() || isHeadingOrDirectiveLine(targetLine.text)) return null

  let startLine = lineNumber
  while (startLine > 1) {
    const prevText = doc.line(startLine - 1).text
    if (!prevText.trim() || isHeadingOrDirectiveLine(prevText)) break
    startLine--
  }

  let endLine = lineNumber
  while (endLine < doc.lines) {
    const nextText = doc.line(endLine + 1).text
    if (!nextText.trim() || isHeadingOrDirectiveLine(nextText)) break
    endLine++
  }

  const from = doc.line(startLine).from
  const to = doc.line(endLine).to
  const text = doc.sliceString(from, to)
  const prevContext = doc.sliceString(Math.max(0, from - 450), from).trim()

  return { from, to, text, prevContext }
}

export function useRealtimeAI(options: UseRealtimeAIOptions) {
  const {
    view,
    onViewUpdate,
    enabled,
    enableCompletion,
    completionPace,
    enableMarginNotes,
    enableLint,
    getTitle,
  } = options

  const isCompletionOn = () => enabled.value && (enableCompletion ? enableCompletion.value : true)
  const isMarginNotesOn = () => enabled.value && (enableMarginNotes ? enableMarginNotes.value : true)
  const isLintOn = () => enabled.value && (enableLint ? enableLint.value : true)
  const getPace = (): CompletionPace => completionPace?.value ?? 'restrained'

  const status = ref<RealtimeAIStatus>('idle')
  const activeLint = ref<ActiveLintPopover | null>(null)
  const activeMarginNote = ref<ActiveMarginNotePopover | null>(null)

  let completionTimer: ReturnType<typeof setTimeout> | null = null
  let lintTimer: ReturnType<typeof setTimeout> | null = null
  let marginTimer: ReturnType<typeof setTimeout> | null = null
  let completionAbort: AbortController | null = null
  let lintAbort: AbortController | null = null
  let marginAbort: AbortController | null = null
  let flashTimer: ReturnType<typeof setTimeout> | null = null

  // ── 书写流速引擎（Ink Pacer）状态 ──
  let pacerRaf: number | null = null
  let pacerAnchorPos = -1
  let pacerRawBuffer = ''
  let pacerConfirmedText = ''
  let pacerDisplayedLen = 0
  let pacerStreamDone = false
  let pacerSkipped = false
  let pacerNextReleaseAt = 0

  // ── 心流节制 & 退避冷却状态 ──
  let cooldownUntil = 0
  let consecutiveDismissals = 0
  let silencedParagraphLine: number | null = null
  const recentInsertTimes: number[] = []

  const checkedParagraphs = new Set<string>()
  const annotatedParagraphs = new Set<string>()
  const ignoredOriginals = new Set<string>()

  function stopInkPacer() {
    if (pacerRaf !== null) {
      cancelAnimationFrame(pacerRaf)
      pacerRaf = null
    }
    pacerAnchorPos = -1
    pacerRawBuffer = ''
    pacerConfirmedText = ''
    pacerDisplayedLen = 0
    pacerStreamDone = false
    pacerSkipped = false
    pacerNextReleaseAt = 0
  }

  function clearTimers() {
    if (completionTimer) {
      clearTimeout(completionTimer)
      completionTimer = null
    }
    if (lintTimer) {
      clearTimeout(lintTimer)
      lintTimer = null
    }
    if (marginTimer) {
      clearTimeout(marginTimer)
      marginTimer = null
    }
  }

  function abortRequests() {
    stopInkPacer()
    if (completionAbort) {
      completionAbort.abort()
      completionAbort = null
    }
    if (lintAbort) {
      lintAbort.abort()
      lintAbort = null
    }
    if (marginAbort) {
      marginAbort.abort()
      marginAbort = null
    }
  }

  function resetAll() {
    clearTimers()
    abortRequests()
    status.value = 'idle'
    activeLint.value = null
    activeMarginNote.value = null
    const currentView = view.value
    if (currentView) {
      currentView.dispatch({
        effects: [
          clearGhostText.of(null),
          clearLintMatches.of(null),
          clearMarginNotes.of(null),
        ],
      })
    }
  }

  /** 判断作者最近是否正处于高速连续输入的心流状态 */
  function isFastTypingFlow(now: number): boolean {
    if (recentInsertTimes.length < 4) return false
    const first = recentInsertTimes[0]!
    const last = recentInsertTimes[recentInsertTimes.length - 1]!
    if (now - last > 1500) return false
    const avgInterval = (last - first) / (recentInsertTimes.length - 1)
    return avgInterval < 280
  }

  function stepInkPacer(now: number) {
    pacerRaf = null
    const v = view.value
    if (!v || v.state.selection.main.head !== pacerAnchorPos || pacerSkipped) {
      stopInkPacer()
      return
    }

    const totalLen = pacerConfirmedText.length
    if (pacerDisplayedLen < totalLen) {
      if (now >= pacerNextReleaseAt) {
        const nextChar = pacerConfirmedText[pacerDisplayedLen]!
        pacerDisplayedLen++
        const backlog = totalLen - pacerDisplayedLen
        const isDone = pacerStreamDone && pacerDisplayedLen >= totalLen

        status.value = 'suggesting'
        v.dispatch({
          effects: setGhostText.of({
            pos: pacerAnchorPos,
            text: pacerConfirmedText.slice(0, pacerDisplayedLen),
            done: isDone,
          }),
        })

        pacerNextReleaseAt = now + charPaceMs(nextChar, backlog)
      }
      pacerRaf = requestAnimationFrame(stepInkPacer)
      return
    }

    if (pacerStreamDone) {
      if (pacerDisplayedLen > 0) {
        // 确保最终状态带有 done: true，让极简 Tab 角标柔和淡入
        const currentGhost = v.state.field(ghostTextField, false)
        if (currentGhost && !currentGhost.done) {
          v.dispatch({
            effects: setGhostText.of({
              pos: pacerAnchorPos,
              text: currentGhost.text,
              done: true,
            }),
          })
        }
        status.value = 'suggesting'
      } else {
        status.value = 'idle'
      }
      pacerRaf = null
    }
  }

  function feedInkPacer(chunk: string, anchorPos: number) {
    if (pacerSkipped) return
    pacerRawBuffer += chunk

    // 检测 <SKIP> 静音指令：前几个字符可能是 "<SKIP>" 的前缀，暂缓释放直到确认不是静音
    const trimmedStart = pacerRawBuffer.trimStart()
    if (trimmedStart.length === 0) return
    if (trimmedStart.startsWith(SKIP_MARKER)) {
      pacerSkipped = true
      stopInkPacer()
      status.value = 'idle'
      return
    }
    if (trimmedStart.length < SKIP_MARKER.length && SKIP_MARKER.startsWith(trimmedStart)) {
      return
    }

    // 去掉模型偶尔带出的前导过多换行
    const cleaned = pacerRawBuffer.replace(/^\n{2,}/, '\n')
    if (!cleaned) return

    pacerAnchorPos = anchorPos
    pacerConfirmedText = cleaned
    if (pacerRaf === null) {
      pacerNextReleaseAt = performance.now()
      pacerRaf = requestAnimationFrame(stepInkPacer)
    }
  }

  function finishInkPacer(anchorPos: number) {
    if (pacerSkipped) {
      status.value = 'idle'
      return
    }
    const trimmed = pacerRawBuffer.trim()
    if (!trimmed || trimmed === SKIP_MARKER || trimmed.startsWith(SKIP_MARKER)) {
      stopInkPacer()
      status.value = 'idle'
      return
    }

    // 如果流已结束但之前因长度小于 SKIP_MARKER 被暂缓，此时补放出来
    const cleaned = pacerRawBuffer.replace(/^\n{2,}/, '\n')
    pacerAnchorPos = anchorPos
    pacerConfirmedText = cleaned
    pacerStreamDone = true

    if (pacerRaf === null) {
      pacerNextReleaseAt = performance.now()
      pacerRaf = requestAnimationFrame(stepInkPacer)
    }
  }

  async function runCompletion(manual = false) {
    const currentView = view.value
    if (!currentView || !isCompletionOn() || !currentView.hasFocus) return
    if (currentView.composing) return

    const { state } = currentView
    const sel = state.selection.main
    if (!sel.empty) return

    const line = state.doc.lineAt(sel.head)
    const afterCursorOnLine = line.text.slice(sel.head - line.from)
    if (afterCursorOnLine.trim().length > 0) return

    // 自动触发时的额外克制门控（手动 Alt+\ 触发可跳过这些限制）
    if (!manual) {
      const now = Date.now()
      if (now < cooldownUntil) return
      if (silencedParagraphLine !== null && silencedParagraphLine === line.number) return
      if (isHeadingOrDirectiveLine(line.text)) return
    }

    const fullText = state.doc.toString()
    if (isInsideCodeFence(fullText, sel.head)) return

    const prefix = state.doc.sliceString(Math.max(0, sel.head - 1200), sel.head)
    if (prefix.trim().length < 15) return

    const suffix = state.doc.sliceString(sel.head, Math.min(state.doc.length, sel.head + 360))

    stopInkPacer()
    completionAbort?.abort()
    completionAbort = new AbortController()
    const signal = completionAbort.signal
    const anchorPos = sel.head

    status.value = 'thinking'

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
          if (!v || v.state.selection.main.head !== anchorPos || v.composing) return
          feedInkPacer(chunk, anchorPos)
        },
        signal,
      )
      if (!signal.aborted) {
        finishInkPacer(anchorPos)
      }
    } catch {
      // AbortError 或网络波动静默处理，避免打断写作心流
      if (!signal.aborted) {
        stopInkPacer()
        status.value = 'idle'
      }
    }
  }

  async function runInlineLint(centerPos: number) {
    const currentView = view.value
    if (!currentView || !isLintOn() || currentView.composing) return

    const { doc } = currentView.state
    const centerLine = doc.lineAt(Math.min(centerPos, doc.length))
    const startLineNum = Math.max(1, centerLine.number - 1)
    const endLineNum = Math.min(doc.lines, centerLine.number + 1)

    const windowFrom = doc.line(startLineNum).from
    const windowTo = doc.line(endLineNum).to
    const snippet = doc.sliceString(windowFrom, windowTo)

    if (snippet.trim().length < 14) return
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

  async function runMarginAnnotation(targetLineNum: number, minChars: number) {
    const currentView = view.value
    if (!currentView || !isMarginNotesOn() || currentView.composing) return

    const para = extractParagraphAt(currentView, targetLineNum)
    if (!para) return

    const trimmed = para.text.trim()
    const nonSpaceLen = trimmed.replace(/\s/g, '').length
    if (nonSpaceLen < minChars) return
    if (isInsideCodeFence(currentView.state.doc.toString(), para.from)) return
    if (annotatedParagraphs.has(trimmed)) return

    marginAbort?.abort()
    marginAbort = new AbortController()
    const signal = marginAbort.signal

    try {
      const resp = await annotateParagraph(
        {
          paragraph: trimmed,
          prevContext: para.prevContext || undefined,
          title: getTitle?.(),
        },
        signal,
      )
      if (signal.aborted) return

      if (annotatedParagraphs.size > 40) {
        annotatedParagraphs.clear()
      }
      annotatedParagraphs.add(trimmed)

      const first = resp.annotations?.[0]
      if (!first) return

      const v = view.value
      if (!v) return
      const latestText = v.state.doc.sliceString(
        para.from,
        Math.min(v.state.doc.length, para.to),
      )
      if (latestText !== para.text) return

      const note: MarginNote = {
        id: `mn:${para.from}:${Date.now()}`,
        paragraphFrom: para.from,
        paragraphTo: para.to,
        paragraphText: para.text,
        kind: first.kind,
        quote: first.quote,
        comment: first.comment,
        suggestion: first.suggestion,
      }

      v.dispatch({
        effects: upsertMarginNote.of(note),
      })
    } catch {
      // 静默失败
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

  function dismissMarginNote(note: MarginNote) {
    view.value?.dispatch({ effects: removeMarginNote.of(note.id) })
    activeMarginNote.value = null
    view.value?.focus()
  }

  function closeMarginNote() {
    activeMarginNote.value = null
  }

  function triggerCompletionNow() {
    if (completionTimer) clearTimeout(completionTimer)
    void runCompletion(true)
  }

  // ── 绑定编辑器 DOM 自定义事件（采纳 / 拒绝 / 手动唤起 / 点击眉批） ──
  function handleGhostAccepted() {
    stopInkPacer()
    status.value = 'idle'
    consecutiveDismissals = 0
    cooldownUntil = 0
    silencedParagraphLine = null
  }

  function handleGhostDismissed() {
    stopInkPacer()
    completionAbort?.abort()
    status.value = 'idle'
    consecutiveDismissals++
    cooldownUntil = Date.now() + 30_000
    if (consecutiveDismissals >= 2 && view.value) {
      silencedParagraphLine = view.value.state.doc.lineAt(view.value.state.selection.main.head).number
    }
  }

  function handleManualTrigger() {
    if (!isCompletionOn()) return
    triggerCompletionNow()
  }

  function handleMarginNoteOpen(event: Event) {
    const detail = (event as CustomEvent<{ noteId: string; pos: { top: number; left: number } }>).detail
    if (!detail || !view.value) return
    const notes = view.value.state.field(marginNotesField, false) || []
    const found = notes.find((n) => n.id === detail.noteId)
    if (!found) return
    activeLint.value = null
    activeMarginNote.value = {
      note: found,
      pos: detail.pos,
    }
  }

  let boundDom: HTMLElement | null = null
  function bindDomEvents(el: HTMLElement) {
    if (boundDom === el) return
    unbindDomEvents()
    boundDom = el
    el.addEventListener('ai-ghost-accepted', handleGhostAccepted)
    el.addEventListener('ai-ghost-dismissed', handleGhostDismissed)
    el.addEventListener('ai-ghost-manual-trigger', handleManualTrigger)
    el.addEventListener('ai-margin-note-open', handleMarginNoteOpen)
  }

  function unbindDomEvents() {
    if (!boundDom) return
    boundDom.removeEventListener('ai-ghost-accepted', handleGhostAccepted)
    boundDom.removeEventListener('ai-ghost-dismissed', handleGhostDismissed)
    boundDom.removeEventListener('ai-ghost-manual-trigger', handleManualTrigger)
    boundDom.removeEventListener('ai-margin-note-open', handleMarginNoteOpen)
    boundDom = null
  }

  watch(
    view,
    (v) => {
      if (v) {
        bindDomEvents(v.dom)
      } else {
        unbindDomEvents()
      }
    },
    { immediate: true },
  )

  const stopUpdateListener = onViewUpdate((update: ViewUpdate) => {
    if (!enabled.value) return

    if (update.docChanged) {
      const hadGhostBefore = !!update.startState.field(ghostTextField, false)
      const hasGhostNow = !!update.state.field(ghostTextField, false)

      // 如果原本有幽灵文本、但用户没有采纳而是自己打了别的字，触发轻度冷却
      if (hadGhostBefore && !hasGhostNow) {
        let wasAcceptedOrMatched = false
        for (const tr of update.transactions) {
          for (const eff of tr.effects) {
            if (eff.is(clearGhostText)) {
              wasAcceptedOrMatched = true
            }
          }
        }
        if (!wasAcceptedOrMatched) {
          cooldownUntil = Math.max(cooldownUntil, Date.now() + 15_000)
        }
      }

      activeLint.value = null
      activeMarginNote.value = null
      stopInkPacer()
      completionAbort?.abort()
      status.value = hasGhostNow ? 'suggesting' : 'idle'

      // 统计本次变更的增删字符与是否换行
      let insertedLen = 0
      let deletedLen = 0
      let insertedNewline = false
      update.changes.iterChanges((fromA, toA, _fromB, _toB, inserted) => {
        deletedLen += toA - fromA
        const str = inserted.toString()
        insertedLen += str.length
        if (str.includes('\n')) {
          insertedNewline = true
        }
      })

      const now = Date.now()
      const head = update.state.selection.main.head
      const curLineNum = update.state.doc.lineAt(head).number

      // 换行开启新段落时，解除旧段落静默状态
      if (insertedNewline || (silencedParagraphLine !== null && curLineNum !== silencedParagraphLine)) {
        silencedParagraphLine = null
        consecutiveDismissals = 0
      }

      if (insertedLen > 0) {
        recentInsertTimes.push(now)
        if (recentInsertTimes.length > 6) recentInsertTimes.shift()
      }

      // ── 1. 幽灵续写调度（高度克制） ──
      if (completionTimer) clearTimeout(completionTimer)
      const pace = getPace()
      const isPureDeletion = insertedLen === 0 && deletedLen > 0
      if (isCompletionOn() && pace !== 'manual' && !isPureDeletion && !update.view.composing) {
        const baseDelay = pace === 'active' ? 1600 : 2800
        const flowBonus = isFastTypingFlow(now) ? 1400 : 0
        completionTimer = setTimeout(() => {
          void runCompletion(false)
        }, baseDelay + flowBonus)
      }

      // ── 2. 错字校对调度 ──
      if (lintTimer) clearTimeout(lintTimer)
      if (isLintOn() && !update.view.composing) {
        lintTimer = setTimeout(() => {
          void runInlineLint(head)
        }, 2600)
      }

      // ── 3. 纸边眉批调度（段落落笔后静默审阅） ──
      if (marginTimer) clearTimeout(marginTimer)
      if (isMarginNotesOn() && !update.view.composing) {
        if (insertedNewline && curLineNum > 1) {
          // 刚按回车结束上一段：稍等 1.2s 审阅刚写完的那一段
          const prevLineNum = curLineNum - 1
          marginTimer = setTimeout(() => {
            void runMarginAnnotation(prevLineNum, 36)
          }, 1200)
        } else if (!isPureDeletion) {
          // 或者在当前段停笔 5.8s 后审阅
          marginTimer = setTimeout(() => {
            void runMarginAnnotation(curLineNum, 45)
          }, 5800)
        }
      }
    } else if (update.selectionSet) {
      const sel = update.state.selection.main
      const hasGhost = !!update.state.field(ghostTextField, false)
      if (!hasGhost && status.value === 'suggesting') {
        status.value = 'idle'
      }
      if (sel.empty) {
        updateActiveLintAtPos(sel.head)
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
        stopInkPacer()
        completionAbort?.abort()
        status.value = 'idle'
        view.value?.dispatch({ effects: clearGhostText.of(null) })
      }
    })
  }

  if (enableMarginNotes) {
    watch(enableMarginNotes, (on) => {
      if (!on) {
        if (marginTimer) clearTimeout(marginTimer)
        marginAbort?.abort()
        activeMarginNote.value = null
        view.value?.dispatch({ effects: clearMarginNotes.of(null) })
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
    unbindDomEvents()
    clearTimers()
    abortRequests()
    if (flashTimer) clearTimeout(flashTimer)
  })

  return {
    status,
    activeLint,
    activeMarginNote,
    acceptLint,
    dismissLint,
    closeLint,
    dismissMarginNote,
    closeMarginNote,
    triggerCompletionNow,
  }
}
