export interface WritingIdeaHandoff {
  title?: string
  angle?: string
  why?: string
}

export type WritingIdeaTarget = 'article' | 'note'

const pendingIdeas: Partial<Record<WritingIdeaTarget, WritingIdeaHandoff>> = {}

export function setPendingWritingIdea(target: WritingIdeaTarget, idea: WritingIdeaHandoff) {
  const cleaned: WritingIdeaHandoff = {
    title: idea.title?.trim() || undefined,
    angle: idea.angle?.trim() || undefined,
    why: idea.why?.trim() || undefined,
  }
  if (!cleaned.title && !cleaned.angle && !cleaned.why) return
  pendingIdeas[target] = cleaned
}

export function consumePendingWritingIdea(target: WritingIdeaTarget): WritingIdeaHandoff | null {
  const idea = pendingIdeas[target]
  if (!idea) return null
  delete pendingIdeas[target]
  return idea
}

export function buildIdeaSeedComment(idea: WritingIdeaHandoff): string {
  const lines = ['<!--']
  if (idea.title) lines.push(`选题方向：${idea.title}`)
  if (idea.angle) lines.push(`切入角度：${idea.angle}`)
  if (idea.why) lines.push(`背景参考：${idea.why}`)
  lines.push('-->')
  return lines.join('\n')
}

export function buildIdeaOutlineInstruction(
  idea: WritingIdeaHandoff,
  kind: WritingIdeaTarget,
): string {
  const topicDesc = [idea.title, idea.angle].filter(Boolean).join('，')
  if (kind === 'note') {
    return `围绕「${topicDesc || '这段灵感'}」整理一份简短的手记写作提纲（2~3 个可展开的观察点与收尾思路）。请把全部内容包裹在一个 Markdown HTML 注释 <!-- ... --> 里输出，不要代写正文，方便我在下方直接动笔。`
  }
  return `围绕「${topicDesc || '这份选题灵感'}」扩写一份简洁实用的文章写作大纲（含核心立意、3~4 个递进小节要点与开篇切入建议）。请把全部内容包裹在一个 Markdown HTML 注释 <!-- ... --> 里输出，不要直接代写正文，方便我在正文里边对照边写。`
}
