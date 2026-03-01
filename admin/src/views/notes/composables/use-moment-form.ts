import { useMessage } from 'naive-ui'
import { reactive, ref, computed, onMounted, toRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useLeaveConfirm } from '@/composables'
import { createMoment, getMoment, updateMoment } from '@/services/moments'
import type { ContentExtInfo } from '@/types/ext-info'
import { useImageExtInfo } from '@/composables/use-image-ext-info'

function joinImages(images?: string[]) {
  return (images ?? []).join('\n')
}

function splitImages(value: string) {
  return value
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean)
}

export function useMomentForm() {
  const route = useRoute()
  const router = useRouter()
  const message = useMessage()

  const momentId = computed(() => {
    const param = route.params.id
    if (!param || param === 'new') return null
    const id = Number(param)
    return Number.isFinite(id) ? id : null
  })

  const isCreating = computed(() => momentId.value === null)
  const loading = ref(false)
  const saving = ref(false)
  const initialSnapshot = ref('')

  const form = reactive({
    title: '',
    summary: '',
    aiSummary: null as string | null,
    content: '',
    image: '',
    columnId: null as number | null,
    topicIds: [] as number[],
    shortUrl: '',
    isPublished: false,
    isTop: false,
    isOriginal: true,
    allowComment: true,
  })

  const baseExtInfo = ref<ContentExtInfo | null>(null)
  const { extInfo, processing } = useImageExtInfo({
    content: toRef(form, 'content'),
    extraImages: toRef(form, 'image'),
    baseExtInfo,
  })

  const takeSnapshot = () => JSON.stringify(form)
  const isDirty = computed(
    () => initialSnapshot.value !== '' && takeSnapshot() !== initialSnapshot.value,
  )

  async function fetch() {
    if (isCreating.value) {
      initialSnapshot.value = takeSnapshot()
      return null
    }

    loading.value = true
    try {
      const data = await getMoment(momentId.value!)

      form.title = data.title
      form.summary = data.summary || ''
      form.aiSummary = data.aiSummary ?? null
      form.content = data.content
      form.image = joinImages(data.image)
      form.columnId = data.columnId ?? null
      form.topicIds = data.topics?.map((t) => t.id) ?? []
      form.shortUrl = data.shortUrl
      form.isPublished = data.isPublished
      form.isTop = data.isTop
      form.isOriginal = data.isOriginal
      form.allowComment = data.allowComment
      baseExtInfo.value = data.extInfo ?? null

      initialSnapshot.value = takeSnapshot()
      return data
    } catch (e) {
      console.error(e)
      message.error('无法加载手记数据')
      router.replace({ name: 'noteList' })
      return null
    } finally {
      loading.value = false
    }
  }

  async function save() {
    if (!form.title.trim()) return message.error('请输入标题')
    if (!form.content.trim()) return message.error('请输入正文内容')
    if (!isCreating.value && !form.shortUrl.trim()) return message.error('短链接不能为空')

    saving.value = true
    try {
      const images = splitImages(form.image)
      const basePayload = {
        title: form.title,
        summary: form.summary,
        aiSummary: form.aiSummary || null,
        content: form.content,
        image: images.length ? images : undefined,
        columnId: form.columnId ?? undefined,
        topicIds: form.topicIds.length ? form.topicIds : undefined,
        isPublished: form.isPublished,
        isTop: form.isTop,
        isOriginal: form.isOriginal,
        allowComment: form.allowComment,
        extInfo: extInfo.value ?? undefined,
      }

      if (isCreating.value) {
        await createMoment({
          ...basePayload,
          shortUrl: form.shortUrl || undefined,
        })
        message.success('创建成功')
      } else {
        await updateMoment(momentId.value!, {
          ...basePayload,
          shortUrl: form.shortUrl,
        })
        message.success('更新成功')
      }

      initialSnapshot.value = takeSnapshot()
      router.push({ name: 'noteList' })
    } catch (e: any) {
      message.error(e.message || '保存失败')
    } finally {
      saving.value = false
    }
  }

  useLeaveConfirm({
    when: isDirty,
    title: '未保存的更改',
    content: '当前内容未保存，确定要离开吗？',
    positiveText: '离开',
    negativeText: '继续编辑',
  })

  onMounted(fetch)

  return {
    form,
    loading,
    saving,
    imageProcessing: processing,
    extInfo,
    isCreating,
    isDirty,
    fetch,
    save,
  }
}
