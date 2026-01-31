import { autocompletion } from '@codemirror/autocomplete'
import { markdown } from '@codemirror/lang-markdown'
import { EditorState, Compartment, type Extension } from '@codemirror/state'
import { GFM, Subscript, Superscript } from '@lezer/markdown'
import { EditorView, basicSetup } from 'codemirror'
import { type Ref, shallowRef, onMounted, onUnmounted } from 'vue'

import { codeMirrorTheme } from './codemirror-theme'
import { componentAttributeSource } from './extensions/component-attrs'
import {
  createComponentEditorExtension,
  type ComponentEditPayload,
} from './extensions/component-editor'
import { customBlockExtension } from './extensions/custom-block'
import { slashCommandSource } from './extensions/slash-command'
import './editor.css'
import { slashHintExtension } from './extensions/slash-hint'
import { addUpload, removeUpload, uploadStateField } from './use-upload-extensions'

import type { ViewUpdate } from '@codemirror/view'

interface UseCodeMirrorProps {
  initialDoc?: string
  onChange?: (doc: string) => void
  readonly?: boolean
  onComponentEdit?: (payload: ComponentEditPayload) => void
  onUploadImage?: (file: File) => Promise<string>
  // 允许传入额外的 extensions (如果需要)
  extensions?: Extension[]
}

// 定义钩子类型
export type UpdateHook = (update: ViewUpdate) => void

export function useCodeMirror(container: Ref<HTMLElement | undefined>, props: UseCodeMirrorProps) {
  const view = shallowRef<EditorView>()
  const readonlyConfig = new Compartment()

  const updateCallbacks = new Set<UpdateHook>()
  const handleImageUpload = (
    file: File,
    editor: EditorView,
    insertPos: number,
  ) => {
    if (!props.onUploadImage) return

    const uploadId =
      typeof crypto !== 'undefined' && 'randomUUID' in crypto
        ? crypto.randomUUID()
        : `upload_${Date.now()}_${Math.random().toString(16).slice(2)}`

    editor.dispatch({
      effects: addUpload.of({ id: uploadId, pos: insertPos }),
    })

    props
      .onUploadImage(file)
      .then((url) => {
        const state = editor.state
        const decorations = state.field(uploadStateField)
        let foundFrom: number | null = null

        decorations.between(0, state.doc.length, (from, _to, value) => {
          if (value.spec.id === uploadId) {
            foundFrom = from
            return false
          }
        })

        if (foundFrom !== null) {
          editor.dispatch({
            changes: {
              from: foundFrom,
              insert: `![${file.name}](${url})`,
            },
            effects: removeUpload.of({ id: uploadId }),
          })
        }
      })
      .catch(() => {
        editor.dispatch({
          effects: removeUpload.of({ id: uploadId }),
        })
      })
  }

  // 对外暴露的注册函数
  const onViewUpdate = (callback: UpdateHook) => {
    updateCallbacks.add(callback)
    return () => updateCallbacks.delete(callback) // 返回清理函数
  }

  // 内部扩展：统一分发更新事件
  const eventBusExtension = EditorView.updateListener.of((update) => {
    // 触发所有订阅者
    updateCallbacks.forEach((cb) => cb(update))

    // 处理原有的 onChange
    if (update.docChanged) {
      props.onChange?.(update.state.doc.toString())
    }
  })

  onMounted(() => {
    if (!container.value) return

    const startState = EditorState.create({
      doc: props.initialDoc || '',
      extensions: [
        uploadStateField,
        basicSetup,
        markdown({
          extensions: [GFM, Subscript, Superscript],
        }),
        EditorView.lineWrapping,
        codeMirrorTheme,
        readonlyConfig.of(EditorState.readOnly.of(!!props.readonly)),

        // 功能扩展
        autocompletion({
          override: [slashCommandSource, componentAttributeSource],
          icons: false,
          defaultKeymap: true,
        }),
        customBlockExtension,
        slashHintExtension,
        createComponentEditorExtension({ onEdit: props.onComponentEdit }),

        // 注册事件总线扩展
        eventBusExtension,

        EditorView.domEventHandlers({
          paste: (event, editor) => {
            if (!props.onUploadImage) return
            const files = event.clipboardData?.files
            if (!files || files.length === 0) return
            const imageFiles = Array.from(files).filter((file) =>
              file.type.startsWith('image/'),
            )
            if (!imageFiles.length) return
            event.preventDefault()
            const insertPos = editor.state.selection.main.head
            imageFiles.forEach((file) => handleImageUpload(file, editor, insertPos))
          },
          drop: (event, editor) => {
            if (!props.onUploadImage) return
            const files = event.dataTransfer?.files
            if (!files || files.length === 0) return
            const imageFiles = Array.from(files).filter((file) =>
              file.type.startsWith('image/'),
            )
            if (!imageFiles.length) return
            event.preventDefault()
            const pos = editor.posAtCoords({ x: event.clientX, y: event.clientY })
            const insertPos = pos ?? editor.state.selection.main.head
            if (pos !== null) {
              editor.dispatch({ selection: { anchor: insertPos } })
            }
            imageFiles.forEach((file) => handleImageUpload(file, editor, insertPos))
          },
        }),

        // 合并外部传入的扩展
        ...(props.extensions || []),
      ],
    })

    view.value = new EditorView({
      state: startState,
      parent: container.value,
    })
  })

  onUnmounted(() => {
    view.value?.destroy()
    updateCallbacks.clear()
  })

  return {
    view,
    onViewUpdate, // 导出钩子
  }
}
