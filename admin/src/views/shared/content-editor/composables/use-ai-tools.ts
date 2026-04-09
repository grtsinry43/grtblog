import { ref } from 'vue'

import { generateSummaryStream, generateTitle } from '@/services/ai'

import type { MessageApi } from 'naive-ui'

interface UseAiTitleGenerationOptions {
  getContent: () => string
  applyResult: (result: { title: string; shortUrl: string }) => void
  message: MessageApi
}

export function useAiTitleGeneration(options: UseAiTitleGenerationOptions) {
  const loading = ref(false)

  async function generate() {
    const content = options.getContent().trim()
    if (!content) {
      options.message.warning('请先输入内容')
      return
    }

    loading.value = true
    try {
      const result = await generateTitle(content)
      options.applyResult(result)
      options.message.success('AI 生成成功')
    } catch (error) {
      options.message.error(error instanceof Error ? error.message : 'AI 生成失败')
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    generate,
  }
}

interface UseAiSummaryGenerationOptions {
  getContent: () => string
  adoptSummary: (summary: string) => void
  message: MessageApi
}

export function useAiSummaryGeneration(options: UseAiSummaryGenerationOptions) {
  const loading = ref(false)
  const result = ref('')
  const done = ref(false)

  async function generate() {
    const content = options.getContent().trim()
    if (!content) {
      options.message.warning('请先输入内容')
      return
    }

    loading.value = true
    result.value = ''
    done.value = false

    try {
      await generateSummaryStream(content, (chunk) => {
        result.value += chunk
      })
      done.value = true
    } catch (error) {
      options.message.error(error instanceof Error ? error.message : 'AI 摘要生成失败')
      result.value = ''
    } finally {
      loading.value = false
    }
  }

  function adopt() {
    options.adoptSummary(result.value)
    result.value = ''
    done.value = false
    options.message.success('已采纳 AI 摘要')
  }

  function dismiss() {
    result.value = ''
    done.value = false
  }

  return {
    loading,
    result,
    done,
    generate,
    adopt,
    dismiss,
  }
}
