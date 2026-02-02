import { reactive, computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { getPage, createPage, updatePage } from '@/services/page'

export function usePageForm() {
  const route = useRoute()
  const router = useRouter()
  const message = useMessage()

  const id = computed(() => route.params.id as string | undefined)
  const isCreating = computed(() => !id.value)
  const loading = ref(false)
  const saving = ref(false)

  const form = reactive({
    title: '',
    description: '',
    content: '',
    shortUrl: '',
    isEnabled: true,
  })

  async function fetch() {
    if (isCreating.value) {
      return null
    }
    loading.value = true
    try {
      const data = await getPage(Number(id.value))
      form.title = data.title
      form.description = data.description || ''
      form.content = data.content
      form.shortUrl = data.shortUrl
      form.isEnabled = data.isEnabled
      return data
    } finally {
      loading.value = false
    }
  }

  async function save() {
    saving.value = true
    try {
      const payload = {
        title: form.title,
        description: form.description,
        content: form.content,
        shortUrl: form.shortUrl,
        isEnabled: form.isEnabled,
      }
      if (isCreating.value) {
        await createPage(payload)
        message.success('页面创建成功')
      } else {
        await updatePage(Number(id.value), payload)
        message.success('页面更新成功')
      }
      router.push({ name: 'pageList' })
    } catch (error) {
      if (error instanceof Error) {
        message.error(error.message)
      } else {
        message.error('保存失败')
      }
    } finally {
      saving.value = false
    }
  }

  return { form, loading, saving, isCreating, fetch, save }
}
