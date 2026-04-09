import { EditorView } from '@codemirror/view'

import { getMarkdownComponent, parseComponentInfo } from '@/composables/markdown/shared/components'

export interface ComponentEditPayload {
  name: string
  attrs: Record<string, string>
  rawAttrs: string
  blockFrom: number
  blockTo: number
  headerFrom: number
  headerTo: number
  body: string
  hasClosingFence: boolean
  isComponentSyntax: boolean
}

interface ComponentEditorOptions {
  onEdit?: (payload: ComponentEditPayload) => void
}

export const createComponentEditorExtension = (options: ComponentEditorOptions) => {
  if (!options.onEdit) return []

  const getHeaderAtLine = (view: EditorView, lineNumber: number) => {
    const line = view.state.doc.line(lineNumber)
    const trimmed = line.text.trim()
    if (!trimmed.startsWith(':::')) return null
    const info = trimmed.slice(3).trim()
    if (!info) return null
    const parsed = parseComponentInfo(info)
    const component = getMarkdownComponent(parsed.name)
    if (!component) return null

    return {
      line,
      parsed,
      isComponentSyntax: /^:::\s*component\s+/.test(trimmed),
    }
  }

  const findHeaderLine = (view: EditorView, clickedLineNumber: number) => {
    let pendingCloseCount = 0

    for (let lineNumber = clickedLineNumber; lineNumber >= 1; lineNumber -= 1) {
      const line = view.state.doc.line(lineNumber)
      const trimmed = line.text.trim()

      if (trimmed === ':::') {
        if (lineNumber === clickedLineNumber) {
          continue
        }
        pendingCloseCount += 1
        continue
      }

      const header = getHeaderAtLine(view, lineNumber)
      if (!header) continue

      if (pendingCloseCount > 0) {
        pendingCloseCount -= 1
        continue
      }

      return header
    }

    return null
  }

  const findClosingLine = (view: EditorView, headerLineNumber: number) => {
    let nestedOpenCount = 0

    for (
      let lineNumber = headerLineNumber + 1;
      lineNumber <= view.state.doc.lines;
      lineNumber += 1
    ) {
      const line = view.state.doc.line(lineNumber)
      const trimmed = line.text.trim()

      if (trimmed === ':::') {
        if (nestedOpenCount === 0) return line
        nestedOpenCount -= 1
        continue
      }

      if (trimmed.startsWith(':::') && getHeaderAtLine(view, lineNumber)) {
        nestedOpenCount += 1
      }
    }

    return null
  }

  const normalizeBodyText = (raw: string) => {
    let next = raw
    if (next.startsWith('\n')) next = next.slice(1)
    if (next.endsWith('\n')) next = next.slice(0, -1)
    return next
  }

  return EditorView.domEventHandlers({
    mousedown: (event, view) => {
      const pos = view.posAtCoords({ x: event.clientX, y: event.clientY })
      if (pos == null) return false

      const clickedLine = view.state.doc.lineAt(pos)
      const header = findHeaderLine(view, clickedLine.number)
      if (!header) return false

      const closingLine = findClosingLine(view, header.line.number)
      const bodyFrom = header.line.to
      const bodyTo = closingLine?.from ?? header.line.to
      const rawBody = view.state.doc.sliceString(bodyFrom, bodyTo)

      options.onEdit?.({
        name: header.parsed.name,
        attrs: header.parsed.attrs,
        rawAttrs: header.parsed.rawAttrs,
        headerFrom: header.line.from,
        headerTo: header.line.to,
        blockFrom: header.line.from,
        blockTo: closingLine?.to ?? header.line.to,
        body: normalizeBodyText(rawBody),
        hasClosingFence: Boolean(closingLine),
        isComponentSyntax: header.isComponentSyntax,
      })

      return false
    },
  })
}
