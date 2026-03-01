import { reactive, type Ref } from 'vue'
import type { EditorView } from '@codemirror/view'
import { getFederationInstances, listFederationInstancePosts } from '@/services/federation-admin'
import type { FederationInstanceResp, FederationCachedPostResp } from '@/types/federation'

export function useCitationPicker(view: Ref<EditorView | undefined>) {
  const state = reactive({
    show: false,
    step: 'instance' as 'instance' | 'post',
    instances: [] as FederationInstanceResp[],
    instanceFilter: '',
    selectedInstance: null as FederationInstanceResp | null,
    posts: [] as FederationCachedPostResp[],
    searchQuery: '',
    loading: false,
  })

  async function open() {
    state.show = true
    state.step = 'instance'
    state.instances = []
    state.instanceFilter = ''
    state.selectedInstance = null
    state.posts = []
    state.searchQuery = ''
    state.loading = true
    try {
      const resp = await getFederationInstances({ pageSize: 100 })
      state.instances = resp.items ?? []
    } catch {
      state.instances = []
    } finally {
      state.loading = false
    }
  }

  function close() {
    state.show = false
  }

  async function selectInstance(inst: FederationInstanceResp) {
    state.selectedInstance = inst
    state.step = 'post'
    state.searchQuery = ''
    await loadPosts('')
  }

  let debounceTimer: ReturnType<typeof setTimeout> | null = null

  async function searchPosts(query: string) {
    state.searchQuery = query
    if (debounceTimer) clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => loadPosts(query), 250)
  }

  async function loadPosts(query: string) {
    if (!state.selectedInstance) return
    state.loading = true
    try {
      const resp = await listFederationInstancePosts(state.selectedInstance.id, query, 20)
      state.posts = resp.items ?? []
    } catch {
      state.posts = []
    } finally {
      state.loading = false
    }
  }

  function insert(post: FederationCachedPostResp) {
    const v = view.value
    if (!v || !state.selectedInstance) return
    const hostname = extractHostname(state.selectedInstance.base_url)
    const postId = post.remotePostId || String(post.id)
    const text = `<cite:${hostname}|${postId}>`
    const pos = v.state.selection.main.head
    v.dispatch({ changes: { from: pos, to: pos, insert: text } })
    v.focus()
    close()
  }

  function back() {
    state.step = 'instance'
    state.selectedInstance = null
    state.posts = []
    state.searchQuery = ''
  }

  function insertRaw(instance: string, postId: string) {
    const v = view.value
    if (!v || !instance.trim() || !postId.trim()) return
    const text = `<cite:${instance.trim()}|${postId.trim()}>`
    const pos = v.state.selection.main.head
    v.dispatch({ changes: { from: pos, to: pos, insert: text } })
    v.focus()
    close()
  }

  return { state, open, close, selectInstance, searchPosts, insert, back, insertRaw }
}

function extractHostname(url: string): string {
  try {
    return new URL(url).hostname
  } catch {
    return url.replace(/^https?:\/\//, '').replace(/\/.*$/, '')
  }
}
