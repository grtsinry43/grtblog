import {
  Decoration,
  type DecorationSet,
  EditorView,
  ViewPlugin,
  type ViewUpdate,
} from '@codemirror/view'
import { RangeSetBuilder } from '@codemirror/state'

const mentionRe = /<@([^\s@<>]+)@([^\s<>]+)>/g
const citationRe = /<cite:([^|<>]+)\|([^<>]+)>/g

const mentionDeco = Decoration.mark({ class: 'cm-federation-mention' })
const citationDeco = Decoration.mark({ class: 'cm-federation-citation' })

function buildDecorations(view: EditorView): DecorationSet {
  const builder = new RangeSetBuilder<Decoration>()
  for (const { from, to } of view.visibleRanges) {
    const text = view.state.doc.sliceString(from, to)
    let m: RegExpExecArray | null

    mentionRe.lastIndex = 0
    while ((m = mentionRe.exec(text)) !== null) {
      builder.add(from + m.index, from + m.index + m[0].length, mentionDeco)
    }

    citationRe.lastIndex = 0
    while ((m = citationRe.exec(text)) !== null) {
      builder.add(from + m.index, from + m.index + m[0].length, citationDeco)
    }
  }
  return builder.finish()
}

export const federationHighlightPlugin = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet
    constructor(view: EditorView) {
      this.decorations = buildDecorations(view)
    }
    update(update: ViewUpdate) {
      if (update.docChanged || update.viewportChanged) {
        this.decorations = buildDecorations(update.view)
      }
    }
  },
  { decorations: (v) => v.decorations },
)
