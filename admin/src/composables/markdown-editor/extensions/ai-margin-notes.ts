import { StateEffect, StateField } from '@codemirror/state'
import { Decoration, EditorView, WidgetType } from '@codemirror/view'

import type { MarginAnnotationKind } from '@/services/ai'
import type { DecorationSet } from '@codemirror/view'

export interface MarginNote {
  id: string
  paragraphFrom: number
  paragraphTo: number
  paragraphText: string
  kind: MarginAnnotationKind
  quote?: string
  comment: string
  suggestion?: string
}

export const upsertMarginNote = StateEffect.define<MarginNote>()
export const removeMarginNote = StateEffect.define<string>()
export const clearMarginNotes = StateEffect.define<null>()

const KIND_LABELS: Record<MarginAnnotationKind, string> = {
  refine: '表达推敲',
  spark: '思路延伸',
  structure: '逻辑结构',
}

class MarginNoteDotWidget extends WidgetType {
  constructor(
    readonly noteId: string,
    readonly kind: MarginAnnotationKind,
  ) {
    super()
  }

  eq(other: MarginNoteDotWidget) {
    return this.noteId === other.noteId && this.kind === other.kind
  }

  toDOM(view: EditorView) {
    const btn = document.createElement('button')
    btn.type = 'button'
    btn.className = `cm-ai-margin-dot cm-ai-margin-dot--${this.kind}`
    btn.setAttribute('aria-label', `AI 眉批：${KIND_LABELS[this.kind]}`)
    btn.title = `${KIND_LABELS[this.kind]} · 点击查看眉批`

    const core = document.createElement('span')
    core.className = 'cm-ai-margin-dot-core'
    btn.appendChild(core)

    btn.addEventListener('mousedown', (e) => {
      e.preventDefault()
      e.stopPropagation()
    })

    btn.addEventListener('click', (e) => {
      e.preventDefault()
      e.stopPropagation()
      const rect = btn.getBoundingClientRect()
      view.dom.dispatchEvent(
        new CustomEvent('ai-margin-note-open', {
          bubbles: true,
          detail: {
            noteId: this.noteId,
            pos: {
              top: rect.top - 8,
              left: rect.left,
            },
          },
        }),
      )
    })

    return btn
  }

  ignoreEvent() {
    return true
  }
}

export const marginNotesField = StateField.define<MarginNote[]>({
  create() {
    return []
  },
  update(notes, tr) {
    let next = notes
    for (const effect of tr.effects) {
      if (effect.is(clearMarginNotes)) {
        return []
      }
      if (effect.is(removeMarginNote)) {
        next = next.filter((n) => n.id !== effect.value)
      }
      if (effect.is(upsertMarginNote)) {
        const incoming = effect.value
        // 同一段落只保留最新的一条眉批
        const filtered = next.filter(
          (n) =>
            n.id !== incoming.id &&
            (n.paragraphTo <= incoming.paragraphFrom || n.paragraphFrom >= incoming.paragraphTo),
        )
        next = [...filtered, incoming]
      }
    }

    if (!tr.docChanged || next.length === 0) {
      return next
    }

    const updated: MarginNote[] = []
    for (const note of next) {
      const from = tr.changes.mapPos(note.paragraphFrom, 1)
      const to = tr.changes.mapPos(note.paragraphTo, -1)
      if (from >= to) continue
      const currentText = tr.state.doc.sliceString(from, to).trim()
      // 如果段落被大幅改写或清空，自动移除过期眉批
      if (currentText.length < 12) continue
      if (Math.abs(currentText.length - note.paragraphText.trim().length) > Math.max(24, note.paragraphText.length * 0.45)) {
        continue
      }
      updated.push({
        ...note,
        paragraphFrom: from,
        paragraphTo: to,
      })
    }
    return updated
  },
  provide: (field) =>
    EditorView.decorations.from(field, (notes): DecorationSet => {
      if (notes.length === 0) return Decoration.none
      const sorted = [...notes].sort((a, b) => a.paragraphTo - b.paragraphTo)
      return Decoration.set(
        sorted.map((note) =>
          Decoration.widget({
            widget: new MarginNoteDotWidget(note.id, note.kind),
            side: 1,
          }).range(note.paragraphTo),
        ),
      )
    }),
})

export const aiMarginNotesExtension = [marginNotesField]
