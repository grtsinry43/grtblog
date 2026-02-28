import { reactive, computed, type Ref } from 'vue'

import {
  getMarkdownComponent,
  serializeComponentAttributes,
} from '@/composables/markdown/shared/components'
import type { ComponentEditPayload } from '@/composables/markdown-editor/extensions/component-editor'

import type { EditorView } from 'codemirror'

export function useComponentInserter(view: Ref<EditorView | undefined>) {
  const state = reactive({
    show: false,
    name: '',
    blockFrom: 0,
    blockTo: 0,
    isComponentSyntax: false,
    hasClosingFence: false,
    attrs: {} as Record<string, string>,
    body: '',
  })

  const touchedKeys = new Set<string>()
  const componentMeta = computed(() => getMarkdownComponent(state.name))

  const open = (payload: ComponentEditPayload) => {
    state.name = payload.name
    state.blockFrom = payload.blockFrom
    state.blockTo = payload.blockTo
    state.isComponentSyntax = payload.isComponentSyntax
    state.hasClosingFence = payload.hasClosingFence
    state.attrs = { ...payload.attrs }
    state.body = payload.body
    touchedKeys.clear()
    state.show = true
  }

  const formatLine = () => {
    const meta = componentMeta.value
    const definedKeys = meta?.attrs?.map((attr) => attr.key) ?? []
    const extraKeys = Object.keys(state.attrs).filter((key) => !definedKeys.includes(key))

    const keys = [...definedKeys, ...extraKeys].filter(
      (key) => key in state.attrs || touchedKeys.has(key),
    )

    const attrsByKey = Object.fromEntries(keys.map((key) => [key, state.attrs[key] ?? '']))
    const attrsString = serializeComponentAttributes(attrsByKey, keys)

    const prefix = state.isComponentSyntax ? `::: component ${state.name}` : `::: ${state.name}`

    return attrsString ? `${prefix} ${attrsString}` : prefix
  }

  const formatBlock = () => {
    const header = formatLine()
    const normalizedBody = state.body.replace(/\r\n?/g, '\n')
    if (normalizedBody) {
      return `${header}\n${normalizedBody}\n:::`
    }
    if (componentMeta.value?.body || state.hasClosingFence) {
      return `${header}\n\n:::`
    }
    return header
  }

  const apply = () => {
    if (!view.value || !state.show) return
    const newLine = formatBlock()
    view.value.dispatch({
      changes: { from: state.blockFrom, to: state.blockTo, insert: newLine },
    })
    // 更新选中范围，防止连续编辑位置错乱
    state.blockTo = state.blockFrom + newLine.length
  }

  const updateAttr = (key: string, value: string | boolean) => {
    touchedKeys.add(key)
    state.attrs[key] = String(value)
    apply()
  }

  const updateBody = (value: string) => {
    state.body = value
    apply()
  }

  return {
    state,
    componentMeta,
    open,
    updateAttr,
    updateBody,
    close: () => (state.show = false),
  }
}
