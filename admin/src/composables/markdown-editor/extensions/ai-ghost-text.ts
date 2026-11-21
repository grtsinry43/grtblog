import { Prec, StateEffect, StateField } from '@codemirror/state'
import { Decoration, EditorView, WidgetType, keymap } from '@codemirror/view'

import type { DecorationSet } from '@codemirror/view'

export interface GhostTextPayload {
  pos: number
  text: string
  /** 是否已完成流式书写（完成后才淡入极简提示角标） */
  done?: boolean
}

export const setGhostText = StateEffect.define<GhostTextPayload>()
export const clearGhostText = StateEffect.define<null>()
export const settleRangeEffect = StateEffect.define<{ from: number; to: number; id: number } | null>()

interface GhostTextState {
  pos: number
  text: string
  done: boolean
}

interface SettleRangeState {
  from: number
  to: number
  id: number
}

function createCharNode(ch: string, animate: boolean): Node {
  if (ch === '\n') {
    return document.createTextNode('\n')
  }
  const span = document.createElement('span')
  span.className = animate ? 'cm-ai-ghost-char cm-ai-ghost-char--fresh' : 'cm-ai-ghost-char'
  span.textContent = ch
  return span
}

class GhostTextWidget extends WidgetType {
  constructor(
    readonly text: string,
    readonly done: boolean,
  ) {
    super()
  }

  eq(other: GhostTextWidget) {
    return this.text === other.text && this.done === other.done
  }

  toDOM() {
    const wrapper = document.createElement('span')
    wrapper.className = 'cm-ai-ghost-wrapper'
    wrapper.setAttribute('aria-hidden', 'true')
    wrapper.dataset.ghostText = this.text

    const content = document.createElement('span')
    content.className = 'cm-ai-ghost-content'
    for (const ch of Array.from(this.text)) {
      content.appendChild(createCharNode(ch, true))
    }
    wrapper.appendChild(content)

    const hint = document.createElement('span')
    hint.className = this.done ? 'cm-ai-ghost-hint is-visible' : 'cm-ai-ghost-hint'
    hint.textContent = 'Tab'
    wrapper.appendChild(hint)

    return wrapper
  }

  /**
   * 复用已有 DOM，仅对新增字符追加带淡入动画的节点，
   * 避免流式追加或逐句采纳时整段闪烁与提示角标重播。
   */
  updateDOM(dom: HTMLElement): boolean {
    const prevText = dom.dataset.ghostText ?? ''
    const content = dom.querySelector('.cm-ai-ghost-content') as HTMLElement | null
    const hint = dom.querySelector('.cm-ai-ghost-hint') as HTMLElement | null
    if (!content || !hint) return false

    if (this.text !== prevText) {
      if (this.text.startsWith(prevText)) {
        // 流式追加：仅给新抵达的字播放柔和显影动画
        const delta = this.text.slice(prevText.length)
        for (const ch of Array.from(delta)) {
          content.appendChild(createCharNode(ch, true))
        }
      } else {
        // 逐句采纳或前缀匹配前移：剩余文字已落定，不再重播动画
        content.textContent = ''
        for (const ch of Array.from(this.text)) {
          content.appendChild(createCharNode(ch, false))
        }
      }
      dom.dataset.ghostText = this.text
    }

    hint.classList.toggle('is-visible', this.done)
    return true
  }

  ignoreEvent() {
    return true
  }
}

export const ghostTextField = StateField.define<GhostTextState | null>({
  create() {
    return null
  },
  update(value, tr) {
    for (const effect of tr.effects) {
      if (effect.is(clearGhostText)) {
        return null
      }
      if (effect.is(setGhostText)) {
        const { pos, text, done = false } = effect.value
        if (!text) return null
        return { pos, text, done }
      }
    }

    if (!value) return null

    if (tr.docChanged) {
      // 如果用户在幽灵文本起点处打字，且输入内容恰好匹配幽灵前缀，则平滑前移
      let matched = false
      let insertedText = ''
      let changeCount = 0
      tr.changes.iterChanges((fromA, toA, _fromB, _toB, inserted) => {
        changeCount++
        if (fromA === value.pos && toA === value.pos) {
          insertedText = inserted.toString()
          matched = true
        }
      })

      if (matched && changeCount === 1 && insertedText.length > 0 && value.text.startsWith(insertedText)) {
        const remaining = value.text.slice(insertedText.length)
        if (!remaining) return null
        return {
          pos: value.pos + insertedText.length,
          text: remaining,
          done: value.done,
        }
      }
      return null
    }

    if (tr.selection) {
      const sel = tr.selection.main
      if (!sel.empty || sel.head !== value.pos) {
        return null
      }
    }

    return value
  },
  provide: (field) =>
    EditorView.decorations.from(field, (state): DecorationSet => {
      if (!state || !state.text) return Decoration.none
      return Decoration.set([
        Decoration.widget({
          widget: new GhostTextWidget(state.text, state.done),
          side: 1,
        }).range(state.pos),
      ])
    }),
})

export const settleHighlightField = StateField.define<SettleRangeState | null>({
  create() {
    return null
  },
  update(value, tr) {
    for (const effect of tr.effects) {
      if (effect.is(settleRangeEffect)) {
        return effect.value
      }
    }
    if (value && tr.docChanged) {
      const from = tr.changes.mapPos(value.from, 1)
      const to = tr.changes.mapPos(value.to, -1)
      return from < to ? { ...value, from, to } : null
    }
    return value
  },
  provide: (field) =>
    EditorView.decorations.from(field, (range): DecorationSet => {
      if (!range || range.from >= range.to) return Decoration.none
      return Decoration.set([
        Decoration.mark({
          class: 'cm-ai-ghost-settle',
          attributes: { 'data-settle-id': String(range.id) },
        }).range(range.from, range.to),
      ])
    }),
})

function scheduleClearSettle(view: EditorView, settleId: number) {
  window.setTimeout(() => {
    if (!view.dom.isConnected) return
    const current = view.state.field(settleHighlightField, false)
    if (current && current.id === settleId) {
      view.dispatch({ effects: settleRangeEffect.of(null) })
    }
  }, 540)
}

/** 提取下一个自然句/词边界长度，用于 Ctrl/Cmd + → 渐进采纳 */
function findNextClauseLength(text: string): number {
  const clauseMatch = text.match(/^.{1,24}?[，。！？；：,.!?;:\n\s]+/)
  if (clauseMatch) {
    return clauseMatch[0].length
  }
  const wordMatch = text.match(/^[\w\u4e00-\u9fa5]{1,8}/)
  if (wordMatch) {
    return wordMatch[0].length
  }
  return Math.min(text.length, 4)
}

function acceptFullGhostText(view: EditorView): boolean {
  const ghost = view.state.field(ghostTextField, false)
  if (!ghost || !ghost.text) return false

  const from = ghost.pos
  const to = ghost.pos + ghost.text.length
  const settleId = Date.now()

  view.dispatch({
    changes: { from, insert: ghost.text },
    selection: { anchor: to },
    effects: [clearGhostText.of(null), settleRangeEffect.of({ from, to, id: settleId })],
  })
  scheduleClearSettle(view, settleId)
  view.dom.dispatchEvent(new CustomEvent('ai-ghost-accepted', { bubbles: true }))
  return true
}

function acceptNextClauseGhostText(view: EditorView): boolean {
  const ghost = view.state.field(ghostTextField, false)
  if (!ghost || !ghost.text) return false

  const len = findNextClauseLength(ghost.text)
  const accepted = ghost.text.slice(0, len)
  const remaining = ghost.text.slice(len)
  const from = ghost.pos
  const to = ghost.pos + accepted.length
  const settleId = Date.now()

  view.dispatch({
    changes: { from, insert: accepted },
    selection: { anchor: to },
    effects: [
      remaining
        ? setGhostText.of({
            pos: to,
            text: remaining,
            done: ghost.done,
          })
        : clearGhostText.of(null),
      settleRangeEffect.of({ from, to, id: settleId }),
    ],
  })
  scheduleClearSettle(view, settleId)
  view.dom.dispatchEvent(new CustomEvent('ai-ghost-accepted', { bubbles: true }))
  return true
}

function dismissGhostText(view: EditorView): boolean {
  const ghost = view.state.field(ghostTextField, false)
  if (!ghost) return false

  view.dispatch({ effects: clearGhostText.of(null) })
  view.dom.dispatchEvent(new CustomEvent('ai-ghost-dismissed', { bubbles: true }))
  return true
}

function triggerManualGhostText(view: EditorView): boolean {
  view.dom.dispatchEvent(new CustomEvent('ai-ghost-manual-trigger', { bubbles: true }))
  return true
}

export const aiGhostTextExtension = [
  ghostTextField,
  settleHighlightField,
  Prec.highest(
    keymap.of([
      { key: 'Tab', run: acceptFullGhostText },
      { key: 'Ctrl-ArrowRight', run: acceptNextClauseGhostText },
      { key: 'Mod-ArrowRight', run: acceptNextClauseGhostText },
      { key: 'Escape', run: dismissGhostText },
      { key: 'Alt-\\', run: triggerManualGhostText },
    ]),
  ),
]
