import { StateEffect, StateField } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'

import type { DecorationSet } from '@codemirror/view'

export interface LintMatch {
  id: string
  from: number
  to: number
  original: string
  suggestion: string
  reason: string
  severity: 'typo' | 'style'
}

export const setLintMatches = StateEffect.define<LintMatch[]>()
export const removeLintMatch = StateEffect.define<string>()
export const clearLintMatches = StateEffect.define<null>()
export const flashRangeEffect = StateEffect.define<{ from: number; to: number } | null>()

export const lintMatchesField = StateField.define<LintMatch[]>({
  create() {
    return []
  },
  update(matches, tr) {
    for (const effect of tr.effects) {
      if (effect.is(clearLintMatches)) {
        return []
      }
      if (effect.is(setLintMatches)) {
        return effect.value
      }
      if (effect.is(removeLintMatch)) {
        return matches.filter((m) => m.id !== effect.value)
      }
    }

    if (!tr.docChanged || matches.length === 0) {
      return matches
    }

    // 文档变更时自动映射位置，并剔除被编辑过的区间
    const updated: LintMatch[] = []
    for (const m of matches) {
      const mappedFrom = tr.changes.mapPos(m.from, 1)
      const mappedTo = tr.changes.mapPos(m.to, -1)
      if (mappedFrom >= mappedTo) continue
      const currentText = tr.state.doc.sliceString(mappedFrom, mappedTo)
      if (currentText !== m.original) continue
      updated.push({ ...m, from: mappedFrom, to: mappedTo })
    }
    return updated
  },
  provide: (field) =>
    EditorView.decorations.from(field, (matches): DecorationSet => {
      if (matches.length === 0) return Decoration.none
      const sorted = [...matches].sort((a, b) => a.from - b.from)
      return Decoration.set(
        sorted.map((m) =>
          Decoration.mark({
            class: m.severity === 'typo' ? 'cm-ai-lint-typo' : 'cm-ai-lint-style',
            attributes: { 'data-lint-id': m.id },
          }).range(m.from, m.to),
        ),
      )
    }),
})

export const flashHighlightField = StateField.define<{ from: number; to: number } | null>({
  create() {
    return null
  },
  update(value, tr) {
    for (const effect of tr.effects) {
      if (effect.is(flashRangeEffect)) {
        return effect.value
      }
    }
    if (value && tr.docChanged) {
      const from = tr.changes.mapPos(value.from, 1)
      const to = tr.changes.mapPos(value.to, -1)
      return from < to ? { from, to } : null
    }
    return value
  },
  provide: (field) =>
    EditorView.decorations.from(field, (range): DecorationSet => {
      if (!range) return Decoration.none
      return Decoration.set([
        Decoration.mark({ class: 'cm-ai-lint-flash' }).range(range.from, range.to),
      ])
    }),
})

export const aiInlineLintExtension = [lintMatchesField, flashHighlightField]
