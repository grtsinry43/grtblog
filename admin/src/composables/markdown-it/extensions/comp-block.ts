import container from 'markdown-it-container'

import { markdownComponents, parseComponentInfo } from '@/composables/markdown/shared/components'

import type { MarkdownExtension } from '../types'

/**
 * 自定义语法扩展
 * 语法: ::: component <name> :::
 * 作用: 不渲染内容，只输出一个带有 data-component="name" 的占位符 div
 */
export const componentBlockExtension: MarkdownExtension = (md) => {
  const escapeRegExp = (value: string) => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

  const registerContainer = (
    containerName: string,
    resolveComponentInfo: (tokenInfo: string) => ReturnType<typeof parseComponentInfo>,
    validate?: (params: string) => boolean,
  ) => {
    md.use(container as any, containerName, {
      validate,
      render: (tokens: any[], idx: number) => {
        if (tokens[idx].nesting === 1) {
          const { name, attrs } = resolveComponentInfo(tokens[idx].info)
          const propsJson = JSON.stringify(attrs)
          const propsAttr =
            propsJson !== '{}' ? ` data-props="${md.utils.escapeHtml(propsJson)}"` : ''
          return `<div class="md-component-placeholder" data-component="${md.utils.escapeHtml(name)}"${propsAttr}>`
        }
        return '</div>\n'
      },
    })
  }

  // ::: component <name>
  registerContainer(
    'component',
    (info) => parseComponentInfo(info),
    (params) => {
      const trimmed = params.trim()
      return /^component\s+/.test(trimmed) && !trimmed.includes('{') && !trimmed.includes('}')
    },
  )

  // ::: <component-name> key="value"
  markdownComponents.forEach((component) => {
    const prefix = component.name
    const escaped = escapeRegExp(prefix)
    const prefixRe = new RegExp(`^(?:${escaped}|\\[${escaped}\\])(?:\\s|$)`)
    registerContainer(
      component.name,
      (info) => parseComponentInfo(info),
      (params) => {
        const trimmed = params.trim()
        return prefixRe.test(trimmed) && !trimmed.includes('{') && !trimmed.includes('}')
      },
    )
  })
}
