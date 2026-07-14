import { flushPromises, shallowMount } from '@vue/test-utils'
import { NTooltip } from 'naive-ui'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import MarkdownPreview from '@/components/markdown-editor/MarkdownPreview.vue'
import { getArticle } from '@/services/articles'
import { getMoment } from '@/services/moments'
import { getPage } from '@/services/page'

import ContentQuickPreview from './ContentQuickPreview.vue'

vi.mock('@/services/articles', () => ({ getArticle: vi.fn() }))
vi.mock('@/services/moments', () => ({ getMoment: vi.fn() }))
vi.mock('@/services/page', () => ({ getPage: vi.fn() }))

describe('ContentQuickPreview', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads article content only when the tooltip is opened', async () => {
    vi.mocked(getArticle).mockResolvedValue({ content: '# 文章预览' } as never)
    const wrapper = shallowMount(ContentQuickPreview, {
      props: { contentType: 'article', contentId: 12 },
      global: { renderStubDefaultSlot: true },
    })

    expect(getArticle).not.toHaveBeenCalled()
    wrapper.findComponent(NTooltip).vm.$emit('update:show', true)
    await flushPromises()

    expect(getArticle).toHaveBeenCalledWith(12)
    expect(wrapper.findComponent(MarkdownPreview).props('source')).toBe('# 文章预览')
  })

  it.each([
    ['moment', getMoment, '手记预览'],
    ['page', getPage, '页面预览'],
  ] as const)('uses the matching %s detail service', async (contentType, loader, content) => {
    vi.mocked(loader).mockResolvedValue({ content } as never)
    const wrapper = shallowMount(ContentQuickPreview, {
      props: { contentType, contentId: 8 },
      global: { renderStubDefaultSlot: true },
    })

    wrapper.findComponent(NTooltip).vm.$emit('update:show', true)
    await flushPromises()

    expect(loader).toHaveBeenCalledWith(8)
    expect(wrapper.findComponent(MarkdownPreview).props('source')).toBe(content)
  })
})
