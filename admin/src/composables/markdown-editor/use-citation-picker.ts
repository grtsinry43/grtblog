import { reactive, type Ref } from 'vue'
import type { EditorView } from '@codemirror/view'
import { getFederationInstances, fetchRemotePosts } from '@/services/federation-admin'
import type { FederationInstanceResp, FederationRemotePostResp } from '@/types/federation'

/** URL 格式校验：必须是合法域名或 http(s) URL */
const URL_PATTERN = /^(https?:\/\/)?[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)+(\:\d{1,5})?(\/.*)?$/

const PAGE_SIZE = 20

export function useCitationPicker(view: Ref<EditorView | undefined>) {
  const state = reactive({
    show: false,
    // URL 输入
    urlInput: '',
    urlValid: false,
    urlError: '',
    // 已有实例（快捷选项）
    instances: [] as FederationInstanceResp[],
    instancesLoading: false,
    // 文章列表
    posts: [] as FederationRemotePostResp[],
    postsLoading: false,
    searchQuery: '',
    // 分页
    page: 1,
    total: 0,
    pageSize: PAGE_SIZE,
    // 当前选中的远端
    resolvedURL: '',
    resolvedName: '',
    // 步骤
    step: 'input' as 'input' | 'posts',
  })

  let searchDebounce: ReturnType<typeof setTimeout> | null = null

  async function open() {
    state.show = true
    state.step = 'input'
    state.urlInput = ''
    state.urlValid = false
    state.urlError = ''
    state.posts = []
    state.searchQuery = ''
    state.page = 1
    state.total = 0
    state.resolvedURL = ''
    state.resolvedName = ''
    // 后台加载已有实例列表（快捷选项）
    state.instancesLoading = true
    try {
      const resp = await getFederationInstances({ pageSize: 50 })
      state.instances = (resp.items ?? []).filter(i => i.status === 'active')
    } catch {
      state.instances = []
    } finally {
      state.instancesLoading = false
    }
  }

  function close() {
    state.show = false
    if (searchDebounce) clearTimeout(searchDebounce)
  }

  function validateURL(input: string): boolean {
    const trimmed = input.trim()
    if (!trimmed) {
      state.urlValid = false
      state.urlError = ''
      return false
    }
    if (!URL_PATTERN.test(trimmed)) {
      state.urlValid = false
      state.urlError = '请输入有效的域名或 URL'
      return false
    }
    state.urlValid = true
    state.urlError = ''
    return true
  }

  function onURLInput(value: string) {
    state.urlInput = value
    validateURL(value)
  }

  /** 拉取远端文章（核心方法） */
  async function loadPosts(url: string, query: string, page: number) {
    const normalized = normalizeURL(url)
    if (!normalized) return
    state.resolvedURL = normalized
    state.resolvedName = extractHostname(normalized)
    state.step = 'posts'
    state.postsLoading = true
    state.page = page
    try {
      const resp = await fetchRemotePosts(normalized, query, page, PAGE_SIZE)
      state.posts = resp.items ?? []
      state.total = resp.total ?? 0
    } catch {
      state.posts = []
      state.total = 0
    } finally {
      state.postsLoading = false
    }
  }

  /** 用户确认 URL 输入后拉取 */
  function submitURL() {
    if (!validateURL(state.urlInput)) return
    state.searchQuery = ''
    loadPosts(state.urlInput.trim(), '', 1)
  }

  /** 选择已有实例快捷选项 */
  function selectInstance(inst: FederationInstanceResp) {
    state.urlInput = inst.base_url
    state.urlValid = true
    state.urlError = ''
    state.resolvedName = inst.name || extractHostname(inst.base_url)
    state.searchQuery = ''
    loadPosts(inst.base_url, '', 1)
  }

  /** 文章搜索（防抖 400ms，重置到第 1 页） */
  function searchPosts(query: string) {
    state.searchQuery = query
    if (searchDebounce) clearTimeout(searchDebounce)
    searchDebounce = setTimeout(() => {
      if (state.resolvedURL) {
        loadPosts(state.resolvedURL, query, 1)
      }
    }, 400)
  }

  /** 翻页 */
  function goToPage(page: number) {
    if (page < 1 || !state.resolvedURL) return
    loadPosts(state.resolvedURL, state.searchQuery, page)
  }

  /** 选择文章插入引用标记 */
  function insert(post: FederationRemotePostResp) {
    const v = view.value
    if (!v) return
    const hostname = extractHostname(state.resolvedURL || post.instance_url)
    const text = `<cite:${hostname}|${post.id}>`
    const pos = v.state.selection.main.head
    v.dispatch({ changes: { from: pos, to: pos, insert: text } })
    v.focus()
    close()
  }

  /** 手动输入插入 */
  function insertRaw(instance: string, postId: string) {
    const v = view.value
    if (!v || !instance.trim() || !postId.trim()) return
    const text = `<cite:${instance.trim()}|${postId.trim()}>`
    const pos = v.state.selection.main.head
    v.dispatch({ changes: { from: pos, to: pos, insert: text } })
    v.focus()
    close()
  }

  function back() {
    state.step = 'input'
    state.posts = []
    state.searchQuery = ''
    state.page = 1
    state.total = 0
  }

  return {
    state,
    open,
    close,
    onURLInput,
    submitURL,
    selectInstance,
    searchPosts,
    goToPage,
    insert,
    insertRaw,
    back,
  }
}

function extractHostname(url: string): string {
  try {
    return new URL(url).hostname
  } catch {
    return url.replace(/^https?:\/\//, '').replace(/[:/].*$/, '')
  }
}

function normalizeURL(raw: string): string {
  const trimmed = raw.trim()
  if (!trimmed) return ''
  if (trimmed.startsWith('http://') || trimmed.startsWith('https://')) {
    return trimmed.replace(/\/+$/, '')
  }
  return 'https://' + trimmed.replace(/\/+$/, '')
}
