<script setup lang="ts">
import {
  NButton,
  NInput,
  NSwitch,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
} from 'naive-ui'
import { onMounted, ref } from 'vue'
import { PreviewLink20Regular } from '@vicons/fluent'
import { SaveOutline } from '@vicons/ionicons5'

// Components
import MarkdownEditor from '@/components/markdown-editor/MarkdownEditor.vue'
import MarkdownPreview from '@/components/markdown-editor/MarkdownPreview.vue'

// Composables
import { usePageForm } from './composables/use-page-form'

defineOptions({ name: 'PageEdit' })

// 1. Initialize form logic
const { form, loading, saving, fetch, save } = usePageForm()

// 2. View state management
const showMeta = ref(false)
const showPreview = ref(false)

// 3. Lifecycle
onMounted(() => {
  fetch()
})
</script>

<template>
  <div class="flex h-full min-h-0 flex-col">
    <header
      class="z-10 flex shrink-0 flex-col gap-3 px-10 py-8 backdrop-blur sm:h-24 sm:flex-row sm:items-center sm:justify-between sm:py-0"
    >
      <div class="flex w-full items-center gap-4 sm:flex-1">
        <NInput
          v-model:value="form.title"
          placeholder="页面标题..."
          :bordered="false"
          class="flex-1 text-xl! leading-tight font-bold sm:text-2xl!"
          style="--n-caret-color: var(--primary-color); background-color: transparent"
        />
      </div>

      <div class="flex w-full flex-wrap items-center gap-3 sm:w-auto sm:flex-nowrap sm:gap-4">
        <div class="flex items-baseline gap-1">
          <div class="iconify self-center ph--link-simple" />
          <span class="text-xs leading-none">/pages/</span>
          <input
            v-model="form.shortUrl"
            placeholder="请填写短链接"
            class="w-24 border-b border-current/30 p-0 pb-0.5 text-[11px] leading-none focus:border-primary focus:outline-none sm:w-32"
          />
        </div>

        <div class="flex items-center gap-2">
          <NButton
            quaternary
            circle
            size="small"
            @click="showMeta = true"
          >
            <template #icon><div class="iconify text-xl ph--sliders-horizontal" /></template>
          </NButton>

          <NButton
            quaternary
            circle
            size="small"
            :type="showPreview ? 'primary' : 'default'"
            @click="showPreview = !showPreview"
          >
            <template #icon><PreviewLink20Regular /></template>
          </NButton>

          <NButton
            type="primary"
            size="medium"
            :loading="saving"
            @click="save"
            class="px-5 font-medium shadow-sm active:scale-95"
          >
            <template #icon><SaveOutline /></template>
            保存
          </NButton>
        </div>
      </div>
    </header>

    <main class="flex min-h-0 flex-1 overflow-hidden">
      <div
        class="editor-container grid h-full min-h-0 w-full"
        :class="showPreview ? 'grid-cols-1 lg:grid-cols-2' : 'grid-cols-1'"
      >
        <div class="pane editor-pane relative h-full overflow-auto">
          <MarkdownEditor
            v-model="form.content"
            class="h-full min-h-full"
          />
        </div>

        <div
          v-if="showPreview"
          class="pane preview-pane relative h-full overflow-auto"
        >
          <MarkdownPreview
            :source="form.content"
            class="p-4 sm:p-8"
          />
        </div>
      </div>
    </main>

    <NDrawer
      v-model:show="showMeta"
      placement="right"
      width="400"
    >
      <NDrawerContent
        title="页面设置"
        :native-scrollbar="false"
        closable
        header-style="padding: 24px;"
        body-style="padding: 24px;"
      >
        <div class="flex flex-col gap-6">
          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--article" />
              <span>元信息</span>
            </div>
            <NForm
              label-placement="top"
              label-width="auto"
              class="space-y-4"
            >
              <NFormItem
                label="描述"
                :show-feedback="false"
              >
                <NInput
                  v-model:value="form.description"
                  type="textarea"
                  placeholder="简短的页面描述..."
                  :autosize="{ minRows: 2, maxRows: 4 }"
                />
              </NFormItem>
            </NForm>
          </div>

          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--toggle-left" />
              <span>属性</span>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="flex items-center justify-between rounded-lg px-4 py-3">
                <span class="text-sm">是否启用</span>
                <NSwitch
                  v-model:value="form.isEnabled"
                  size="small"
                />
              </div>
            </div>
          </div>
        </div>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped>
.pane::-webkit-scrollbar,
.editor-pane :deep(.cm-scroller::-webkit-scrollbar),
.preview-pane :deep(.markdown-preview::-webkit-scrollbar) {
  width: 5px;
  height: 5px;
}
.pane::-webkit-scrollbar-track,
.editor-pane :deep(.cm-scroller::-webkit-scrollbar-track),
.preview-pane :deep(.markdown-preview::-webkit-scrollbar-track) {
  background: transparent;
}
:global(.dark) .pane::-webkit-scrollbar-thumb,
:global(.dark) .editor-pane :deep(.cm-scroller::-webkit-scrollbar-thumb),
:global(.dark) .preview-pane :deep(.markdown-preview::-webkit-scrollbar-thumb) {
  background-color: #374151;
}
.pane::-webkit-scrollbar-thumb:hover,
.editor-pane :deep(.cm-scroller::-webkit-scrollbar-thumb:hover),
.preview-pane :deep(.markdown-preview::-webkit-scrollbar-thumb:hover) {
  background-color: #d1d5db;
}
:global(.dark) .pane::-webkit-scrollbar-thumb:hover,
:global(.dark) .editor-pane :deep(.cm-scroller::-webkit-scrollbar-thumb:hover),
:global(.dark) .preview-pane :deep(.markdown-preview::-webkit-scrollbar-thumb:hover) {
  background-color: #4b5563;
}
.editor-pane :deep(.cm-editor) {
  height: 100% !important;
  font-family: inherit;
}
.editor-pane :deep(.cm-scroller) {
  padding-bottom: 50vh;
  font-family: 'JetBrains Mono', monospace;
  line-height: 1.6;
}
.preview-pane :deep(.markdown-preview) {
  padding-bottom: 50vh;
}
</style>