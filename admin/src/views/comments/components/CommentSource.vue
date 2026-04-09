<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { DocumentTextOutline, NewspaperOutline, ChatbubbleEllipsesOutline } from '@vicons/ionicons5'
import { NSpin, NTag, NText, NIcon } from 'naive-ui'
import { computed } from 'vue'

import { getArticle } from '@/services/articles'
import { getMoment } from '@/services/moments'
import { getPage } from '@/services/page'

const props = defineProps<{
  type?: string
  id?: number
  initialTitle?: string
}>()

const isArticle = computed(() => props.type === 'posts' || props.type === 'article')
const isPage = computed(() => props.type === 'pages' || props.type === 'page')
const isMoment = computed(() => props.type === 'moments' || props.type === 'moment')

const { data: article, isLoading: isLoadingArticle } = useQuery({
  queryKey: ['article', props.id],
  queryFn: () => getArticle(props.id!),
  enabled: computed(() => !!props.id && isArticle.value),
  staleTime: 1000 * 60 * 5, // 5 mins
})

const { data: page, isLoading: isLoadingPage } = useQuery({
  queryKey: ['page', props.id],
  queryFn: () => getPage(props.id!),
  enabled: computed(() => !!props.id && isPage.value),
  staleTime: 1000 * 60 * 5,
})

const { data: moment, isLoading: isLoadingMoment } = useQuery({
  queryKey: ['moment', props.id],
  queryFn: () => getMoment(props.id!),
  enabled: computed(() => !!props.id && isMoment.value),
  staleTime: 1000 * 60 * 5,
})

const displayTitle = computed(() => {
  if (isArticle.value && article.value) return article.value.title
  if (isPage.value && page.value) return page.value.title
  if (isMoment.value && moment.value) return moment.value.title || '动态'
  return props.initialTitle || '未知来源'
})

const displayLink = computed(() => {
  // In a real app, you might want to link to the public site
  // For now, we can perhaps link to the admin edit page or just perform no action
  // If we had the public URL base, we could link there.
  return null
})

const typeTag = computed(() => {
  if (isArticle.value) return { type: 'info' as const, icon: DocumentTextOutline, label: '文章' }
  if (isPage.value) return { type: 'success' as const, icon: NewspaperOutline, label: '页面' }
  if (isMoment.value)
    return { type: 'warning' as const, icon: ChatbubbleEllipsesOutline, label: '动态' }
  return { type: 'default' as const, icon: undefined, label: props.type || '其他' }
})
</script>

<template>
  <div class="flex items-center gap-2">
    <n-tag
      :type="typeTag.type"
      size="small"
      :bordered="false"
      class="flex items-center"
    >
      {{ typeTag.label }}
      <template
        #icon
        v-if="typeTag.icon"
      >
        <n-icon :component="typeTag.icon" />
      </template>
    </n-tag>

    <n-spin
      v-if="isLoadingArticle || isLoadingPage || isLoadingMoment"
      size="small"
    />

    <n-text
      v-else
      class="max-w-[200px] truncate text-sm"
      :title="displayTitle"
    >
      {{ displayTitle }}
    </n-text>
  </div>
</template>
