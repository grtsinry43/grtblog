import { describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'

import {
  findContentTagByName,
  mergeContentTagOptions,
} from '@/views/shared/content-editor/model/content-tags'

import { useContentTagSelect } from './use-content-tag-select'

function createMessage() {
  return {
    error: vi.fn(),
    success: vi.fn(),
  }
}

describe('content tag model', () => {
  it('normalizes, de-duplicates and sorts options', () => {
    expect(
      mergeContentTagOptions(
        [{ label: ' Vue ', value: 1 }],
        [
          { id: 2, name: '随笔' },
          { id: 1, name: 'Vue 3' },
          { id: 3, name: '  前端   开发 ' },
        ],
      ),
    ).toEqual([
      { label: '前端 开发', value: 3 },
      { label: '随笔', value: 2 },
      { label: 'Vue 3', value: 1 },
    ])
  })

  it('finds an exact name without case or surrounding-space differences', () => {
    expect(
      findContentTagByName(
        [
          { label: 'Vue', value: 1 },
          { label: 'TypeScript', value: 2 },
        ],
        '  typescript ',
      ),
    ).toEqual({ label: 'TypeScript', value: 2 })
  })
})

describe('useContentTagSelect', () => {
  it('keeps initial items when the global option request finishes later', async () => {
    const selectedIds = ref<number[]>([])
    const selector = useContentTagSelect({
      selectedIds,
      noun: '标签',
      message: createMessage(),
      api: {
        list: vi.fn().mockResolvedValue([{ id: 1, name: 'Vue' }]),
        create: vi.fn(),
      },
    })

    selector.setInitialItems([{ id: 9, name: '已归档标签' }])
    await selector.loadOptions()

    expect(selectedIds.value).toEqual([9])
    expect(selector.options.value).toEqual([
      { label: '已归档标签', value: 9 },
      { label: 'Vue', value: 1 },
    ])
  })

  it('selects an existing item without creating a duplicate', async () => {
    const selectedIds = ref<number[]>([])
    const create = vi.fn()
    const selector = useContentTagSelect({
      selectedIds,
      noun: '话题',
      message: createMessage(),
      api: {
        list: vi.fn().mockResolvedValue([{ id: 2, name: 'TypeScript' }]),
        create,
      },
    })
    await selector.loadOptions()

    await selector.createAndSelect(' typescript ')
    await selector.createAndSelect('TypeScript')

    expect(create).not.toHaveBeenCalled()
    expect(selectedIds.value).toEqual([2])
  })

  it('adds a created item to both options and selected IDs', async () => {
    const selectedIds = ref([1])
    const message = createMessage()
    const create = vi.fn().mockResolvedValue({ id: 3, name: '新话题' })
    const selector = useContentTagSelect({
      selectedIds,
      noun: '话题',
      message,
      api: {
        list: vi.fn().mockResolvedValue([{ id: 1, name: '日常' }]),
        create,
      },
    })
    await selector.loadOptions()

    await selector.createAndSelect('  新话题  ')

    expect(create).toHaveBeenCalledWith('新话题')
    expect(selectedIds.value).toEqual([1, 3])
    expect(selector.selectedItems.value).toEqual([
      { id: 1, name: '日常' },
      { id: 3, name: '新话题' },
    ])
    expect(message.success).toHaveBeenCalledOnce()
  })

  it('keeps the selection unchanged after a failed creation so it can be retried', async () => {
    const selectedIds = ref([1])
    const message = createMessage()
    const selector = useContentTagSelect({
      selectedIds,
      noun: '标签',
      message,
      api: {
        list: vi.fn().mockResolvedValue([{ id: 1, name: 'Vue' }]),
        create: vi.fn().mockRejectedValue(new Error('network error')),
      },
    })
    await selector.loadOptions()

    expect(await selector.createAndSelect('新标签')).toBeNull()
    expect(selectedIds.value).toEqual([1])
    expect(message.error).toHaveBeenCalledWith('创建标签“新标签”失败，请重试')
    expect(selector.creating.value).toBe(false)
  })
})
