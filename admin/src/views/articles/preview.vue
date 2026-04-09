<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import MarkdownPreview from '@/components/markdown-editor/MarkdownPreview.vue'
import { getArticle } from '@/services/articles'

const content = ref<string>('')

// 组件prop传入的文章ID
const props = defineProps<{
  articleId: number | null
}>()

const isLoading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
  if (props.articleId === null) {
    error.value = '无效的文章 ID'
    isLoading.value = false
    return
  }

  try {
    const article = await getArticle(props.articleId)
    content.value = article.content
  } catch (err) {
    error.value = '获取文章失败'
  } finally {
    isLoading.value = false
  }
})
</script>

<template>
  <MarkdownPreview
    :source="content"
    v-if="!isLoading && !error"
    class="p-4 text-sm sm:p-8"
  />
  <div
    v-else-if="isLoading"
    class="p-4 text-center text-gray-500 sm:p-8"
  >
    加载中...
  </div>
  <div
    v-else
    class="p-4 text-center text-red-500 sm:p-8"
  >
    {{ error }}
  </div>
</template>
