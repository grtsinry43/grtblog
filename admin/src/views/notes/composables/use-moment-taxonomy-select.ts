import { ref, onMounted, reactive, computed } from 'vue'

import { listColumns, listTags, createTag, createColumn } from '@/services/taxonomy'

import type { MomentTopic } from '@/services/moments'
import type { SelectOption, MessageApi } from 'naive-ui'
import type { Ref } from 'vue'

export function useMomentTaxonomySelect(
  formTopicIds: Ref<number[]>,
  formColumnId: Ref<number | null>,
  message: MessageApi,
) {
  const columnOptions = ref<SelectOption[]>([])
  const topicOptions = ref<SelectOption[]>([])
  const dynamicTopics = ref<string[]>([])
  const topicSearchValue = ref('')

  const newColumnModal = reactive({
    show: false,
    name: '',
    slug: '',
    loading: false,
  })

  async function fetchOptions() {
    try {
      const [columns, topics] = await Promise.all([listColumns(), listTags()])
      columnOptions.value = columns.map((c) => ({ label: c.name, value: c.id }))
      topicOptions.value = topics.map((t) => ({ label: t.name, value: t.id }))
    } catch (e) {
      console.error('Fetch moment taxonomy failed', e)
    }
  }

  function setInitialTopics(topics: MomentTopic[]) {
    dynamicTopics.value = topics.map((t) => t.name)
    formTopicIds.value = topics.map((t) => t.id)
  }

  async function handleTopicsChange(newTopics: string[]) {
    const ids: number[] = []
    const nextDynamicTopics: string[] = []

    for (const topicStr of newTopics) {
      const trimmed = topicStr.trim()
      if (!trimmed) continue

      const existing = topicOptions.value.find((t) => t.label === trimmed)
      if (existing) {
        ids.push(existing.value as number)
        nextDynamicTopics.push(trimmed)
      } else {
        try {
          const created = await createTag(trimmed)
          topicOptions.value.push({ label: created.name, value: created.id })
          ids.push(created.id)
          nextDynamicTopics.push(created.name)
        } catch (e) {
          message.error('创建话题失败')
        }
      }
    }

    dynamicTopics.value = nextDynamicTopics
    formTopicIds.value = ids
  }

  function addTopicFromSearch(value: string) {
    if (!value?.trim()) return
    if (!dynamicTopics.value.includes(value)) {
      handleTopicsChange([...dynamicTopics.value, value])
    }
    topicSearchValue.value = ''
  }

  async function createNewColumn() {
    if (!newColumnModal.name.trim()) return message.error('请输入分区名称')
    if (!newColumnModal.slug.trim()) return message.error('请输入分区短链接')

    newColumnModal.loading = true
    try {
      const res = await createColumn({
        name: newColumnModal.name,
        shortUrl: newColumnModal.slug,
      })
      columnOptions.value.push({ label: res.name, value: res.id })
      formColumnId.value = res.id
      newColumnModal.show = false
      newColumnModal.name = ''
      newColumnModal.slug = ''
    } catch (e) {
      message.error('创建分区失败')
    } finally {
      newColumnModal.loading = false
    }
  }

  const autoCompleteOptions = computed(() => {
    return topicOptions.value.map((t) => ({ label: t.label as string, value: t.label as string }))
  })

  onMounted(fetchOptions)

  return {
    columnOptions,
    topicOptions,
    dynamicTopics,
    topicSearchValue,
    autoCompleteOptions,
    newColumnModal,
    setInitialTopics,
    handleTopicsChange,
    addTopicFromSearch,
    createNewColumn,
  }
}
