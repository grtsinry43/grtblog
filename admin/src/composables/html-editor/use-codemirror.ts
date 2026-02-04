import { acceptCompletion } from '@codemirror/autocomplete'
import { html } from '@codemirror/lang-html'
import { EditorState, Compartment, type Extension } from '@codemirror/state'
import { EditorView, basicSetup } from 'codemirror'
import { type Ref, shallowRef, onMounted, onUnmounted } from 'vue'

import { codeMirrorTheme } from '@/composables/markdown-editor/codemirror-theme'
import '@/composables/markdown-editor/editor.css'
import {
    setTemplateVariables,
    templateHighlightExtension,
    templateTooltipExtension,
    templateVariablesField,
} from './template-highlight'

import { keymap, type ViewUpdate } from '@codemirror/view'

interface UseHtmlCodeMirrorProps {
    initialDoc?: string
    onChange?: (doc: string) => void
    readonly?: boolean
    extensions?: Extension[]
}

export type UpdateHook = (update: ViewUpdate) => void

export function useHtmlCodeMirror(container: Ref<HTMLElement | undefined>, props: UseHtmlCodeMirrorProps) {
    const view = shallowRef<EditorView>()
    const readonlyConfig = new Compartment()
    const updateCallbacks = new Set<UpdateHook>()

    const onViewUpdate = (callback: UpdateHook) => {
        updateCallbacks.add(callback)
        return () => updateCallbacks.delete(callback)
    }

    const eventBusExtension = EditorView.updateListener.of((update) => {
        updateCallbacks.forEach((cb) => cb(update))
        if (update.docChanged) {
            props.onChange?.(update.state.doc.toString())
        }
    })

    // Expose method to update valid variables
    const setVariables = (variables: string[]) => {
        view.value?.dispatch({
            effects: setTemplateVariables.of(variables),
        })
    }

    onMounted(() => {
        if (!container.value) return

        const startState = EditorState.create({
            doc: props.initialDoc || '',
            extensions: [
                basicSetup,
                keymap.of([{ key: 'Tab', run: acceptCompletion }]),
                html(),
                EditorView.lineWrapping,
                codeMirrorTheme,
                readonlyConfig.of(EditorState.readOnly.of(!!props.readonly)),
                templateVariablesField,
                templateHighlightExtension,
                templateTooltipExtension,
                eventBusExtension,
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
        onViewUpdate,
        setVariables,
    }
}
