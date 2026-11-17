import { Prec, StateEffect, StateField } from '@codemirror/state'
import { Decoration, EditorView, WidgetType, keymap } from '@codemirror/view'

import type { DecorationSet } from '@codemirror/view'

export interface GhostTextPayload {
  pos: number
  text: string
  /** 仅对 [animFrom, text.length] 播放墨迹浮现动画，避免流式追加时整段闪烁 */
  animFrom?: number
}

export const setGhostText = StateEffect.define<GhostTextPayload>()
export const clearGhostText = StateEffect.define<null>()

interface GhostTextState {
  pos: number
  text: string
  animFrom: number
}

class GhostTextWidget extends WidgetType {
  constructor(
    readonly text: string,
    readonly animFrom: number,
  ) {
    super()
  }

  eq(other: GhostTextWidget) {
    return this.text === other.text && this.animFrom === other.animFrom
  }

  toDOM() {
    const wrapper = document.createElement('span')
    wrapper.className = 'cm-ai-ghost-wrapper'
    wrapper.setAttribute('aria-hidden', 'true')

    const settledText = this.text.slice(0, this.animFrom)
    const freshText = this.text.slice(this.animFrom)

    if (settledText) {
      const settledSpan = document.createElement('span')
      settledSpan.className = 'cm-ai-ghost-text'
      settledSpan.textContent = settledText
      wrapper.appendChild(settledSpan)
    }

    if (freshText) {
      // 按字/短片段错峰淡入，营造纸面墨迹浮现的阅读节奏
      const chars = Array.from(freshText)
      const stepMs = Math.min(22, Math.floor(180 / Math.max(chars.length, 1)))
      chars.forEach((ch, idx) => {
        if (ch === '\n') {
          wrapper.appendChild(document.createTextNode('\n'))
          return
        }
        const charSpan = document.createElement('span')
        charSpan.className = 'cm-ai-ghost-text cm-ai-ghost-ink'
        charSpan.style.animationDelay = `${Math.min(idx * stepMs, 200)}ms`
        charSpan.textContent = ch
        wrapper.appendChild(charSpan)
      })
    }

    const hint = document.createElement('span')
    hint.className = 'cm-ai-ghost-hint'
    hint.textContent = 'Tab 采纳 · Ctrl+→ 逐句 · Esc'
    wrapper.appendChild(hint)

    return wrapper
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
        const { pos, text, animFrom = 0 } = effect.value
        if (!text) return null
        return { pos, text, animFrom }
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
          animFrom: remaining.length,
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
          widget: new GhostTextWidget(state.text, state.animFrom),
          side: 1,
        }).range(state.pos),
      ])
    }),
})

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

  view.dispatch({
    changes: { from: ghost.pos, insert: ghost.text },
    selection: { anchor: ghost.pos + ghost.text.length },
    effects: clearGhostText.of(null),
  })
  return true
}

function acceptNextClauseGhostText(view: EditorView): boolean {
  const ghost = view.state.field(ghostTextField, false)
  if (!ghost || !ghost.text) return false

  const len = findNextClauseLength(ghost.text)
  const accepted = ghost.text.slice(0, len)
  const remaining = ghost.text.slice(len)

  view.dispatch({
    changes: { from: ghost.pos, insert: accepted },
    selection: { anchor: ghost.pos + accepted.length },
    effects: remaining
      ? setGhostText.of({
          pos: ghost.pos + accepted.length,
          text: remaining,
          animFrom: remaining.length,
        })
      : clearGhostText.of(null),
  })
  return true
}

function dismissGhostText(view: EditorView): boolean {
  const ghost = view.state.field(ghostTextField, false)
  if (!ghost) return false

  view.dispatch({ effects: clearGhostText.of(null) })
  return true
}

export const aiGhostTextExtension = [
  ghostTextField,
  Prec.highest(
    keymap.of([
      { key: 'Tab', run: acceptFullGhostText },
      { key: 'Ctrl-ArrowRight', run: acceptNextClauseGhostText },
      { key: 'Mod-ArrowRight', run: acceptNextClauseGhostText },
      { key: 'Escape', run: dismissGhostText },
    ]),
  ),
]
