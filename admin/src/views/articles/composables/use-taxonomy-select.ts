import { onMounted, reactive, shallowRef, type Ref } from 'vue'

import { createCategory, listCategories } from '@/services/taxonomy'
import { useContentTagSelect } from '@/views/shared/content-editor/composables/use-content-tag-select'

import type { ArticleTag } from '@/services/articles'
import type { SelectOption, useMessage } from 'naive-ui'

export interface NewCategoryModalState {
  show: boolean
  name: string
  slug: string
  loading: boolean
}

export function useTaxonomySelect(
  formTagIds: Ref<number[]>,
  formCategoryId: Ref<number | null>,
  message: ReturnType<typeof useMessage>,
) {
  const categoryOptions = shallowRef<SelectOption[]>([])
  const {
    options: tagOptions,
    loading: tagsLoading,
    creating: tagCreating,
    selectedItems: selectedTags,
    loadOptions: loadTagOptions,
    setInitialItems,
    createAndSelect: createAndSelectTag,
  } = useContentTagSelect({ selectedIds: formTagIds, noun: '标签', message })

  onMounted(async () => {
    try {
      const [categories] = await Promise.all([listCategories(), loadTagOptions()])
      categoryOptions.value = categories.map((category) => ({
        label: category.name,
        value: category.id,
      }))
    } catch (error) {
      console.error('Fetch taxonomy failed', error)
      message.error('加载分类失败')
    }
  })

  function setInitialTags(tags: ArticleTag[]) {
    setInitialItems(tags)
  }

  const newCatModal = reactive<NewCategoryModalState>({
    show: false,
    name: '',
    slug: '',
    loading: false,
  })

  async function createNewCategory() {
    if (!newCatModal.name || !newCatModal.slug) return message.error('请填写完整')

    newCatModal.loading = true
    try {
      const res = await createCategory({ name: newCatModal.name, shortUrl: newCatModal.slug })
      categoryOptions.value = [...categoryOptions.value, { label: res.name, value: res.id }]
      formCategoryId.value = res.id
      message.success('分类创建成功')
      newCatModal.show = false
      newCatModal.name = ''
      newCatModal.slug = ''
    } catch {
      message.error('分类创建失败')
    } finally {
      newCatModal.loading = false
    }
  }

  return {
    categoryOptions,
    tagOptions,
    tagsLoading,
    tagCreating,
    selectedTags,
    newCatModal,
    setInitialTags,
    createAndSelectTag,
    createNewCategory,
  }
}
