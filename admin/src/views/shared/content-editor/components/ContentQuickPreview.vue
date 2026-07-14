<script setup lang="ts">
import { NSpin, NTooltip } from 'naive-ui'
import { computed, shallowRef, watch } from 'vue'

import { ScrollContainer } from '@/components'
import MarkdownPreview from '@/components/markdown-editor/MarkdownPreview.vue'
import { getArticle } from '@/services/articles'
import { getMoment } from '@/services/moments'
import { getPage } from '@/services/page'

export type QuickPreviewContentType = 'article' | 'moment' | 'page'

const props = defineProps<{
  contentId: number
  contentType: QuickPreviewContentType
}>()

const source = shallowRef('')
const status = shallowRef<'idle' | 'loading' | 'success' | 'error'>('idle')
const loadedKey = shallowRef('')
let requestVersion = 0

const contentKey = computed(() => `${props.contentType}:${props.contentId}`)

async function loadContent() {
  const key = contentKey.value
  if (status.value === 'loading' || loadedKey.value === key) return

  const version = ++requestVersion
  status.value = 'loading'
  try {
    const content = await loadContentByType(props.contentType, props.contentId)
    if (version !== requestVersion) return
    source.value = content
    loadedKey.value = key
    status.value = 'success'
  } catch {
    if (version !== requestVersion) return
    status.value = 'error'
  }
}

function handleShow(show: boolean) {
  if (show) void loadContent()
}

watch(contentKey, () => {
  requestVersion++
  source.value = ''
  loadedKey.value = ''
  status.value = 'idle'
})

async function loadContentByType(type: QuickPreviewContentType, id: number) {
  switch (type) {
    case 'article':
      return (await getArticle(id)).content
    case 'moment':
      return (await getMoment(id)).content
    case 'page':
      return (await getPage(id)).content
  }
}
</script>

<template>
  <NTooltip
    trigger="hover"
    placement="bottom-start"
    :show-arrow="false"
    @update:show="handleShow"
  >
    <template #trigger>
      <button
        type="button"
        aria-label="快捷预览"
        class="ml-2 inline-flex cursor-help items-center align-middle text-neutral-500 transition-colors hover:text-primary focus-visible:text-primary focus-visible:outline-none dark:text-neutral-400"
      >
        <span class="iconify size-4 ph--file-search" />
      </button>
    </template>

    <ScrollContainer class="max-h-80 w-[min(30rem,80vw)] overflow-auto text-sm">
      <div
        v-if="status === 'idle' || status === 'loading'"
        class="flex min-h-24 items-center justify-center"
      >
        <NSpin size="small" />
      </div>
      <div
        v-else-if="status === 'error'"
        class="flex min-h-24 items-center justify-center text-sm opacity-65"
      >
        预览加载失败
      </div>
      <MarkdownPreview
        v-else
        :source="source"
        class="p-2 text-sm"
      />
    </ScrollContainer>
  </NTooltip>
</template>
