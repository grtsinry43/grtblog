import { type CompletionContext, type CompletionResult } from '@codemirror/autocomplete'

import {
  getMarkdownComponent,
  parseComponentAttributes,
  parseComponentInfo,
} from '@/composables/markdown/shared/components'

const COMPONENT_HEADER_RE = /^\s*:::\s*(.+)$/

export const componentAttributeSource = (context: CompletionContext): CompletionResult | null => {
  const { state, pos } = context
  const line = state.doc.lineAt(pos)
  const offset = pos - line.from
  const textBefore = line.text.slice(0, offset)

  const headerMatch = COMPONENT_HEADER_RE.exec(textBefore)
  if (!headerMatch) return null

  const { name } = parseComponentInfo(headerMatch[1] ?? '')
  const component = getMarkdownComponent(name)
  if (!component || component.attrs.length === 0) return null

  const headerInfo = headerMatch[1] ?? ''
  const parsed = parseComponentInfo(headerInfo)
  if (!parsed.rawAttrs && !/\s$/.test(headerInfo)) return null

  const attrTextBefore = parsed.rawAttrs
  if (attrTextBefore.includes('{') || attrTextBefore.includes('}')) return null

  const tokenStart = Math.max(attrTextBefore.lastIndexOf(' ') + 1, 0)
  const currentToken = attrTextBefore.slice(tokenStart)
  if (currentToken.includes('=')) return null

  const existingPart = attrTextBefore.slice(0, tokenStart)
  const existingKeys = new Set(Object.keys(parseComponentAttributes(existingPart)))
  const query = currentToken
  const replacementFrom = pos - query.length
  const beforeFrom = replacementFrom > 0 ? state.sliceDoc(replacementFrom - 1, replacementFrom) : ''
  const insertPrefix =
    replacementFrom > line.from && beforeFrom && !/\s/.test(beforeFrom) ? ' ' : ''

  const queryLower = query.toLowerCase()
  const options = component.attrs
    .filter((attr) => !existingKeys.has(attr.key))
    .filter((attr) => (queryLower ? attr.key.toLowerCase().includes(queryLower) : true))
    .map((attr) => ({
      label: attr.key,
      type: 'property',
      detail: attr.label,
      apply: `${insertPrefix}${attr.key}=""`,
    }))

  if (!options.length) return null

  return {
    from: replacementFrom,
    to: pos,
    options,
    filter: false,
  }
}
