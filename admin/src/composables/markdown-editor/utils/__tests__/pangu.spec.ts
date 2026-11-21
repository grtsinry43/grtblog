import { describe, expect, it } from 'vitest'

import { formatPanguMarkdown } from '../pangu'

describe('formatPanguMarkdown', () => {
  it('adds spaces between CJK and Latin/digits', () => {
    expect(formatPanguMarkdown('我在使用Vue3和Go1.24开发博客')).toBe(
      '我在使用 Vue3 和 Go1.24 开发博客',
    )
  })

  it('adds spaces around inline code adjacent to CJK without modifying code content', () => {
    expect(formatPanguMarkdown('调用`useState(中文123)`来更新状态')).toBe(
      '调用 `useState(中文123)` 来更新状态',
    )
  })

  it('preserves fenced code blocks and URLs while formatting link text', () => {
    const input = [
      '请看[Vue3官方文档](https://vuejs.org/guide/中文123)说明。',
      '```ts',
      'const msg = "中文123不加空格"',
      '```',
      '以及裸链接https://example.com/test123和后面2026年。',
    ].join('\n')

    const expected = [
      '请看[Vue3 官方文档](https://vuejs.org/guide/中文123)说明。',
      '```ts',
      'const msg = "中文123不加空格"',
      '```',
      '以及裸链接https://example.com/test123和后面 2026 年。',
    ].join('\n')

    expect(formatPanguMarkdown(input)).toBe(expected)
  })

  it('preserves markdown headings, lists, and custom directives', () => {
    const input = [
      '## 第1章Vue3入门',
      '- 支持TypeScript6类型检查',
      ':::note title="中文123"',
      '正文包含HTML标签<span>中文123</span>内容',
      ':::',
    ].join('\n')

    const expected = [
      '## 第 1 章 Vue3 入门',
      '- 支持 TypeScript6 类型检查',
      ':::note title="中文123"',
      '正文包含 HTML 标签<span>中文 123</span>内容',
      ':::',
    ].join('\n')

    expect(formatPanguMarkdown(input)).toBe(expected)
  })
})
