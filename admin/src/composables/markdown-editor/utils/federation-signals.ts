const mentionPatternSource = '<@([^\\s@<>]+)@([^\\s<>]+)>'
const citationPatternSource = '<cite:([^|<>]+)\\|([^<>]+)>'

export function createFederationMentionRegExp(flags = 'g') {
  return new RegExp(mentionPatternSource, flags)
}

export function createFederationCitationRegExp(flags = 'g') {
  return new RegExp(citationPatternSource, flags)
}

function normalizeSignalHost(value: string) {
  const trimmed = value.trim()
  if (!trimmed) return ''
  return (
    trimmed
      .replace(/^https?:\/\//i, '')
      .replace(/\/+$/, '')
      .split('/')[0] || ''
  )
}

export interface FederationMentionSignal {
  user: string
  instance: string
  key: string
  marker: string
}

export interface FederationCitationSignal {
  instance: string
  postId: string
  key: string
  marker: string
}

export function parseFederationSignals(content: string) {
  const mentions: FederationMentionSignal[] = []
  const citations: FederationCitationSignal[] = []
  const mentionSeen = new Set<string>()
  const citationSeen = new Set<string>()
  const mentionRe = createFederationMentionRegExp('g')
  const citationRe = createFederationCitationRegExp('g')

  let mentionMatch: RegExpExecArray | null
  while ((mentionMatch = mentionRe.exec(content)) !== null) {
    const user = mentionMatch[1]?.trim()
    const instance = normalizeSignalHost(mentionMatch[2] ?? '')
    if (!user || !instance) continue
    const key = `${user}@${instance}`
    if (mentionSeen.has(key)) continue
    mentionSeen.add(key)
    mentions.push({
      user,
      instance,
      key,
      marker: `<@${user}@${instance}>`,
    })
  }

  let citationMatch: RegExpExecArray | null
  while ((citationMatch = citationRe.exec(content)) !== null) {
    const instance = normalizeSignalHost(citationMatch[1] ?? '')
    const postId = citationMatch[2]?.trim()
    if (!instance || !postId) continue
    const key = `${instance}|${postId}`
    if (citationSeen.has(key)) continue
    citationSeen.add(key)
    citations.push({
      instance,
      postId,
      key,
      marker: `<cite:${instance}|${postId}>`,
    })
  }

  return { mentions, citations }
}
