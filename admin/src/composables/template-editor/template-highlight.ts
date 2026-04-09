import { RangeSetBuilder, StateEffect, StateField } from '@codemirror/state'
import {
  Decoration,
  EditorView,
  ViewPlugin,
  hoverTooltip,
  type DecorationSet,
  type ViewUpdate,
} from '@codemirror/view'

export const setTemplateVariables = StateEffect.define<string[]>()

export const templateVariablesField = StateField.define<string[]>({
  create() {
    return []
  },
  update(value, tr) {
    for (const effect of tr.effects) {
      if (effect.is(setTemplateVariables)) {
        return effect.value
      }
    }
    return value
  },
})

const templateToken = Decoration.mark({ class: 'cm-template-token' })
const templateInvalid = Decoration.mark({
  class: 'cm-template-invalid',
  attributes: { title: '未知模板变量' },
})

function normalizeExpr(expr: string) {
  return expr.replace(/\s+/g, ' ').trim()
}

const DEFAULT_VALID_VARIABLES = ['.Name', '.OccurredAt', '.SiteURL']

function isValidTemplateExpression(expression: string, validFields: string[]): boolean {
  const normalized = expression.trim()
  if (DEFAULT_VALID_VARIABLES.includes(normalized)) return true

  // Support both .Field and .Event.Field formats
  if (normalized.startsWith('.')) {
    let core = normalized.slice(1) // Remove leading dot
    if (core.startsWith('Event.')) {
      core = core.slice(6) // Remove 'Event.' prefix
    }
    const fieldName = core.split('.')[0]
    return fieldName ? validFields.includes(fieldName) : false
  }

  return false
}

function buildDecorations(view: EditorView) {
  const builder = new RangeSetBuilder<Decoration>()
  const pattern = /{{[^}]*}}/g
  const validFields = view.state.field(templateVariablesField)

  for (const range of view.visibleRanges) {
    const text = view.state.doc.sliceString(range.from, range.to)
    pattern.lastIndex = 0
    for (let match = pattern.exec(text); match; match = pattern.exec(text)) {
      const start = range.from + match.index
      const end = start + match[0].length
      const expr = match[0].slice(2, -2)
      const decoration = isValidTemplateExpression(expr, validFields)
        ? templateToken
        : templateInvalid
      builder.add(start, end, decoration)
    }
  }

  return builder.finish()
}

type TemplateMatch = {
  from: number
  to: number
  expr: string
}

function findTemplateMatch(lineText: string, lineFrom: number, pos: number): TemplateMatch | null {
  const pattern = /{{[^}]*}}/g
  let match: RegExpExecArray | null
  while ((match = pattern.exec(lineText))) {
    const from = lineFrom + match.index
    const to = from + match[0].length
    if (pos >= from && pos <= to) {
      return {
        from,
        to,
        expr: match[0].slice(2, -2),
      }
    }
  }
  return null
}

export const templateHighlightExtension = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet

    constructor(view: EditorView) {
      this.decorations = buildDecorations(view)
    }

    update(update: ViewUpdate) {
      // Re-build decorations if doc changed OR if the variables field changed (configuration update)
      const varsChanged =
        update.startState.field(templateVariablesField) !==
        update.state.field(templateVariablesField)
      if (update.docChanged || update.viewportChanged || varsChanged) {
        this.decorations = buildDecorations(update.view)
      }
    }
  },
  {
    decorations: (instance) => instance.decorations,
  },
)

export const templateTooltipExtension = hoverTooltip((view, pos, side) => {
  const line = view.state.doc.lineAt(pos)
  const match = findTemplateMatch(line.text, line.from, pos)
  if (!match) return null
  if (match.from === pos && side < 0) return null
  if (match.to === pos && side > 0) return null

  const validFields = view.state.field(templateVariablesField)
  if (isValidTemplateExpression(match.expr, validFields)) return null

  return {
    pos: match.from,
    end: match.to,
    above: true,
    create() {
      const dom = document.createElement('div')
      dom.className = 'cm-template-tooltip'
      dom.textContent = '未知模板变量'
      return { dom }
    },
  }
})
