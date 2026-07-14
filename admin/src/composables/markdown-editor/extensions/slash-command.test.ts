import { CompletionContext } from '@codemirror/autocomplete'
import { markdown } from '@codemirror/lang-markdown'
import { EditorState } from '@codemirror/state'
import { describe, expect, it } from 'vitest'

import { slashCommandSource } from './slash-command'

const complete = (doc: string) => {
  const state = EditorState.create({ doc, extensions: [markdown()] })
  return slashCommandSource(new CompletionContext(state, doc.length, true))
}

describe('slash command completion', () => {
  it('keeps dedicated federation actions but hides server-generated component blocks', () => {
    const labels = complete('/')?.options.map((option) => option.label)

    expect(labels).toContain('@mention')
    expect(labels).toContain('Citation')
    expect(labels).not.toContain('Federation Mention')
    expect(labels).not.toContain('Federation Citation')
  })

  it('does not offer components without an insert template', () => {
    expect(complete('/fed')).toBeNull()
  })
})
