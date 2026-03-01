import type { MarkdownExtension } from '../types'

/**
 * Inline rule: matches <@user@instance> and renders as a colored badge.
 */
export const federationMentionExtension: MarkdownExtension = (md) => {
  const mentionRe = /^<@([^\s@<>]+)@([^\s<>]+)>/

  md.inline.ruler.before('html_inline', 'federation_mention', (state, silent) => {
    const src = state.src.slice(state.pos)
    const match = mentionRe.exec(src)
    if (!match) return false
    if (silent) return true

    const token = state.push('federation_mention', '', 0)
    token.markup = match[0]
    token.meta = { user: match[1], instance: match[2] }
    state.pos += match[0].length
    return true
  })

  md.renderer.rules.federation_mention = (tokens, idx) => {
    const { user, instance } = tokens[idx].meta
    const esc = md.utils.escapeHtml
    return `<span class="fed-mention">@${esc(user)}<span class="fed-mention-host">@${esc(instance)}</span></span>`
  }
}
