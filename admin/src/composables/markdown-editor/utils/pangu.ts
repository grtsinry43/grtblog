const CJK =
  '\u2e80-\u2eff\u2f00-\u2fdf\u3040-\u309f\u30a0-\u30ff\u3100-\u312f\u3200-\u32ff\u3400-\u4dbf\u4e00-\u9fff\uf900-\ufaff'

const RE_CJK_ANS = new RegExp(`([${CJK}])([A-Za-z0-9])`, 'g')
const RE_ANS_CJK = new RegExp(`([A-Za-z0-9])([${CJK}])`, 'g')
const RE_CJK_CURRENCY = new RegExp(`([${CJK}])([$¥€£])(?=\\d)`, 'g')
const RE_PERCENT_CJK = new RegExp(`(\\d%)([${CJK}])`, 'g')
const RE_CJK_OPEN_PAREN = new RegExp(`([${CJK}])(\\([A-Za-z0-9])`, 'g')
const RE_CLOSE_PAREN_CJK = new RegExp(`([A-Za-z0-9]\\))([${CJK}])`, 'g')
const RE_ENDS_WITH_CJK = new RegExp(`[${CJK}]$`)
const RE_STARTS_WITH_CJK = new RegExp(`^[${CJK}]`)

/** 对纯文本片段执行盘古之白（中英文、数字与常用单位/括号间加空格） */
function spacingPlainText(text: string): string {
  if (!text) return text
  return text
    .replace(RE_CJK_ANS, '$1 $2')
    .replace(RE_ANS_CJK, '$1 $2')
    .replace(RE_CJK_CURRENCY, '$1 $2')
    .replace(RE_PERCENT_CJK, '$1 $2')
    .replace(RE_CJK_OPEN_PAREN, '$1 $2')
    .replace(RE_CLOSE_PAREN_CJK, '$1 $2')
}

/**
 * 匹配单行内需要保护的 Markdown 行内语法：
 * 1. 行内代码 `...`（支持多反引号）
 * 2. 行内公式 $...$
 * 3. 图片与链接 ![alt](url) / [text](url)（仅格式化方括号内文案，保护 URL）
 * 4. 裸 URL https://...
 * 5. HTML 标签 <tag ...>
 */
const INLINE_TOKEN_RE =
  /(`+)(?:[^`]|(?!\1)`)*?\1|\$[^$\n]+\$|(!?\[)([^\]]*)(\]\([^)\n]*\))|https?:\/\/[A-Za-z0-9\-._~:/?#[\]@!$&'*+,;=%]+|<\/?[a-zA-Z][^>\n]*>/g

function formatMarkdownLine(line: string): string {
  // 跳过自定义容器指令行与引用式链接定义
  const trimmed = line.trim()
  if (!trimmed || trimmed.startsWith(':::') || /^\[[^\]]+\]:\s*\S+/.test(trimmed)) {
    return line
  }

  // 保护行首 Markdown 缩进与列表/标题/引用前缀
  const prefixMatch = /^(\s*(?:#{1,6}\s+|>\s+|[-*+]\s+|\d+\.\s+)?)/.exec(line)
  const prefix = prefixMatch ? prefixMatch[1]! : ''
  const body = line.slice(prefix.length)

  let result = ''
  let lastIndex = 0
  let prevWasInlineCode = false

  INLINE_TOKEN_RE.lastIndex = 0
  let match: RegExpExecArray | null
  while ((match = INLINE_TOKEN_RE.exec(body)) !== null) {
    const token = match[0]
    const plain = body.slice(lastIndex, match.index)
    let formattedPlain = spacingPlainText(plain)

    const isInlineCode = token.startsWith('`')
    // 如果上一个 token 是行内代码，且当前纯文本以 CJK 开头，补一个空格
    if (prevWasInlineCode && formattedPlain && RE_STARTS_WITH_CJK.test(formattedPlain)) {
      formattedPlain = ` ${formattedPlain}`
    }
    // 如果当前 token 是行内代码，且前面纯文本以 CJK 结尾，补一个空格
    if (isInlineCode && formattedPlain && RE_ENDS_WITH_CJK.test(formattedPlain)) {
      formattedPlain = `${formattedPlain} `
    } else if (isInlineCode && !formattedPlain && result && RE_ENDS_WITH_CJK.test(result)) {
      result = `${result} `
    }

    result += formattedPlain

    // 链接 / 图片：保留 URL 不动，只格式化方括号内的文字
    if (match[2] !== undefined && match[3] !== undefined && match[4] !== undefined) {
      result += `${match[2]}${spacingPlainText(match[3])}${match[4]}`
    } else {
      result += token
    }

    prevWasInlineCode = isInlineCode
    lastIndex = match.index + token.length
  }

  let tail = spacingPlainText(body.slice(lastIndex))
  if (prevWasInlineCode && tail && RE_STARTS_WITH_CJK.test(tail)) {
    tail = ` ${tail}`
  }
  result += tail

  return prefix + result
}

/**
 * Markdown 感知的盘古格式化：
 * - 自动在中文与英文/数字/行内代码之间插入空格
 * - 完整保护围栏代码块、行内代码、数学公式、URL、HTML 标签与自定义指令行
 */
export function formatPanguMarkdown(input: string): string {
  if (!input) return input

  const lines = input.split('\n')
  const output: string[] = []
  let inCodeFence = false
  let fenceMarker = ''
  let inMathBlock = false

  for (const line of lines) {
    const trimmed = line.trim()

    // 围栏代码块状态切换（支持 ``` 和 ~~~）
    const fenceMatch = /^(`{3,}|~{3,})/.exec(trimmed)
    if (fenceMatch && !inMathBlock) {
      const marker = fenceMatch[1]!
      if (!inCodeFence) {
        inCodeFence = true
        fenceMarker = marker[0]!
      } else if (marker[0] === fenceMarker && trimmed.replace(new RegExp(`^${fenceMarker}+`), '').trim() === '') {
        inCodeFence = false
        fenceMarker = ''
      }
      output.push(line)
      continue
    }

    if (inCodeFence) {
      output.push(line)
      continue
    }

    // 多行公式块 $$ ... $$
    if (trimmed === '$$') {
      inMathBlock = !inMathBlock
      output.push(line)
      continue
    }

    if (inMathBlock) {
      output.push(line)
      continue
    }

    output.push(formatMarkdownLine(line))
  }

  return output.join('\n')
}
