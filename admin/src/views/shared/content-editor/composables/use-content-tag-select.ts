import { computed, shallowRef } from 'vue'

import { createTag, listTags } from '@/services/taxonomy'
import {
  appendUniqueId,
  findContentTagByName,
  mergeContentTagOptions,
  normalizeContentTagName,
} from '@/views/shared/content-editor/model/content-tags'

import type { ContentTagItem } from '@/views/shared/content-editor/model/content-tags'
import type { MessageApi } from 'naive-ui'
import type { Ref } from 'vue'

interface ContentTagApi {
  list: () => Promise<ContentTagItem[]>
  create: (name: string) => Promise<ContentTagItem>
}

interface UseContentTagSelectOptions {
  selectedIds: Ref<number[]>
  noun: '标签' | '话题'
  message: Pick<MessageApi, 'error' | 'success'>
  api?: ContentTagApi
}

const defaultApi: ContentTagApi = {
  list: listTags,
  create: createTag,
}

export function useContentTagSelect({
  selectedIds,
  noun,
  message,
  api = defaultApi,
}: UseContentTagSelectOptions) {
  const options = shallowRef<ReturnType<typeof mergeContentTagOptions>>([])
  const loading = shallowRef(false)
  const creating = shallowRef(false)

  const selectedItems = computed<ContentTagItem[]>(() => {
    const optionById = new Map(options.value.map((option) => [option.value, option]))
    return selectedIds.value.map((id) => ({
      id,
      name: optionById.get(id)?.label ?? `${noun} ${id}`,
    }))
  })

  function mergeItems(items: ContentTagItem[]) {
    options.value = mergeContentTagOptions(options.value, items)
  }

  async function loadOptions({ silent = false }: { silent?: boolean } = {}) {
    loading.value = true
    try {
      mergeItems(await api.list())
      return true
    } catch {
      if (!silent) message.error(`加载${noun}失败`)
      return false
    } finally {
      loading.value = false
    }
  }

  function setInitialItems(items: ContentTagItem[]) {
    mergeItems(items)
    selectedIds.value = items.map((item) => item.id)
  }

  async function createAndSelect(rawName: string) {
    const name = normalizeContentTagName(rawName)
    if (!name || creating.value) return null

    const existing = findContentTagByName(options.value, name)
    if (existing) {
      selectedIds.value = appendUniqueId(selectedIds.value, existing.value)
      return { id: existing.value, name: existing.label }
    }

    creating.value = true
    try {
      const created = await api.create(name)
      mergeItems([created])
      selectedIds.value = appendUniqueId(selectedIds.value, created.id)
      message.success(`${noun}“${created.name}”已创建并选中`)
      return created
    } catch {
      // 可能是其他页面刚创建了同名项：刷新一次，避免用户重复提交。
      await loadOptions({ silent: true })
      const concurrentlyCreated = findContentTagByName(options.value, name)
      if (concurrentlyCreated) {
        selectedIds.value = appendUniqueId(selectedIds.value, concurrentlyCreated.value)
        return { id: concurrentlyCreated.value, name: concurrentlyCreated.label }
      }
      message.error(`创建${noun}“${name}”失败，请重试`)
      return null
    } finally {
      creating.value = false
    }
  }

  return {
    options,
    loading,
    creating,
    selectedItems,
    loadOptions,
    setInitialItems,
    createAndSelect,
  }
}
