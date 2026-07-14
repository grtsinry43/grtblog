import { onMounted, reactive, shallowRef } from 'vue'

import { createColumn, listColumns } from '@/services/taxonomy'
import { useContentTagSelect } from '@/views/shared/content-editor/composables/use-content-tag-select'

import type { MomentTopic } from '@/services/moments'
import type { MessageApi, SelectOption } from 'naive-ui'
import type { Ref } from 'vue'

export function useMomentTaxonomySelect(
  formTopicIds: Ref<number[]>,
  formColumnId: Ref<number | null>,
  message: MessageApi,
) {
  const columnOptions = shallowRef<SelectOption[]>([])
  const {
    options: topicOptions,
    loading: topicsLoading,
    creating: topicCreating,
    selectedItems: selectedTopics,
    loadOptions: loadTopicOptions,
    setInitialItems,
    createAndSelect: createAndSelectTopic,
  } = useContentTagSelect({ selectedIds: formTopicIds, noun: '话题', message })

  const newColumnModal = reactive({
    show: false,
    name: '',
    slug: '',
    loading: false,
  })

  async function fetchOptions() {
    try {
      const [columns] = await Promise.all([listColumns(), loadTopicOptions()])
      columnOptions.value = columns.map((column) => ({
        label: column.name,
        value: column.id,
      }))
    } catch (error) {
      console.error('Fetch moment taxonomy failed', error)
      message.error('加载分区失败')
    }
  }

  function setInitialTopics(topics: MomentTopic[]) {
    setInitialItems(topics)
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
      columnOptions.value = [...columnOptions.value, { label: res.name, value: res.id }]
      formColumnId.value = res.id
      newColumnModal.show = false
      newColumnModal.name = ''
      newColumnModal.slug = ''
    } catch {
      message.error('创建分区失败')
    } finally {
      newColumnModal.loading = false
    }
  }

  onMounted(fetchOptions)

  return {
    columnOptions,
    topicOptions,
    topicsLoading,
    topicCreating,
    selectedTopics,
    newColumnModal,
    setInitialTopics,
    createAndSelectTopic,
    createNewColumn,
  }
}
