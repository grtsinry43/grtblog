import { StateEffect, StateField } from '@codemirror/state'
import { Decoration, EditorView, WidgetType, type DecorationSet } from '@codemirror/view'

export const addUpload = StateEffect.define<{ id: string; pos: number }>()
export const removeUpload = StateEffect.define<{ id: string }>()

class UploadWidget extends WidgetType {
  toDOM() {
    const container = document.createElement('span')
    container.className = 'cm-uploading-widget'

    const text = document.createElement('span')
    text.className = 'cm-upload-shimmer'
    text.textContent = '图片上传中...'

    container.appendChild(text)
    return container
  }

  ignoreEvent() {
    return false
  }
}

export const uploadStateField = StateField.define<DecorationSet>({
  create() {
    return Decoration.none
  },
  update(uploads, tr) {
    uploads = uploads.map(tr.changes)

    for (const effect of tr.effects) {
      if (effect.is(addUpload)) {
        const decoration = Decoration.widget({
          widget: new UploadWidget(),
          side: 1,
          id: effect.value.id,
        })
        uploads = uploads.update({
          add: [decoration.range(effect.value.pos)],
        })
      } else if (effect.is(removeUpload)) {
        uploads = uploads.update({
          filter: (_from, _to, value) => value.spec.id !== effect.value.id,
        })
      }
    }
    return uploads
  },
  provide: (f) => EditorView.decorations.from(f),
})
