import { reactive, type Ref } from 'vue'
import type { EditorView } from '@codemirror/view'
import { searchFederationAuthors } from '@/services/federation-admin'
import type { FederationAuthorResp } from '@/types/federation'

export function useMentionPicker(view: Ref<EditorView | undefined>) {
  const state = reactive({
    show: false,
    query: '',
    results: [] as FederationAuthorResp[],
    loading: false,
  })

  function open() {
    state.show = true
    state.query = ''
    state.results = []
  }

  function close() {
    state.show = false
  }

  let debounceTimer: ReturnType<typeof setTimeout> | null = null

  async function search(query: string) {
    state.query = query
    if (debounceTimer) clearTimeout(debounceTimer)
    debounceTimer = setTimeout(async () => {
      state.loading = true
      try {
        const resp = await searchFederationAuthors(query, 20)
        state.results = resp.items ?? []
      } catch {
        state.results = []
      } finally {
        state.loading = false
      }
    }, 250)
  }

  function insert(author: FederationAuthorResp) {
    const v = view.value
    if (!v) return
    const hostname = extractHostname(author.instanceUrl)
    const text = `<@${author.name}@${hostname}>`
    const pos = v.state.selection.main.head
    v.dispatch({ changes: { from: pos, to: pos, insert: text } })
    v.focus()
    close()
  }

  function insertRaw(user: string, instance: string) {
    const v = view.value
    if (!v || !user.trim() || !instance.trim()) return
    const text = `<@${user.trim()}@${instance.trim()}>`
    const pos = v.state.selection.main.head
    v.dispatch({ changes: { from: pos, to: pos, insert: text } })
    v.focus()
    close()
  }

  return { state, open, close, search, insert, insertRaw }
}

function extractHostname(url: string): string {
  try {
    return new URL(url).hostname
  } catch {
    return url.replace(/^https?:\/\//, '').replace(/\/.*$/, '')
  }
}
