import type { MarkdownExtension } from '../types'

/**
 * Inline rule: matches <cite:instance|post-id> and renders as a preview placeholder card.
 */
export const federationCitationExtension: MarkdownExtension = (md) => {
  const citationRe = /^<cite:([^|<>]+)\|([^<>]+)>/

  md.inline.ruler.before('html_inline', 'federation_citation', (state, silent) => {
    const src = state.src.slice(state.pos)
    const match = citationRe.exec(src)
    if (!match) return false
    if (silent) return true

    const token = state.push('federation_citation', '', 0)
    token.markup = match[0]
    token.meta = { instance: match[1], postId: match[2] }
    state.pos += match[0].length
    return true
  })

  md.renderer.rules.federation_citation = (tokens, idx) => {
    const { instance, postId } = tokens[idx]!.meta!
    const esc = md.utils.escapeHtml
    return `<div class="fed-citation-preview">\uD83D\uDD17 引用: ${esc(instance)} / ${esc(postId)}</div>`
  }
}
